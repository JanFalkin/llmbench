package config

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseConcurrencyList(s string) ([]int, error) {
	if strings.TrimSpace(s) == "" {
		return nil, fmt.Errorf("empty concurrency list")
	}

	parts := strings.Split(s, ",")
	out := make([]int, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("invalid concurrency value %q: %w", p, err)
		}
		if n <= 0 {
			return nil, fmt.Errorf("concurrency must be > 0: %d", n)
		}
		out = append(out, n)
	}

	return out, nil
}
