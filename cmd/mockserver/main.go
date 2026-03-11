package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"context"
	"time"

	"github.com/basedalex/load-tester/pkg/mockserver"
)

func main() {
	logger := log.New(os.Stdout, "[mock-server] ", log.LstdFlags)

	srv := mockserver.New(":8080", logger)
	srv.RegisterDefaults()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			logger.Printf("server stopped: %v", err)
		}
	}()

	<-quit
	logger.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("shutdown error: %v", err)
	}
}