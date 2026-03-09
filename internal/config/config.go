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
}

