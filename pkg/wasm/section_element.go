package wasm

func ParseElementSection(content []byte, baseOffset int) ([]ElementSegment, error) {
	p := &parser{data: content}

	count, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+p.offset), err, "element count")
	}

	segments := make([]ElementSegment, 0, count)

	for i := uint32(0); i < count; i++ {
		tableIdx, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+p.offset), err, "table index")
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

		funcCount, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+p.offset), err, "func count")
		}

		funcIdxs := make([]uint32, funcCount)
		for j := uint32(0); j < funcCount; j++ {
			idx, err := p.readU32()
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+p.offset), err, "func index")
			}
			funcIdxs[j] = idx
		}

		segments = append(segments, ElementSegment{
			TableIndex: tableIdx,
			Offset:     offsetInstrs,
			FuncIdxs:   funcIdxs,
		})
	}

	return segments, nil
}
