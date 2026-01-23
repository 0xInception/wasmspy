package decompile

import "github.com/0xInception/wasmspy/pkg/wasm"

func Simplify(e Expr) Expr {
	if e == nil {
		return nil
	}

	switch v := e.(type) {
	case *BinaryExpr:
		return simplifyBinary(v)
	case *UnaryExpr:
		return simplifyUnary(v)
	case *TernaryExpr:
		return simplifyTernary(v)
	}
	return e
}

func simplifyBinary(e *BinaryExpr) Expr {
	left := Simplify(e.Left)
	right := Simplify(e.Right)

	lconst, lok := left.(*ConstExpr)
	rconst, rok := right.(*ConstExpr)

	if lok && rok {
		if result := foldBinary(e.Op, lconst, rconst); result != nil {
			return result
		}
	}

	switch e.Op {
	case wasm.OpI32Sub, wasm.OpI64Sub:
		if lok && isZero(lconst) {
			return &NegExpr{Arg: right, Type: e.Type}
		}
		if rok && isZero(rconst) {
			return left
		}

	case wasm.OpI32Add, wasm.OpI64Add:
		if lok && isZero(lconst) {
			return right
		}
		if rok && isZero(rconst) {
			return left
		}

	case wasm.OpI32Mul, wasm.OpI64Mul:
		if lok && isZero(lconst) {
			return lconst
		}
		if rok && isZero(rconst) {
			return rconst
		}
		if lok && isOne(lconst) {
			return right
		}
		if rok && isOne(rconst) {
			return left
		}

	case wasm.OpI32DivS, wasm.OpI32DivU, wasm.OpI64DivS, wasm.OpI64DivU:
		if rok && isOne(rconst) {
			return left
		}

	case wasm.OpI32And, wasm.OpI64And:
		if (lok && isZero(lconst)) || (rok && isZero(rconst)) {
			return &ConstExpr{Value: int32(0), Type: e.Type}
		}

	case wasm.OpI32Or, wasm.OpI64Or:
		if lok && isZero(lconst) {
			return right
		}
		if rok && isZero(rconst) {
			return left
		}

	case wasm.OpI32Shl, wasm.OpI32ShrS, wasm.OpI32ShrU, wasm.OpI64Shl, wasm.OpI64ShrS, wasm.OpI64ShrU:
		if rok && isZero(rconst) {
			return left
		}
	}

	if left != e.Left || right != e.Right {
		return &BinaryExpr{Op: e.Op, Left: left, Right: right, Type: e.Type}
	}
	return e
}

func simplifyUnary(e *UnaryExpr) Expr {
	arg := Simplify(e.Arg)

	if c, ok := arg.(*ConstExpr); ok {
		switch e.Op {
		case wasm.OpI32Eqz:
			if isZero(c) {
				return &ConstExpr{Value: int32(1), Type: wasm.ValI32}
			}
			return &ConstExpr{Value: int32(0), Type: wasm.ValI32}
		}
	}

	if neg, ok := arg.(*NegExpr); ok {
		switch e.Op {
		case wasm.OpI32Eqz, wasm.OpI64Eqz:
			return &UnaryExpr{Op: e.Op, Arg: neg.Arg, Type: e.Type}
		}
	}

	if arg != e.Arg {
		return &UnaryExpr{Op: e.Op, Arg: arg, Type: e.Type}
	}
	return e
}

func simplifyTernary(e *TernaryExpr) Expr {
	cond := Simplify(e.Cond)
	then := Simplify(e.ThenResult)
	els := Simplify(e.ElseResult)

	if c, ok := cond.(*ConstExpr); ok {
		if !isZero(c) {
			return then
		}
		return els
	}

	if cond != e.Cond || then != e.ThenResult || els != e.ElseResult {
		return &TernaryExpr{Cond: cond, ThenResult: then, ElseResult: els, Type: e.Type}
	}
	return e
}

func foldBinary(op wasm.Opcode, left, right *ConstExpr) Expr {
	lv, lok := toInt64(left.Value)
	rv, rok := toInt64(right.Value)
	if !lok || !rok {
		return nil
	}

	var result int64
	switch op {
	case wasm.OpI32Add, wasm.OpI64Add:
		result = lv + rv
	case wasm.OpI32Sub, wasm.OpI64Sub:
		result = lv - rv
	case wasm.OpI32Mul, wasm.OpI64Mul:
		result = lv * rv
	case wasm.OpI32And, wasm.OpI64And:
		result = lv & rv
	case wasm.OpI32Or, wasm.OpI64Or:
		result = lv | rv
	case wasm.OpI32Xor, wasm.OpI64Xor:
		result = lv ^ rv
	default:
		return nil
	}

	if left.Type == wasm.ValI32 {
		return &ConstExpr{Value: int32(result), Type: wasm.ValI32}
	}
	return &ConstExpr{Value: result, Type: wasm.ValI64}
}

func isZero(c *ConstExpr) bool {
	v, ok := toInt64(c.Value)
	return ok && v == 0
}

func isOne(c *ConstExpr) bool {
	v, ok := toInt64(c.Value)
	return ok && v == 1
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int32:
		return int64(x), true
	case int64:
		return x, true
	case uint32:
		return int64(x), true
	case uint64:
		return int64(x), true
	}
	return 0, false
}

func SimplifyBody(body *FuncBody) {
	for i := range body.Stmts {
		body.Stmts[i] = simplifyStmt(body.Stmts[i])
	}
	if body.Return != nil {
		body.Return = Simplify(body.Return)
	}
}

func simplifyStmt(s Stmt) Stmt {
	switch v := s.(type) {
	case *AssignStmt:
		return &AssignStmt{Target: v.Target, Value: Simplify(v.Value)}
	case *StoreStmt:
		return &StoreStmt{Op: v.Op, Addr: Simplify(v.Addr), Value: Simplify(v.Value), Offset: v.Offset}
	case *ReturnStmt:
		if v.Value != nil {
			return &ReturnStmt{Value: Simplify(v.Value)}
		}
	case *DropStmt:
		return &DropStmt{Value: Simplify(v.Value)}
	case *IfStmt:
		then := make([]Stmt, len(v.Then))
		for i := range v.Then {
			then[i] = simplifyStmt(v.Then[i])
		}
		els := make([]Stmt, len(v.Else))
		for i := range v.Else {
			els[i] = simplifyStmt(v.Else[i])
		}
		return &IfStmt{Cond: Simplify(v.Cond), Then: then, Else: els}
	case *LoopStmt:
		body := make([]Stmt, len(v.Body))
		for i := range v.Body {
			body[i] = simplifyStmt(v.Body[i])
		}
		return &LoopStmt{Label: v.Label, Body: body}
	case *BlockStmt:
		body := make([]Stmt, len(v.Body))
		for i := range v.Body {
			body[i] = simplifyStmt(v.Body[i])
		}
		return &BlockStmt{Label: v.Label, Body: body}
	case *BreakStmt:
		if v.Cond != nil {
			return &BreakStmt{Label: v.Label, Cond: Simplify(v.Cond)}
		}
	case *SwitchStmt:
		return &SwitchStmt{Value: Simplify(v.Value), Cases: v.Cases, Default: v.Default}
	case *WhileStmt:
		body := make([]Stmt, len(v.Body))
		for i := range v.Body {
			body[i] = simplifyStmt(v.Body[i])
		}
		return &WhileStmt{Cond: Simplify(v.Cond), Body: body}
	}
	return s
}
