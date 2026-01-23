package wasm

func ParseMemorySection(content []byte, baseOffset int) ([]Limits, error) {
	p := &parser{data: content}

	count, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read memory count")
	}

	memories := make([]Limits, 0, count)

	for i := 0; i < int(count); i++ {
		limits, err := p.readLimits(baseOffset)
		if err != nil {
			return nil, err
		}
		memories = append(memories, *limits)
	}

	return memories, nil
}
