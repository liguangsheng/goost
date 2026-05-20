// Package singleflight is a generic thin wrapper around
// golang.org/x/sync/singleflight: concurrent calls to Do with the same key
// share a single execution of fn.
package singleflight

import (
	"golang.org/x/sync/singleflight"
)

// Group deduplicates in-flight calls to fn keyed by K, returning value V.
//
// All callers that hit Do with the same key while a call is in flight
// observe the same result (value, err, shared).
type Group[K comparable, V any] struct {
	g  singleflight.Group
	mk func(K) string
}

// New returns a Group whose keys are turned into strings by stringify. For
// string-keyed groups, use NewString.
func New[K comparable, V any](stringify func(K) string) *Group[K, V] {
	return &Group[K, V]{mk: stringify}
}

// NewString returns a Group whose key type is string.
func NewString[V any]() *Group[string, V] {
	return &Group[string, V]{mk: func(s string) string { return s }}
}

// Do executes fn for key, suppressing duplicate concurrent calls.
// The returned shared bool indicates whether v was given to multiple callers.
func (g *Group[K, V]) Do(key K, fn func() (V, error)) (v V, err error, shared bool) {
	raw, e, s := g.g.Do(g.mk(key), func() (any, error) {
		return fn()
	})
	if raw != nil {
		v = raw.(V)
	}
	return v, e, s
}

// Forget removes any in-flight call record for key, so the next Do for the
// same key starts a fresh execution. Has no effect once the existing call
// has completed.
func (g *Group[K, V]) Forget(key K) {
	g.g.Forget(g.mk(key))
}
