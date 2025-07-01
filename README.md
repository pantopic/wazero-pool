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
BenchmarkModule/add/linear-16                    1181882              1009 ns/op
BenchmarkModule/add/parallel-2-16                1038873              1042 ns/op
BenchmarkModule/add/parallel-4-16                1000000              1070 ns/op
BenchmarkModule/add/parallel-16-16               1000000              1307 ns/op
BenchmarkModule/add/parallel-0-16                3431593               325.1 ns/op
BenchmarkModule/microsleep/linear-16                6487            156801 ns/op
BenchmarkModule/microsleep/parallel-2-16           93465             14592 ns/op
BenchmarkModule/microsleep/parallel-4-16          696016              2530 ns/op
BenchmarkModule/microsleep/parallel-16-16        1000000              1678 ns/op
BenchmarkModule/microsleep/parallel-0-16         1000000              1417 ns/op
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
