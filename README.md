# llmbench

Benchmark and analyze performance of OpenAI-compatible LLM endpoints.

Metrics:
- latency — end-to-end request latency
- time-to-first-token — delay before first token arrives
- token throughput — tokens generated per second
- concurrency scaling — throughput vs parallel requests
- streaming token timing — inter-token latency

Designed for systems like:
- vLLM
- llama.cpp
- Ollama
- OpenAI-compatible APIs

## Quick Start

Install:

```bash
go install github.com/JanFalkin/llmbench@latest
```

Run a benchmark:

```bash
llmbench benchmark \
  --url http://localhost:11434 \
  --model llama3 \
  --requests 5 \
  --concurrency 2 \
  --completion-tokens 16
```

For authenticated endpoints pass your API key via flag or environment variable:

```bash
# via flag
llmbench benchmark --api-key sk-... --url https://api.openai.com --model gpt-4o-mini

# via environment variable
export LLMBENCH_API_KEY=sk-...
llmbench benchmark --url https://api.openai.com --model gpt-4o-mini
```

Example output:

```text
Requests:           5
Successful:         5
Failed:             0
Elapsed:            13.332880403s
Requests/sec:       0.38
Output tokens/sec:  6.00
URL:                http://localhost:11434
Model:              llama3
Prompt tokens:      512
Completion tokens:  16
Concurrency:        2

Avg Latency:        5.070435294s
Latency p50/p95:    2.708161626s / 7.976055943s
TTFT p50/p95:       1.44212132s / 6.737083255s

```

Run a sweep:

```bash
llmbench sweep \
  --url http://localhost:11434 \
  --model llama3 \
  --requests 16 \
  --completion-tokens 16 \
  --concurrency 1,2,4
```

Example output:

```text
Concurrency   Success   Req/sec   Tok/sec   Lat p50       Lat p95
-----------------------------------------------------------------
1             16/16     0.72      11.52     1.387377436s 1.399095169s
2             16/16     0.74      11.77     2.659929337s 2.800495294s
4             16/16     0.70      11.21     5.614021364s 5.874859306s
```
