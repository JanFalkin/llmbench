package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/JanFalkin/llmbench/internal/client"
	"github.com/JanFalkin/llmbench/internal/config"
	"github.com/JanFalkin/llmbench/internal/report"
	"github.com/JanFalkin/llmbench/internal/runner"
	"github.com/JanFalkin/llmbench/internal/workload"
)

func runBenchmark(args []string) {
	fs := flag.NewFlagSet("benchmark", flag.ExitOnError)

	var cfg config.BenchmarkConfig
	var format string
	fs.StringVar(&format, "format", "table", "Output format: table or json")
	fs.StringVar(&cfg.URL, "url", "http://localhost:11434", "Base URL of OpenAI-compatible endpoint")
	fs.StringVar(&cfg.Model, "model", "", "Model name")
	fs.StringVar(&cfg.APIKey, "api-key", "", "API key (or set LLMBENCH_API_KEY or OPENAI_API_KEY env var)")
	fs.IntVar(&cfg.PromptTokens, "prompt-tokens", 512, "Approximate prompt token count")
	fs.IntVar(&cfg.CompletionTokens, "completion-tokens", 128, "Max completion tokens")
	fs.IntVar(&cfg.Concurrency, "concurrency", 1, "Number of concurrent workers")
	fs.IntVar(&cfg.Requests, "requests", 1, "Number of measured requests")
	fs.IntVar(&cfg.WarmupRequests, "warmup-requests", 0, "Number of warmup requests")
	fs.BoolVar(&cfg.Stream, "stream", true, "Use streaming responses")
	fs.DurationVar(&cfg.Timeout, "timeout", 60*time.Second, "HTTP timeout")

	_ = fs.Parse(args)

	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("LLMBENCH_API_KEY")
	}

	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("OPENAI_API_KEY")
	}

	if cfg.Model == "" {
		fmt.Fprintln(os.Stderr, "error: --model is required")
		os.Exit(1)
	}

	cl := client.New(cfg.URL, cfg.APIKey, cfg.Timeout, nil)
	gen := workload.NewGenerator()
	r := runner.New(cfg, cl, gen)

	rep, err := r.Run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "benchmark failed:", err)
		os.Exit(1)
	}
	switch format {
	case "json":
		data, err := report.RenderBenchmarkJSON(*rep)
		if err != nil {
			fmt.Fprintln(os.Stderr, "render json failed:", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	case "table":
		fmt.Print(report.RenderTable(*rep))
	case "csv":
		data, err := report.RenderBenchmarkCSV(*rep)
		if err != nil {
			fmt.Fprintln(os.Stderr, "render csv failed:", err)
			os.Exit(1)
		}
		fmt.Print(string(data))
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported format %q (expected \"table\" or \"json\")\n", format)
		os.Exit(1)
	}
}
