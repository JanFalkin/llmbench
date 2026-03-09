package stream

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

func ParseChatCompletionsStream(r io.Reader, now func() time.Time) (*Capture, error) {
	br := bufio.NewReader(r)

	cap := &Capture{}
	var sawFirstByte bool
	var lastTokenTime time.Time
	var estimated int

	for {
		line, err := br.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("read stream line: %w", err)
		}

		if !sawFirstByte && len(line) > 0 {
			cap.FirstByteAt = now()
			sawFirstByte = true
		}

		line = strings.TrimSpace(line)
		if line == "" {
			if err == io.EOF {
				break
			}
			continue
		}

		if !strings.HasPrefix(line, "data:") {
			if err == io.EOF {
				break
			}
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" {
			if err == io.EOF {
				break
			}
			continue
		}

		if payload == "[DONE]" {
			break
		}

		var chunk chatCompletionChunk
		if uerr := json.Unmarshal([]byte(payload), &chunk); uerr != nil {
			return nil, fmt.Errorf("unmarshal stream chunk: %w", uerr)
		}

		ts := now()

		if chunk.Usage != nil {
			cap.UsageBlockAt = ts
			if chunk.Usage.CompletionTokens > 0 {
				cap.UsageOutputTokens = chunk.Usage.CompletionTokens
			}
		}

		tokenText := extractDeltaContent(chunk)
		if tokenText != "" {
			if cap.FirstTokenAt.IsZero() {
				cap.FirstTokenAt = ts
			} else if !lastTokenTime.IsZero() {
				cap.InterTokenLatencies = append(cap.InterTokenLatencies, ts.Sub(lastTokenTime))
			}

			cap.LastTokenAt = ts
			lastTokenTime = ts

			// v1 approximation: count each content-bearing chunk as one token
			// replace later with tokenizer-aware logic if needed
			estimated++
		}

		if err == io.EOF {
			break
		}
	}

	cap.EstimatedOutputTokens = estimated
	// Prefer server-reported usage when present; fall back to estimate.
	if cap.UsageOutputTokens > 0 {
		cap.OutputTokens = cap.UsageOutputTokens
	} else {
		cap.OutputTokens = estimated
	}

	return cap, nil
}
