package goost

import (
	"io"
	"testing"
	"time"

	"github.com/liguangsheng/goost/clock"
	"github.com/liguangsheng/goost/httpx"
	"github.com/liguangsheng/goost/ratelimit"
	"github.com/liguangsheng/goost/rotatingwriter"
)

// Compile-time interface satisfaction checks.

var _ io.Writer = (*rotatingwriter.RotatingWriter)(nil)
var _ rotatingwriter.Rotater = (*rotatingwriter.DailyRotater)(nil)
var _ rotatingwriter.Rotater = (*rotatingwriter.SizeRotater)(nil)
var _ clock.Clock = (*clock.Mock)(nil)
var _ httpx.Limiter = (*ratelimit.Bucket)(nil)

func TestConsumerContractTypes(t *testing.T) {
	t.Parallel()

	t.Run("rotatingwriter implements io.Writer", func(t *testing.T) {
		t.Parallel()
		w := &rotatingwriter.RotatingWriter{}
		_ = w
	})

	t.Run("clock.Mock implements clock.Clock", func(t *testing.T) {
		var c clock.Clock = clock.NewMock(time.Now())
		_ = c.Now()
	})

	t.Run("ratelimit.Bucket implements httpx.Limiter", func(t *testing.T) {
		var _ httpx.Limiter = ratelimit.NewBucket(1, 1)
	})
}
