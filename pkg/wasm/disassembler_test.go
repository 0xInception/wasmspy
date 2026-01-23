package wasm

import (
	"reflect"
	"testing"
)

func TestDisassembleCode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []Instruction
	}{
		{
			name: "Simple Instructions",
			input: []byte{
				0x01, // nop
				0x6a, // i32.add
				0x0b, // end
			},
			expected: []Instruction{
				{Offset: 0, Opcode: 0x01, Name: "nop", Immediates: nil},
				{Offset: 1, Opcode: 0x6a, Name: "i32.add", Immediates: nil},
				{Offset: 2, Opcode: 0x0b, Name: "end", Immediates: nil},
			},
		},
		{
			name: "Instructions with Arguments",
			input: []byte{
				0x41, 0x0A, // i32.const 10
				0x41, 0x80, 0x01, // i32.const 128
				0x20, 0x00, // local.get 0
			},
			expected: []Instruction{
				{Offset: 0, Opcode: OpI32Const, Name: "i32.const", Immediates: []any{uint32(10)}},
				{Offset: 2, Opcode: OpI32Const, Name: "i32.const", Immediates: []any{uint32(128)}},
				{Offset: 5, Opcode: OpLocalGet, Name: "local.get", Immediates: []any{uint32(0)}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DisassembleCode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d instructions, want %d", len(got), len(tt.expected))
			}

			for i := range got {
				if got[i].Offset != tt.expected[i].Offset {
					t.Errorf("inst[%d] offset mismatch: got %d, want %d", i, got[i].Offset, tt.expected[i].Offset)
				}
				if got[i].Opcode != tt.expected[i].Opcode {
					t.Errorf("inst[%d] opcode mismatch: got %s, want %s", i, got[i].Name, tt.expected[i].Name)
				}

				if !reflect.DeepEqual(got[i].Immediates, tt.expected[i].Immediates) {
					t.Errorf("inst[%d] args mismatch: got %v, want %v", i, got[i].Immediates, tt.expected[i].Immediates)
				}
			}
		})
	}
}
