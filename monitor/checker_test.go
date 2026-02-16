package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPCheckerCheck(t *testing.T) {
	t.Parallel()

	t.Run("returns status 200 on successful request", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		checker := NewHTTPChecker(Target{
			ID:      "ok-target",
			URL:     ts.URL,
			Timeout: 1,
		})

		res := checker.Check(context.Background())

		if res.Err != nil {
			t.Fatalf("expected no error, got %v", res.Err)
		}
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, res.StatusCode)
		}
		if res.TargetID != "ok-target" {
			t.Fatalf("expected target id %q, got %q", "ok-target", res.TargetID)
		}
		if res.TargetURL != ts.URL {
			t.Fatalf("expected target url %q, got %q", ts.URL, res.TargetURL)
		}
		if res.Latency < 0 {
			t.Fatalf("expected non-negative latency, got %v", res.Latency)
		}
	})

	t.Run("returns response status even for server error", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		checker := NewHTTPChecker(Target{
			ID:      "err-target",
			URL:     ts.URL,
			Timeout: 1,
		})

		res := checker.Check(context.Background())

		if res.Err != nil {
			t.Fatalf("expected no transport error, got %v", res.Err)
		}
		if res.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, res.StatusCode)
		}
	})

	t.Run("returns error when request times out", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		checker := NewHTTPChecker(Target{
			ID:      "timeout-target",
			URL:     ts.URL,
			Timeout: 0,
		})
		checker.Timeout = 50 * time.Millisecond

		res := checker.Check(context.Background())

		if res.Err == nil {
			t.Fatal("expected timeout error, got nil")
		}
		if res.StatusCode != 0 {
			t.Fatalf("expected status 0 on timeout, got %d", res.StatusCode)
		}
	})

	t.Run("returns error on invalid URL", func(t *testing.T) {
		t.Parallel()

		checker := NewHTTPChecker(Target{
			ID:      "bad-url-target",
			URL:     "://bad-url",
			Timeout: 1,
		})

		res := checker.Check(context.Background())

		if res.Err == nil {
			t.Fatal("expected URL parse error, got nil")
		}
		if res.StatusCode != 0 {
			t.Fatalf("expected status 0 on URL error, got %d", res.StatusCode)
		}
	})
}
