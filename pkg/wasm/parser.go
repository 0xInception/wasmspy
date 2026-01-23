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

type LocalEntry struct {
	Count uint32
	Type  byte
}

type FunctionBody struct {
	Offset       int
	Locals       []LocalEntry
	Instructions []Instruction
}

func ParseCodeSection(content []byte, baseOffset int) ([]FunctionBody, error) {
	p := &parser{data: content}

	numBodies, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidSection, int64(baseOffset), err, "failed to read function count")
	}

	bodies := make([]FunctionBody, 0, numBodies)

	for i := 0; i < int(numBodies); i++ {
		funcOffset := baseOffset + p.offset

		bodySize, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(funcOffset), err, "failed to read body size for function %d", i)
		}

		if p.remaining() < int(bodySize) {
			return nil, newError(ErrSectionOverflow, int64(funcOffset), "function %d body exceeds section bounds", i)
		}

		bodyData, _ := p.readBytes(int(bodySize))

		body, err := parseFunctionBody(bodyData, funcOffset)
		if err != nil {
			return nil, err
		}
		body.Offset = funcOffset

		bodies = append(bodies, body)
	}

	return bodies, nil
}

func parseFunctionBody(data []byte, baseOffset int) (FunctionBody, error) {
	p := &parser{data: data}
	var locals []LocalEntry

	numLocalDecls, err := p.readU32()
	if err != nil {
		return FunctionBody{}, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read local count")
	}

	for j := 0; j < int(numLocalDecls); j++ {
		count, err := p.readU32()
		if err != nil {
			return FunctionBody{}, err
		}

		valType, err := p.readByte()
		if err != nil {
			return FunctionBody{}, newError(ErrTruncated, int64(baseOffset+p.offset), "unexpected end reading local type")
		}

		locals = append(locals, LocalEntry{
			Count: count,
			Type:  valType,
		})
	}

	codeBytes := p.data[p.offset:]
	codeOffset := baseOffset + p.offset

	instrs, err := DisassembleCode(codeBytes, codeOffset)
	if err != nil {
		return FunctionBody{}, err
	}

	return FunctionBody{
		Locals:       locals,
		Instructions: instrs,
	}, nil
}

func ParseFromReader(r io.Reader) (*Module, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}
