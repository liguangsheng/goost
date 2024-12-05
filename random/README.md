# random

生成随机字符串，并发安全，性能超高，内存占用超小

# example

```go
random.String(16, random.HumanAlphanumeric)

or

s := random.New()
s.Next(16, random.HumanAlphanumeric)
```

# benchmark
```
goos: linux
goarch: amd64
pkg: github.com/liguangsheng/goost/random
cpu: Intel(R) Core(TM) i7-9700 CPU @ 3.00GHz
Benchmark_String
Benchmark_String-4   	9837387	      110.3 ns/op	     16 B/op	      1 allocs/op
PASS
ok  	github.com/liguangsheng/goost/random	1.241s
```
