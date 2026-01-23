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

func ReadLEB128S32(r io.ByteReader) (int32, int, error) {
	var result int32
	var shift uint
	var count int

	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, count, err
		}
		count++

		result |= int32(b&0x7F) << shift
		shift += 7

		if (b & 0x80) == 0 {
			if shift < 32 && (b&0x40) != 0 {
				result |= ^0 << shift
			}
			break
		}

		if shift >= 35 {
			return 0, count, fmt.Errorf("signed leb128 too large")
		}
	}
	return result, count, nil
}

func ReadLEB128S32FromSlice(data []byte) (int32, int, error) {
	reader := bytes.NewReader(data)
	return ReadLEB128S32(reader)
}

func ReadLEB128S64(r io.ByteReader) (int64, int, error) {
	var result int64
	var shift uint
	var count int

	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, count, err
		}
		count++

		result |= int64(b&0x7F) << shift
		shift += 7

		if (b & 0x80) == 0 {
			if shift < 64 && (b&0x40) != 0 {
				result |= ^0 << shift
			}
			break
		}

		if shift >= 70 {
			return 0, count, fmt.Errorf("signed leb128 too large")
		}
	}
	return result, count, nil
}

func ReadLEB128S64FromSlice(data []byte) (int64, int, error) {
	reader := bytes.NewReader(data)
	return ReadLEB128S64(reader)
}
