package wasm

import (
	"testing"
)

func TestParseFunctionSection(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []uint32
	}{
		{
			name: "single function",
			input: []byte{
				0x01, // 1 function
				0x00, // type index 0
			},
			expected: []uint32{0},
		},
		{
			name: "multiple functions",
			input: []byte{
				0x03,       // 3 functions
				0x00,       // func 0 -> type 0
				0x01,       // func 1 -> type 1
				0x00,       // func 2 -> type 0
			},
			expected: []uint32{0, 1, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFunctionSection(tt.input, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("count mismatch: got %d, want %d", len(got), len(tt.expected))
			}

			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("func[%d]: got %d, want %d", i, got[i], tt.expected[i])
				}
			}
		})
	}
}
