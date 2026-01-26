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

func CollapseSwitchBlocks(body *FuncBody) {
	body.Stmts = collapseSwitchInStmts(body.Stmts)
}

func collapseSwitchInStmts(stmts []Stmt) []Stmt {
	result := make([]Stmt, 0, len(stmts))
	for _, stmt := range stmts {
		result = append(result, collapseSwitchInStmt(stmt))
	}
	return result
}

func collapseSwitchInStmt(stmt Stmt) Stmt {
	switch s := stmt.(type) {
	case *BlockStmt:
		if flat := tryCollapseSwitch(s); flat != nil {
			return flat
		}
		return &BlockStmt{Label: s.Label, Body: collapseSwitchInStmts(s.Body), SrcOffset: s.SrcOffset, EndOffset: s.EndOffset, Offsets: s.Offsets}
	case *IfStmt:
		return &IfStmt{Cond: s.Cond, Then: collapseSwitchInStmts(s.Then), Else: collapseSwitchInStmts(s.Else), SrcOffset: s.SrcOffset, EndOffset: s.EndOffset, Offsets: s.Offsets}
	case *LoopStmt:
		return &LoopStmt{Label: s.Label, Body: collapseSwitchInStmts(s.Body), SrcOffset: s.SrcOffset, EndOffset: s.EndOffset, Offsets: s.Offsets}
	case *WhileStmt:
		return &WhileStmt{Cond: s.Cond, Body: collapseSwitchInStmts(s.Body), Offsets: s.Offsets}
	case *FlatSwitchStmt:
		cases := make([]SwitchCase, len(s.Cases))
		for i, c := range s.Cases {
			cases[i] = SwitchCase{Value: c.Value, Body: collapseSwitchInStmts(c.Body)}
		}
		return &FlatSwitchStmt{Value: s.Value, Cases: cases, Default: collapseSwitchInStmts(s.Default), Offsets: s.Offsets}
	}
	return stmt
}

func tryCollapseSwitch(outerBlock *BlockStmt) *FlatSwitchStmt {
	blocks := []*BlockStmt{outerBlock}
	bodies := [][]Stmt{nil}
	current := outerBlock

	for {
		if len(current.Body) == 0 {
			break
		}
		innerBlock, ok := current.Body[0].(*BlockStmt)
		if !ok {
			break
		}
		afterInner := current.Body[1:]
		blocks = append(blocks, innerBlock)
		bodies = append(bodies, afterInner)
		current = innerBlock
	}

	if len(blocks) < 2 {
		return nil
	}

	var sw *SwitchStmt
	for _, stmt := range current.Body {
		if s, ok := stmt.(*SwitchStmt); ok {
			sw = s
			break
		}
	}
	if sw == nil {
		return nil
	}

	labelToIdx := make(map[int]int)
	for i, b := range blocks {
		labelToIdx[b.Label] = i
	}

	outerLabel := blocks[0].Label
	cases := make([]SwitchCase, 0, len(sw.Cases))
	for i, label := range sw.Cases {
		idx, ok := labelToIdx[label]
		if !ok {
			return nil
		}
		body := extractCaseBody(bodies[idx], outerLabel)
		cases = append(cases, SwitchCase{Value: i, Body: body})
	}

	var defaultBody []Stmt
	if defIdx, ok := labelToIdx[sw.Default]; ok {
		defaultBody = extractCaseBody(bodies[defIdx], outerLabel)
	}

	return &FlatSwitchStmt{
		Value:   sw.Value,
		Cases:   cases,
		Default: defaultBody,
		Offsets: sw.Offsets,
	}
}

func extractCaseBody(stmts []Stmt, outerLabel int) []Stmt {
	result := make([]Stmt, 0, len(stmts))
	for _, stmt := range stmts {
		if br, ok := stmt.(*BreakStmt); ok && br.Label == outerLabel && br.Cond == nil {
			continue
		}
		result = append(result, collapseSwitchInStmt(stmt))
	}
	return result
}

func simplifyStmt(s Stmt) Stmt {
	switch v := s.(type) {
	case *AssignStmt:
		return &AssignStmt{Target: v.Target, Value: Simplify(v.Value), SrcOffset: v.SrcOffset, Offsets: v.Offsets}
	case *StoreStmt:
		return &StoreStmt{Op: v.Op, Addr: Simplify(v.Addr), Value: Simplify(v.Value), Offset: v.Offset, SrcOffset: v.SrcOffset, Offsets: v.Offsets}
	case *ReturnStmt:
		if v.Value != nil {
			return &ReturnStmt{Value: Simplify(v.Value), SrcOffset: v.SrcOffset, Offsets: v.Offsets}
		}
	case *DropStmt:
		return &DropStmt{Value: Simplify(v.Value), SrcOffset: v.SrcOffset, Offsets: v.Offsets}
	case *IfStmt:
		then := make([]Stmt, len(v.Then))
		for i := range v.Then {
			then[i] = simplifyStmt(v.Then[i])
		}
		els := make([]Stmt, len(v.Else))
		for i := range v.Else {
			els[i] = simplifyStmt(v.Else[i])
		}
		cond := Simplify(v.Cond)
		if len(then) == 0 && len(els) > 0 {
			return &IfStmt{Cond: negateCond(cond), Then: els, Else: nil, SrcOffset: v.SrcOffset, EndOffset: v.EndOffset, Offsets: v.Offsets}
		}
		return &IfStmt{Cond: cond, Then: then, Else: els, SrcOffset: v.SrcOffset, EndOffset: v.EndOffset, Offsets: v.Offsets}
	case *LoopStmt:
		body := make([]Stmt, len(v.Body))
		for i := range v.Body {
			body[i] = simplifyStmt(v.Body[i])
		}
		return &LoopStmt{Label: v.Label, Body: body, SrcOffset: v.SrcOffset, EndOffset: v.EndOffset, Offsets: v.Offsets}
	case *BlockStmt:
		body := make([]Stmt, len(v.Body))
		for i := range v.Body {
			body[i] = simplifyStmt(v.Body[i])
		}
		return &BlockStmt{Label: v.Label, Body: body, SrcOffset: v.SrcOffset, EndOffset: v.EndOffset, Offsets: v.Offsets}
	case *BreakStmt:
		if v.Cond != nil {
			return &BreakStmt{Label: v.Label, Cond: Simplify(v.Cond), SrcOffset: v.SrcOffset, Offsets: v.Offsets}
		}
	case *SwitchStmt:
		return &SwitchStmt{Value: Simplify(v.Value), Cases: v.Cases, Default: v.Default, Offsets: v.Offsets}
	case *FlatSwitchStmt:
		cases := make([]SwitchCase, len(v.Cases))
		for i, c := range v.Cases {
			body := make([]Stmt, len(c.Body))
			for j := range c.Body {
				body[j] = simplifyStmt(c.Body[j])
			}
			cases[i] = SwitchCase{Value: c.Value, Body: body}
		}
		def := make([]Stmt, len(v.Default))
		for i := range v.Default {
			def[i] = simplifyStmt(v.Default[i])
		}
		return &FlatSwitchStmt{Value: Simplify(v.Value), Cases: cases, Default: def, Offsets: v.Offsets}
	case *WhileStmt:
		body := make([]Stmt, len(v.Body))
		for i := range v.Body {
			body[i] = simplifyStmt(v.Body[i])
		}
		return &WhileStmt{Cond: Simplify(v.Cond), Body: body, Offsets: v.Offsets}
	}
	return s
}
