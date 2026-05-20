// Package itertools collects small, generic helpers for slice manipulation.
package itertools

import "slices"

// SafeSlice returns arr[begin:end] clamped to the slice bounds. begin is
// clamped to [0, len(arr)] and end to [begin, len(arr)]. A nil input
// returns nil; an empty result returns an empty (non-nil) slice.
func SafeSlice[T any](arr []T, begin, end int) []T {
	if arr == nil {
		return nil
	}
	if begin < 0 {
		begin = 0
	}
	if end > len(arr) {
		end = len(arr)
	}
	if begin >= len(arr) || begin >= end {
		return []T{}
	}
	return arr[begin:end]
}

// Difference returns elements in a that are not in b.
func Difference[T comparable](a, b []T) []T {
	if len(b) == 0 {
		out := make([]T, len(a))
		copy(out, a)
		return out
	}

	m := make(map[T]struct{}, len(b))
	for _, i := range b {
		m[i] = struct{}{}
	}
	out := make([]T, 0, len(a))
	for _, i := range a {
		if _, ok := m[i]; !ok {
			out = append(out, i)
		}
	}
	return out
}

// Intersection returns elements present in both a and b, preserving the order
// of a and de-duplicating against b.
func Intersection[T comparable](a, b []T) []T {
	if len(b) == 0 {
		return []T{}
	}
	m := make(map[T]struct{}, len(b))
	for _, i := range b {
		m[i] = struct{}{}
	}
	out := make([]T, 0, min(len(a), len(b)))
	for _, i := range a {
		if _, ok := m[i]; ok {
			out = append(out, i)
		}
	}
	return out
}

// Filter returns elements of collection that satisfy predicate.
func Filter[V any](collection []V, predicate func(item V, index int) bool) []V {
	out := make([]V, 0, len(collection))
	for i, item := range collection {
		if predicate(item, i) {
			out = append(out, item)
		}
	}
	return out
}

// Map applies iteratee to each element of collection.
func Map[T, R any](collection []T, iteratee func(item T, index int) R) []R {
	out := make([]R, len(collection))
	for i, item := range collection {
		out[i] = iteratee(item, i)
	}
	return out
}

// Reduce folds collection from left to right using accumulator and initial.
func Reduce[T, R any](collection []T, accumulator func(agg R, item T, index int) R, initial R) R {
	for i, item := range collection {
		initial = accumulator(initial, item, i)
	}
	return initial
}

// Uniq returns collection with duplicates removed, preserving first occurrence.
func Uniq[T comparable](collection []T) []T {
	out := make([]T, 0, len(collection))
	seen := make(map[T]struct{}, len(collection))
	for _, item := range collection {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

// Contains reports whether collection has at least one element equal to v.
// Thin wrapper over slices.Contains for parity with the rest of this package.
func Contains[T comparable](collection []T, v T) bool {
	return slices.Contains(collection, v)
}

// Chunk splits collection into chunks of at most size elements. size must be > 0.
func Chunk[T any](collection []T, size int) [][]T {
	if size <= 0 {
		return nil
	}
	out := make([][]T, 0, (len(collection)+size-1)/size)
	for i := 0; i < len(collection); i += size {
		end := min(i+size, len(collection))
		out = append(out, collection[i:end])
	}
	return out
}
