// internal/client/client.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JanFalkin/llmbench/internal/stream"
	"github.com/JanFalkin/llmbench/internal/workload"
)

type ResponseCapture struct {
	HTTPStatus          int
	FirstByteAt         time.Time
	FirstTokenAt        time.Time
	LastTokenAt         time.Time
	UsageBlockAt        time.Time
	OutputTokens        int
	InterTokenLatencies []time.Duration
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	headers    map[string]string
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`

	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func New(baseURL, apiKey string, timeout time.Duration, headers map[string]string) *Client {
	if headers == nil {
		headers = map[string]string{}
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		apiKey:  apiKey,
		headers: headers,
	}
}

func (c *Client) Do(ctx context.Context, req workload.RequestSpec) (*ResponseCapture, error) {

	body, err := c.buildRequest(ctx, req)

	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected http status: %s", resp.Status)
	}

	// capture := &ResponseCapture{
	// 	HTTPStatus: resp.StatusCode,
	// }

	// capture.FirstByteAt = time.Now()

	if req.Stream {
		cap, err := stream.ParseChatCompletionsStream(resp.Body, time.Now)
		if err != nil {
			return nil, err
		}
		return &ResponseCapture{
			HTTPStatus:          resp.StatusCode,
			FirstByteAt:         cap.FirstByteAt,
			FirstTokenAt:        cap.FirstTokenAt,
			LastTokenAt:         cap.LastTokenAt,
			UsageBlockAt:        cap.UsageBlockAt,
			OutputTokens:        cap.OutputTokens,
			InterTokenLatencies: cap.InterTokenLatencies,
		}, nil
	} else {
		return c.parseNonStreaming(resp)
	}

}

func estimateOutputTokens(resp chatCompletionResponse) int {
	totalChars := 0
	for _, ch := range resp.Choices {
		totalChars += len(ch.Message.Content)
	}

	if totalChars == 0 {
		return 0
	}

	// Very rough approximation for v1.
	// Replace later with tokenizer-aware logic if needed.
	estimated := totalChars / 4
	if estimated == 0 {
		return 1
	}
	return estimated
}

func (c *Client) parseNonStreaming(resp *http.Response) (*ResponseCapture, error) {
	var decoded chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode non-streaming response: %w", err)
	}

	now := time.Now()

	cap := &ResponseCapture{
		HTTPStatus:   resp.StatusCode,
		FirstByteAt:  now,
		FirstTokenAt: now,
		LastTokenAt:  now,
		UsageBlockAt: now,
	}

	if decoded.Usage != nil && decoded.Usage.CompletionTokens > 0 {
		cap.OutputTokens = decoded.Usage.CompletionTokens
	} else {
		cap.OutputTokens = estimateOutputTokens(decoded)
	}

	return cap, nil
}
