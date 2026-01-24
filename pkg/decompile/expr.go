package decompile

import "github.com/0xInception/wasmspy/pkg/wasm"

func ValueToExpr(v *Value) Expr {
	if v == nil {
		return nil
	}

	switch v.Source {
	case SourceLocal:
		return &LocalExpr{Index: v.Index, Type: v.Type}
	case SourceParam:
		return &ParamExpr{Index: v.Index, Type: v.Type}
	case SourceGlobal:
		return &GlobalExpr{Index: v.Index, Type: v.Type}
	case SourceConst:
		return &ConstExpr{Value: v.Const, Type: v.Type}
	case SourceOp:
		return opToExpr(v)
	case SourceMemory:
		return opToExpr(v)
	case SourceError:
		if v.Error != nil {
			return &ErrorExpr{
				Message: v.Error.Message,
				Offset:  v.Error.Offset,
				Opcode:  v.Error.Opcode,
			}
		}
	}
	return nil
}

func opToExpr(v *Value) Expr {
	if v.Op == nil {
		return nil
	}

	if v.Op.Ternary != nil {
		return &TernaryExpr{
			Cond:       v.Op.Ternary.Cond,
			ThenResult: v.Op.Ternary.ThenResult,
			ElseResult: v.Op.Ternary.ElseResult,
			Type:       v.Type,
		}
	}

	if v.Op.Instr == nil {
		return nil
	}

	op := v.Op.Instr.Opcode
	inputs := v.Op.Inputs

	if isBinaryOp(op) && len(inputs) >= 2 {
		return &BinaryExpr{
			Op:    op,
			Left:  ValueToExpr(inputs[0]),
			Right: ValueToExpr(inputs[1]),
			Type:  v.Type,
		}
	}

	if isUnaryOp(op) && len(inputs) >= 1 {
		return &UnaryExpr{
			Op:   op,
			Arg:  ValueToExpr(inputs[0]),
			Type: v.Type,
		}
	}

	if isLoadOp(op) && len(inputs) >= 1 {
		var offset uint32
		if len(v.Op.Instr.Immediates) >= 2 {
			offset = getU32(v.Op.Instr.Immediates, 1)
		}
		return &LoadExpr{
			Op:     op,
			Addr:   ValueToExpr(inputs[0]),
			Offset: offset,
			Type:   v.Type,
		}
	}

	if op == wasm.OpCall && len(v.Op.Instr.Immediates) >= 1 {
		idx := getU32(v.Op.Instr.Immediates, 0)
		args := make([]Expr, len(inputs))
		for i, in := range inputs {
			args[i] = ValueToExpr(in)
		}
		return &CallExpr{
			FuncIndex: idx,
			Args:      args,
			Type:      v.Type,
		}
	}

	if op == wasm.OpCallIndirect {
		args := make([]Expr, len(inputs))
		for i, in := range inputs {
			args[i] = ValueToExpr(in)
		}
		return &CallExpr{
			FuncIndex: 0xFFFFFFFF,
			Args:      args,
			Type:      v.Type,
		}
	}

	return &UnaryExpr{
		Op:   op,
		Arg:  safeValueToExpr(inputs, 0),
		Type: v.Type,
	}
}

func safeValueToExpr(inputs []*Value, idx int) Expr {
	if idx < len(inputs) {
		return ValueToExpr(inputs[idx])
	}
	return nil
}

func isBinaryOp(op wasm.Opcode) bool {
	switch op {
	case wasm.OpI32Add, wasm.OpI32Sub, wasm.OpI32Mul, wasm.OpI32DivS, wasm.OpI32DivU,
		wasm.OpI32RemS, wasm.OpI32RemU, wasm.OpI32And, wasm.OpI32Or, wasm.OpI32Xor,
		wasm.OpI32Shl, wasm.OpI32ShrS, wasm.OpI32ShrU, wasm.OpI32Rotl, wasm.OpI32Rotr,
		wasm.OpI32Eq, wasm.OpI32Ne, wasm.OpI32LtS, wasm.OpI32LtU,
		wasm.OpI32GtS, wasm.OpI32GtU, wasm.OpI32LeS, wasm.OpI32LeU,
		wasm.OpI32GeS, wasm.OpI32GeU,
		wasm.OpI64Add, wasm.OpI64Sub, wasm.OpI64Mul, wasm.OpI64DivS, wasm.OpI64DivU,
		wasm.OpI64RemS, wasm.OpI64RemU, wasm.OpI64And, wasm.OpI64Or, wasm.OpI64Xor,
		wasm.OpI64Shl, wasm.OpI64ShrS, wasm.OpI64ShrU, wasm.OpI64Rotl, wasm.OpI64Rotr,
		wasm.OpI64Eq, wasm.OpI64Ne, wasm.OpI64LtS, wasm.OpI64LtU,
		wasm.OpI64GtS, wasm.OpI64GtU, wasm.OpI64LeS, wasm.OpI64LeU,
		wasm.OpI64GeS, wasm.OpI64GeU:
		return true
	}
	return false
}

func isUnaryOp(op wasm.Opcode) bool {
	switch op {
	case wasm.OpI32Eqz, wasm.OpI32Clz, wasm.OpI32Ctz, wasm.OpI32Popcnt,
		wasm.OpI64Eqz, wasm.OpI64Clz, wasm.OpI64Ctz, wasm.OpI64Popcnt,
		wasm.OpI32WrapI64, wasm.OpI64ExtendI32S, wasm.OpI64ExtendI32U:
		return true
	}
	return false
}

func isLoadOp(op wasm.Opcode) bool {
	switch op {
	case wasm.OpI32Load, wasm.OpI64Load, wasm.OpF32Load, wasm.OpF64Load,
		wasm.OpI32Load8S, wasm.OpI32Load8U, wasm.OpI32Load16S, wasm.OpI32Load16U,
		wasm.OpI64Load8S, wasm.OpI64Load8U, wasm.OpI64Load16S, wasm.OpI64Load16U,
		wasm.OpI64Load32S, wasm.OpI64Load32U:
		return true
	}
	return false
}
