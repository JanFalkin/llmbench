package config

import (
	"testing"
)

func TestParseConcurrencyList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{
			name:  "single value",
			input: "1",
			want:  []int{1},
		},
		{
			name:  "multiple values",
			input: "1,2,4",
			want:  []int{1, 2, 4},
		},
		{
			name:  "values with spaces",
			input: "1, 2, 4",
			want:  []int{1, 2, 4},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "non-numeric value",
			input:   "1,foo,4",
			wantErr: true,
		},
		{
			name:    "zero value",
			input:   "1,0,4",
			wantErr: true,
		},
		{
			name:    "negative value",
			input:   "1,-2,4",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseConcurrencyList(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result=%v)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("index %d: got %d, want %d", i, got[i], tc.want[i])
				}
			}
		})
	}
}
