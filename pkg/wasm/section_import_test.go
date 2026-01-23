package wasm

import (
	"testing"
)

func TestParseImportSection(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []Import
	}{
		{
			name: "single func import",
			input: []byte{
				0x01,                   // 1 import
				0x03, 'e', 'n', 'v',    // module: "env"
				0x03, 'l', 'o', 'g',    // name: "log"
				0x00,                   // kind: func
				0x00,                   // type index: 0
			},
			expected: []Import{
				{Module: "env", Name: "log", Kind: ImportFunc, TypeIdx: 0},
			},
		},
		{
			name: "memory import",
			input: []byte{
				0x01,                   // 1 import
				0x02, 'j', 's',         // module: "js"
				0x03, 'm', 'e', 'm',    // name: "mem"
				0x02,                   // kind: memory
				0x00,                   // flags: no max
				0x01,                   // min: 1
			},
			expected: []Import{
				{Module: "js", Name: "mem", Kind: ImportMemory, Memory: &Limits{Min: 1, HasMax: false}},
			},
		},
		{
			name: "memory import with max",
			input: []byte{
				0x01,                   // 1 import
				0x02, 'j', 's',         // module: "js"
				0x03, 'm', 'e', 'm',    // name: "mem"
				0x02,                   // kind: memory
				0x01,                   // flags: has max
				0x01,                   // min: 1
				0x10,                   // max: 16
			},
			expected: []Import{
				{Module: "js", Name: "mem", Kind: ImportMemory, Memory: &Limits{Min: 1, Max: 16, HasMax: true}},
			},
		},
		{
			name: "global import",
			input: []byte{
				0x01,                      // 1 import
				0x03, 'e', 'n', 'v',       // module: "env"
				0x05, 'c', 'o', 'u', 'n', 't', // name: "count"
				0x03,                      // kind: global
				0x7F,                      // type: i32
				0x01,                      // mutable: yes
			},
			expected: []Import{
				{Module: "env", Name: "count", Kind: ImportGlobal, Global: &GlobalType{Type: ValI32, Mutable: true}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseImportSection(tt.input, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("count mismatch: got %d, want %d", len(got), len(tt.expected))
			}

			for i := range got {
				if got[i].Module != tt.expected[i].Module {
					t.Errorf("import[%d] module: got %q, want %q", i, got[i].Module, tt.expected[i].Module)
				}
				if got[i].Name != tt.expected[i].Name {
					t.Errorf("import[%d] name: got %q, want %q", i, got[i].Name, tt.expected[i].Name)
				}
				if got[i].Kind != tt.expected[i].Kind {
					t.Errorf("import[%d] kind: got %d, want %d", i, got[i].Kind, tt.expected[i].Kind)
				}

				switch tt.expected[i].Kind {
				case ImportFunc:
					if got[i].TypeIdx != tt.expected[i].TypeIdx {
						t.Errorf("import[%d] typeIdx: got %d, want %d", i, got[i].TypeIdx, tt.expected[i].TypeIdx)
					}
				case ImportMemory:
					if got[i].Memory.Min != tt.expected[i].Memory.Min {
						t.Errorf("import[%d] memory.min: got %d, want %d", i, got[i].Memory.Min, tt.expected[i].Memory.Min)
					}
					if got[i].Memory.HasMax != tt.expected[i].Memory.HasMax {
						t.Errorf("import[%d] memory.hasMax: got %v, want %v", i, got[i].Memory.HasMax, tt.expected[i].Memory.HasMax)
					}
					if tt.expected[i].Memory.HasMax && got[i].Memory.Max != tt.expected[i].Memory.Max {
						t.Errorf("import[%d] memory.max: got %d, want %d", i, got[i].Memory.Max, tt.expected[i].Memory.Max)
					}
				case ImportGlobal:
					if got[i].Global.Type != tt.expected[i].Global.Type {
						t.Errorf("import[%d] global.type: got %v, want %v", i, got[i].Global.Type, tt.expected[i].Global.Type)
					}
					if got[i].Global.Mutable != tt.expected[i].Global.Mutable {
						t.Errorf("import[%d] global.mutable: got %v, want %v", i, got[i].Global.Mutable, tt.expected[i].Global.Mutable)
					}
				}
			}
		})
	}
}
