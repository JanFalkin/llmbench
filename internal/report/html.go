package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

type sweepHTMLData struct {
	Title            string
	Model            string
	URL              string
	PromptTokens     int
	CompletionTokens int
	Requests         int
	ConcurrencyJSON  template.JS
	TokensPerSecJSON template.JS
	AvgLatencyJSON   template.JS
	LatencyP95JSON   template.JS
	TTFTP50JSON      template.JS
}

func GenerateHTMLReport(inputPath, outputPath string) error {
	raw, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	var probe struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return fmt.Errorf("decode report kind: %w", err)
	}

	switch probe.Kind {
	case "sweep":
		return generateSweepHTML(raw, outputPath)
	default:
		return fmt.Errorf("unsupported report kind: %q", probe.Kind)
	}
}

func generateSweepHTML(raw []byte, outputPath string) error {
	var rep JSONSweepReport
	if err := json.Unmarshal(raw, &rep); err != nil {
		return fmt.Errorf("decode sweep report: %w", err)
	}

	var concurrency []int
	var tokPerSec []float64
	var avgLatency []int64
	var latencyP95 []int64
	var ttftP50 []int64

	for _, run := range rep.Runs {
		concurrency = append(concurrency, run.Concurrency)
		tokPerSec = append(tokPerSec, run.Summary.OutputTokensPerSec)
		avgLatency = append(avgLatency, run.Summary.AvgLatencyMS)
		latencyP95 = append(latencyP95, run.Summary.LatencyP95MS)
		ttftP50 = append(ttftP50, run.Summary.TTFTP50MS)
	}

	concurrencyJSON, err := json.Marshal(concurrency)
	if err != nil {
		return fmt.Errorf("marshal concurrency: %w", err)
	}
	tokPerSecJSON, err := json.Marshal(tokPerSec)
	if err != nil {
		return fmt.Errorf("marshal tokens per second: %w", err)
	}
	avgLatencyJSON, err := json.Marshal(avgLatency)
	if err != nil {
		return fmt.Errorf("marshal average latency: %w", err)
	}
	latencyP95JSON, err := json.Marshal(latencyP95)
	if err != nil {
		return fmt.Errorf("marshal latency p95: %w", err)
	}
	ttftP50JSON, err := json.Marshal(ttftP50)
	if err != nil {
		return fmt.Errorf("marshal ttft p50: %w", err)
	}

	var (
		model            string
		url              string
		promptTokens     int
		completionTokens int
		requests         int
	)

	if rep.BaseConfig != nil {
		model = rep.BaseConfig.Model
		url = rep.BaseConfig.URL
		promptTokens = rep.BaseConfig.PromptTokens
		completionTokens = rep.BaseConfig.CompletionTokens
		requests = rep.BaseConfig.Requests
	}

	data := sweepHTMLData{
		Title:            "llmbench Sweep Report",
		Model:            model,
		URL:              url,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		Requests:         requests,
		ConcurrencyJSON:  template.JS(concurrencyJSON),
		TokensPerSecJSON: template.JS(tokPerSecJSON),
		AvgLatencyJSON:   template.JS(avgLatencyJSON),
		LatencyP95JSON:   template.JS(latencyP95JSON),
		TTFTP50JSON:      template.JS(ttftP50JSON),
	}

	tmpl, err := template.New("report").Parse(sweepHTMLTemplate)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	dir := filepath.Dir(outputPath)
	tmp, err := os.CreateTemp(dir, "llmbench-report-*.html")
	if err != nil {
		return fmt.Errorf("create temp output: %w", err)
	}
	tmpName := tmp.Name()

	if err := tmpl.Execute(tmp, data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("render template: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close temp output: %w", err)
	}

	if err := os.Rename(tmpName, outputPath); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("finalize output: %w", err)
	}

	return nil
}
