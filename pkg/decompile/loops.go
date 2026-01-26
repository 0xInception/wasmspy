package decompile

func RecoverLoops(body *FuncBody) {
	body.Stmts = recoverLoopsInStmts(body.Stmts)
}

func recoverLoopsInStmts(stmts []Stmt) []Stmt {
	result := make([]Stmt, 0, len(stmts))
	for _, stmt := range stmts {
		result = append(result, recoverLoopsInStmt(stmt))
	}
	return result
}

func recoverLoopsInStmt(stmt Stmt) Stmt {
	switch s := stmt.(type) {
	case *BlockStmt:
		if w := tryConvertToWhile(s); w != nil {
			return w
		}
		return &BlockStmt{Label: s.Label, Body: recoverLoopsInStmts(s.Body), SrcOffset: s.SrcOffset, EndOffset: s.EndOffset, Offsets: s.Offsets}

	case *LoopStmt:
		return &LoopStmt{Label: s.Label, Body: recoverLoopsInStmts(s.Body), SrcOffset: s.SrcOffset, EndOffset: s.EndOffset, Offsets: s.Offsets}

	case *IfStmt:
		return &IfStmt{
			Cond:      s.Cond,
			Then:      recoverLoopsInStmts(s.Then),
			Else:      recoverLoopsInStmts(s.Else),
			SrcOffset: s.SrcOffset,
			EndOffset: s.EndOffset,
			Offsets:   s.Offsets,
		}

	case *WhileStmt:
		return &WhileStmt{Cond: s.Cond, Body: recoverLoopsInStmts(s.Body), Offsets: s.Offsets}
	}
	return stmt
}

func tryConvertToWhile(block *BlockStmt) *WhileStmt {
	if len(block.Body) != 1 {
		return nil
	}

	loop, ok := block.Body[0].(*LoopStmt)
	if !ok || len(loop.Body) < 2 {
		return nil
	}

	brk, ok := loop.Body[0].(*BreakStmt)
	if !ok || brk.Cond == nil || brk.Label != block.Label {
		return nil
	}

	lastIdx := len(loop.Body) - 1
	cont, ok := loop.Body[lastIdx].(*BreakStmt)
	if !ok || cont.Cond != nil || cont.Label != loop.Label {
		return nil
	}

	cond := negateCond(brk.Cond)
	body := recoverLoopsInStmts(loop.Body[1:lastIdx])

	body = convertBreaksInLoop(body, loop.Label, block.Label)

	return &WhileStmt{Cond: cond, Body: body, Offsets: block.Offsets}
}

func negateCond(e Expr) Expr {
	switch v := e.(type) {
	case *NotExpr:
		return v.Arg
	case *BinaryExpr:
		switch v.Op {
		case 0x46, 0x51: // i32.eq, i64.eq
			return &BinaryExpr{Op: 0x47, Left: v.Left, Right: v.Right, Type: v.Type} // i32.ne
		case 0x47, 0x52: // i32.ne, i64.ne
			return &BinaryExpr{Op: 0x46, Left: v.Left, Right: v.Right, Type: v.Type} // i32.eq
		case 0x48, 0x53: // i32.lt_s, i64.lt_s
			return &BinaryExpr{Op: 0x4e, Left: v.Left, Right: v.Right, Type: v.Type} // i32.ge_s
		case 0x4a, 0x55: // i32.gt_s, i64.gt_s
			return &BinaryExpr{Op: 0x4c, Left: v.Left, Right: v.Right, Type: v.Type} // i32.le_s
		case 0x4c, 0x57: // i32.le_s, i64.le_s
			return &BinaryExpr{Op: 0x4a, Left: v.Left, Right: v.Right, Type: v.Type} // i32.gt_s
		case 0x4e, 0x59: // i32.ge_s, i64.ge_s
			return &BinaryExpr{Op: 0x48, Left: v.Left, Right: v.Right, Type: v.Type} // i32.lt_s
		case 0x49, 0x54: // i32.lt_u, i64.lt_u
			return &BinaryExpr{Op: 0x4f, Left: v.Left, Right: v.Right, Type: v.Type} // i32.ge_u
		case 0x4b, 0x56: // i32.gt_u, i64.gt_u
			return &BinaryExpr{Op: 0x4d, Left: v.Left, Right: v.Right, Type: v.Type} // i32.le_u
		case 0x4d, 0x58: // i32.le_u, i64.le_u
			return &BinaryExpr{Op: 0x4b, Left: v.Left, Right: v.Right, Type: v.Type} // i32.gt_u
		case 0x4f, 0x5a: // i32.ge_u, i64.ge_u
			return &BinaryExpr{Op: 0x49, Left: v.Left, Right: v.Right, Type: v.Type} // i32.lt_u
		}
	case *UnaryExpr:
		if v.Op == 0x45 || v.Op == 0x50 { // i32.eqz, i64.eqz
			return v.Arg
		}
	}
	return &NotExpr{Arg: e}
}

func convertBreaksInLoop(stmts []Stmt, loopLabel, blockLabel int) []Stmt {
	result := make([]Stmt, 0, len(stmts))
	for _, stmt := range stmts {
		result = append(result, convertBreakInStmt(stmt, loopLabel, blockLabel))
	}
	return result
}

func convertBreakInStmt(stmt Stmt, loopLabel, blockLabel int) Stmt {
	switch s := stmt.(type) {
	case *BreakStmt:
		if s.Label == loopLabel {
			if s.Cond != nil {
				return &IfStmt{Cond: s.Cond, Then: []Stmt{&ContinueStmt{Offsets: s.Offsets}}, Offsets: s.Offsets}
			}
			return &ContinueStmt{Offsets: s.Offsets}
		}
		if s.Label == blockLabel {
			if s.Cond != nil {
				return &IfStmt{Cond: s.Cond, Then: []Stmt{&BreakStmt{Label: 0, Offsets: s.Offsets}}, Offsets: s.Offsets}
			}
			return &BreakStmt{Label: 0, Offsets: s.Offsets}
		}

	case *IfStmt:
		return &IfStmt{
			Cond:      s.Cond,
			Then:      convertBreaksInLoop(s.Then, loopLabel, blockLabel),
			Else:      convertBreaksInLoop(s.Else, loopLabel, blockLabel),
			SrcOffset: s.SrcOffset,
			EndOffset: s.EndOffset,
			Offsets:   s.Offsets,
		}

	case *LoopStmt:
		return &LoopStmt{Label: s.Label, Body: convertBreaksInLoop(s.Body, loopLabel, blockLabel), SrcOffset: s.SrcOffset, EndOffset: s.EndOffset, Offsets: s.Offsets}

	case *BlockStmt:
		return &BlockStmt{Label: s.Label, Body: convertBreaksInLoop(s.Body, loopLabel, blockLabel), SrcOffset: s.SrcOffset, EndOffset: s.EndOffset, Offsets: s.Offsets}

	case *WhileStmt:
		return &WhileStmt{Cond: s.Cond, Body: convertBreaksInLoop(s.Body, loopLabel, blockLabel), Offsets: s.Offsets}
	}
	return stmt
}

func RecoverIfElse(body *FuncBody) {
	body.Stmts = recoverIfElseInStmts(body.Stmts)
}

func recoverIfElseInStmts(stmts []Stmt) []Stmt {
	result := make([]Stmt, 0, len(stmts))
	for _, stmt := range stmts {
		result = append(result, recoverIfElseInStmt(stmt))
	}
	return result
}

func recoverIfElseInStmt(stmt Stmt) Stmt {
	switch s := stmt.(type) {
	case *BlockStmt:
		if converted := tryConvertBlockToIf(s); converted != nil {
			return converted
		}
		return &BlockStmt{Label: s.Label, Body: recoverIfElseInStmts(s.Body), SrcOffset: s.SrcOffset, EndOffset: s.EndOffset, Offsets: s.Offsets}

	case *LoopStmt:
		return &LoopStmt{Label: s.Label, Body: recoverIfElseInStmts(s.Body), SrcOffset: s.SrcOffset, EndOffset: s.EndOffset, Offsets: s.Offsets}

	case *IfStmt:
		return &IfStmt{
			Cond:      s.Cond,
			Then:      recoverIfElseInStmts(s.Then),
			Else:      recoverIfElseInStmts(s.Else),
			SrcOffset: s.SrcOffset,
			EndOffset: s.EndOffset,
			Offsets:   s.Offsets,
		}

	case *WhileStmt:
		return &WhileStmt{Cond: s.Cond, Body: recoverIfElseInStmts(s.Body), Offsets: s.Offsets}
	}
	return stmt
}

func tryConvertBlockToIf(block *BlockStmt) Stmt {
	if len(block.Body) < 2 {
		return nil
	}
	brk, ok := block.Body[0].(*BreakStmt)
	if !ok || brk.Cond == nil || brk.Label != block.Label {
		return nil
	}
	thenBody := recoverIfElseInStmts(block.Body[1:])
	return &IfStmt{
		Cond:      negateCond(brk.Cond),
		Then:      thenBody,
		Else:      nil,
		SrcOffset: block.SrcOffset,
		EndOffset: block.EndOffset,
		Offsets:   block.Offsets,
	}
}
