package engine

import (
	"context"
	"sync"
	"time"

	"github.com/basedalex/load-tester/internal/config"
	"github.com/basedalex/load-tester/internal/httpclient"
	"github.com/basedalex/load-tester/internal/metrics"
	"github.com/basedalex/load-tester/internal/worker"
)

type Engine struct {
	cfg       *config.Config
	client    httpclient.Client
	collector *metrics.Collector
}

func New(cfg *config.Config, client httpclient.Client) *Engine {
	return &Engine{
		cfg:       cfg,
		client:    client,
		collector: metrics.NewCollector(),
	}
}

func (e *Engine) Run(ctx context.Context) (metrics.Summary, error) {
	if err := e.cfg.Validate(); err != nil {
		return metrics.Summary{}, err
	}

	runCtx, cancel := e.buildContext(ctx)
	defer cancel()

	jobs := make(chan struct{}, e.cfg.Concurrency)

	request := httpclient.Request{
		URL:     e.cfg.TargetURL,
		Method:  e.cfg.Method,
		Headers: e.cfg.Headers,
		Body:    e.cfg.Body,
		Timeout: e.cfg.RequestTimeout,
	}

	var wg sync.WaitGroup
	for i := 0; i < e.cfg.Concurrency; i++ {
		wg.Add(1)
		w := worker.New(i, e.client, request, e.collector)
		go func() {
			defer wg.Done()
			w.Run(runCtx, jobs)
		}()
	}

	e.dispatchJobs(runCtx, jobs)

	wg.Wait()

	return e.collector.Summarize(), nil
}

func (e *Engine) Collector() *metrics.Collector {
	return e.collector
}

func (e *Engine) buildContext(parent context.Context) (context.Context, context.CancelFunc) {
	if e.cfg.Duration > 0 {
		return context.WithTimeout(parent, e.cfg.Duration)
	}
	return context.WithCancel(parent)
}

func (e *Engine) dispatchJobs(ctx context.Context, jobs chan<- struct{}) {
	defer close(jobs)

	if e.cfg.TotalRequests > 0 {
		// фиксированные запросы
		for i := 0; i < e.cfg.TotalRequests; i++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- struct{}{}:
			}
		}
		return
	}

	// запросы по времени
	for {
		select {
		case <-ctx.Done():
			return
		case jobs <- struct{}{}:
		}
	}
}

func (e *Engine) LiveStats(ctx context.Context, interval time.Duration, fn func(int64)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn(e.collector.TotalRequests())
		}
	}
}