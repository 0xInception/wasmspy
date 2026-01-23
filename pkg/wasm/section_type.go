package wasm

const funcTypeMarker = 0x60

func ParseTypeSection(content []byte, baseOffset int) ([]FuncType, error) {
	p := &parser{data: content}

	count, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read type count")
	}

	types := make([]FuncType, 0, count)

	for i := 0; i < int(count); i++ {
		marker, err := p.readByte()
		if err != nil {
			return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read functype marker")
		}
		if marker != funcTypeMarker {
			return nil, newError(ErrInvalidSection, int64(baseOffset+p.offset-1), "expected functype marker 0x60, got 0x%02x", marker)
		}

		paramCount, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read param count")
		}

		params := make([]ValType, paramCount)
		for j := 0; j < int(paramCount); j++ {
			b, err := p.readByte()
			if err != nil {
				return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read param type")
			}
			params[j] = ValType(b)
		}

		resultCount, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read result count")
		}

		results := make([]ValType, resultCount)
		for j := 0; j < int(resultCount); j++ {
			b, err := p.readByte()
			if err != nil {
				return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read result type")
			}
			results[j] = ValType(b)
		}

		types = append(types, FuncType{
			Params:  params,
			Results: results,
		})
	}

	return types, nil
}
