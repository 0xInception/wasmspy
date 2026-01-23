package wasm

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func ParseFile(path string) (*Module, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	return Parse(r)
}

func Parse(r *bufio.Reader) (*Module, error) {
	mod := &Module{}

	// Magic Number: \0asm
	magic := make([]byte, 4)
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, err
	}
	if string(magic) != "\x00asm" {
		return nil, fmt.Errorf("invalid magic: %x", magic)
	}

	if err := binary.Read(r, binary.LittleEndian, &mod.Version); err != nil {
		return nil, err
	}

	for {
		idByte, err := r.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		size, _, err := ReadLEB128U32(r)
		if err != nil {
			return nil, err
		}

		content := make([]byte, size)
		if _, err := io.ReadFull(r, content); err != nil {
			return nil, err
		}

		mod.Sections = append(mod.Sections, Section{
			ID:      SectionID(idByte),
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
	Locals       []LocalEntry
	Instructions []Instruction
}

func ParseCodeSection(content []byte) ([]FunctionBody, error) {
	cursor := 0

	numBodies, bytesRead, err := ReadLEB128U32FromSlice(content[cursor:])
	if err != nil {
		return nil, fmt.Errorf("failed to read code section vector size: %w", err)
	}
	cursor += bytesRead

	bodies := make([]FunctionBody, 0, numBodies)

	for i := 0; i < int(numBodies); i++ {
		if cursor >= len(content) {
			return nil, fmt.Errorf("unexpected EOF while reading code section body %d", i)
		}

		bodySize, bytesRead, err := ReadLEB128U32FromSlice(content[cursor:])
		if err != nil {
			return nil, fmt.Errorf("failed to read body size for function %d: %w", i, err)
		}
		cursor += bytesRead

		endOfBody := cursor + int(bodySize)
		if endOfBody > len(content) {
			return nil, fmt.Errorf("body size for function %d exceeds section bounds", i)
		}

		rawBody := content[cursor:endOfBody]

		body, err := parseFunctionBody(rawBody)
		if err != nil {
			return nil, fmt.Errorf("failed to parse function %d: %w", i, err)
		}

		bodies = append(bodies, body)

		cursor = endOfBody
	}

	return bodies, nil
}

func parseFunctionBody(rawBody []byte) (FunctionBody, error) {
	localCursor := 0
	var locals []LocalEntry

	numLocalDecls, bytesRead, err := ReadLEB128U32FromSlice(rawBody[localCursor:])
	if err != nil {
		return FunctionBody{}, fmt.Errorf("failed to read local declaration count: %w", err)
	}
	localCursor += bytesRead

	for j := 0; j < int(numLocalDecls); j++ {
		count, bytesRead, err := ReadLEB128U32FromSlice(rawBody[localCursor:])
		if err != nil {
			return FunctionBody{}, fmt.Errorf("failed to read local count: %w", err)
		}
		localCursor += bytesRead

		if localCursor >= len(rawBody) {
			return FunctionBody{}, fmt.Errorf("unexpected EOF reading local type")
		}

		valType := rawBody[localCursor]
		localCursor++

		locals = append(locals, LocalEntry{
			Count: count,
			Type:  valType,
		})
	}

	codeBytes := rawBody[localCursor:]

	instrs, err := DisassembleCode(codeBytes)
	if err != nil {
		return FunctionBody{}, fmt.Errorf("disassembly failed: %w", err)
	}

	return FunctionBody{
		Locals:       locals,
		Instructions: instrs,
	}, nil
}
