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
		m.inuse = make(chan *wrapper, n)
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
	inuse    chan *wrapper
	pool     *sync.Pool
	runtime  wazero.Runtime
}

// Module instances can't be garbage collected directly. This wrapper type has no external references so it can be
// garbage collected.
type wrapper struct {
	mod api.Module
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
		for range cap(m.limit) {
			m.limit <- struct{}{}
		}
	}
	m.ctx = ctx
	m.pool = &sync.Pool{
		New: func() any {
			mod, _ := r.InstantiateModule(ctx, compiled, cfg.WithName(""))
			w := &wrapper{mod}
			runtime.AddCleanup(w, func(mod api.Module) { mod.Close(ctx) }, mod)
			return w
		},
	}
	w := &wrapper{mod}
	runtime.AddCleanup(w, func(mod api.Module) { mod.Close(ctx) }, mod)
	m.pool.Put(w)
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
	mod := w.mod
	if m.limit != nil {
		w.mod = nil
		m.inuse <- w
	}
	return mod
}

// Put puts a module instance back into the pool.
func (m *instancePool) Put(mod api.Module) {
	var w *wrapper
	if m.limit != nil {
		w = <-m.inuse
		w.mod = mod
		m.limit <- struct{}{}
	} else {
		w = &wrapper{mod}
	}
	m.pool.Put(w)
}

// With automatically [Get]s a module instance from the pool and [Put]s it back after the function returns.
func (m *instancePool) With(fn func(mod api.Module)) {
	mod := m.Get()
	defer m.Put(mod)
	fn(mod)
}
