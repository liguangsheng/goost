// Package httpx assembles http.Client with goost building blocks:
// exponential-backoff retry, optional rate limiting, optional circuit
// breaker, and optional request logging.
//
// All knobs are off by default; New returns a plain *http.Client with
// the wrapped transport, ready to be tuned via the option setters.
package httpx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/liguangsheng/goost/backoff"
	"github.com/liguangsheng/goost/circuitbreaker"
	"github.com/liguangsheng/goost/ratelimit"
)

// RetryPolicy controls Retry.
type RetryPolicy struct {
	MaxAttempts int              // 0 = no retry
	Backoff     *backoff.Backoff // required when MaxAttempts > 0
	RetryOn     func(*http.Response, error) bool
}

// Limiter is the minimal interface httpx needs from a rate limiter.
type Limiter interface {
	Wait(ctx context.Context, n int) error
}

// Options configures NewClient. The zero value is valid.
type Options struct {
	// Base is the underlying RoundTripper. Defaults to http.DefaultTransport.
	Base http.RoundTripper
	// Timeout is the per-request timeout (across retries). 0 disables.
	Timeout time.Duration
	// Retry, when non-nil and MaxAttempts > 0, retries failed responses.
	Retry *RetryPolicy
	// Limiter, when non-nil, blocks before each request via Wait(ctx,1).
	Limiter Limiter
	// Breaker, when non-nil, short-circuits requests after enough failures.
	Breaker *circuitbreaker.Breaker
	// Logger, when non-nil, logs one summary line after each RoundTrip.
	// URL query strings and request/response bodies are not logged.
	Logger *slog.Logger
}

// New returns an *http.Client whose Transport is wrapped according to opts.
func New(opts Options) *http.Client {
	base := opts.Base
	if base == nil {
		base = http.DefaultTransport
	}
	var rt http.RoundTripper = &transport{base: base, opts: opts}
	return &http.Client{Transport: rt, Timeout: opts.Timeout}
}

type transport struct {
	base http.RoundTripper
	opts Options
}

// errOpenBreaker preserves circuitbreaker.ErrOpen as the root cause.
var errOpenBreaker = circuitbreaker.ErrOpen

// DefaultRetryOn retries 5xx responses, 429, and any transport error.
var DefaultRetryOn = func(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp == nil {
		return false
	}
	return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	attemptsMade := 0
	defer func() {
		t.logRoundTrip(req, start, attemptsMade, resp, err)
	}()

	// Apply rate limit first to avoid spending retry budget on rejections.
	if t.opts.Limiter != nil {
		if waitErr := t.opts.Limiter.Wait(req.Context(), 1); waitErr != nil {
			return nil, waitErr
		}
	}

	policy := t.opts.Retry
	retryOn := DefaultRetryOn
	if policy != nil && policy.RetryOn != nil {
		retryOn = policy.RetryOn
	}

	body, err := snapshotBody(req)
	if err != nil {
		return nil, err
	}

	attempts := 1
	if policy != nil && policy.MaxAttempts > 1 {
		attempts = policy.MaxAttempts
	}
	if policy != nil && policy.Backoff != nil {
		policy.Backoff.Reset()
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		attemptsMade++
		if body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}
		resp, lastErr = t.callOnce(req)
		if !retryOn(resp, lastErr) {
			return resp, lastErr
		}
		drain(resp)
		if i == attempts-1 || policy == nil || policy.Backoff == nil {
			break
		}
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(policy.Backoff.Next()):
		}
	}
	return resp, lastErr
}

func (t *transport) logRoundTrip(req *http.Request, start time.Time, attempts int, resp *http.Response, err error) {
	if t.opts.Logger == nil {
		return
	}
	status := 0
	if resp != nil {
		status = resp.StatusCode
	}
	attrs := []slog.Attr{
		slog.String("method", req.Method),
		slog.String("scheme", req.URL.Scheme),
		slog.String("host", req.URL.Host),
		slog.String("path", req.URL.Path),
		slog.Int("status", status),
		slog.Int("attempts", attempts),
		slog.Duration("duration", time.Since(start)),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	t.opts.Logger.LogAttrs(req.Context(), slog.LevelInfo, "httpx request", attrs...)
}

func (t *transport) callOnce(req *http.Request) (*http.Response, error) {
	if t.opts.Breaker == nil {
		return t.base.RoundTrip(req)
	}
	var resp *http.Response
	err := t.opts.Breaker.Do(req.Context(), func(_ context.Context) error {
		r, e := t.base.RoundTrip(req)
		resp = r
		if e != nil {
			return e
		}
		if r.StatusCode >= 500 {
			// 5xx counts against the breaker.
			return errBadStatus
		}
		return nil
	})
	if errors.Is(err, errOpenBreaker) {
		return nil, errOpenBreaker
	}
	if errors.Is(err, errBadStatus) {
		// breaker recorded a failure but we still want the response back
		return resp, nil
	}
	return resp, err
}

var errBadStatus = errors.New("httpx: 5xx response")

func snapshotBody(req *http.Request) ([]byte, error) {
	if req.Body == nil || req.GetBody != nil {
		return nil, nil
	}
	all, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	_ = req.Body.Close()
	return all, nil
}

func drain(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
}

var _ = ratelimit.NewBucket // keep import in case Options.Limiter docs reference it
