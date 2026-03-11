package metrics_test

import (
	"errors"
	"testing"
	"time"

	"github.com/basedalex/load-tester/internal/metrics"
)

func TestCollector_Add_CountsRequests(t *testing.T) {
	c := metrics.NewCollector()

	c.Add(metrics.Result{StatusCode: 200, Latency: 10 * time.Millisecond})
	c.Add(metrics.Result{StatusCode: 200, Latency: 20 * time.Millisecond})
	c.Add(metrics.Result{StatusCode: 500, Latency: 5 * time.Millisecond})
	c.Add(metrics.Result{Error: errors.New("timeout")})

	if got := c.TotalRequests(); got != 4 {
		t.Errorf("TotalRequests = %d, want 4", got)
	}
}

func TestCollector_Summarize_Latency(t *testing.T) {
	c := metrics.NewCollector()

	latencies := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
	}

	for _, l := range latencies {
		c.Add(metrics.Result{StatusCode: 200, Latency: l})
	}

	s := c.Summarize()

	if s.MinLatency != 10*time.Millisecond {
		t.Errorf("MinLatency = %v, want 10ms", s.MinLatency)
	}
	if s.MaxLatency != 30*time.Millisecond {
		t.Errorf("MaxLatency = %v, want 30ms", s.MaxLatency)
	}
	if s.AvgLatency != 20*time.Millisecond {
		t.Errorf("AvgLatency = %v, want 20ms", s.AvgLatency)
	}
}

func TestCollector_Summarize_StatusCodes(t *testing.T) {
	c := metrics.NewCollector()

	c.Add(metrics.Result{StatusCode: 200})
	c.Add(metrics.Result{StatusCode: 200})
	c.Add(metrics.Result{StatusCode: 404})
	c.Add(metrics.Result{StatusCode: 500})

	s := c.Summarize()

	if s.StatusCodes[200] != 2 {
		t.Errorf("StatusCodes[200] = %d, want 2", s.StatusCodes[200])
	}
	if s.StatusCodes[404] != 1 {
		t.Errorf("StatusCodes[404] = %d, want 1", s.StatusCodes[404])
	}
}

func TestCollector_Summarize_Empty(t *testing.T) {
	c := metrics.NewCollector()
	s := c.Summarize()

	if s.TotalRequests != 0 {
		t.Errorf("expected empty summary")
	}
}

func TestCollector_ConcurrentAdd(t *testing.T) {
	c := metrics.NewCollector()
	done := make(chan struct{})

	// 100 горутин добавляют результаты одновременно
	for i := 0; i < 100; i++ {
		go func() {
			c.Add(metrics.Result{StatusCode: 200, Latency: time.Millisecond})
			done <- struct{}{}
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	if got := c.TotalRequests(); got != 100 {
		t.Errorf("TotalRequests = %d, want 100", got)
	}
}