package httpclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/basedalex/load-tester/internal/httpclient"
)

// хелпер для создания тестового сервера
func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestHTTPClient_Do_Success(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	client := httpclient.New(5 * time.Second)
	resp, err := client.Do(context.Background(), httpclient.Request{
		URL:    srv.URL,
		Method: http.MethodGet,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if resp.Latency <= 0 {
		t.Error("Latency should be greater than 0")
	}
}

func TestHTTPClient_Do_StatusCodes(t *testing.T) {
	tests := []struct {
		name           string
		serverStatus   int
		wantStatusCode int
	}{
		{
			name:           "200 OK",
			serverStatus:   http.StatusOK,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "404 Not Found",
			serverStatus:   http.StatusNotFound,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "500 Internal Server Error",
			serverStatus:   http.StatusInternalServerError,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "201 Created",
			serverStatus:   http.StatusCreated,
			wantStatusCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			})

			client := httpclient.New(5 * time.Second)
			resp, err := client.Do(context.Background(), httpclient.Request{
				URL:    srv.URL,
				Method: http.MethodGet,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("StatusCode = %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}
		})
	}
}

func TestHTTPClient_Do_Methods(t *testing.T) {
	tests := []struct {
		method string
	}{
		{http.MethodGet},
		{http.MethodPost},
		{http.MethodPut},
		{http.MethodDelete},
		{http.MethodPatch},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				// Проверяем что сервер получил правильный метод
				if r.Method != tt.method {
					t.Errorf("server got method %s, want %s", r.Method, tt.method)
				}
				w.WriteHeader(http.StatusOK)
			})

			client := httpclient.New(5 * time.Second)
			_, err := client.Do(context.Background(), httpclient.Request{
				URL:    srv.URL,
				Method: tt.method,
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestHTTPClient_Do_SendsHeaders(t *testing.T) {
	wantHeaders := map[string]string{
		"X-Custom-Header": "test-value",
		"Authorization":   "Bearer token123",
		"Content-Type":    "application/json",
	}

	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		for key, want := range wantHeaders {
			if got := r.Header.Get(key); got != want {
				t.Errorf("header %s = %q, want %q", key, got, want)
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	client := httpclient.New(5 * time.Second)
	_, err := client.Do(context.Background(), httpclient.Request{
		URL:     srv.URL,
		Method:  http.MethodGet,
		Headers: wantHeaders,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClient_Do_SendsBody(t *testing.T) {
	wantBody := `{"key":"value"}`

	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, len(wantBody))
		n, _ := r.Body.Read(buf)

		if got := string(buf[:n]); got != wantBody {
			t.Errorf("body = %q, want %q", got, wantBody)
		}
		w.WriteHeader(http.StatusOK)
	})

	client := httpclient.New(5 * time.Second)
	_, err := client.Do(context.Background(), httpclient.Request{
		URL:    srv.URL,
		Method: http.MethodPost,
		Body:   []byte(wantBody),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClient_Do_MeasuresLatency(t *testing.T) {
	delay := 50 * time.Millisecond

	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.WriteHeader(http.StatusOK)
	})

	client := httpclient.New(5 * time.Second)
	resp, err := client.Do(context.Background(), httpclient.Request{
		URL:    srv.URL,
		Method: http.MethodGet,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Latency < delay {
		t.Errorf("Latency = %v, want >= %v", resp.Latency, delay)
	}
}

func TestHTTPClient_Do_Timeout(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Сервер намеренно висит дольше таймаута клиента
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	// Таймаут клиента меньше задержки сервера
	client := httpclient.New(50 * time.Millisecond)
	_, err := client.Do(context.Background(), httpclient.Request{
		URL:    srv.URL,
		Method: http.MethodGet,
	})

	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestHTTPClient_Do_ContextCancel(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Отменяем контекст сразу после запуска
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	client := httpclient.New(5 * time.Second)
	_, err := client.Do(ctx, httpclient.Request{
		URL:    srv.URL,
		Method: http.MethodGet,
	})

	if err == nil {
		t.Error("expected context cancel error, got nil")
	}
}

func TestHTTPClient_Do_InvalidURL(t *testing.T) {
	client := httpclient.New(5 * time.Second)
	_, err := client.Do(context.Background(), httpclient.Request{
		URL:    "://invalid-url",
		Method: http.MethodGet,
	})

	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestHTTPClient_Do_UnreachableHost(t *testing.T) {
	client := httpclient.New(100 * time.Millisecond)
	_, err := client.Do(context.Background(), httpclient.Request{
		// TODO: вынести в конфиг?
		URL:    "http://localhost:19999",
		Method: http.MethodGet,
	})

	if err == nil {
		t.Error("expected connection error, got nil")
	}
}