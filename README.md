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
Requests:           5
Successful:         5
Failed:             0
Elapsed:            6.922719974s
Requests/sec:       0.72
Output tokens/sec:  11.56
URL:                http://localhost:11434
Model:              llama3
```
