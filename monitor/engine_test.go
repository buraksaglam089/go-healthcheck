package monitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type memoryResultLogger struct {
	mu      sync.Mutex
	results []Result
}

func (l *memoryResultLogger) SaveLog(r Result) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.results = append(l.results, r)
	return nil
}

func (l *memoryResultLogger) snapshot() []Result {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Result, len(l.results))
	copy(out, l.results)
	return out
}

func TestMonitorRunLogsResultsAndStopsOnCancel(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	targets := []Target{
		{ID: "t1", URL: server.URL, Interval: 1, Timeout: 1},
		{ID: "t2", URL: server.URL, Interval: 1, Timeout: 1},
	}

	logger := &memoryResultLogger{}
	m := NewMonitor(targets, logger)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		m.Run(ctx)
		close(done)
	}()

	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("monitor did not stop after cancel")
	}

	results := logger.snapshot()
	if len(results) < len(targets) {
		t.Fatalf("expected at least %d results, got %d", len(targets), len(results))
	}

	for _, r := range results {
		if r.TargetURL != server.URL {
			t.Fatalf("expected target url %q, got %q", server.URL, r.TargetURL)
		}
	}
}

