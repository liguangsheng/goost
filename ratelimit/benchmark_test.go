package ratelimit

import (
	"context"
	"testing"
)

func BenchmarkBucketAllow(b *testing.B) {
	bk := NewBucket(1_000_000, b.N+1)
	for i := 0; i < b.N; i++ {
		bk.Allow()
	}
}

func BenchmarkBucketWait(b *testing.B) {
	bk := NewBucket(1_000_000, b.N+1)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		_ = bk.Wait(ctx, 1)
	}
}
