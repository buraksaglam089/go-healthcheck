package monitor

import (
	"context"
	"net/http"
	"time"
)

type Target struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout"`
}

type Result struct {
	TargetID   string
	TargetURL  string
	StatusCode int
	Latency    time.Duration
	Timestamp  time.Time
	Err        error
}

type HTTPChecker struct {
	ID         string        `json:"id"`
	URL        string        `json:"url"`
	Timeout    time.Duration `json:"timeout"`
	httpClient *http.Client
}

type Checker interface {
	Check(ctx context.Context) Result
}

func NewHTTPChecker(t Target) *HTTPChecker {
	return &HTTPChecker{
		ID:         t.ID,
		URL:        t.URL,
		Timeout:    time.Duration(t.Timeout) * time.Second,
		httpClient: &http.Client{},
	}
}

func (h *HTTPChecker) Check(ctx context.Context) Result {
	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	start := time.Now()
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.URL, nil)
	if err != nil {
		return Result{
			TargetID:   h.ID,
			TargetURL:  h.URL,
			StatusCode: 0,
			Latency:    time.Since(start),
			Timestamp:  start,
			Err:        err,
		}
	}
	resp, err := h.httpClient.Do(req)
	if err != nil {
		return Result{
			TargetID:   h.ID,
			TargetURL:  h.URL,
			StatusCode: 0,
			Latency:    time.Since(start),
			Timestamp:  start,
			Err:        err,
		}
	}
	defer resp.Body.Close()

	return Result{
		TargetID:   h.ID,
		TargetURL:  h.URL,
		StatusCode: resp.StatusCode,
		Latency:    time.Since(start),
		Timestamp:  start,
		Err:        nil,
	}
}
