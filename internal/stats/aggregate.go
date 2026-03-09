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
	Results            []RequestResult
}

func Aggregate(cfg config.BenchmarkConfig, results []RequestResult, elapsed time.Duration) BenchmarkReport {
	var okCount int
	var outputTokens int

	for _, r := range results {
		if r.Error == "" {
			okCount++
		}
		outputTokens += r.OutputTokens
	}

	reqPerSec := 0.0
	tokPerSec := 0.0
	if elapsed > 0 {
		reqPerSec = float64(len(results)) / elapsed.Seconds()
		tokPerSec = float64(outputTokens) / elapsed.Seconds()
	}

	return BenchmarkReport{
		Config:             cfg,
		TotalRequests:      len(results),
		SuccessfulRequests: okCount,
		FailedRequests:     len(results) - okCount,
		Elapsed:            elapsed,
		RequestsPerSecond:  reqPerSec,
		OutputTokensPerSec: tokPerSec,
		Results:            results,
	}
}
