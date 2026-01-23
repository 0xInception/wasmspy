package wasm

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
