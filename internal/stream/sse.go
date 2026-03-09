package stream

import (
	"time"
)

type Capture struct {
	HTTPStatus          int
	FirstByteAt         time.Time
	FirstTokenAt        time.Time
	LastTokenAt         time.Time
	UsageBlockAt        time.Time
	OutputTokens        int
	ChunkCount          int
	InterTokenLatencies []time.Duration
}

type chatCompletionChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`

	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func extractDeltaContent(chunk chatCompletionChunk) string {
	for _, ch := range chunk.Choices {
		if ch.Delta.Content != "" {
			return ch.Delta.Content
		}
	}
	return ""
}
