package config

import (
	"errors"
	"time"
)

// Config содержит все настройки для нагрузочного теста
type Config struct {
	// Target — URL цели
	TargetURL string

	// Concurrency — количество параллельных воркеров
	Concurrency int

	// TotalRequests — общее количество запросов (0 = неограничено)
	TotalRequests int

	// Duration — длительность теста (0 = неограничено)
	Duration time.Duration

	// RequestTimeout — таймаут одного запроса
	RequestTimeout time.Duration

	// Method — HTTP метод
	Method string

	// Headers — заголовки запроса
	Headers map[string]string

	// Body — тело запроса
	Body []byte
}

func (c *Config) Validate() error {
	if c.TargetURL == "" {
		return errors.New("target URL is required")
	}
	if c.Concurrency <= 0 {
		return errors.New("concurrency must be greater than 0")
	}
	if c.TotalRequests == 0 && c.Duration == 0 {
		return errors.New("either total_requests or duration must be set")
	}
	if c.RequestTimeout == 0 {
		c.RequestTimeout = 30 * time.Second
	}
	if c.Method == "" {
		c.Method = "GET"
	}
	return nil
}

func Default() *Config {
	return &Config{
		Concurrency:    10,
		Duration:       30 * time.Second,
		RequestTimeout: 10 * time.Second,
		Method:         "GET",
		Headers:        make(map[string]string),
	}
}