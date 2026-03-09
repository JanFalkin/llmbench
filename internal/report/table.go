package report

import (
	"fmt"
	"strings"

	"github.com/JanFalkin/llmbench/internal/stats"
)

func RenderTable(rep stats.BenchmarkReport) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Requests:           %d\n", rep.TotalRequests)
	fmt.Fprintf(&b, "Successful:         %d\n", rep.SuccessfulRequests)
	fmt.Fprintf(&b, "Failed:             %d\n", rep.FailedRequests)
	fmt.Fprintf(&b, "Elapsed:            %s\n", rep.Elapsed)
	fmt.Fprintf(&b, "Requests/sec:       %.2f\n", rep.RequestsPerSecond)
	fmt.Fprintf(&b, "Output tokens/sec:  %.2f\n", rep.OutputTokensPerSec)
	fmt.Fprintf(&b, "URL:                %s\n", rep.Config.URL)
	fmt.Fprintf(&b, "Model:              %s\n", rep.Config.Model)

	if rep.FailedRequests > 0 {
		fmt.Fprintf(&b, "\nErrors:\n")
		shown := 0
		for _, r := range rep.Results {
			if r.Error == "" {
				continue
			}
			fmt.Fprintf(&b, "  %s: %s\n", r.RequestID, r.Error)
			shown++
			if shown >= 5 {
				break
			}
		}
	}

	return b.String()
}
