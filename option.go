package wazeropool

import (
	"github.com/tetratelabs/wazero"
)

// Option represents a constructor option.
type Option func(*instance)

// WithMemoryLimit sets the maximum memory size. Instances will be recycled when they exceed this limit.
func WithMemoryLimit(n uint32) Option {
	return func(m *instance) {
		m.memcap = n
	}
}

// WithLimit sets the maximum pool size. Set to 0 for unlimited pool (default)
func WithLimit(n int) Option {
	return func(m *instance) {
		if n < 1 {
			m.limit = nil
			return
		}
		m.limit = make(chan struct{}, n)
	}
}

// WithBurst adds an additional burst limit for high priority workloads.
// Useful where you want to prevent read workloads from blocking writes.
func WithBurst(n int) Option {
	return func(m *instance) {
		if n < 1 {
			m.burst = nil
			return
		}
		m.burst = make(chan struct{}, n)
	}
}

// WithmemCap sets the memory limit after which instances will be recycled.
func WithMemCap(n uint32) Option {
	return func(m *instance) {
		m.memcap = n
	}
}

// WithName sets an optional name for the pool. Useful for debugging.
func WithName(name string) Option {
	return func(m *instance) {
		m.name = name
	}
}

// WithVersion sets an optional version for the pool. Useful for debugging.
func WithVersion(version uint64) Option {
	return func(m *instance) {
		m.version = version
	}
}

// WithModuleConfig sets the module config.
func WithModuleConfig(cfg wazero.ModuleConfig) Option {
	return func(m *instance) {
		m.config = cfg
	}
}
