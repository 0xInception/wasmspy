package wasm

import (
	"testing"
)

func TestParseMemorySection(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []Limits
	}{
		{
			name: "single memory no max",
			input: []byte{
				0x01, // 1 memory
				0x00, // flags: no max
				0x01, // min: 1 page
			},
			expected: []Limits{
				{Min: 1, HasMax: false},
			},
		},
		{
			name: "single memory with max",
			input: []byte{
				0x01, // 1 memory
				0x01, // flags: has max
				0x01, // min: 1 page
				0x10, // max: 16 pages
			},
			expected: []Limits{
				{Min: 1, Max: 16, HasMax: true},
			},
		},
		{
			name: "multiple memories",
			input: []byte{
				0x02, // 2 memories
				0x00, // flags: no max
				0x01, // min: 1
				0x01, // flags: has max
				0x02, // min: 2
				0x08, // max: 8
			},
			expected: []Limits{
				{Min: 1, HasMax: false},
				{Min: 2, Max: 8, HasMax: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMemorySection(tt.input, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("count mismatch: got %d, want %d", len(got), len(tt.expected))
			}

			for i := range got {
				if got[i].Min != tt.expected[i].Min {
					t.Errorf("memory[%d] min: got %d, want %d", i, got[i].Min, tt.expected[i].Min)
				}
				if got[i].HasMax != tt.expected[i].HasMax {
					t.Errorf("memory[%d] hasMax: got %v, want %v", i, got[i].HasMax, tt.expected[i].HasMax)
				}
				if tt.expected[i].HasMax && got[i].Max != tt.expected[i].Max {
					t.Errorf("memory[%d] max: got %d, want %d", i, got[i].Max, tt.expected[i].Max)
				}
			}
		})
	}
}
