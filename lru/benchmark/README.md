# lru/benchmark

Benchmarks for comparing `goost/lru` with selected external cache libraries.

This is a nested module so benchmark-only dependencies stay out of the root
`goost` module. It is not part of the library API surface.

## Run

```sh
go test -bench=. ./...
```

Run benchmarks from this directory. Treat results as local performance evidence;
do not use them as correctness checks.

