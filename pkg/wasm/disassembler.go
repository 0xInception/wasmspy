package wasm

func DisassembleCode(code []byte, baseOffset int) ([]Instruction, error) {
	var instructions []Instruction
	pc := 0

	for pc < len(code) {
		startPC := pc
		op := Opcode(code[pc])
		pc++

		instr := Instruction{
			Offset: uint64(baseOffset + startPC),
			Opcode: op,
		}

		if name, ok := OpcodeNames[op]; ok {
			instr.Name = name
		} else {
			instr.Name = "unknown"
		}

		switch op {
		case OpI32Const:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading i32.const")
			}
			val, n, err := ReadLEB128S32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid i32.const immediate")
			}
			instr.Immediates = append(instr.Immediates, val)
			pc += n

		case OpI64Const:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading i64.const")
			}
			val, n, err := ReadLEB128S64FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid i64.const immediate")
			}
			instr.Immediates = append(instr.Immediates, val)
			pc += n

		case OpLocalGet, OpLocalSet, OpLocalTee, OpCall, OpBr, OpBrIf:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading index")
			}
			val, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid index immediate")
			}
			instr.Immediates = append(instr.Immediates, val)
			pc += n

		case OpBlock, OpLoop:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading block type")
			}
			blockType := code[pc]
			pc++
			instr.Immediates = append(instr.Immediates, blockType)
		}

		instructions = append(instructions, instr)
	}

	return instructions, nil
}
