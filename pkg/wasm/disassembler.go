package wasm

import (
	"encoding/binary"
	"fmt"
	"math"
)

func DisassembleCode(code []byte, baseOffset int) ([]Instruction, error) {
	var instructions []Instruction
	pc := 0

	for pc < len(code) {
		startPC := pc
		op := Opcode(code[pc])
		pc++

		if op == OpMiscPrefix {
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading misc opcode")
			}
			subOp, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid misc sub-opcode")
			}
			pc += n
			op = Opcode(0xfc00 | (subOp & 0xff))
		}

		instr := Instruction{
			Offset: uint64(baseOffset + startPC),
			Opcode: op,
		}

		if name, ok := OpcodeNames[op]; ok {
			instr.Name = name
		} else {
			instr.Name = fmt.Sprintf("unknown (0x%x)", op)
		}

		switch op {
		case OpBlock, OpLoop, OpIf:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading block type")
			}
			blockType := code[pc]
			pc++
			instr.Immediates = append(instr.Immediates, blockType)

		case OpBr, OpBrIf:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading branch index")
			}
			val, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid branch index")
			}
			instr.Immediates = append(instr.Immediates, val)
			pc += n

		case OpBrTable:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading br_table")
			}
			count, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid br_table count")
			}
			pc += n

			labels := make([]uint32, count+1)
			for i := uint32(0); i <= count; i++ {
				label, n, err := ReadLEB128U32FromSlice(code[pc:])
				if err != nil {
					return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid br_table label")
				}
				labels[i] = label
				pc += n
			}
			instr.Immediates = append(instr.Immediates, labels)

		case OpCall:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading call index")
			}
			val, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid call index")
			}
			instr.Immediates = append(instr.Immediates, val)
			pc += n

		case OpCallIndirect:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading call_indirect")
			}
			typeIdx, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid type index")
			}
			pc += n

			tableIdx, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid table index")
			}
			pc += n

			instr.Immediates = append(instr.Immediates, typeIdx, tableIdx)

		case OpLocalGet, OpLocalSet, OpLocalTee, OpGlobalGet, OpGlobalSet:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading index")
			}
			val, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid index")
			}
			instr.Immediates = append(instr.Immediates, val)
			pc += n

		case OpI32Load, OpI64Load, OpF32Load, OpF64Load,
			OpI32Load8S, OpI32Load8U, OpI32Load16S, OpI32Load16U,
			OpI64Load8S, OpI64Load8U, OpI64Load16S, OpI64Load16U,
			OpI64Load32S, OpI64Load32U,
			OpI32Store, OpI64Store, OpF32Store, OpF64Store,
			OpI32Store8, OpI32Store16,
			OpI64Store8, OpI64Store16, OpI64Store32:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading memarg")
			}
			align, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid memarg align")
			}
			pc += n

			offset, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid memarg offset")
			}
			pc += n

			instr.Immediates = append(instr.Immediates, align, offset)

		case OpMemorySize, OpMemoryGrow:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading memory index")
			}
			pc++

		case OpI32Const:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading i32.const")
			}
			val, n, err := ReadLEB128S32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid i32.const")
			}
			instr.Immediates = append(instr.Immediates, val)
			pc += n

		case OpI64Const:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading i64.const")
			}
			val, n, err := ReadLEB128S64FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid i64.const")
			}
			instr.Immediates = append(instr.Immediates, val)
			pc += n

		case OpF32Const:
			if pc+4 > len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading f32.const")
			}
			bits := binary.LittleEndian.Uint32(code[pc:])
			val := math.Float32frombits(bits)
			instr.Immediates = append(instr.Immediates, val)
			pc += 4

		case OpF64Const:
			if pc+8 > len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading f64.const")
			}
			bits := binary.LittleEndian.Uint64(code[pc:])
			val := math.Float64frombits(bits)
			instr.Immediates = append(instr.Immediates, val)
			pc += 8

		case OpMemoryInit:
			dataIdx, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid memory.init data index")
			}
			pc += n
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading memory.init")
			}
			pc++
			instr.Immediates = append(instr.Immediates, dataIdx)

		case OpDataDrop:
			dataIdx, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid data.drop index")
			}
			pc += n
			instr.Immediates = append(instr.Immediates, dataIdx)

		case OpMemoryCopy:
			if pc+2 > len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading memory.copy")
			}
			pc += 2

		case OpMemoryFill:
			if pc >= len(code) {
				return nil, newError(ErrTruncated, int64(baseOffset+pc), "unexpected end reading memory.fill")
			}
			pc++

		case OpTableInit:
			elemIdx, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid table.init elem index")
			}
			pc += n
			tableIdx, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid table.init table index")
			}
			pc += n
			instr.Immediates = append(instr.Immediates, elemIdx, tableIdx)

		case OpElemDrop:
			elemIdx, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid elem.drop index")
			}
			pc += n
			instr.Immediates = append(instr.Immediates, elemIdx)

		case OpTableCopy:
			dstIdx, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid table.copy dst index")
			}
			pc += n
			srcIdx, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid table.copy src index")
			}
			pc += n
			instr.Immediates = append(instr.Immediates, dstIdx, srcIdx)

		case OpTableGrow, OpTableSize, OpTableFill:
			tableIdx, n, err := ReadLEB128U32FromSlice(code[pc:])
			if err != nil {
				return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+pc), err, "invalid table index")
			}
			pc += n
			instr.Immediates = append(instr.Immediates, tableIdx)
		}

		instructions = append(instructions, instr)
	}

	return instructions, nil
}
