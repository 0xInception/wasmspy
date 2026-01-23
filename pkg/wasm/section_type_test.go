package wasm

import (
	"testing"
)

func TestParseTypeSection(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []FuncType
	}{
		{
			name: "single func (i32, i32) -> i32",
			input: []byte{
				0x01,             // 1 type
				0x60,             // functype marker
				0x02, 0x7F, 0x7F, // 2 params: i32, i32
				0x01, 0x7F,       // 1 result: i32
			},
			expected: []FuncType{
				{Params: []ValType{ValI32, ValI32}, Results: []ValType{ValI32}},
			},
		},
		{
			name: "no params no results",
			input: []byte{
				0x01, // 1 type
				0x60, // functype marker
				0x00, // 0 params
				0x00, // 0 results
			},
			expected: []FuncType{
				{Params: []ValType{}, Results: []ValType{}},
			},
		},
		{
			name: "multiple types",
			input: []byte{
				0x02,       // 2 types
				0x60,       // type 0
				0x00,       // 0 params
				0x01, 0x7F, // 1 result: i32
				0x60,             // type 1
				0x01, 0x7E,       // 1 param: i64
				0x02, 0x7D, 0x7C, // 2 results: f32, f64
			},
			expected: []FuncType{
				{Params: []ValType{}, Results: []ValType{ValI32}},
				{Params: []ValType{ValI64}, Results: []ValType{ValF32, ValF64}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypeSection(tt.input, 0)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.expected) {
				t.Fatalf("count mismatch: got %d, want %d", len(got), len(tt.expected))
			}

			for i := range got {
				if len(got[i].Params) != len(tt.expected[i].Params) {
					t.Errorf("type[%d] param count: got %d, want %d", i, len(got[i].Params), len(tt.expected[i].Params))
				}
				for j := range got[i].Params {
					if got[i].Params[j] != tt.expected[i].Params[j] {
						t.Errorf("type[%d] param[%d]: got %v, want %v", i, j, got[i].Params[j], tt.expected[i].Params[j])
					}
				}

				if len(got[i].Results) != len(tt.expected[i].Results) {
					t.Errorf("type[%d] result count: got %d, want %d", i, len(got[i].Results), len(tt.expected[i].Results))
				}
				for j := range got[i].Results {
					if got[i].Results[j] != tt.expected[i].Results[j] {
						t.Errorf("type[%d] result[%d]: got %v, want %v", i, j, got[i].Results[j], tt.expected[i].Results[j])
					}
				}
			}
		})
	}
}

func TestFuncTypeString(t *testing.T) {
	tests := []struct {
		ft       FuncType
		expected string
	}{
		{
			ft:       FuncType{Params: []ValType{ValI32, ValI32}, Results: []ValType{ValI32}},
			expected: "(func (param i32 i32) (result i32))",
		},
		{
			ft:       FuncType{Params: []ValType{}, Results: []ValType{}},
			expected: "(func)",
		},
		{
			ft:       FuncType{Params: []ValType{ValI64}, Results: []ValType{ValF32, ValF64}},
			expected: "(func (param i64) (result f32 f64))",
		},
	}

	for _, tt := range tests {
		got := tt.ft.String()
		if got != tt.expected {
			t.Errorf("got %q, want %q", got, tt.expected)
		}
	}
}

func TestParseTypeSectionError(t *testing.T) {
	input := []byte{
		0x01, // 1 type
		0x61, // wrong marker (should be 0x60)
	}

	_, err := ParseTypeSection(input, 0x50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	t.Logf("Error: %v", err)

	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T", err)
	}

	if pe.Code != ErrInvalidSection {
		t.Errorf("expected ErrInvalidSection, got %d", pe.Code)
	}
}
