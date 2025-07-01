# Wazero Pool

A wazero module instance pool.

[![Go Reference](https://godoc.org/github.com/pantopic/wazero-pool?status.svg)](https://godoc.org/github.com/pantopic/wazero-pool)
[![Go Report Card](https://goreportcard.com/badge/github.com/pantopic/wazero-pool?1)](https://goreportcard.com/report/github.com/pantopic/wazero-pool)
[![Go Coverage](https://github.com/pantopic/wazero-pool/wiki/coverage.svg)](https://raw.githack.com/wiki/pantopic/wazero-pool/coverage.html)

When working with webassembly modules, it is frequently more advantageous to use a homogeneous set of module instances
rather than a single instance. This package provides a common approach and interface to manage pools of
wazero module instances.

```go
package main

import (
	"context"
	_ "embed"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"

	"github.com/pantopic/wazero-pool"
)

//go:embed test\.wasm
var src []byte

func main() {
	ctx := context.Background()
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig())
    wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
    cfg := wazero.NewModuleConfig()

    pool, err := wazeropool.New(ctx, runtime, src, cfg)
    if err != nil {
        panic(err)
    }
    pool.With(func(mod api.Module) {
        _, err := mod.ExportedFunction(`test1`).Call(ctx)
        if err != nil {
            panic(err)
        }
        _, err := mod.ExportedFunction(`test2`).Call(ctx)
        if err != nil {
            panic(err)
        }
    })

    // ...
}
```

## Options

The constructor has one functional option.

### Limit

Supports explicit `limit` to prevent unbounded memory growth.

```go
pool, err := wazeropool.New(ctx, runtime, src, cfg, 
    wazeropool.WithLimit(2))
if err != nil {
    panic(err)
}
mod1 := pool.Get()
mod2 := pool.Get()
mod3 := pool.Get() // Blocks until a module is returned to the pool with `pool.Put`
```

## Performance

Setting `limit` explicitly is recommended unless you are already strictly limiting the concurrency of the calling code
in other ways.

```go
> make bench
BenchmarkModule/add/raw-16                      31112986                38.95 ns/op
BenchmarkModule/add/wrapped-16                  26988614                45.59 ns/op
BenchmarkModule/add/pooled-16                    7362813               160.9 ns/op
BenchmarkModule/add/parallel-2-16                2847218               445.3 ns/op
BenchmarkModule/add/parallel-4-16                2397896               508.7 ns/op
BenchmarkModule/add/parallel-8-16                1846426               656.6 ns/op
BenchmarkModule/add/parallel-16-16               1676278               745.8 ns/op
BenchmarkModule/add/parallel-0-16                3551059               331.8 ns/op
BenchmarkModule/microsleep/raw-16                  30891             40574 ns/op
BenchmarkModule/microsleep/wrapped-16              25706             48160 ns/op
BenchmarkModule/microsleep/pooled-16               12202             94872 ns/op
BenchmarkModule/microsleep/parallel-2-16           83138             14738 ns/op
BenchmarkModule/microsleep/parallel-4-16          109393             11234 ns/op
BenchmarkModule/microsleep/parallel-8-16          731980              1714 ns/op
BenchmarkModule/microsleep/parallel-16-16        1258392               969.3 ns/op
BenchmarkModule/microsleep/parallel-0-16         2359347               517.5 ns/op
```

## Roadmap

This project is in `Alpha`. Breaking API changes should be expected until Beta.

- `v0.0.x` - Alpha
  - [ ] Stabilize API
- `v0.x.x` - Beta
  - [ ] Finalize API
  - [ ] Test in production
- `v1.x.x` - General Availability
  - [ ] Proven long term stability in production
