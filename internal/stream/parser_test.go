package stream

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// buildSSE constructs a minimal SSE stream string from a slice of JSON payloads.
func buildSSE(payloads []string) string {
	var sb strings.Builder
	for _, p := range payloads {
		sb.WriteString("data: ")
		sb.WriteString(p)
		sb.WriteString("\n\n")
	}
	sb.WriteString("data: [DONE]\n\n")
	return sb.String()
}

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestParseChatCompletionsStream_EstimateOnly(t *testing.T) {
	// Three content-bearing chunks, no usage block → OutputTokens should equal EstimatedOutputTokens = 3.
	chunks := []string{
		`{"choices":[{"delta":{"content":"Hello"}}]}`,
		`{"choices":[{"delta":{"content":" world"}}]}`,
		`{"choices":[{"delta":{"content":"!"}}]}`,
	}
	r := strings.NewReader(buildSSE(chunks))
	cap, err := ParseChatCompletionsStream(r, fixedNow(time.Now()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.EstimatedOutputTokens != 3 {
		t.Errorf("EstimatedOutputTokens: got %d, want 3", cap.EstimatedOutputTokens)
	}
	if cap.UsageOutputTokens != 0 {
		t.Errorf("UsageOutputTokens: got %d, want 0", cap.UsageOutputTokens)
	}
	if cap.OutputTokens != 3 {
		t.Errorf("OutputTokens: got %d, want 3", cap.OutputTokens)
	}
}

func TestParseChatCompletionsStream_UsagePreferred(t *testing.T) {
	// Three content chunks + a usage block with completion_tokens=7.
	// OutputTokens must equal 7 (usage), not 3 (estimate) — no double-counting.
	chunks := []string{
		`{"choices":[{"delta":{"content":"Hello"}}]}`,
		`{"choices":[{"delta":{"content":" world"}}]}`,
		`{"choices":[{"delta":{"content":"!"}}]}`,
		`{"choices":[],"usage":{"prompt_tokens":10,"completion_tokens":7,"total_tokens":17}}`,
	}
	r := strings.NewReader(buildSSE(chunks))
	cap, err := ParseChatCompletionsStream(r, fixedNow(time.Now()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.EstimatedOutputTokens != 3 {
		t.Errorf("EstimatedOutputTokens: got %d, want 3", cap.EstimatedOutputTokens)
	}
	if cap.UsageOutputTokens != 7 {
		t.Errorf("UsageOutputTokens: got %d, want 7", cap.UsageOutputTokens)
	}
	// Must not double-count: should be 7, not 10 (3+7).
	if cap.OutputTokens != 7 {
		t.Errorf("OutputTokens: got %d, want 7 (no double-count)", cap.OutputTokens)
	}
}

func TestParseChatCompletionsStream_FirstAndLastToken(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tick := 0
	clk := func() time.Time {
		t := base.Add(time.Duration(tick) * time.Millisecond)
		tick++
		return t
	}

	chunks := []string{
		`{"choices":[{"delta":{"content":"A"}}]}`,
		`{"choices":[{"delta":{"content":"B"}}]}`,
	}
	r := strings.NewReader(buildSSE(chunks))
	cap, err := ParseChatCompletionsStream(r, clk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.FirstTokenAt.IsZero() {
		t.Error("FirstTokenAt should not be zero")
	}
	if cap.LastTokenAt.IsZero() {
		t.Error("LastTokenAt should not be zero")
	}
	if !cap.LastTokenAt.After(cap.FirstTokenAt) {
		t.Errorf("LastTokenAt (%v) should be after FirstTokenAt (%v)", cap.LastTokenAt, cap.FirstTokenAt)
	}
}

func TestParseChatCompletionsStream_EmptyStream(t *testing.T) {
	r := strings.NewReader("data: [DONE]\n\n")
	cap, err := ParseChatCompletionsStream(r, fixedNow(time.Now()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.OutputTokens != 0 {
		t.Errorf("OutputTokens: got %d, want 0", cap.OutputTokens)
	}
}

func TestParseChatCompletionsStream_InterTokenLatencies(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	calls := []time.Time{
		base,                             // first byte
		base.Add(10 * time.Millisecond), // first token (chunk 0 parse)
		base.Add(20 * time.Millisecond), // second token (chunk 1 parse)
		base.Add(30 * time.Millisecond), // third token (chunk 2 parse)
	}
	idx := 0
	clk := func() time.Time {
		if idx >= len(calls) {
			return calls[len(calls)-1]
		}
		t := calls[idx]
		idx++
		return t
	}

	chunks := []string{
		`{"choices":[{"delta":{"content":"A"}}]}`,
		`{"choices":[{"delta":{"content":"B"}}]}`,
		`{"choices":[{"delta":{"content":"C"}}]}`,
	}
	r := strings.NewReader(buildSSE(chunks))
	cap, err := ParseChatCompletionsStream(r, clk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 3 tokens → 2 inter-token gaps
	if len(cap.InterTokenLatencies) != 2 {
		t.Errorf("InterTokenLatencies: got %d, want 2; values: %v", len(cap.InterTokenLatencies), cap.InterTokenLatencies)
	}
	_ = fmt.Sprintf("latencies: %v", cap.InterTokenLatencies)
}
