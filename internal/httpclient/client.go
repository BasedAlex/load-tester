package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

type Request struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    []byte
	Timeout time.Duration
}

type Response struct {
	StatusCode int
	Latency    time.Duration
}

type Client interface {
	Do(ctx context.Context, req Request) (Response, error)
}

type HTTPClient struct {
	client *http.Client
}

func New(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (c *HTTPClient) Do(ctx context.Context, req Request) (Response, error) {
	var body *bytes.Reader
	if len(req.Body) > 0 {
		body = bytes.NewReader(req.Body)
	} else {
		body = bytes.NewReader(nil)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, body)
	if err != nil {
		return Response{}, fmt.Errorf("create request: %w", err)
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := c.client.Do(httpReq)
	latency := time.Since(start)

	if err != nil {
		return Response{Latency: latency}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	return Response{
		StatusCode: resp.StatusCode,
		Latency:    latency,
	}, nil
}