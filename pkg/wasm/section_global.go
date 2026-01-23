package wasm

func ParseGlobalSection(content []byte, baseOffset int) ([]Global, error) {
	p := &parser{data: content}

	count, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read global count")
	}

	globals := make([]Global, 0, count)

	for i := 0; i < int(count); i++ {
		valType, err := p.readByte()
		if err != nil {
			return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read global type")
		}

		mut, err := p.readByte()
		if err != nil {
			return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read global mutability")
		}

		initStart := p.offset
		initBytes, err := p.readInitExpr()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+initStart), err, "failed to read global init expr")
		}

		initInstrs, err := DisassembleCode(initBytes, baseOffset+initStart)
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+initStart), err, "failed to disassemble init expr")
		}

		globals = append(globals, Global{
			Type: GlobalType{
				Type:    ValType(valType),
				Mutable: mut == 1,
			},
			Init: initInstrs,
		})
	}

	return globals, nil
}

func (p *parser) readInitExpr() ([]byte, error) {
	start := p.offset

	for p.offset < len(p.data) {
		op := p.data[p.offset]
		p.offset++

		if op == 0x0B {
			return p.data[start:p.offset], nil
		}

		switch Opcode(op) {
		case OpI32Const, OpI64Const:
			_, n, err := ReadLEB128S64FromSlice(p.data[p.offset:])
			if err != nil {
				return nil, err
			}
			p.offset += n
		case OpGlobalGet:
			_, n, err := ReadLEB128U32FromSlice(p.data[p.offset:])
			if err != nil {
				return nil, err
			}
			p.offset += n
		}
	}

	return nil, newError(ErrTruncated, int64(p.offset), "init expr missing end")
}
