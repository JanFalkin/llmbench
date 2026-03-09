package report

import (
	"fmt"
	"strings"

	"github.com/JanFalkin/llmbench/internal/stats"
)

func RenderSweepTable(reports []stats.BenchmarkReport) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Concurrency   Success   Req/sec   Tok/sec   Lat p50       Lat p95\n")
	fmt.Fprintf(&b, "-----------------------------------------------------------------\n")

	for _, rep := range reports {
		fmt.Fprintf(
			&b,
			"%-12d  %d/%d     %-8.2f  %-8.2f  %-8s %-8s\n",
			rep.Config.Concurrency,
			rep.SuccessfulRequests,
			rep.TotalRequests,
			rep.RequestsPerSecond,
			rep.OutputTokensPerSec,
			rep.LatencyP50,
			rep.LatencyP95,
		)
	}

	return b.String()
}
