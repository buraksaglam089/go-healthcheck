package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ResultLogger interface {
	SaveLog(Result) error
}

type Monitor struct {
	targets []Target
	results chan Result
	logger  ResultLogger
	wg      sync.WaitGroup
}

func NewMonitor(targets []Target, logger ResultLogger) *Monitor {
	return &Monitor{
		targets: targets,
		results: make(chan Result, len(targets)),
		logger:  logger,
	}
}

func (m *Monitor) Run(ctx context.Context) {
	fmt.Println("Run started")
	for _, t := range m.targets {
		m.wg.Add(1)
		go func(t Target) {
			defer m.wg.Done()
			m.worker(ctx, t)
		}(t)
	}

	go func() {
		m.wg.Wait()
		close(m.results)
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Run cancelled")
			return
		case result, ok := <-m.results:
			if !ok {
				fmt.Println("Run finished")
				return
			}
			if m.logger != nil {
				if err := m.logger.SaveLog(result); err != nil {
					fmt.Printf("failed to save log for %s: %v\n", result.TargetURL, err)
				}
			}
		}
	}
}

func (m *Monitor) worker(ctx context.Context, t Target) {
	ticker := time.NewTicker(time.Duration(t.Interval) * time.Second)
	httpChecker := NewHTTPChecker(t)
	defer ticker.Stop()
	m.results <- httpChecker.Check(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r := httpChecker.Check(ctx)
			m.results <- r
		}
	}

}
