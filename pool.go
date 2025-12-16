package wazeropool

import (
	"context"
	"runtime"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// New returns a new module instance pool.
func New(ctx context.Context, r wazero.Runtime, src []byte, opts ...Option) (m *instance, err error) {
	compiled, err := r.CompileModule(ctx, src)
	if err != nil {
		return
	}
	m = &instance{
		compiled: compiled,
		config:   wazero.NewModuleConfig(),
		runtime:  r,
	}
	for _, opt := range opts {
		opt(m)
	}
	w := newWrapper()
	w.Module, err = r.InstantiateModule(ctx, compiled, m.config.WithName(""))
	if err != nil {
		return
	}
	m.pool = &sync.Pool{
		New: func() any {
			w := newWrapper()
			w.Module, _ = r.InstantiateModule(ctx, compiled, m.config.WithName(""))
			return w
		},
	}
	if m.limit != nil {
		m.limit <- struct{}{}
	}
	m.Put(w)
	return
}

// Instance represents a module instance pool
type Instance interface {
	// Compiled returns the compiled module for this instance.
	Compiled() wazero.CompiledModule

	// Get returns a module instance from the pool.
	// If limit is non-zero, [Get] will block until an instance becomes available.
	// If limit is non-zero and the module is not [Put] back, a deadlock may occur.
	Get() api.Module

	// Put puts a module instance back into the pool.
	Put(mod api.Module)

	// Run is a conveience method.
	// It [Get]s a module instance from the pool and [Put]s it back after the function returns.
	Run(fn func(mod api.Module))
}

type instance struct {
	compiled wazero.CompiledModule
	config   wazero.ModuleConfig
	limit    chan struct{}
	pool     *sync.Pool
	runtime  wazero.Runtime
}

func (m *instance) Get() api.Module {
	if m.limit != nil {
		m.limit <- struct{}{}
	}
	w := m.pool.Get().(*wrapper)
	w.cleanup.Stop()
	return w
}

func (m *instance) Put(mod api.Module) {
	w := mod.(*wrapper)
	w.cleanup = runtime.AddCleanup(w, func(m api.Module) { m.Close(context.Background()) }, w.Module)
	m.pool.Put(w)
	if m.limit != nil {
		<-m.limit
	}
}

func (m *instance) Run(fn func(mod api.Module)) {
	mod := m.Get()
	defer m.Put(mod)
	fn(mod)
}

func (m *instance) Compiled() wazero.CompiledModule {
	return m.compiled
}

// Module instances can't be garbage collected directly. This wrapper type has no external references so it can be
// garbage collected. The wrapper also caches exported function references automatically to reduce call latency.
type wrapper struct {
	api.Module

	cache   map[string]api.Function
	cleanup runtime.Cleanup
}

func newWrapper() *wrapper {
	return &wrapper{cache: make(map[string]api.Function)}
}

func (w *wrapper) ExportedFunction(name string) (fn api.Function) {
	fn, ok := w.cache[name]
	if !ok {
		fn = w.Module.ExportedFunction(name)
		w.cache[name] = fn
	}
	return
}
