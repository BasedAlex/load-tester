package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/basedalex/load-tester/internal/config"
	"github.com/basedalex/load-tester/internal/engine"
	"github.com/basedalex/load-tester/internal/httpclient"
	"github.com/basedalex/load-tester/internal/reporter"
)

func main() {
	cfg := parseFlags()

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	client := httpclient.New(cfg.RequestTimeout)
	eng := engine.New(cfg, client)
	rep := reporter.New(os.Stdout)

	log.Printf("starting load test: url=%s concurrency=%d", cfg.TargetURL, cfg.Concurrency)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go eng.LiveStats(ctx, time.Second, func(total int64) {
		log.Printf("requests sent: %d", total)
	})

	summary, err := eng.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run error: %v\n", err)
		os.Exit(1)
	}

	rep.Print(summary)
}

func parseFlags() *config.Config {
	cfg := config.Default()

	flag.StringVar(&cfg.TargetURL, "url", "", "target URL (required)")
	flag.IntVar(&cfg.Concurrency, "c", cfg.Concurrency, "concurrency (workers)")
	flag.IntVar(&cfg.TotalRequests, "n", 0, "total requests (0 = use duration)")
	flag.DurationVar(&cfg.Duration, "d", cfg.Duration, "test duration")
	flag.DurationVar(&cfg.RequestTimeout, "timeout", cfg.RequestTimeout, "request timeout")
	flag.StringVar(&cfg.Method, "method", cfg.Method, "HTTP method")

	flag.Parse()
	return cfg
}