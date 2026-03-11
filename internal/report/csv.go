package report

import (
	"bytes"
	"encoding/csv"
	"strconv"

	"github.com/JanFalkin/llmbench/internal/config"
	"github.com/JanFalkin/llmbench/internal/stats"
)

func RenderSweepCSV(reports []stats.BenchmarkReport) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := []string{
		"model",
		"url",
		"label",
		"concurrency",
		"total_requests",
		"successful_requests",
		"failed_requests",
		"elapsed_ms",
		"requests_per_second",
		"output_tokens_per_second",
		"avg_latency_ms",
		"latency_p50_ms",
		"latency_p95_ms",
		"ttft_p50_ms",
		"ttft_p95_ms",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	for _, rep := range reports {
		model, url, label := csvMetaFromConfig(rep.Config)

		row := []string{
			model,
			url,
			label,
			strconv.Itoa(rep.Config.Concurrency),
			strconv.Itoa(rep.TotalRequests),
			strconv.Itoa(rep.SuccessfulRequests),
			strconv.Itoa(rep.FailedRequests),
			strconv.FormatInt(rep.Elapsed.Milliseconds(), 10),
			strconv.FormatFloat(rep.RequestsPerSecond, 'f', 6, 64),
			strconv.FormatFloat(rep.OutputTokensPerSec, 'f', 6, 64),
			strconv.FormatInt(rep.AvgLatency.Milliseconds(), 10),
			strconv.FormatInt(rep.LatencyP50.Milliseconds(), 10),
			strconv.FormatInt(rep.LatencyP95.Milliseconds(), 10),
			strconv.FormatInt(rep.TTFTP50.Milliseconds(), 10),
			strconv.FormatInt(rep.TTFTP95.Milliseconds(), 10),
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}

func RenderBenchmarkCSV(rep stats.BenchmarkReport) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	header := []string{
		"model",
		"url",
		"label",
		"request_id",
		"success",
		"http_status",
		"input_tokens",
		"output_tokens",
		"end_to_end_ms",
		"ttft_ms",
		"decode_ms",
		"error",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	model, url, label := csvMetaFromConfig(rep.Config)

	for _, r := range rep.Results {
		success := "false"
		if r.Error == "" {
			success = "true"
		}

		row := []string{
			model,
			url,
			label,
			r.RequestID,
			success,
			strconv.Itoa(r.HTTPStatus),
			strconv.Itoa(r.InputTokens),
			strconv.Itoa(r.OutputTokens),
			strconv.FormatInt(r.EndToEnd.Milliseconds(), 10),
			strconv.FormatInt(r.TTFT.Milliseconds(), 10),
			strconv.FormatInt(r.Decode.Milliseconds(), 10),
			r.Error,
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}

func csvMetaFromConfig(cfg config.BenchmarkConfig) (model, url, label string) {
	return cfg.Model, cfg.URL, cfg.Label
}
