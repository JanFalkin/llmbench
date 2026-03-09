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
	"github.com/JanFalkin/llmbench/internal/stats"
	"github.com/JanFalkin/llmbench/internal/workload"
)

func runSweep(args []string) {
	fs := flag.NewFlagSet("sweep", flag.ExitOnError)

	var cfg config.BenchmarkConfig
	var concurrencyList string

	fs.StringVar(&cfg.URL, "url", "http://localhost:11434", "Base URL of OpenAI-compatible endpoint")
	fs.StringVar(&cfg.Model, "model", "", "Model name")
	fs.IntVar(&cfg.PromptTokens, "prompt-tokens", 512, "Approximate prompt token count")
	fs.IntVar(&cfg.CompletionTokens, "completion-tokens", 128, "Max completion tokens")
	fs.IntVar(&cfg.Requests, "requests", 1, "Number of measured requests per sweep step")
	fs.IntVar(&cfg.WarmupRequests, "warmup-requests", 0, "Number of warmup requests per sweep step")
	fs.BoolVar(&cfg.Stream, "stream", true, "Use streaming responses")
	fs.DurationVar(&cfg.Timeout, "timeout", 60*time.Second, "HTTP timeout")
	fs.StringVar(&concurrencyList, "concurrency", "1,2,4", "Comma-separated concurrency levels")

	_ = fs.Parse(args)

	if cfg.Model == "" {
		fmt.Fprintln(os.Stderr, "error: --model is required")
		os.Exit(1)
	}

	levels, err := config.ParseConcurrencyList(concurrencyList)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	var reports []stats.BenchmarkReport

	for _, level := range levels {
		stepCfg := cfg
		stepCfg.Concurrency = level

		if stepCfg.Requests < stepCfg.Concurrency {
			fmt.Fprintf(
				os.Stderr,
				"warning: requests (%d) < concurrency (%d); results may underutilize workers\n",
				stepCfg.Requests,
				stepCfg.Concurrency,
			)
		}

		cl := client.New(stepCfg.URL, stepCfg.APIKey, stepCfg.Timeout, nil)
		gen := workload.NewGenerator()
		r := runner.New(stepCfg, cl, gen)

		rep, err := r.Run(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "sweep step failed at concurrency=%d: %v\n", level, err)
			continue
		}

		reports = append(reports, *rep)
	}

	if len(reports) == 0 {
		fmt.Fprintln(os.Stderr, "no successful sweep steps")
		os.Exit(1)
	}

	fmt.Print(report.RenderSweepTable(reports))
}
