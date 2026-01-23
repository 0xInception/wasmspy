package wasm

func ParseDataSection(content []byte, baseOffset int) ([]DataSegment, error) {
	p := &parser{data: content}

	count, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+p.offset), err, "data segment count")
	}

	segments := make([]DataSegment, 0, count)

	for i := uint32(0); i < count; i++ {
		memIdx, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+p.offset), err, "memory index")
		}

		initStart := p.offset
		initBytes, err := p.readInitExpr()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+initStart), err, "offset expr")
		}

		offsetInstrs, err := DisassembleCode(initBytes, baseOffset+initStart)
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+initStart), err, "disassemble offset")
		}

		size, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+p.offset), err, "data size")
		}

		if p.offset+int(size) > len(p.data) {
			return nil, newError(ErrTruncated, int64(baseOffset+p.offset), "data bytes truncated")
		}

		data := make([]byte, size)
		copy(data, p.data[p.offset:p.offset+int(size)])
		p.offset += int(size)

		segments = append(segments, DataSegment{
			MemoryIndex: memIdx,
			Offset:      offsetInstrs,
			Data:        data,
		})
	}

	return segments, nil
}
