package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JanFalkin/llmbench/internal/client"
	"github.com/JanFalkin/llmbench/internal/config"
	"github.com/JanFalkin/llmbench/internal/stats"
	"github.com/JanFalkin/llmbench/internal/workload"
)

type Runner struct {
	cfg    config.BenchmarkConfig
	client *client.Client
	gen    *workload.Generator
}

func New(cfg config.BenchmarkConfig, c *client.Client, gen *workload.Generator) *Runner {
	return &Runner{
		cfg:    cfg,
		client: c,
		gen:    gen,
	}
}

func (r *Runner) Run(ctx context.Context) (*stats.BenchmarkReport, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}

	if r.cfg.WarmupRequests > 0 {
		if err := r.runWarmup(ctx); err != nil {
			return nil, fmt.Errorf("warmup failed: %w", err)
		}
	}

	start := time.Now()

	jobs := make(chan workload.RequestSpec)
	results := make(chan stats.RequestResult, r.cfg.Requests)

	var wg sync.WaitGroup
	for i := 0; i < r.cfg.Concurrency; i++ {
		wg.Add(1)
		go r.worker(ctx, i, jobs, results, &wg)
	}

	go func() {
		defer close(jobs)
		for i := 0; i < r.cfg.Requests; i++ {
			req := r.gen.MakeRequest(
				r.cfg.Model,
				r.cfg.PromptTokens,
				r.cfg.CompletionTokens,
				r.cfg.Stream,
			)

			select {
			case <-ctx.Done():
				return
			case jobs <- req:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	allResults := make([]stats.RequestResult, 0, r.cfg.Requests)
	for res := range results {
		allResults = append(allResults, res)
	}

	elapsed := time.Since(start)
	report := stats.Aggregate(r.cfg, allResults, elapsed)
	return &report, nil
}

func (r *Runner) runWarmup(ctx context.Context) error {
	jobs := make(chan workload.RequestSpec)
	results := make(chan stats.RequestResult, r.cfg.WarmupRequests)

	var wg sync.WaitGroup
	for i := 0; i < r.cfg.Concurrency; i++ {
		wg.Add(1)
		go r.worker(ctx, i, jobs, results, &wg)
	}

	go func() {
		defer close(jobs)
		for i := 0; i < r.cfg.WarmupRequests; i++ {
			req := r.gen.MakeRequest(
				r.cfg.Model,
				r.cfg.PromptTokens,
				r.cfg.CompletionTokens,
				r.cfg.Stream,
			)

			select {
			case <-ctx.Done():
				return
			case jobs <- req:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for res := range results {
		if res.Error != "" {
			return fmt.Errorf("warmup request %s failed: %s", res.RequestID, res.Error)
		}
	}

	return nil
}

func (r *Runner) worker(
	ctx context.Context,
	workerID int,
	jobs <-chan workload.RequestSpec,
	results chan<- stats.RequestResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-jobs:
			if !ok {
				return
			}

			start := time.Now()
			resp, err := r.client.Do(ctx, req)
			end := time.Now()

			result := stats.RequestResult{
				RequestID:    req.ID,
				Model:        req.Model,
				StartTime:    start,
				EndTime:      end,
				EndToEnd:     end.Sub(start),
				InputTokens:  req.PromptTokens,
				OutputTokens: 0,
			}

			if err != nil {
				result.Error = err.Error()
			} else if resp != nil {
				result.HTTPStatus = resp.HTTPStatus
				result.FirstByteAt = resp.FirstByteAt
				result.FirstTokenAt = resp.FirstTokenAt
				result.LastTokenAt = resp.LastTokenAt
				result.UsageBlockAt = resp.UsageBlockAt
				result.OutputTokens = resp.OutputTokens
				result.InterTokenLatencies = resp.InterTokenLatencies

				if !resp.FirstTokenAt.IsZero() {
					result.TTFT = resp.FirstTokenAt.Sub(start)
				}
				if !resp.FirstTokenAt.IsZero() && !resp.LastTokenAt.IsZero() {
					result.Decode = resp.LastTokenAt.Sub(resp.FirstTokenAt)
				}
			}

			select {
			case <-ctx.Done():
				return
			case results <- result:
			}
		}
	}
}

func (r *Runner) validate() error {
	if r.cfg.URL == "" {
		return fmt.Errorf("url is required")
	}
	if r.cfg.Model == "" {
		return fmt.Errorf("model is required")
	}
	if r.cfg.Requests <= 0 {
		return fmt.Errorf("requests must be > 0")
	}
	if r.cfg.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be > 0")
	}
	if r.cfg.PromptTokens <= 0 {
		return fmt.Errorf("prompt-tokens must be > 0")
	}
	if r.cfg.CompletionTokens <= 0 {
		return fmt.Errorf("completion-tokens must be > 0")
	}
	if r.client == nil {
		return fmt.Errorf("client is nil")
	}
	if r.gen == nil {
		return fmt.Errorf("workload generator is nil")
	}
	return nil
}

