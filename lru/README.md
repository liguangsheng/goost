# lru

Go lru cache implement, faster, less alloc.

```go
c := lru.New().Cap(1000).Safe(true).Build()
c.Set("hello", "world")
c.SetWithExpire("hello", "world", time.Now.Add(time.Minute))
c.SetWithDuration("hello", "world", time.Minute)

v, ok := c.Get("hello")
if ok {
	fmt.Println(v) // world
}
```

```
goos: darwin
goarch: amd64
pkg: github.com/liguangsheng/go-lru/benchmark
Benchmark_gcache_lru_Set-8      	 5000000	       225 ns/op	      56 B/op	       2 allocs/op
Benchmark_gcache_lru_Get-8      	10000000	       144 ns/op	      19 B/op	       1 allocs/op
Benchmark_gcache_arc_Set-8      	 1000000	      2070 ns/op	     379 B/op	       4 allocs/op
Benchmark_gcache_arc_Get-8      	 2000000	       581 ns/op	      64 B/op	       2 allocs/op
Benchmark_golanglru_lru_Set-8   	 3000000	       349 ns/op	      90 B/op	       2 allocs/op
Benchmark_golanglru_lru_Get-8   	10000000	       121 ns/op	      19 B/op	       1 allocs/op
Benchmark_golanglru_arc_Set-8   	 1000000	      1056 ns/op	     223 B/op	       4 allocs/op
Benchmark_golanglru_arc_Get-8   	 3000000	       404 ns/op	      55 B/op	       2 allocs/op
Benchmark_goostlru_Set-8           	20000000	        79.4 ns/op	      11 B/op	       1 allocs/op
Benchmark_goostlru_Get-8           	30000000	        40.5 ns/op	       0 B/op	       0 allocs/op
Benchmark_goostlru_UnsafeSet-8     	20000000	        71.5 ns/op	      11 B/op	       1 allocs/op
Benchmark_goostlru_UnsafeGet-8     	30000000	        34.8 ns/op	       0 B/op	       0 allocs/op
```
