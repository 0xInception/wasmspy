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
		return &BlockStmt{Label: s.Label, Body: recoverLoopsInStmts(s.Body)}

	case *LoopStmt:
		return &LoopStmt{Label: s.Label, Body: recoverLoopsInStmts(s.Body)}

	case *IfStmt:
		return &IfStmt{
			Cond: s.Cond,
			Then: recoverLoopsInStmts(s.Then),
			Else: recoverLoopsInStmts(s.Else),
		}

	case *WhileStmt:
		return &WhileStmt{Cond: s.Cond, Body: recoverLoopsInStmts(s.Body)}
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

	return &WhileStmt{Cond: cond, Body: body}
}

func negateCond(e Expr) Expr {
	switch v := e.(type) {
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
		}
	case *UnaryExpr:
		if v.Op == 0x45 || v.Op == 0x50 { // i32.eqz, i64.eqz
			return v.Arg
		}
	}
	return &UnaryExpr{Op: 0x45, Arg: e, Type: e.(*BinaryExpr).Type} // i32.eqz as fallback
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
				return &IfStmt{Cond: s.Cond, Then: []Stmt{&ContinueStmt{}}}
			}
			return &ContinueStmt{}
		}
		if s.Label == blockLabel {
			if s.Cond != nil {
				return &IfStmt{Cond: s.Cond, Then: []Stmt{&BreakStmt{Label: 0}}}
			}
			return &BreakStmt{Label: 0}
		}

	case *IfStmt:
		return &IfStmt{
			Cond: s.Cond,
			Then: convertBreaksInLoop(s.Then, loopLabel, blockLabel),
			Else: convertBreaksInLoop(s.Else, loopLabel, blockLabel),
		}

	case *LoopStmt:
		return &LoopStmt{Label: s.Label, Body: convertBreaksInLoop(s.Body, loopLabel, blockLabel)}

	case *BlockStmt:
		return &BlockStmt{Label: s.Label, Body: convertBreaksInLoop(s.Body, loopLabel, blockLabel)}

	case *WhileStmt:
		return &WhileStmt{Cond: s.Cond, Body: convertBreaksInLoop(s.Body, loopLabel, blockLabel)}
	}
	return stmt
}
