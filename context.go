package wazeropool

import (
	"context"
)

var DefaultCtxKeyWazeroPool = `__wazero_pool`

func ContextSet(ctx context.Context, pool Instance) context.Context {
	return context.WithValue(ctx, DefaultCtxKeyWazeroPool, pool)
}

func ContextCopy(dst, src context.Context) context.Context {
	return context.WithValue(dst, DefaultCtxKeyWazeroPool, Context(src))
}

func Context(ctx context.Context) Instance {
	return ctx.Value(DefaultCtxKeyWazeroPool).(Instance)
}
