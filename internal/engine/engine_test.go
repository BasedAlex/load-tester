package engine_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/basedalex/load-tester/internal/config"
	"github.com/basedalex/load-tester/internal/engine"
	"github.com/basedalex/load-tester/internal/httpclient"
)

func TestEngine_Run_FixedRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := &config.Config{
		TargetURL:      srv.URL,
		Concurrency:    5,
		TotalRequests:  50,
		RequestTimeout: 5 * time.Second,
		Method:         "GET",
		Headers:        map[string]string{},
	}

	client := httpclient.New(cfg.RequestTimeout)
	eng := engine.New(cfg, client)

	summary, err := eng.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if summary.TotalRequests != 50 {
		t.Errorf("TotalRequests = %d, want 50", summary.TotalRequests)
	}
	if summary.SuccessRequests != 50 {
		t.Errorf("SuccessRequests = %d, want 50", summary.SuccessRequests)
	}
	if summary.FailedRequests != 0 {
		t.Errorf("FailedRequests = %d, want 0", summary.FailedRequests)
	}
}

func TestEngine_Run_Duration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := &config.Config{
		TargetURL:      srv.URL,
		Concurrency:    3,
		Duration:       200 * time.Millisecond,
		RequestTimeout: 5 * time.Second,
		Method:         "GET",
		Headers:        map[string]string{},
	}

	client := httpclient.New(cfg.RequestTimeout)
	eng := engine.New(cfg, client)

	summary, err := eng.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// За 200ms должны выполниться хоть какие-то запросы
	if summary.TotalRequests == 0 {
		t.Error("expected at least one request")
	}
	if summary.TotalDuration < 200*time.Millisecond {
		t.Errorf("duration too short: %v", summary.TotalDuration)
	}
}

func TestEngine_Run_InvalidConfig(t *testing.T) {
	cfg := &config.Config{}
	client := httpclient.New(time.Second)
	eng := engine.New(cfg, client)

	_, err := eng.Run(context.Background())
	if err == nil {
		t.Error("expected validation error")
	}
}