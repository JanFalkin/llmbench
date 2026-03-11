// internal/config/config.go
package config

import "time"

type BenchmarkConfig struct {
	URL              string
	Model            string
	APIKey           string
	PromptTokens     int
	CompletionTokens int
	Concurrency      int
	Requests         int
	WarmupRequests   int
	Stream           bool
	Timeout          time.Duration
	// Optional user-supplied run label (included in CSV output only; not included in JSON reports).
	Label string
}
