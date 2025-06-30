wasm:
	@cd test && tinygo build -buildmode=wasi-legacy -target=wasi -opt=2 -gc=conservative -scheduler=none -o ../test.wasm module.go

wasm-invalid:
	@cd test-invalid && tinygo build -buildmode=wasi-legacy -target=wasi -opt=2 -gc=conservative -scheduler=none -o ../test.invalid.wasm module.go

test:
	@go test .

bench:
	@go test -bench=. -v -run=Benchmark.*

cover:
	@mkdir -p _dist
	@go test . -coverprofile=_dist/coverage.out -v
	@go tool cover -html=_dist/coverage.out -o _dist/coverage.html

cloc:
	@cloc . --exclude-dir=_dist

.PHONY: test
