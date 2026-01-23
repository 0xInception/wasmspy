package wasm

import (
	"bytes"
	"fmt"
	"io"
)

func ReadLEB128U32(r io.ByteReader) (uint32, int, error) {
	var result uint32
	var shift uint
	var count int

	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, count, err
		}
		count++

		result |= uint32(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
		if shift >= 32 {
			return 0, count, fmt.Errorf("leb128 too large")
		}
	}
	return result, count, nil
}

func ReadLEB128U32FromSlice(data []byte) (uint32, int, error) {
	reader := bytes.NewReader(data)
	return ReadLEB128U32(reader)
}
