package wasm

import (
	"bytes"
	"testing"
)

func TestReadLEB128U32(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		expectedVal   uint32
		expectedCount int
		expectError   bool
	}{
		{"Zero", []byte{0x00}, 0, 1, false},
		{"One byte (127)", []byte{0x7F}, 127, 1, false},
		{"Two bytes (128)", []byte{0x80, 0x01}, 128, 2, false},
		{"Two bytes (255)", []byte{0xFF, 0x01}, 255, 2, false},
		{"Max Uint32", []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x0F}, 0xFFFFFFFF, 5, false},

		// Edge Cases
		{"Unexpected EOF", []byte{0x80}, 0, 1, true},                         // Indicates more bytes coming, but buffer ends
		{"Overflow", []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x01}, 0, 0, true}, // Too many bytes for u32
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.input)
			val, count, err := ReadLEB128U32(r)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if val != tt.expectedVal {
				t.Errorf("value mismatch: got %d, want %d", val, tt.expectedVal)
			}

			if count != tt.expectedCount {
				t.Errorf("byte count mismatch: got %d, want %d", count, tt.expectedCount)
			}
		})
	}
}
