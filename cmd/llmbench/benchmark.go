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
	fs.StringVar(&cfg.URL, "url", "http://localhost:8000", "Base URL of OpenAI-compatible endpoint")
	fs.StringVar(&cfg.Model, "model", "", "Model name")
	fs.IntVar(&cfg.PromptTokens, "prompt-tokens", 512, "Approximate prompt token count")
	fs.IntVar(&cfg.CompletionTokens, "completion-tokens", 128, "Max completion tokens")
	fs.IntVar(&cfg.Concurrency, "concurrency", 1, "Number of concurrent workers")
	fs.IntVar(&cfg.Requests, "requests", 1, "Number of measured requests")
	fs.IntVar(&cfg.WarmupRequests, "warmup-requests", 0, "Number of warmup requests")
	fs.BoolVar(&cfg.Stream, "stream", true, "Use streaming responses")
	fs.DurationVar(&cfg.Timeout, "timeout", 60*time.Second, "HTTP timeout")

	_ = fs.Parse(args)

	if cfg.Model == "" {
		fmt.Fprintln(os.Stderr, "error: --model is required")
		os.Exit(1)
	}

	cl := client.New(cfg.URL, "", cfg.Timeout, nil)
	gen := workload.NewGenerator()
	r := runner.New(cfg, cl, gen)

	rep, err := r.Run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "benchmark failed:", err)
		os.Exit(1)
	}

	fmt.Print(report.RenderTable(*rep))
}
