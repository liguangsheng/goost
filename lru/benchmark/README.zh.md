# lru/benchmark

用于对比 `goost/lru` 与若干外部 cache 库的 benchmark。

这是一个 nested module，因此 benchmark-only 依赖不会进入 root `goost` module。它不是库的公开 API 表面。

## 运行

```sh
go test -bench=. ./...
```

请在本目录运行 benchmark。结果只作为本地性能证据，不作为正确性检查。

