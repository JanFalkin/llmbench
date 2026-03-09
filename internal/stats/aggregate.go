// internal/stats/aggregate.go
package stats

import (
	"time"

	"github.com/JanFalkin/llmbench/internal/config"
)

type RequestResult struct {
	RequestID string
	Model     string

	StartTime    time.Time
	EndTime      time.Time
	FirstByteAt  time.Time
	FirstTokenAt time.Time
	LastTokenAt  time.Time
	UsageBlockAt time.Time

	InputTokens         int
	OutputTokens        int
	HTTPStatus          int
	Error               string
	TTFT                time.Duration
	Decode              time.Duration
	EndToEnd            time.Duration
	InterTokenLatencies []time.Duration
}

type BenchmarkReport struct {
	Config             config.BenchmarkConfig
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	Elapsed            time.Duration
	RequestsPerSecond  float64
	OutputTokensPerSec float64
	AvgLatency         time.Duration
	LatencyP50         time.Duration
	LatencyP95         time.Duration

	TTFTP50 time.Duration
	TTFTP95 time.Duration
	Results []RequestResult
}

func Aggregate(cfg config.BenchmarkConfig, results []RequestResult, elapsed time.Duration) BenchmarkReport {
	var okCount int
	var outputTokens int
	var totalLatency time.Duration
	var latencies []time.Duration
	var ttfts []time.Duration

	for _, r := range results {
		if r.Error == "" {
			okCount++
			totalLatency += r.EndToEnd
			latencies = append(latencies, r.EndToEnd)
			ttfts = append(ttfts, r.TTFT)
		}
		outputTokens += r.OutputTokens
	}

	reqPerSec := 0.0
	tokPerSec := 0.0
	if elapsed > 0 && okCount > 0 {
		reqPerSec = float64(okCount) / elapsed.Seconds()
		tokPerSec = float64(outputTokens) / elapsed.Seconds()
	}

	var avgLatency time.Duration
	if okCount > 0 {
		avgLatency = totalLatency / time.Duration(okCount)
	}

	return BenchmarkReport{
		Config:             cfg,
		TotalRequests:      len(results),
		SuccessfulRequests: okCount,
		FailedRequests:     len(results) - okCount,
		Elapsed:            elapsed,
		RequestsPerSecond:  reqPerSec,
		OutputTokensPerSec: tokPerSec,
		AvgLatency:         avgLatency,
		LatencyP50:         Percentile(latencies, 0.5),
		LatencyP95:         Percentile(latencies, 0.95),
		TTFTP50:            Percentile(ttfts, 0.5),
		TTFTP95:            Percentile(ttfts, 0.95),
		Results:            results,
	}
}
