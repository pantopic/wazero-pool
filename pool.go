package wazeropool

import (
	"context"
	"runtime"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Option represents a constructor option.
type Option func(*instancePool)

// WithLimit sets the maximum pool size. Set to 0 for unlimited pool (default)
func WithLimit(n int) Option {
	return func(m *instancePool) {
		if n < 1 {
			m.limit = nil
			return
		}
		m.limit = make(chan struct{}, n)
	}
}

// InstancePool represents a module instance pool
type InstancePool interface {
	Get() api.Module
	Put(mod api.Module)
	With(fn func(mod api.Module))
}

type instancePool struct {
	compiled wazero.CompiledModule
	ctx      context.Context
	limit    chan struct{}
	pool     *sync.Pool
	runtime  wazero.Runtime
	wrappers *sync.Pool
}

// Module instances can't be garbage collected directly. This wrapper type has no external references so it can be
// garbage collected.
type wrapper struct {
	mod     api.Module
	cleanup runtime.Cleanup
}

// New returns a new module instance pool.
func New(ctx context.Context, r wazero.Runtime, src []byte, cfg wazero.ModuleConfig, opts ...Option) (m *instancePool, err error) {
	compiled, err := r.CompileModule(ctx, src)
	if err != nil {
		return
	}
	mod, err := r.InstantiateModule(ctx, compiled, cfg.WithName(""))
	if err != nil {
		return
	}
	m = &instancePool{
		compiled: compiled,
		runtime:  r,
	}
	for _, opt := range opts {
		opt(m)
	}
	if m.limit != nil {
		for range cap(m.limit) - 1 {
			m.limit <- struct{}{}
		}
	}
	m.ctx = ctx
	m.wrappers = &sync.Pool{New: func() any { return new(wrapper) }}
	m.pool = &sync.Pool{
		New: func() any {
			mod, _ := r.InstantiateModule(ctx, compiled, cfg.WithName(""))
			w := m.wrappers.Get().(*wrapper)
			w.mod = mod
			return w
		},
	}
	m.Put(mod)
	return
}

// Get returns a module instance from the pool.
// If limit is non-zero, [Get] will block until an instance becomes available.
// If limit is non-zero and the module is not [Put] back, a deadlock may occur.
func (m *instancePool) Get() api.Module {
	if m.limit != nil {
		<-m.limit
	}
	w := m.pool.Get().(*wrapper)
	w.cleanup.Stop()
	defer m.wrappers.Put(w)
	return w.mod
}

// Put puts a module instance back into the pool.
func (m *instancePool) Put(mod api.Module) {
	var w = m.wrappers.Get().(*wrapper)
	w.mod = mod
	w.cleanup = runtime.AddCleanup(w, func(mod api.Module) { mod.Close(context.Background()) }, mod)
	m.pool.Put(w)
	if m.limit != nil {
		m.limit <- struct{}{}
	}
}

// With automatically [Get]s a module instance from the pool and [Put]s it back after the function returns.
func (m *instancePool) With(fn func(mod api.Module)) {
	mod := m.Get()
	defer m.Put(mod)
	fn(mod)
}
