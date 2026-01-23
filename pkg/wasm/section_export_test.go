package wasm

import (
	"testing"
)

func TestParseExportSection(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []Export
	}{
		{
			name: "single function export",
			input: []byte{
				0x01,                   // 1 export
				0x03, 'a', 'd', 'd',    // name: "add"
				0x00,                   // kind: func
				0x00,                   // index: 0
			},
			expected: []Export{
				{Name: "add", Kind: ExportFunc, Index: 0},
			},
		},
		{
			name: "multiple exports",
			input: []byte{
				0x02,                         // 2 exports
				0x03, 'a', 'd', 'd',          // name: "add"
				0x00,                         // kind: func
				0x00,                         // index: 0
				0x06, 'm', 'e', 'm', 'o', 'r', 'y', // name: "memory"
				0x02,                         // kind: memory
				0x00,                         // index: 0
			},
			expected: []Export{
				{Name: "add", Kind: ExportFunc, Index: 0},
				{Name: "memory", Kind: ExportMemory, Index: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExportSection(tt.input, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("count mismatch: got %d, want %d", len(got), len(tt.expected))
			}

			for i := range got {
				if got[i].Name != tt.expected[i].Name {
					t.Errorf("export[%d] name: got %q, want %q", i, got[i].Name, tt.expected[i].Name)
				}
				if got[i].Kind != tt.expected[i].Kind {
					t.Errorf("export[%d] kind: got %d, want %d", i, got[i].Kind, tt.expected[i].Kind)
				}
				if got[i].Index != tt.expected[i].Index {
					t.Errorf("export[%d] index: got %d, want %d", i, got[i].Index, tt.expected[i].Index)
				}
			}
		})
	}
}
