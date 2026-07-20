package wazeropool

import (
	"github.com/tetratelabs/wazero"
)

// Option represents a constructor option.
type Option func(*instance)

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
