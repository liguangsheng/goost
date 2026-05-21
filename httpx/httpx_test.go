package httpx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/liguangsheng/goost/backoff"
	"github.com/liguangsheng/goost/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

func Test_PlainRoundtrip(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer s.Close()

	c := New(Options{})
	resp, err := c.Get(s.URL)
	assert.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "ok", string(body))
}

func Test_RetryOn5xx(t *testing.T) {
	var calls atomic.Int64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = io.WriteString(w, "ok")
	}))
	defer s.Close()

	c := New(Options{
		Retry: &RetryPolicy{
			MaxAttempts: 5,
			Backoff:     &backoff.Backoff{Initial: time.Millisecond, Max: 10 * time.Millisecond},
		},
	})
	resp, err := c.Get(s.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.EqualValues(t, 3, calls.Load())
}

func Test_OnRetryReportsRetryableAttempts(t *testing.T) {
	var calls atomic.Int64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	var events []RetryEvent
	c := New(Options{
		Retry: &RetryPolicy{
			MaxAttempts: 3,
			Backoff:     &backoff.Backoff{Initial: time.Millisecond, Max: time.Millisecond},
			OnRetry: func(e RetryEvent) {
				events = append(events, e)
			},
		},
	})

	resp, err := c.Get(s.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Len(t, events, 2)
	assert.Equal(t, RetryEvent{Attempt: 1, MaxAttempts: 3, StatusCode: http.StatusTooManyRequests, Delay: time.Millisecond}, events[0])
	assert.Equal(t, RetryEvent{Attempt: 2, MaxAttempts: 3, StatusCode: http.StatusTooManyRequests, Delay: time.Millisecond}, events[1])
}

func Test_OnRetryReportsTransportError(t *testing.T) {
	want := errors.New("network down")
	var events []RetryEvent
	c := New(Options{
		Base: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, want
		}),
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Millisecond},
			OnRetry: func(e RetryEvent) {
				events = append(events, e)
			},
		},
	})

	_, err := c.Get("https://api.example.com/users")
	assert.ErrorIs(t, err, want)
	assert.Len(t, events, 1)
	assert.Equal(t, 1, events[0].Attempt)
	assert.Equal(t, 2, events[0].MaxAttempts)
	assert.Equal(t, 0, events[0].StatusCode)
	assert.ErrorIs(t, events[0].Err, want)
	assert.Equal(t, time.Millisecond, events[0].Delay)
}

func Test_RetryGivesUp(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer s.Close()

	c := New(Options{
		Retry: &RetryPolicy{
			MaxAttempts: 3,
			Backoff:     &backoff.Backoff{Initial: time.Millisecond},
		},
	})
	resp, err := c.Get(s.URL)
	assert.NoError(t, err) // 502 is still a "response"; not an error
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func Test_BreakerOpens(t *testing.T) {
	var calls atomic.Int64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer s.Close()

	b := circuitbreaker.New(circuitbreaker.Config{FailureThreshold: 2, CooldownPeriod: time.Minute})
	c := New(Options{Breaker: b})

	for range 5 {
		resp, _ := c.Get(s.URL)
		if resp != nil {
			_ = resp.Body.Close()
		}
	}
	// After 2 5xx the breaker opens; subsequent calls bypass the server.
	assert.LessOrEqual(t, calls.Load(), int64(2))
}

func Test_LimiterWaits(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ok")
	}))
	defer s.Close()

	lim := &fakeLimiter{}
	c := New(Options{Limiter: lim})
	_, _ = c.Get(s.URL)
	assert.EqualValues(t, 1, lim.calls.Load())
}

func Test_LimiterWaitsBeforeEachRetryAttempt(t *testing.T) {
	var calls atomic.Int64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	lim := &fakeLimiter{}
	c := New(Options{
		Limiter: lim,
		Retry:   &RetryPolicy{MaxAttempts: 3, Backoff: &backoff.Backoff{Initial: time.Millisecond}},
	})

	resp, err := c.Get(s.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.EqualValues(t, 3, calls.Load())
	assert.EqualValues(t, 3, lim.calls.Load())
}

func Test_LimiterErrorBeforeRetryStopsWithoutAnotherRequest(t *testing.T) {
	want := errors.New("blocked retry")
	var calls atomic.Int64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer s.Close()

	lim := &failAfterLimiter{failAfter: 1, err: want}
	c := New(Options{
		Limiter: lim,
		Retry:   &RetryPolicy{MaxAttempts: 3, Backoff: &backoff.Backoff{Initial: time.Millisecond}},
	})

	_, err := c.Get(s.URL)
	assert.ErrorIs(t, err, want)
	assert.EqualValues(t, 1, calls.Load())
	assert.EqualValues(t, 2, lim.calls.Load())
}

type fakeLimiter struct{ calls atomic.Int64 }

func (f *fakeLimiter) Wait(_ context.Context, _ int) error {
	f.calls.Add(1)
	return nil
}

type failAfterLimiter struct {
	calls     atomic.Int64
	failAfter int64
	err       error
}

func (f *failAfterLimiter) Wait(_ context.Context, _ int) error {
	if f.calls.Add(1) > f.failAfter {
		return f.err
	}
	return nil
}

func Test_LimiterErrorAborts(t *testing.T) {
	c := New(Options{Limiter: errLimiter{err: errors.New("blocked")}})
	_, err := c.Get("http://localhost")
	assert.Error(t, err)
}

type errLimiter struct{ err error }

func (e errLimiter) Wait(_ context.Context, _ int) error { return e.err }

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func Test_BodyResnapshotForRetry(t *testing.T) {
	var calls atomic.Int64
	var lastBody atomic.Value
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		lastBody.Store(string(body))
		if calls.Add(1) < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = io.WriteString(w, "ok")
	}))
	defer s.Close()

	c := New(Options{
		Retry: &RetryPolicy{MaxAttempts: 3, Backoff: &backoff.Backoff{Initial: time.Millisecond}},
	})
	resp, err := c.Post(s.URL, "text/plain", strings.NewReader("payload"))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "payload", lastBody.Load())
}

func Test_BodyGetBodyReplayForRetry(t *testing.T) {
	var bodies []string
	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			bodies = append(bodies, string(body))
			status := http.StatusOK
			if len(bodies) == 1 {
				status = http.StatusInternalServerError
			}
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    req,
			}, nil
		}),
		Retry: &RetryPolicy{MaxAttempts: 2, Backoff: &backoff.Backoff{Initial: time.Millisecond}},
	})

	req, err := http.NewRequest(http.MethodPost, "https://api.example.com/users", strings.NewReader("payload"))
	assert.NoError(t, err)

	resp, err := c.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []string{"payload", "payload"}, bodies)
}

func Test_BodyGetBodyErrorStopsRetryAfterFirstAttempt(t *testing.T) {
	want := errors.New("cannot replay body")
	var calls atomic.Int64
	var bodies []string
	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			calls.Add(1)
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			bodies = append(bodies, string(body))
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    req,
			}, nil
		}),
		Retry: &RetryPolicy{MaxAttempts: 2, Backoff: &backoff.Backoff{Initial: time.Millisecond}},
	})

	req, err := http.NewRequest(http.MethodPost, "https://api.example.com/users", nil)
	assert.NoError(t, err)
	req.Body = io.NopCloser(strings.NewReader("payload"))
	req.GetBody = func() (io.ReadCloser, error) {
		return nil, want
	}

	_, err = c.Do(req)
	assert.ErrorIs(t, err, want)
	assert.EqualValues(t, 1, calls.Load())
	assert.Equal(t, []string{"payload"}, bodies)
}

func Test_LoggerRecordsFinalRequestSummary(t *testing.T) {
	var calls atomic.Int64
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 2 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer s.Close()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	c := New(Options{
		Logger: logger,
		Retry:  &RetryPolicy{MaxAttempts: 2, Backoff: &backoff.Backoff{Initial: time.Millisecond}},
	})

	resp, err := c.Get(s.URL + "/users?token=secret")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	out := buf.String()
	assert.Contains(t, out, "msg=\"httpx request\"")
	assert.Contains(t, out, "method=GET")
	assert.Contains(t, out, "path=/users")
	assert.Contains(t, out, "status=201")
	assert.Contains(t, out, "attempts=2")
	assert.NotContains(t, out, "token=secret")
}

func Test_LoggerRecordsLimiterError(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	c := New(Options{
		Logger:  logger,
		Limiter: errLimiter{err: errors.New("blocked")},
	})

	_, err := c.Get("https://api.example.com/users")
	assert.Error(t, err)

	out := buf.String()
	assert.Contains(t, out, "status=0")
	assert.Contains(t, out, "attempts=0")
	assert.Contains(t, out, "error=blocked")
}
