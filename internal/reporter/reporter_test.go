package reporter_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/basedalex/load-tester/internal/metrics"
	"github.com/basedalex/load-tester/internal/reporter"
)

// buildSummary — хелпер для создания тестовой статистики
func buildSummary() metrics.Summary {
	return metrics.Summary{
		TotalRequests:     1000,
		SuccessRequests:   970,
		FailedRequests:    30,
		TotalDuration:     5 * time.Second,
		RequestsPerSecond: 200.0,
		MinLatency:        1 * time.Millisecond,
		AvgLatency:        48 * time.Millisecond,
		MaxLatency:        312 * time.Millisecond,
		P50:               45 * time.Millisecond,
		P90:               98 * time.Millisecond,
		P95:               145 * time.Millisecond,
		P99:               287 * time.Millisecond,
		StatusCodes: map[int]int64{
			200: 970,
			500: 30,
		},
		Errors: map[string]int64{},
	}
}

func TestReporter_Print_ContainsRequests(t *testing.T) {
	var buf bytes.Buffer
	rep := reporter.New(&buf)

	rep.Print(buildSummary())
	output := buf.String()

	checks := []struct {
		name    string
		contain string
	}{
		{"total requests", "1000"},
		{"success requests", "970"},
		{"failed requests", "30"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(output, c.contain) {
				t.Errorf("output does not contain %q\noutput:\n%s", c.contain, output)
			}
		})
	}
}

func TestReporter_Print_ContainsLatency(t *testing.T) {
	var buf bytes.Buffer
	rep := reporter.New(&buf)

	rep.Print(buildSummary())
	output := buf.String()

	// Проверяем наличие секции и ключевых меток перцентилей
	sections := []string{"LATENCY", "P50", "P90", "P95", "P99"}

	for _, section := range sections {
		t.Run(section, func(t *testing.T) {
			if !strings.Contains(output, section) {
				t.Errorf("output missing section %q\noutput:\n%s", section, output)
			}
		})
	}
}

func TestReporter_Print_ContainsStatusCodes(t *testing.T) {
	var buf bytes.Buffer
	rep := reporter.New(&buf)

	rep.Print(buildSummary())
	output := buf.String()

	if !strings.Contains(output, "STATUS CODES") {
		t.Errorf("output missing STATUS CODES section\noutput:\n%s", output)
	}
	if !strings.Contains(output, "200") {
		t.Errorf("output missing status code 200\noutput:\n%s", output)
	}
	if !strings.Contains(output, "500") {
		t.Errorf("output missing status code 500\noutput:\n%s", output)
	}
}

func TestReporter_Print_NoErrorsSection_WhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	rep := reporter.New(&buf)

	summary := buildSummary()
	summary.Errors = map[string]int64{} // пустые ошибки

	rep.Print(summary)
	output := buf.String()

	// Секция ERRORS не должна появляться если ошибок нет
	if strings.Contains(output, "ERRORS") {
		t.Errorf("output should not contain ERRORS section when errors map is empty\noutput:\n%s", output)
	}
}

func TestReporter_Print_ShowsErrorsSection_WhenPresent(t *testing.T) {
	var buf bytes.Buffer
	rep := reporter.New(&buf)

	summary := buildSummary()
	summary.Errors = map[string]int64{
		"connection refused": 10,
		"context deadline":   5,
	}

	rep.Print(summary)
	output := buf.String()

	checks := []string{"ERRORS", "connection refused", "context deadline"}
	for _, c := range checks {
		t.Run(c, func(t *testing.T) {
			if !strings.Contains(output, c) {
				t.Errorf("output missing %q\noutput:\n%s", c, output)
			}
		})
	}
}

func TestReporter_Print_ReqPerSecond(t *testing.T) {
	var buf bytes.Buffer
	rep := reporter.New(&buf)

	rep.Print(buildSummary())
	output := buf.String()

	// 200.00 req/sec
	if !strings.Contains(output, "200.00") {
		t.Errorf("output missing req/sec value\noutput:\n%s", output)
	}
}

func TestReporter_Print_WritesToCustomWriter(t *testing.T) {
	// Проверяем что reporter пишет именно в переданный writer
	var buf1, buf2 bytes.Buffer

	rep1 := reporter.New(&buf1)
	rep2 := reporter.New(&buf2)

	summary := buildSummary()
	rep1.Print(summary)
	rep2.Print(summary)

	// Оба вывода должны быть идентичны
	if buf1.String() != buf2.String() {
		t.Error("outputs from two reporters with same summary should be identical")
	}

	// И оба не пустые
	if buf1.Len() == 0 {
		t.Error("reporter output should not be empty")
	}
}

func TestReporter_Print_EmptySummary(t *testing.T) {
	var buf bytes.Buffer
	rep := reporter.New(&buf)

	// Не должен паниковать на пустой статистике
	rep.Print(metrics.Summary{
		StatusCodes: map[int]int64{},
		Errors:      map[string]int64{},
	})

	if buf.Len() == 0 {
		t.Error("reporter should write something even for empty summary")
	}
}