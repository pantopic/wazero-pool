package wazeropool

import (
	"context"
	_ "embed"
	"fmt"
	goruntime "runtime"
	"testing"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

//go:embed test\.wasm
var src []byte

//go:embed test\.invalid\.wasm
var srcInvalid []byte

func TestModule(t *testing.T) {
	ctx := context.Background()
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithMemoryLimitPages(256).
		WithMemoryCapacityFromMax(true))
	wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
	cfg := wazero.NewModuleConfig()
	t.Run(`base`, func(t *testing.T) {
		pool, err := New(ctx, runtime, src, cfg)
		if err != nil {
			t.Fatalf(`%v`, err)
		}
		for range 2 {
			mod := pool.Get()
			defer pool.Put(mod)
			stack, err := mod.ExportedFunction("add").Call(ctx, 1, 1)
			if err != nil {
				t.Fatalf(`%v`, err)
			}
			if len(stack) < 1 || stack[0] != 2 {
				t.Fatalf(`Incorrect response: %v`, stack)
			}
		}
		t.Run(`with`, func(t *testing.T) {
			pool.With(func(mod api.Module) {
				stack, err := mod.ExportedFunction("add").Call(ctx, 1, 1)
				if err != nil {
					t.Fatalf(`%v`, err)
				}
				if len(stack) < 1 || stack[0] != 2 {
					t.Fatalf(`Incorrect response: %v`, stack)
				}
			})
		})
	})
	t.Run(`invalid`, func(t *testing.T) {
		t.Run(`src`, func(t *testing.T) {
			_, err := New(ctx, runtime, []byte(`invalid`), cfg)
			if err == nil {
				t.Fatal(`Pool instantiation should have failed`)
			}
		})
		t.Run(`main`, func(t *testing.T) {
			_, err := New(ctx, runtime, srcInvalid, cfg)
			if err == nil {
				t.Fatal(`Pool instantiation should have failed`)
			}
		})
	})
	t.Run(`limit`, func(t *testing.T) {
		t.Run(`zero`, func(t *testing.T) {
			_, err := New(ctx, runtime, src, cfg, WithLimit(-1))
			if err != nil {
				t.Fatalf(`%v`, err)
			}
		})
		t.Run(`block`, func(t *testing.T) {
			pool, err := New(ctx, runtime, src, cfg, WithLimit(1))
			if err != nil {
				t.Fatalf(`%v`, err)
			}
			mod := pool.Get()
			defer pool.Put(mod)
			got := make(chan bool)
			go func() {
				mod2 := pool.Get()
				defer pool.Put(mod2)
				got <- true
			}()
			select {
			case <-got:
				t.Fatalf("Got second module intance")
			case <-time.After(time.Millisecond):
			}
		})
	})
	t.Run(`cleanup`, func(t *testing.T) {
		pool, err := New(ctx, runtime, src, cfg, WithLimit(2))
		if err != nil {
			t.Fatalf(`%v`, err)
		}
		goruntime.GC()
		goruntime.GC()
		for range 5 {
			var mod = pool.Get()
			goruntime.GC()
			goruntime.GC()
			if mod.IsClosed() {
				t.Fatal(`Module should not be closed.`)
			}
			pool.Put(mod)
			goruntime.GC()
			goruntime.GC()
			if !mod.IsClosed() {
				t.Fatal(`Module should be closed.`)
			}
		}
	})
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
