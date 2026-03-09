package wazeropool

import (
	"context"
)

var ctxKeyWazeroPool = `__wazero_pool`

func ContextSet(ctx context.Context, pool Instance) context.Context {
	return context.WithValue(ctx, ctxKeyWazeroPool, pool)
}

func ContextCopy(dst, src context.Context) context.Context {
	return context.WithValue(dst, ctxKeyWazeroPool, FromContext(src))
}

func FromContext(ctx context.Context) Instance {
	return ctx.Value(ctxKeyWazeroPool).(Instance)
}
