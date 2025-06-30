package wazeropool

import (
	"context"
	_ "embed"
	"fmt"
	"testing"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed test\.wasm
var src []byte

func TestModule(t *testing.T) {
	ctx := context.Background()
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithMemoryLimitPages(256).
		WithMemoryCapacityFromMax(true))
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	cfg := wazero.NewModuleConfig()
	pool, err := New(ctx, runtime, src, cfg)
	if err != nil {
		t.Fatalf(`%v`, err)
	}
	mod := pool.Get()
	stack, err := mod.ExportedFunction("add").Call(ctx, 1, 1)
	if len(stack) < 1 || stack[0] != 2 {
		t.Fatalf(`Incorrect response: %v`, stack)
	}
}

func BenchmarkModule(b *testing.B) {
	ctx := context.Background()
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithMemoryLimitPages(256).
		WithMemoryCapacityFromMax(true))
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	cfg := wazero.NewModuleConfig().WithSysNanosleep()
	for _, name := range []string{
		"add",
		"microsleep",
	} {
		b.Run(name, func(b *testing.B) {
			b.Run(`linear`, func(b *testing.B) {
				pool, err := New(ctx, runtime, src, cfg)
				if err != nil {
					b.Fatalf(`%v`, err)
				}
				for b.Loop() {
					pool.With(func(mod api.Module) {
						stack, err := mod.ExportedFunction(name).Call(ctx, 1, 1)
						if err != nil {
							b.Fatalf(`%v`, err)
						}
						if len(stack) < 1 || stack[0] != 2 {
							b.Fatalf(`Incorrect response: %v`, stack)
						}
					})
				}
			})
			for _, n := range []int{2, 4, 16, 0} {
				pool, err := New(ctx, runtime, src, cfg, WithLimit(n))
				if err != nil {
					b.Fatalf(`%v`, err)
				}
				b.Run(fmt.Sprintf(`parallel-%d`, n), func(b *testing.B) {
					b.SetParallelism(n)
					b.RunParallel(func(pb *testing.PB) {
						for pb.Next() {
							pool.With(func(mod api.Module) {
								stack, err := mod.ExportedFunction(name).Call(ctx, 1, 1)
								if err != nil {
									b.Fatalf(`%v`, err)
								}
								if len(stack) < 1 || stack[0] != 2 {
									b.Fatalf(`Incorrect response: %v`, stack)
								}
							})
						}
					})
				})
			}
		})
	}
}
