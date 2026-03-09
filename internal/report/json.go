package report

import (
	"encoding/json"
	"time"

	"github.com/JanFalkin/llmbench/internal/config"
	"github.com/JanFalkin/llmbench/internal/stats"
)

const schemaVersion = "0.1.0"

type JSONConfig struct {
	URL              string `json:"url"`
	Model            string `json:"model"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	Concurrency      int    `json:"concurrency,omitempty"`
	Requests         int    `json:"requests"`
	WarmupRequests   int    `json:"warmup_requests"`
	Stream           bool   `json:"stream"`
	TimeoutMS        int64  `json:"timeout_ms"`
}

type JSONSummary struct {
	TotalRequests      int `json:"total_requests"`
	SuccessfulRequests int `json:"successful_requests"`
	FailedRequests     int `json:"failed_requests"`

	ElapsedMS          int64   `json:"elapsed_ms"`
	RequestsPerSecond  float64 `json:"requests_per_second"`
	OutputTokensPerSec float64 `json:"output_tokens_per_second"`

	AvgLatencyMS int64 `json:"avg_latency_ms"`
	LatencyP50MS int64 `json:"latency_p50_ms"`
	LatencyP95MS int64 `json:"latency_p95_ms"`

	TTFTP50MS int64 `json:"ttft_p50_ms"`
	TTFTP95MS int64 `json:"ttft_p95_ms"`
}

type JSONRequestResult struct {
	RequestID string `json:"request_id"`
	Success   bool   `json:"success"`

	HTTPStatus int    `json:"http_status,omitempty"`
	Error      string `json:"error,omitempty"`

	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`

	EndToEndMS int64 `json:"end_to_end_ms,omitempty"`
	TTFTMS     int64 `json:"ttft_ms,omitempty"`
	DecodeMS   int64 `json:"decode_ms,omitempty"`

	InterTokenLatenciesMS []int64 `json:"inter_token_latencies_ms,omitempty"`
}

type JSONBenchmarkReport struct {
	Version   string              `json:"version"`
	Kind      string              `json:"kind"`
	Timestamp time.Time           `json:"timestamp"`
	Config    JSONConfig          `json:"config"`
	Summary   JSONSummary         `json:"summary"`
	Results   []JSONRequestResult `json:"results,omitempty"`
}

type JSONSweepRun struct {
	Concurrency int         `json:"concurrency"`
	Summary     JSONSummary `json:"summary"`
}

type JSONSweepReport struct {
	Version           string         `json:"version"`
	Kind              string         `json:"kind"`
	Timestamp         time.Time      `json:"timestamp"`
	BaseConfig        JSONConfig     `json:"base_config"`
	ConcurrencyLevels []int          `json:"concurrency_levels"`
	Runs              []JSONSweepRun `json:"runs"`
}

func RenderBenchmarkJSON(rep stats.BenchmarkReport) ([]byte, error) {
	doc := JSONBenchmarkReport{
		Version:   schemaVersion,
		Kind:      "benchmark",
		Timestamp: time.Now().UTC(),
		Config:    toJSONConfig(rep.Config),
		Summary:   toJSONSummary(rep),
		Results:   toJSONRequestResults(rep.Results),
	}

	return json.MarshalIndent(doc, "", "  ")
}

func RenderSweepJSON(reports []stats.BenchmarkReport) ([]byte, error) {
	if len(reports) == 0 {
		doc := JSONSweepReport{
			Version:           schemaVersion,
			Kind:              "sweep",
			Timestamp:         time.Now().UTC(),
			ConcurrencyLevels: []int{},
			Runs:              []JSONSweepRun{},
		}
		return json.MarshalIndent(doc, "", "  ")
	}

	baseCfg := reports[0].Config
	baseCfg.Concurrency = 0

	levels := make([]int, 0, len(reports))
	runs := make([]JSONSweepRun, 0, len(reports))

	for _, rep := range reports {
		levels = append(levels, rep.Config.Concurrency)
		runs = append(runs, JSONSweepRun{
			Concurrency: rep.Config.Concurrency,
			Summary:     toJSONSummary(rep),
		})
	}

	doc := JSONSweepReport{
		Version:           schemaVersion,
		Kind:              "sweep",
		Timestamp:         time.Now().UTC(),
		BaseConfig:        toJSONConfig(baseCfg),
		ConcurrencyLevels: levels,
		Runs:              runs,
	}

	return json.MarshalIndent(doc, "", "  ")
}

func toJSONConfig(cfg config.BenchmarkConfig) JSONConfig {
	return JSONConfig{
		URL:              cfg.URL,
		Model:            cfg.Model,
		PromptTokens:     cfg.PromptTokens,
		CompletionTokens: cfg.CompletionTokens,
		Concurrency:      cfg.Concurrency,
		Requests:         cfg.Requests,
		WarmupRequests:   cfg.WarmupRequests,
		Stream:           cfg.Stream,
		TimeoutMS:        cfg.Timeout.Milliseconds(),
	}
}

func toJSONSummary(rep stats.BenchmarkReport) JSONSummary {
	return JSONSummary{
		TotalRequests:      rep.TotalRequests,
		SuccessfulRequests: rep.SuccessfulRequests,
		FailedRequests:     rep.FailedRequests,

		ElapsedMS:          rep.Elapsed.Milliseconds(),
		RequestsPerSecond:  rep.RequestsPerSecond,
		OutputTokensPerSec: rep.OutputTokensPerSec,

		AvgLatencyMS: rep.AvgLatency.Milliseconds(),
		LatencyP50MS: rep.LatencyP50.Milliseconds(),
		LatencyP95MS: rep.LatencyP95.Milliseconds(),

		TTFTP50MS: rep.TTFTP50.Milliseconds(),
		TTFTP95MS: rep.TTFTP95.Milliseconds(),
	}
}

func toJSONRequestResults(results []stats.RequestResult) []JSONRequestResult {
	if len(results) == 0 {
		return nil
	}

	out := make([]JSONRequestResult, 0, len(results))
	for _, r := range results {
		out = append(out, JSONRequestResult{
			RequestID:  r.RequestID,
			Success:    r.Error == "",
			HTTPStatus: r.HTTPStatus,
			Error:      r.Error,

			InputTokens:  r.InputTokens,
			OutputTokens: r.OutputTokens,

			EndToEndMS: r.EndToEnd.Milliseconds(),
			TTFTMS:     r.TTFT.Milliseconds(),
			DecodeMS:   r.Decode.Milliseconds(),

			InterTokenLatenciesMS: durationsToMilliseconds(r.InterTokenLatencies),
		})
	}

	return out
}

func durationsToMilliseconds(in []time.Duration) []int64 {
	if len(in) == 0 {
		return nil
	}

	out := make([]int64, 0, len(in))
	for _, d := range in {
		out = append(out, d.Milliseconds())
	}
	return out
}
