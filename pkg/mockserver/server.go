package mockserver

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Route struct {
	Path       string
	StatusCode int
	Response   any
	Delay time.Duration
	DelayJitter time.Duration
	ErrorRate float64
}

type Server struct {
	server *http.Server
	mux    *http.ServeMux
	logger *log.Logger
}

func New(addr string, logger *log.Logger) *Server {
	mux := http.NewServeMux()
	return &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		mux:    mux,
		logger: logger,
	}
}

func (s *Server) RegisterRoute(route Route) {
	s.mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
		delay := route.Delay
		if route.DelayJitter > 0 {
			delay += time.Duration(rand.Int63n(int64(route.DelayJitter)))
		}
		if delay > 0 {
			time.Sleep(delay)
		}

		if route.ErrorRate > 0 && rand.Float64() < route.ErrorRate {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		statusCode := route.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		w.WriteHeader(statusCode)

		if route.Response != nil {
			if err := json.NewEncoder(w).Encode(route.Response); err != nil {
				s.logger.Printf("encode response: %v", err)
			}
		}
	})

	s.logger.Printf("registered route: %s", route.Path)
}

func (s *Server) RegisterDefaults() {
	// fast
	s.RegisterRoute(Route{
		Path:       "/fast",
		StatusCode: http.StatusOK,
		Response:   map[string]string{"status": "ok", "endpoint": "fast"},
	})

	// slow
	s.RegisterRoute(Route{
		Path:       "/slow",
		StatusCode: http.StatusOK,
		Delay:      100 * time.Millisecond,
		DelayJitter: 50 * time.Millisecond,
		Response:   map[string]string{"status": "ok", "endpoint": "slow"},
	})

	// errs
	s.RegisterRoute(Route{
		Path:       "/flaky",
		StatusCode: http.StatusOK,
		ErrorRate:  0.3,
		Response:   map[string]string{"status": "ok", "endpoint": "flaky"},
	})

	// Health check
	s.RegisterRoute(Route{
		Path:       "/health",
		StatusCode: http.StatusOK,
		Response:   map[string]string{"status": "healthy"},
	})
}

func (s *Server) Start() error {
	s.logger.Printf("mock server starting on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) Addr() string {
	return s.server.Addr
}