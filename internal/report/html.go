package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
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

	concurrencyJSON, _ := json.Marshal(concurrency)
	tokPerSecJSON, _ := json.Marshal(tokPerSec)
	avgLatencyJSON, _ := json.Marshal(avgLatency)
	latencyP95JSON, _ := json.Marshal(latencyP95)
	ttftP50JSON, _ := json.Marshal(ttftP50)

	data := sweepHTMLData{
		Title:            "llmbench Sweep Report",
		Model:            rep.BaseConfig.Model,
		URL:              rep.BaseConfig.URL,
		PromptTokens:     rep.BaseConfig.PromptTokens,
		CompletionTokens: rep.BaseConfig.CompletionTokens,
		Requests:         rep.BaseConfig.Requests,
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

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	return nil
}
