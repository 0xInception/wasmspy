package wasm

import (
	"testing"
)

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectError bool
	}{
		{
			name:        "Valid Header v1",
			input:       []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}, // \0asm + v1
			expectError: false,
		},
		{
			name:        "Invalid Magic Number",
			input:       []byte{0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00},
			expectError: true,
		},
		{
			name:        "Truncated File",
			input:       []byte{0x00, 0x61},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
