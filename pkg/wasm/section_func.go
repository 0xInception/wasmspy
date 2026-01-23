package wasm

func ParseFunctionSection(content []byte, baseOffset int) ([]uint32, error) {
	p := &parser{data: content}

	count, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read function count")
	}

	typeIndices := make([]uint32, 0, count)

	for i := 0; i < int(count); i++ {
		idx, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read type index for function %d", i)
		}
		typeIndices = append(typeIndices, idx)
	}

	return typeIndices, nil
}
