package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/basedalex/load-tester/internal/metrics"
)

type Reporter struct {
	out io.Writer
}

func New(out io.Writer) *Reporter {
	return &Reporter{out: out}
}

func (r *Reporter) Print(s metrics.Summary) {
	fmt.Fprintln(r.out, strings.Repeat("─", 50))
	fmt.Fprintln(r.out, "           LOAD TEST RESULTS")
	fmt.Fprintln(r.out, strings.Repeat("─", 50))

	fmt.Fprintf(r.out, "  Total Requests:    %d\n", s.TotalRequests)
	fmt.Fprintf(r.out, "  Success:           %d\n", s.SuccessRequests)
	fmt.Fprintf(r.out, "  Failed:            %d\n", s.FailedRequests)
	fmt.Fprintf(r.out, "  Duration:          %s\n", s.TotalDuration.Round(1*1000*1000))
	fmt.Fprintf(r.out, "  Req/sec:           %.2f\n", s.RequestsPerSecond)

	fmt.Fprintln(r.out, strings.Repeat("─", 50))
	fmt.Fprintln(r.out, "  LATENCY")
	fmt.Fprintf(r.out, "  Min:               %s\n", s.MinLatency)
	fmt.Fprintf(r.out, "  Avg:               %s\n", s.AvgLatency)
	fmt.Fprintf(r.out, "  Max:               %s\n", s.MaxLatency)
	fmt.Fprintf(r.out, "  P50:               %s\n", s.P50)
	fmt.Fprintf(r.out, "  P90:               %s\n", s.P90)
	fmt.Fprintf(r.out, "  P95:               %s\n", s.P95)
	fmt.Fprintf(r.out, "  P99:               %s\n", s.P99)

	if len(s.StatusCodes) > 0 {
		fmt.Fprintln(r.out, strings.Repeat("─", 50))
		fmt.Fprintln(r.out, "  STATUS CODES")
		for code, count := range s.StatusCodes {
			fmt.Fprintf(r.out, "  [%d]:               %d\n", code, count)
		}
	}

	if len(s.Errors) > 0 {
		fmt.Fprintln(r.out, strings.Repeat("─", 50))
		fmt.Fprintln(r.out, "  ERRORS")
		for err, count := range s.Errors {
			fmt.Fprintf(r.out, "  %s: %d\n", err, count)
		}
	}

	fmt.Fprintln(r.out, strings.Repeat("─", 50))
}