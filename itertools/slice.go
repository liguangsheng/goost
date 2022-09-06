package itertools

// SafeSlice return a slice from begin to end, if begin or end is out of range, return empty slice
func SafeSlice[T any](arr []T, begin, end int) []T {
	if arr == nil {
		return nil
	}
	if begin < 0 {
		begin = 0
	}
	if end <= 0 || begin >= len(arr) || begin >= end {
		return []T{}
	}
	if end > len(arr) {
		end = len(arr)
	}
	return arr[begin:end]
}

// Difference return a slice of elements that are only in a but not in b
func Difference[T comparable](a, b []T) []T {
	if len(b) == 0 {
		return a
	}

	var res []T
	m := make(map[T]struct{})
	for _, i := range b {
		m[i] = struct{}{}
	}
	for _, i := range a {
		if _, ok := m[i]; !ok {
			res = append(res, i)
		}
	}
	return res
}

// Reject return a slice of elements that are in a but not in b
func Reject[T comparable](a, b []T) []T {
	if len(b) == 0 {
		return a
	}

	var res []T
	m := make(map[T]struct{})
	for _, i := range b {
		m[i] = struct{}{}
	}
	for _, i := range a {
		if _, ok := m[i]; ok {
			res = append(res, i)
		}
	}
	return res
}

// Filter return a slice of elements that satisfy the predicate
func Filter[V any](collection []V, predicate func(item V, index int) bool) []V {
	result := []V{}

	for i, item := range collection {
		if predicate(item, i) {
			result = append(result, item)
		}
	}

	return result
}

// Map return a slice of elements that are the result of applying the iteratee to each element of the collection
func Map[T any, R any](collection []T, iteratee func(item T, index int) R) []R {
	result := make([]R, len(collection))

	for i, item := range collection {
		result[i] = iteratee(item, i)
	}

	return result
}

// Reduce applies a function against an accumulator and each element in the array (from left to right) to reduce it to a single value.
func Reduce[T any, R any](collection []T, accumulator func(agg R, item T, index int) R, initial R) R {
	for i, item := range collection {
		initial = accumulator(initial, item, i)
	}

	return initial
}

// Uniq return a slice of unique elements
func Uniq[T comparable](collection []T) []T {
	result := make([]T, 0, len(collection))
	seen := make(map[T]struct{}, len(collection))

	for _, item := range collection {
		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		result = append(result, item)
	}

	return result
}

