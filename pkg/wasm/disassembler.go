package wasm

import "fmt"

func DisassembleCode(code []byte) ([]Instruction, error) {
	var instructions []Instruction
	pc := 0

	for pc < len(code) {
		startPC := pc
		op := Opcode(code[pc])
		pc++

		instr := Instruction{
			Offset: uint64(startPC),
			Opcode: op,
			Name:   fmt.Sprintf("0x%02x", byte(op)),
		}

		if name, ok := OpcodeNames[op]; ok {
			instr.Name = name
		} else {
			instr.Name = "unknown"
		}

		switch op {
		case OpI32Const:
			val, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, val)
			pc += n

		case OpLocalGet, OpLocalSet, OpLocalTee, OpCall, OpBr, OpBrIf:
			val, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, err
			}
			instr.Immediates = append(instr.Immediates, val)
			pc += n
		}

		instructions = append(instructions, instr)
	}

	return instructions, nil
}
