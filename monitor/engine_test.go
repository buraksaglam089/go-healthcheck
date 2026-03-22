package monitor

import (
	"context"
	"fmt"
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

func TestMonitorRunRespectsWorkerLimit(t *testing.T) {
	t.Parallel()

	const (
		workerCount = 2
		targetCount = 5
	)

	release := make(chan struct{})
	started := make(chan struct{}, targetCount)

	var mu sync.Mutex
	startedCount := 0
	inFlight := 0
	maxInFlight := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		startedCount++
		inFlight++
		if inFlight > maxInFlight {
			maxInFlight = inFlight
		}
		mu.Unlock()

		select {
		case started <- struct{}{}:
		default:
		}

		defer func() {
			mu.Lock()
			inFlight--
			mu.Unlock()
		}()

		select {
		case <-release:
		case <-r.Context().Done():
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	targets := make([]Target, 0, targetCount)
	for i := 0; i < targetCount; i++ {
		targets = append(targets, Target{
			ID:       fmt.Sprintf("t%d", i),
			URL:      server.URL,
			Interval: 10,
			Timeout:  1,
		})
	}

	logger := &memoryResultLogger{}
	m := NewMonitor(targets, logger, WithWorkerCount(workerCount))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		m.Run(ctx)
		close(done)
	}()

	for i := 0; i < workerCount; i++ {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatalf("expected %d checks to start", workerCount)
		}
	}

	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	currentStarted := startedCount
	currentMax := maxInFlight
	mu.Unlock()

	if currentStarted != workerCount {
		t.Fatalf("expected only %d checks to start before workers were released, got %d", workerCount, currentStarted)
	}
	if currentMax != workerCount {
		t.Fatalf("expected max concurrent checks to be %d, got %d", workerCount, currentMax)
	}

	cancel()
	close(release)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("monitor did not stop after cancel")
	}
}
