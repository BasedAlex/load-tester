package worker

import (
	"context"
	"time"

	"github.com/basedalex/load-tester/internal/httpclient"
	"github.com/basedalex/load-tester/internal/metrics"
)

type Worker struct {
	id        int
	client    httpclient.Client
	request   httpclient.Request
	collector *metrics.Collector
}

func New(
	id int,
	client httpclient.Client,
	req httpclient.Request,
	collector *metrics.Collector,
) *Worker {
	return &Worker{
		id:        id,
		client:    client,
		request:   req,
		collector: collector,
	}
}

// jobs — канал с токенами, один токен = один запрос
func (w *Worker) Run(ctx context.Context, jobs <-chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-jobs:
			if !ok {
				return
			}
			w.doRequest(ctx)
		}
	}
}

func (w *Worker) doRequest(ctx context.Context) {
	resp, err := w.client.Do(ctx, w.request)

	result := metrics.Result{
		StatusCode: resp.StatusCode,
		Latency:    resp.Latency,
		Error:      err,
		Timestamp:  time.Now(),
	}

	w.collector.Add(result)
}