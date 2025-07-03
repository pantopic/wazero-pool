wasm:
	@cd test && tinygo build -buildmode=wasi-legacy -target=wasi -opt=2 -gc=leaking -scheduler=none -o ../test.wasm module.go
	@cd test-invalid && tinygo build -buildmode=wasi-legacy -target=wasi -opt=2 -gc=leaking -scheduler=none -o ../test.invalid.wasm module.go

test:
	@go test . -v -cover

bench:
	@go test -bench=. -v -run=Benchmark.*

cover:
	@mkdir -p _dist
	@go test . -coverprofile=_dist/coverage.out -v
	@go tool cover -html=_dist/coverage.out -o _dist/coverage.html

cloc:
	@cloc . --exclude-dir=_dist

.PHONY: test
