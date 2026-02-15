package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Monitor struct {
	targets []Target
	results chan Result
	wg      sync.WaitGroup
}

func NewMonitor(targets []Target) *Monitor {
	return &Monitor{
		targets: targets,
		results: make(chan Result, len(targets)),
	}
}

func (m *Monitor) Run(ctx context.Context) {
	fmt.Println("Run started")
	for i, t := range m.targets {
		m.wg.Add(1)
		go func(t Target) {
			defer m.wg.Done()
			m.worker(ctx, t)
			fmt.Println(i)
		}(t)
	}

	go func() {
		for result := range m.results {
			fmt.Println(result)
		}
	}()

	<-ctx.Done()
	m.wg.Wait()
	close(m.results)
	fmt.Println("Run finished")
}

func (m *Monitor) worker(ctx context.Context, t Target) {
	ticker := time.NewTicker(time.Duration(t.Interval) * time.Second)
	httpChecker := NewHTTPChecker(t)
	defer ticker.Stop()
	m.results <- httpChecker.Check(ctx)

	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r := httpChecker.Check(ctx)
			fmt.Println(r)
			m.results <- r
		}
	}

}
