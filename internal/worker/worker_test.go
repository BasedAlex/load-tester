package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/basedalex/load-tester/internal/httpclient"
	"github.com/basedalex/load-tester/internal/metrics"
	"github.com/basedalex/load-tester/internal/worker"
)

type mockClient struct {
	response httpclient.Response
	err      error
	calls    int
}

func (m *mockClient) Do(_ context.Context, _ httpclient.Request) (httpclient.Response, error) {
	m.calls++
	return m.response, m.err
}

func TestWorker_Run_ProcessesJobs(t *testing.T) {
	client := &mockClient{
		response: httpclient.Response{StatusCode: 200, Latency: 5 * time.Millisecond},
	}
	collector := metrics.NewCollector()

	w := worker.New(1, client, httpclient.Request{}, collector)

	jobs := make(chan struct{}, 3)
	jobs <- struct{}{}
	jobs <- struct{}{}
	jobs <- struct{}{}
	close(jobs)

	w.Run(context.Background(), jobs)

	if client.calls != 3 {
		t.Errorf("client.calls = %d, want 3", client.calls)
	}
	if collector.TotalRequests() != 3 {
		t.Errorf("TotalRequests = %d, want 3", collector.TotalRequests())
	}
}

func TestWorker_Run_StopsOnContextCancel(t *testing.T) {
	client := &mockClient{
		response: httpclient.Response{StatusCode: 200},
	}
	collector := metrics.NewCollector()
	w := worker.New(1, client, httpclient.Request{}, collector)

	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan struct{}, 100)

	done := make(chan struct{})
	go func() {
		w.Run(ctx, jobs)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		t.Error("worker did not stop after context cancel")
	}
}

func TestWorker_Run_HandlesErrors(t *testing.T) {
	client := &mockClient{
		err: errors.New("connection refused"),
	}
	collector := metrics.NewCollector()
	w := worker.New(1, client, httpclient.Request{}, collector)

	jobs := make(chan struct{}, 1)
	jobs <- struct{}{}
	close(jobs)

	w.Run(context.Background(), jobs)

	s := collector.Summarize()
	if s.FailedRequests != 1 {
		t.Errorf("FailedRequests = %d, want 1", s.FailedRequests)
	}
}