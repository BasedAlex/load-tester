package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type Result struct {
	StatusCode int
	Latency    time.Duration
	Error      error
	Timestamp  time.Time
}

type Summary struct {
	TotalRequests   int64
	SuccessRequests int64
	FailedRequests  int64

	MinLatency time.Duration
	MaxLatency time.Duration
	AvgLatency time.Duration

	// Перцентили
	P50 time.Duration
	P90 time.Duration
	P95 time.Duration
	P99 time.Duration

	RequestsPerSecond float64
	TotalDuration     time.Duration

	StatusCodes map[int]int64
	Errors      map[string]int64
}

type Collector struct {
	mu      sync.Mutex
	results []Result

	// атомарные метрики
	totalRequests   atomic.Int64
	successRequests atomic.Int64
	failedRequests  atomic.Int64

	startTime time.Time
}

func NewCollector() *Collector {
	return &Collector{
		results:   make([]Result, 0, 1000),
		startTime: time.Now(),
	}
}

func (c *Collector) Add(r Result) {
	c.totalRequests.Add(1)

	if r.Error != nil || r.StatusCode >= 500 {
		c.failedRequests.Add(1)
	} else {
		c.successRequests.Add(1)
	}

	c.mu.Lock()
	c.results = append(c.results, r)
	c.mu.Unlock()
}

func (c *Collector) TotalRequests() int64 {
	return c.totalRequests.Load()
}

func (c *Collector) Summarize() Summary {
	c.mu.Lock()
	results := make([]Result, len(c.results))
	copy(results, c.results)
	c.mu.Unlock()

	if len(results) == 0 {
		return Summary{}
	}

	summary := Summary{
		TotalRequests:   c.totalRequests.Load(),
		SuccessRequests: c.successRequests.Load(),
		FailedRequests:  c.failedRequests.Load(),
		StatusCodes:     make(map[int]int64),
		Errors:          make(map[string]int64),
		MinLatency:      results[0].Latency,
	}

	var totalLatency time.Duration

	latencies := make([]time.Duration, 0, len(results))

	for _, r := range results {
		if r.StatusCode > 0 {
			summary.StatusCodes[r.StatusCode]++
		}

		if r.Error != nil {
			summary.Errors[r.Error.Error()]++
			continue
		}

		latencies = append(latencies, r.Latency)
		totalLatency += r.Latency

		if r.Latency < summary.MinLatency {
			summary.MinLatency = r.Latency
		}
		if r.Latency > summary.MaxLatency {
			summary.MaxLatency = r.Latency
		}
	}

	if len(latencies) > 0 {
		summary.AvgLatency = totalLatency / time.Duration(len(latencies))

		sorted := sortDurations(latencies)
		summary.P50 = percentile(sorted, 50)
		summary.P90 = percentile(sorted, 90)
		summary.P95 = percentile(sorted, 95)
		summary.P99 = percentile(sorted, 99)
	}

	summary.TotalDuration = time.Since(c.startTime)
	if summary.TotalDuration > 0 {
		summary.RequestsPerSecond = float64(summary.TotalRequests) /
			summary.TotalDuration.Seconds()
	}

	return summary
}

func sortDurations(d []time.Duration) []time.Duration {
	sorted := make([]time.Duration, len(d))
	copy(sorted, d)

	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}
	return sorted
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)) * p / 100)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}