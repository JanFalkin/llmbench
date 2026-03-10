package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateHTMLReportContentSweep(t *testing.T) {
	t.Parallel()

	inputPath := filepath.Join(t.TempDir(), "sweep.json")

	doc := JSONSweepReport{
		Version:           "0.1.0",
		Kind:              "sweep",
		Timestamp:         time.Now().UTC(),
		BaseConfig:        &JSONConfig{URL: "http://localhost:11434", Model: "llama3", PromptTokens: 512, CompletionTokens: 16, Requests: 8},
		ConcurrencyLevels: []int{1, 2},
		Runs: []JSONSweepRun{
			{Concurrency: 1, Summary: JSONSummary{OutputTokensPerSec: 10.5, AvgLatencyMS: 1000, LatencyP95MS: 1400, TTFTP50MS: 500}},
			{Concurrency: 2, Summary: JSONSummary{OutputTokensPerSec: 18.2, AvgLatencyMS: 1700, LatencyP95MS: 2200, TTFTP50MS: 700}},
		},
	}

	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal report: %v", err)
	}

	if err := os.WriteFile(inputPath, raw, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	html, err := GenerateHTMLReportContent(inputPath)
	if err != nil {
		t.Fatalf("GenerateHTMLReportContent() error = %v", err)
	}

	out := string(html)
	if !strings.Contains(out, "<!DOCTYPE html>") {
		t.Fatalf("expected HTML document, got: %q", out)
	}
	if !strings.Contains(out, "llmbench Sweep Report") {
		t.Fatalf("expected report title in output")
	}
	if !strings.Contains(out, "llama3") {
		t.Fatalf("expected model name in output")
	}
	if !strings.Contains(out, "[1,2]") {
		t.Fatalf("expected concurrency data in output")
	}
}

func TestGenerateHTMLReportContentUnsupportedKind(t *testing.T) {
	t.Parallel()

	inputPath := filepath.Join(t.TempDir(), "benchmark.json")
	if err := os.WriteFile(inputPath, []byte(`{"kind":"benchmark"}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	_, err := GenerateHTMLReportContent(inputPath)
	if err == nil {
		t.Fatal("expected error for unsupported kind, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported report kind") {
		t.Fatalf("expected unsupported kind error, got: %v", err)
	}
}
func TestGenerateHTMLReportContentEmptySweep(t *testing.T) {
	t.Parallel()

	inputPath := filepath.Join(t.TempDir(), "sweep-empty.json")

	doc := JSONSweepReport{
		Version:           "0.1.0",
		Kind:              "sweep",
		Timestamp:         time.Now().UTC(),
		BaseConfig:        &JSONConfig{URL: "http://localhost:11434", Model: "llama3", PromptTokens: 512, CompletionTokens: 16, Requests: 8},
		ConcurrencyLevels: nil,
		Runs:              nil,
	}

	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal empty sweep report: %v", err)
	}

	if err := os.WriteFile(inputPath, raw, 0o644); err != nil {
		t.Fatalf("write empty sweep input: %v", err)
	}

	html, err := GenerateHTMLReportContent(inputPath)
	if err != nil {
		t.Fatalf("GenerateHTMLReportContent() error for empty sweep = %v", err)
	}

	out := string(html)
	if !strings.Contains(out, "<!DOCTYPE html>") {
		t.Fatalf("expected HTML document for empty sweep, got: %q", out)
	}
	if !strings.Contains(out, "llmbench Sweep Report") {
		t.Fatalf("expected report title in output for empty sweep")
	}
	if !strings.Contains(out, "llama3") {
		t.Fatalf("expected model name in output for empty sweep")
	}
	// Nil runs must produce empty JS arrays, not null, to avoid breaking Chart.js.
	if !strings.Contains(out, "const labels = []") {
		t.Fatalf("expected empty JS array for labels in output for empty sweep, got: %q", out)
	}
	if strings.Contains(out, "= null") {
		t.Fatalf("expected no null JS variable assignments in generated output for empty sweep, got: %q", out)
	}
}
