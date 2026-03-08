# llmbench

Benchmark and analyze performance of OpenAI-compatible LLM endpoints.

Metrics:
- latency
- time-to-first-token
- token throughput
- concurrency scaling
- streaming token timing

Designed for systems like:
- vLLM
- llama.cpp
- Ollama
- OpenAI-compatible APIs

## Quick Start

Install:

go install github.com/JanFalkin/llmbench@latest

Run a benchmark:

llmbench benchmark \
  --url http://localhost:8000/v1/chat/completions \
  --model llama3 \
  --prompt-tokens 512 \
  --completion-tokens 128 \
  --concurrency 16
