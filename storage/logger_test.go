package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/buraksaglam089/go-healthcheck/monitor"
)

func TestFileLoggerSaveLogAppendsJSONLine(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "results.jsonl")

	logger := NewFileLogger(logFile)
	input := monitor.Result{
		TargetID:   "google",
		TargetURL:  "https://google.com",
		StatusCode: 200,
		Latency:    25 * time.Millisecond,
		Timestamp:  time.Now().UTC().Truncate(time.Second),
		Err:        nil,
	}

	if err := logger.SaveLog(input); err != nil {
		t.Fatalf("SaveLog returned error: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 log line, got %d", len(lines))
	}

	var got monitor.Result
	if err := json.Unmarshal([]byte(lines[0]), &got); err != nil {
		t.Fatalf("failed to unmarshal log line: %v", err)
	}

	if got.TargetID != input.TargetID {
		t.Fatalf("expected target id %q, got %q", input.TargetID, got.TargetID)
	}
	if got.TargetURL != input.TargetURL {
		t.Fatalf("expected target url %q, got %q", input.TargetURL, got.TargetURL)
	}
	if got.StatusCode != input.StatusCode {
		t.Fatalf("expected status code %d, got %d", input.StatusCode, got.StatusCode)
	}
	if got.Err != nil {
		t.Fatalf("expected nil error, got %v", got.Err)
	}
}

