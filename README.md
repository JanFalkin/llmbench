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

Example output:

```text
Requests:           5
Successful:         5
Failed:             0
Elapsed:            6.922719974s
Requests/sec:       0.72
Output tokens/sec:  11.56
URL:                http://localhost:11434
Model:              llama3
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
Concurrency   Success   Req/sec   Tok/sec   Avg Latency
-------------------------------------------------------
1             16/16     0.71      11.38     1.406416821s
2             16/16     0.75      12.00     2.584097372s
4             16/16     0.74      11.85     4.888931437s
```
