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
	"sync"
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

	resp, err := c.Get(s.URL + "/users?token=secret")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Len(t, events, 2)
	assert.Equal(t, RetryEvent{
		Method:      http.MethodGet,
		Scheme:      "http",
		Host:        strings.TrimPrefix(s.URL, "http://"),
		Path:        "/users",
		Attempt:     1,
		MaxAttempts: 3,
		StatusCode:  http.StatusTooManyRequests,
		Delay:       time.Millisecond,
	}, events[0])
	assert.Equal(t, RetryEvent{
		Method:      http.MethodGet,
		Scheme:      "http",
		Host:        strings.TrimPrefix(s.URL, "http://"),
		Path:        "/users",
		Attempt:     2,
		MaxAttempts: 3,
		StatusCode:  http.StatusTooManyRequests,
		Delay:       time.Millisecond,
	}, events[1])
	assert.NotContains(t, events[0].Path, "token=secret")
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
		_, _ = io.WriteString(w, "still failing")
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
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "still failing", string(body))
	assert.NoError(t, resp.Body.Close())
}

func Test_OnGiveUpReportsFinalRetryableResponse(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer s.Close()

	var retries []RetryEvent
	var giveUps []RetryEvent
	c := New(Options{
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Millisecond},
			OnRetry: func(e RetryEvent) {
				retries = append(retries, e)
			},
			OnGiveUp: func(e RetryEvent) {
				giveUps = append(giveUps, e)
			},
		},
	})

	resp, err := c.Get(s.URL + "/users?token=secret")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
	assert.NoError(t, resp.Body.Close())
	assert.Len(t, retries, 1)
	assert.Len(t, giveUps, 1)
	assert.Equal(t, RetryEvent{
		Method:      http.MethodGet,
		Scheme:      "http",
		Host:        strings.TrimPrefix(s.URL, "http://"),
		Path:        "/users",
		Attempt:     2,
		MaxAttempts: 2,
		StatusCode:  http.StatusBadGateway,
	}, giveUps[0])
	assert.NotContains(t, giveUps[0].Path, "token=secret")
	assert.Zero(t, giveUps[0].Delay)
}

func Test_OnGiveUpReportsFinalTransportError(t *testing.T) {
	want := errors.New("network down")
	var giveUps []RetryEvent
	c := New(Options{
		Base: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, want
		}),
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Millisecond},
			OnGiveUp: func(e RetryEvent) {
				giveUps = append(giveUps, e)
			},
		},
	})

	_, err := c.Get("https://api.example.com/users")
	assert.ErrorIs(t, err, want)
	assert.Len(t, giveUps, 1)
	assert.Equal(t, http.MethodGet, giveUps[0].Method)
	assert.Equal(t, "https", giveUps[0].Scheme)
	assert.Equal(t, "api.example.com", giveUps[0].Host)
	assert.Equal(t, "/users", giveUps[0].Path)
	assert.Equal(t, 2, giveUps[0].Attempt)
	assert.Equal(t, 2, giveUps[0].MaxAttempts)
	assert.Equal(t, 0, giveUps[0].StatusCode)
	assert.ErrorIs(t, giveUps[0].Err, want)
	assert.Zero(t, giveUps[0].Delay)
}

func Test_OnRetryRunsBeforeIntermediateResponseIsClosed(t *testing.T) {
	firstBody := trackingBody("retryable")
	finalBody := trackingBody("ok")
	var calls atomic.Int64
	var bodyClosedInCallback atomic.Bool

	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			status := http.StatusOK
			body := finalBody
			if calls.Add(1) == 1 {
				status = http.StatusBadGateway
				body = firstBody
			}
			return testResponse(req, status, body), nil
		}),
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Nanosecond},
			OnRetry: func(RetryEvent) {
				bodyClosedInCallback.Store(firstBody.closed.Load())
			},
		},
	})

	resp, err := c.Get("https://api.example.com/users")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.False(t, bodyClosedInCallback.Load())
	assert.True(t, firstBody.closed.Load())
	assert.False(t, finalBody.closed.Load())
	assert.NoError(t, resp.Body.Close())
}

func Test_OnGiveUpRunsBeforeFinalResponseIsReturnedOpen(t *testing.T) {
	body := trackingBody("final")
	var bodyClosedInCallback atomic.Bool

	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return testResponse(req, http.StatusBadGateway, body), nil
		}),
		Retry: &RetryPolicy{
			MaxAttempts: 1,
			Backoff:     &backoff.Backoff{Initial: time.Nanosecond},
			OnGiveUp: func(RetryEvent) {
				bodyClosedInCallback.Store(body.closed.Load())
			},
		},
	})

	resp, err := c.Get("https://api.example.com/users")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
	assert.False(t, bodyClosedInCallback.Load())
	assert.False(t, body.closed.Load())

	got, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "final", string(got))
	assert.NoError(t, resp.Body.Close())
	assert.True(t, body.closed.Load())
}

func Test_OnGiveUpIgnoresNonRetryableResponseAndSuccessAfterRetry(t *testing.T) {
	t.Run("non retryable", func(t *testing.T) {
		var giveUps atomic.Int64
		c := New(Options{
			Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("")), Request: req}, nil
			}),
			Retry: &RetryPolicy{
				MaxAttempts: 2,
				Backoff:     &backoff.Backoff{Initial: time.Millisecond},
				OnGiveUp: func(RetryEvent) {
					giveUps.Add(1)
				},
			},
		})

		resp, err := c.Get("https://api.example.com/users")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.EqualValues(t, 0, giveUps.Load())
	})

	t.Run("success after retry", func(t *testing.T) {
		var calls atomic.Int64
		var giveUps atomic.Int64
		c := New(Options{
			Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				status := http.StatusOK
				if calls.Add(1) == 1 {
					status = http.StatusBadGateway
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("")), Request: req}, nil
			}),
			Retry: &RetryPolicy{
				MaxAttempts: 2,
				Backoff:     &backoff.Backoff{Initial: time.Millisecond},
				OnGiveUp: func(RetryEvent) {
					giveUps.Add(1)
				},
			},
		})

		resp, err := c.Get("https://api.example.com/users")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.EqualValues(t, 0, giveUps.Load())
	})
}

func Test_RetryBackoffSequenceIsPerRequest(t *testing.T) {
	var firstAttempts atomic.Int64
	firstAttemptsDone := make(chan struct{})
	firstRetryDone := make(chan struct{})
	var closeFirstRetryDone sync.Once

	var mu sync.Mutex
	delaysByPath := map[string]time.Duration{}

	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("attempt") == "" {
				req.Header.Set("attempt", "retried")
				if firstAttempts.Add(1) == 2 {
					close(firstAttemptsDone)
				}
				return testResponse(req, http.StatusBadGateway, stringBody("")), nil
			}
			return testResponse(req, http.StatusNoContent, stringBody("")), nil
		}),
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Nanosecond, Factor: 1000, Max: time.Hour},
			RetryOn: func(resp *http.Response, err error) bool {
				if resp == nil || resp.StatusCode != http.StatusBadGateway {
					return DefaultRetryOn(resp, err)
				}
				switch resp.Request.URL.Path {
				case "/first":
					<-firstAttemptsDone
					return true
				case "/second":
					<-firstRetryDone
					return true
				default:
					return true
				}
			},
			OnRetry: func(e RetryEvent) {
				mu.Lock()
				delaysByPath[e.Path] = e.Delay
				mu.Unlock()
				if e.Path == "/first" {
					closeFirstRetryDone.Do(func() { close(firstRetryDone) })
				}
			},
		},
	})

	errs := make(chan error, 2)
	for _, path := range []string{"/first", "/second"} {
		go func() {
			resp, err := c.Get("https://api.example.com" + path)
			if resp != nil {
				_ = resp.Body.Close()
			}
			errs <- err
		}()
	}

	assert.NoError(t, <-errs)
	assert.NoError(t, <-errs)
	assert.Equal(t, map[string]time.Duration{
		"/first":  time.Nanosecond,
		"/second": time.Nanosecond,
	}, delaysByPath)
}

func Test_RetryDelayStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int64
	body := trackingBody("retryable")

	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			calls.Add(1)
			return testResponse(req, http.StatusBadGateway, body), nil
		}),
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Hour},
			OnRetry: func(RetryEvent) {
				cancel()
			},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/users", nil)
	assert.NoError(t, err)

	resp, err := c.Do(req)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, context.Canceled)
	assert.EqualValues(t, 1, calls.Load())
	assert.True(t, body.closed.Load())
}

func Test_RetryClosesIntermediateResponsesAndLeavesFinalBodyOpen(t *testing.T) {
	firstBody := trackingBody("retryable")
	finalBody := trackingBody("ok")
	var calls atomic.Int64

	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			status := http.StatusOK
			body := finalBody
			if calls.Add(1) == 1 {
				status = http.StatusBadGateway
				body = firstBody
			}
			return testResponse(req, status, body), nil
		}),
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Nanosecond},
		},
	})

	resp, err := c.Get("https://api.example.com/users")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.EqualValues(t, 2, calls.Load())
	assert.True(t, firstBody.closed.Load())
	assert.False(t, finalBody.closed.Load())

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "ok", string(body))
	assert.NoError(t, resp.Body.Close())
	assert.True(t, finalBody.closed.Load())
}

func Test_RetryGiveUpLeavesFinalResponseBodyOpen(t *testing.T) {
	firstBody := trackingBody("first")
	finalBody := trackingBody("final")
	var calls atomic.Int64

	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			body := firstBody
			if calls.Add(1) == 2 {
				body = finalBody
			}
			return testResponse(req, http.StatusBadGateway, body), nil
		}),
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Nanosecond},
		},
	})

	resp, err := c.Get("https://api.example.com/users")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
	assert.EqualValues(t, 2, calls.Load())
	assert.True(t, firstBody.closed.Load())
	assert.False(t, finalBody.closed.Load())

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "final", string(body))
	assert.NoError(t, resp.Body.Close())
	assert.True(t, finalBody.closed.Load())
}

func Test_CustomRetryOnClosesIntermediateResponses(t *testing.T) {
	firstBody := trackingBody("conflict")
	finalBody := trackingBody("created")
	var calls atomic.Int64

	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			status := http.StatusCreated
			body := finalBody
			if calls.Add(1) == 1 {
				status = http.StatusConflict
				body = firstBody
			}
			return testResponse(req, status, body), nil
		}),
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Nanosecond},
			RetryOn: func(resp *http.Response, err error) bool {
				return err != nil || resp.StatusCode == http.StatusConflict
			},
		},
	})

	resp, err := c.Get("https://api.example.com/users")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.EqualValues(t, 2, calls.Load())
	assert.True(t, firstBody.closed.Load())
	assert.False(t, finalBody.closed.Load())

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "created", string(body))
	assert.NoError(t, resp.Body.Close())
	assert.True(t, finalBody.closed.Load())
}

func Test_CustomRetryOnCanReturnRetryableStatusBodyToCaller(t *testing.T) {
	body := trackingBody("server error")
	var calls atomic.Int64

	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			calls.Add(1)
			return testResponse(req, http.StatusInternalServerError, body), nil
		}),
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Nanosecond},
			RetryOn: func(*http.Response, error) bool {
				return false
			},
		},
	})

	resp, err := c.Get("https://api.example.com/users")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.EqualValues(t, 1, calls.Load())
	assert.False(t, body.closed.Load())

	got, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "server error", string(got))
	assert.NoError(t, resp.Body.Close())
	assert.True(t, body.closed.Load())
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

func stringBody(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

type trackingReadCloser struct {
	*strings.Reader
	closed atomic.Bool
}

func trackingBody(s string) *trackingReadCloser {
	return &trackingReadCloser{Reader: strings.NewReader(s)}
}

func (r *trackingReadCloser) Close() error {
	r.closed.Store(true)
	return nil
}

func testResponse(req *http.Request, status int, body io.ReadCloser) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       body,
		Request:    req,
	}
}

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

func Test_LoggerRecordsRetryGiveUpSummary(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return testResponse(req, http.StatusBadGateway, stringBody("")), nil
		}),
		Logger: logger,
		Retry:  &RetryPolicy{MaxAttempts: 2, Backoff: &backoff.Backoff{Initial: time.Nanosecond}},
	})

	resp, err := c.Get("https://api.example.com/users?token=secret")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
	assert.NoError(t, resp.Body.Close())

	out := buf.String()
	assert.Contains(t, out, "status=502")
	assert.Contains(t, out, "attempts=2")
	assert.NotContains(t, out, "error=")
	assert.NotContains(t, out, "token=secret")
}

func Test_LoggerRecordsContextCancelDuringRetryDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	c := New(Options{
		Base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return testResponse(req, http.StatusBadGateway, stringBody("")), nil
		}),
		Logger: logger,
		Retry: &RetryPolicy{
			MaxAttempts: 2,
			Backoff:     &backoff.Backoff{Initial: time.Hour},
			OnRetry: func(RetryEvent) {
				cancel()
			},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/users?token=secret", nil)
	assert.NoError(t, err)

	resp, err := c.Do(req)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, context.Canceled)

	out := buf.String()
	assert.Contains(t, out, "status=502")
	assert.Contains(t, out, "attempts=1")
	assert.Contains(t, out, "error=\"context canceled\"")
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
