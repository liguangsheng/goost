# itertools

Small, generic slice helpers.

```go
itertools.Map([]int{1, 2, 3}, func(v, _ int) int { return v * 2 })
// [2 4 6]

itertools.Filter([]int{1, 2, 3, 4}, func(v, _ int) bool { return v%2 == 0 })
// [2 4]

itertools.Chunk([]int{1, 2, 3, 4, 5}, 2)
// [[1 2] [3 4] [5]]

itertools.Difference([]int{1, 2, 3}, []int{2}) // [1 3]
itertools.Intersection([]int{1, 2, 3}, []int{2, 3, 4}) // [2 3]
```
