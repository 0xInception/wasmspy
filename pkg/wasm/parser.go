package wasm

import (
	"encoding/binary"
	"io"
	"os"
)

type parser struct {
	data   []byte
	offset int
}

func (p *parser) remaining() int {
	return len(p.data) - p.offset
}

func (p *parser) readBytes(n int) ([]byte, error) {
	if p.remaining() < n {
		return nil, newError(ErrTruncated, int64(p.offset), "need %d bytes, have %d", n, p.remaining())
	}
	b := p.data[p.offset : p.offset+n]
	p.offset += n
	return b, nil
}

func (p *parser) readByte() (byte, error) {
	if p.remaining() < 1 {
		return 0, newError(ErrTruncated, int64(p.offset), "unexpected end of data")
	}
	b := p.data[p.offset]
	p.offset++
	return b, nil
}

func (p *parser) readU32() (uint32, error) {
	val, n, err := ReadLEB128U32FromSlice(p.data[p.offset:])
	if err != nil {
		return 0, wrapError(ErrInvalidLEB128, int64(p.offset), err, "invalid leb128")
	}
	p.offset += n
	return val, nil
}

func (p *parser) readS32() (int32, error) {
	val, n, err := ReadLEB128S32FromSlice(p.data[p.offset:])
	if err != nil {
		return 0, wrapError(ErrInvalidLEB128, int64(p.offset), err, "invalid signed leb128")
	}
	p.offset += n
	return val, nil
}

func ParseFile(path string) (*Module, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

func Parse(data []byte) (*Module, error) {
	p := &parser{data: data}
	mod := &Module{}

	// Magic Number: \0asm
	magic, err := p.readBytes(4)
	if err != nil {
		return nil, err
	}
	if string(magic) != "\x00asm" {
		return nil, newError(ErrInvalidMagic, 0, "invalid magic number: %x", magic)
	}

	if p.remaining() < 4 {
		return nil, newError(ErrTruncated, int64(p.offset), "missing version")
	}
	mod.Version = binary.LittleEndian.Uint32(p.data[p.offset:])
	p.offset += 4

	for p.remaining() > 0 {
		sectionStart := p.offset

		idByte, err := p.readByte()
		if err != nil {
			return nil, err
		}

		size, err := p.readU32()
		if err != nil {
			return nil, err
		}

		if p.remaining() < int(size) {
			return nil, newError(ErrSectionOverflow, int64(sectionStart), "section %d claims %d bytes, only %d available", idByte, size, p.remaining())
		}

		content, _ := p.readBytes(int(size))

		mod.Sections = append(mod.Sections, Section{
			ID:      SectionID(idByte),
			Offset:  uint64(sectionStart),
			Size:    size,
			Content: content,
		})
	}

	return mod, nil
}

func ParseFromReader(r io.Reader) (*Module, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}
