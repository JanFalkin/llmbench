package stats

import (
	"sort"
	"time"
)

func Percentile(d []time.Duration, p float64) time.Duration {
	if len(d) == 0 {
		return 0
	}

	s := make([]time.Duration, len(d))
	copy(s, d)

	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})

	idx := int(float64(len(s)-1) * p)

	return s[idx]
}
