package report

import (
	"bytes"
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
	html, err := GenerateHTMLReportContent(inputPath)
	if err != nil {
		return err
	}

	return writeHTMLReport(outputPath, html)
}

func GenerateHTMLReportContent(inputPath string) ([]byte, error) {
	raw, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}

	return renderHTMLReport(raw)
}

func renderHTMLReport(raw []byte) ([]byte, error) {

	var probe struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("decode report kind: %w", err)
	}

	switch probe.Kind {
	case "sweep":
		return generateSweepHTML(raw)
	default:
		return nil, fmt.Errorf("unsupported report kind: %q", probe.Kind)
	}
}

func generateSweepHTML(raw []byte) ([]byte, error) {
	var rep JSONSweepReport
	if err := json.Unmarshal(raw, &rep); err != nil {
		return nil, fmt.Errorf("decode sweep report: %w", err)
	}

	concurrency := make([]int, 0, len(rep.Runs))
	tokPerSec := make([]float64, 0, len(rep.Runs))
	avgLatency := make([]int64, 0, len(rep.Runs))
	latencyP95 := make([]int64, 0, len(rep.Runs))
	ttftP50 := make([]int64, 0, len(rep.Runs))

	for _, run := range rep.Runs {
		concurrency = append(concurrency, run.Concurrency)
		tokPerSec = append(tokPerSec, run.Summary.OutputTokensPerSec)
		avgLatency = append(avgLatency, run.Summary.AvgLatencyMS)
		latencyP95 = append(latencyP95, run.Summary.LatencyP95MS)
		ttftP50 = append(ttftP50, run.Summary.TTFTP50MS)
	}

	concurrencyJSON, err := json.Marshal(concurrency)
	if err != nil {
		return nil, fmt.Errorf("marshal concurrency: %w", err)
	}
	tokPerSecJSON, err := json.Marshal(tokPerSec)
	if err != nil {
		return nil, fmt.Errorf("marshal tokens per second: %w", err)
	}
	avgLatencyJSON, err := json.Marshal(avgLatency)
	if err != nil {
		return nil, fmt.Errorf("marshal average latency: %w", err)
	}
	latencyP95JSON, err := json.Marshal(latencyP95)
	if err != nil {
		return nil, fmt.Errorf("marshal latency p95: %w", err)
	}
	ttftP50JSON, err := json.Marshal(ttftP50)
	if err != nil {
		return nil, fmt.Errorf("marshal ttft p50: %w", err)
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
		return nil, fmt.Errorf("parse template: %w", err)
	}

	buf := bytes.Buffer{}
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}

	return buf.Bytes(), nil
}

func writeHTMLReport(outputPath string, html []byte) error {
	dir := filepath.Dir(outputPath)
	tmp, err := os.CreateTemp(dir, "llmbench-report-*.html")
	if err != nil {
		return fmt.Errorf("create temp output: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(html); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write temp output: %w", err)
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
