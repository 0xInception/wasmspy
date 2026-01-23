package wasm

import (
	"testing"
)

func TestParseGlobalSection(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []Global
	}{
		{
			name: "single i32 const global",
			input: []byte{
				0x01,       // 1 global
				0x7F,       // type: i32
				0x00,       // mutable: no
				0x41, 0x2A, // i32.const 42
				0x0B,       // end
			},
			expected: []Global{
				{
					Type: GlobalType{Type: ValI32, Mutable: false},
					Init: []Instruction{
						{Offset: 0x3, Opcode: OpI32Const, Name: "i32.const", Immediates: []any{int32(42)}},
						{Offset: 0x5, Opcode: OpEnd, Name: "end"},
					},
				},
			},
		},
		{
			name: "mutable i64 global",
			input: []byte{
				0x01,       // 1 global
				0x7E,       // type: i64
				0x01,       // mutable: yes
				0x42, 0x00, // i64.const 0
				0x0B,       // end
			},
			expected: []Global{
				{
					Type: GlobalType{Type: ValI64, Mutable: true},
					Init: []Instruction{
						{Offset: 0x3, Opcode: OpI64Const, Name: "i64.const", Immediates: []any{int64(0)}},
						{Offset: 0x5, Opcode: OpEnd, Name: "end"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGlobalSection(tt.input, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("count mismatch: got %d, want %d", len(got), len(tt.expected))
			}

			for i := range got {
				if got[i].Type.Type != tt.expected[i].Type.Type {
					t.Errorf("global[%d] type: got %v, want %v", i, got[i].Type.Type, tt.expected[i].Type.Type)
				}
				if got[i].Type.Mutable != tt.expected[i].Type.Mutable {
					t.Errorf("global[%d] mutable: got %v, want %v", i, got[i].Type.Mutable, tt.expected[i].Type.Mutable)
				}
				if len(got[i].Init) == 0 {
					t.Errorf("global[%d] init: empty", i)
				}
			}
		})
	}
}
