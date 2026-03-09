// internal/workload/generator.go
package workload

import (
	"fmt"
	"sync/atomic"
)

type RequestSpec struct {
	ID               string
	Model            string
	Prompt           string
	PromptTokens     int
	CompletionTokens int
	Stream           bool
}

type Generator struct {
	counter atomic.Uint64
}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) MakeRequest(model string, promptTokens, completionTokens int, stream bool) RequestSpec {
	id := g.counter.Add(1)
	return RequestSpec{
		ID:               fmt.Sprintf("req-%d", id),
		Model:            model,
		Prompt:           fakePrompt(promptTokens),
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		Stream:           stream,
	}
}

func fakePrompt(n int) string {
	return fmt.Sprintf("Generate a response for a synthetic prompt of approximately %d tokens.", n)
}

