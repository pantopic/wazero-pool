package wazeropool

import (
	"context"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Option represents a constructor option.
type Option func(*module)

// WithLimit sets the maximum pool size. Set to 0 for unlimited pool (default)
func WithLimit(n int) Option {
	return func(m *module) {
		if n < 1 {
			m.limit = nil
			return
		}
		m.limit = make(chan struct{}, n)
	}
}

// Module represents a module instance pool
type Module interface {
	Get() api.Module
	Put(mod api.Module)
	With(fn func(mod api.Module))
}

type module struct {
	compiled wazero.CompiledModule
	ctx      context.Context
	limit    chan struct{}
	pool     *sync.Pool
	runtime  wazero.Runtime
}

// New returns a new module instance pool.
func New(ctx context.Context, r wazero.Runtime, src []byte, cfg wazero.ModuleConfig, opts ...Option) (m *module, err error) {
	compiled, err := r.CompileModule(ctx, src)
	if err != nil {
		return
	}
	mod, err := r.InstantiateModule(ctx, compiled, cfg.WithName(""))
	if err != nil {
		return
	}
	m = &module{
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
			mod, err := r.InstantiateModule(ctx, compiled, cfg.WithName(""))
			if err != nil {
				panic(err)
			}
			return mod
		},
	}
	m.pool.Put(mod)
	return
}

// Get returns a module instance from the pool.
// If limit is non-zero, [Get] will block until an instance becomes available.
// If limit is non-zero and the module is not [Put] back, a deadlock may occur.
func (m *module) Get() api.Module {
	if m.limit != nil {
		<-m.limit
	}
	return m.pool.Get().(api.Module)
}

// Put puts a module instance back into the pool.
func (m *module) Put(mod api.Module) {
	m.pool.Put(mod)
	if m.limit != nil {
		m.limit <- struct{}{}
	}
}

// With automatically [Get]s a module instance from the pool and [Put]s it back after the function returns.
func (m *module) With(fn func(mod api.Module)) {
	mod := m.Get()
	defer m.Put(mod)
	fn(mod)
}
