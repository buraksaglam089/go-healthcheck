package monitor

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type ResultLogger interface {
	SaveLog(Result) error
}

type MonitorOption func(*Monitor)

type Monitor struct {
	targets     []Target
	logger      ResultLogger
	workerCount int
}

type checkJob struct {
	checker Checker
	done    chan struct{}
}

func NewMonitor(targets []Target, logger ResultLogger, opts ...MonitorOption) *Monitor {
	m := &Monitor{
		targets:     targets,
		logger:      logger,
		workerCount: defaultWorkerCount(len(targets)),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func WithWorkerCount(n int) MonitorOption {
	return func(m *Monitor) {
		if n > 0 {
			m.workerCount = n
		}
	}
}

func (m *Monitor) Run(ctx context.Context) {
	if len(m.targets) == 0 || m.workerCount == 0 {
		return
	}

	jobs := make(chan checkJob)
	results := make(chan Result, m.workerCount)

	var workerWg sync.WaitGroup
	workerWg.Add(m.workerCount)
	for i := 0; i < m.workerCount; i++ {
		go func() {
			defer workerWg.Done()
			runWorker(ctx, jobs, results)
		}()
	}

	var schedulerWg sync.WaitGroup
	schedulerWg.Add(len(m.targets))
	for _, target := range m.targets {
		checker := NewHTTPChecker(target)
		go func(target Target, checker Checker) {
			defer schedulerWg.Done()
			scheduleChecks(ctx, target, checker, jobs)
		}(target, checker)
	}

	go func() {
		schedulerWg.Wait()
		close(jobs)
	}()

	go func() {
		workerWg.Wait()
		close(results)
	}()

	for result := range results {
		if m.logger == nil {
			continue
		}

		if err := m.logger.SaveLog(result); err != nil {
			fmt.Printf("failed to save log for %s: %v\n", result.TargetURL, err)
		}
	}
}

func scheduleChecks(ctx context.Context, target Target, checker Checker, jobs chan<- checkJob) {
	ticker := time.NewTicker(time.Duration(target.Interval) * time.Second)
	defer ticker.Stop()

	if !dispatchCheck(ctx, checker, jobs) {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !dispatchCheck(ctx, checker, jobs) {
				return
			}
		}
	}
}

func dispatchCheck(ctx context.Context, checker Checker, jobs chan<- checkJob) bool {
	done := make(chan struct{})
	job := checkJob{
		checker: checker,
		done:    done,
	}

	select {
	case <-ctx.Done():
		return false
	case jobs <- job:
	}

	select {
	case <-ctx.Done():
		return false
	case <-done:
		return true
	}
}

func runWorker(ctx context.Context, jobs <-chan checkJob, results chan<- Result) {
	for job := range jobs {
		result := job.checker.Check(ctx)
		close(job.done)
		results <- result
	}
}

func defaultWorkerCount(targetCount int) int {
	if targetCount == 0 {
		return 0
	}

	return min(targetCount, max(1, runtime.GOMAXPROCS(0)))
}
