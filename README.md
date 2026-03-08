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
  --url http://localhost:8000/v1/chat/completions \
  --model llama3 \
  --prompt-tokens 512 \
  --completion-tokens 128 \
  --concurrency 16
```

Example output:

```text
Model: llama3
Concurrency: 16

Requests/sec:       8.2
Output tokens/sec:  1045

TTFT p50/p95:       182ms / 391ms
Inter-token p50/p95: 19ms / 44ms
Latency p50/p95:    1.24s / 2.06s
```
