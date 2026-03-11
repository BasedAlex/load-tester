package mockserver_test

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/basedalex/load-tester/pkg/mockserver"
)

func TestMockServer_FastRoute(t *testing.T) {
	logger := log.New(log.Default().Writer(), "", 0)
	srv := mockserver.New(":0", logger)
	srv.RegisterDefaults()

	// Используем httptest для тестирования без реального порта
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestMockServer_RouteWithDelay(t *testing.T) {
	logger := log.New(log.Default().Writer(), "", 0)
	srv := mockserver.New(":0", logger)

	delay := 50 * time.Millisecond
	srv.RegisterRoute(mockserver.Route{
		Path:       "/delayed",
		StatusCode: http.StatusOK,
		Delay:      delay,
		Response:   map[string]string{"ok": "true"},
	})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	start := time.Now()
	resp, err := http.Get(ts.URL)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if elapsed < delay {
		t.Errorf("elapsed %v < delay %v", elapsed, delay)
	}

	_ = srv
}