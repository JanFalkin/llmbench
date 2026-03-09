package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/JanFalkin/llmbench/internal/workload"
)

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *Client) buildRequest(ctx context.Context, req workload.RequestSpec) (*http.Request, error) {
	payload := chatCompletionRequest{
		Model: req.Model,
		Messages: []chatMessage{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		MaxTokens:   req.CompletionTokens,
		Stream:      req.Stream,
		Temperature: 0.0,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal chat completion request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.chatCompletionsURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if req.Stream {
		httpReq.Header.Set("Accept", "text/event-stream")
	}

	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	for k, v := range c.headers {
		httpReq.Header.Set(k, v)
	}

	return httpReq, nil
}

func (c *Client) chatCompletionsURL() string {
	if strings.HasSuffix(c.baseURL, "/v1/chat/completions") {
		return c.baseURL
	}
	return strings.TrimRight(c.baseURL, "/") + "/v1/chat/completions"
}
