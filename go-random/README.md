# go-random

生成随机字符串，线程安全

字符集的长度越接近2的n次方，效率越高

# example

```go
random.String(32)
random.Number(32) 

// or 
var seq = random.NewSequence("abcdefg")
seq.Next(32)
```

# benchmark

```
Benchmark_String-8   	 5000000	       253 ns/op	      32 B/op	       1 allocs/op
Benchmark_Number-8   	10000000	       176 ns/op	      32 B/op	       1 allocs/op
```