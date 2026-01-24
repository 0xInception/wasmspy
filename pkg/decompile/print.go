package decompile

import (
	"fmt"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func (e *LocalExpr) String() string {
	return fmt.Sprintf("v%d", e.Index)
}

func (e *ParamExpr) String() string {
	return fmt.Sprintf("p%d", e.Index)
}

func (e *GlobalExpr) String() string {
	return fmt.Sprintf("global%d", e.Index)
}

func (e *ConstExpr) String() string {
	return fmt.Sprintf("%v", e.Value)
}

func (e *BinaryExpr) String() string {
	op := opSymbol(e.Op)
	left := exprString(e.Left)
	right := exprString(e.Right)
	return fmt.Sprintf("(%s %s %s)", left, op, right)
}

func (e *UnaryExpr) String() string {
	op := opName(e.Op)
	arg := exprString(e.Arg)
	return fmt.Sprintf("%s(%s)", op, arg)
}

func (e *CallExpr) String() string {
	name := e.FuncName
	if name == "" {
		if e.FuncIndex == 0xFFFFFFFF {
			name = "indirect"
		} else {
			name = fmt.Sprintf("func%d", e.FuncIndex)
		}
	}
	args := ""
	for i, arg := range e.Args {
		if i > 0 {
			args += ", "
		}
		args += exprString(arg)
	}
	return fmt.Sprintf("%s(%s)", name, args)
}

func (e *LoadExpr) String() string {
	addr := exprString(e.Addr)
	if e.Offset > 0 {
		return fmt.Sprintf("mem[%s + %d]", addr, e.Offset)
	}
	return fmt.Sprintf("mem[%s]", addr)
}

func (e *TernaryExpr) String() string {
	return fmt.Sprintf("(%s ? %s : %s)", exprString(e.Cond), exprString(e.ThenResult), exprString(e.ElseResult))
}

func (e *NegExpr) String() string {
	return fmt.Sprintf("-%s", exprString(e.Arg))
}

func (s *AssignStmt) String() string {
	return fmt.Sprintf("%s = %s", exprString(s.Target), exprString(s.Value))
}

func (s *StoreStmt) String() string {
	addr := exprString(s.Addr)
	val := exprString(s.Value)
	if s.Offset > 0 {
		return fmt.Sprintf("mem[%s + %d] = %s", addr, s.Offset, val)
	}
	return fmt.Sprintf("mem[%s] = %s", addr, val)
}

func (s *CallStmt) String() string {
	return s.Call.String()
}

func (s *ReturnStmt) String() string {
	if s.Value != nil {
		return fmt.Sprintf("return %s", exprString(s.Value))
	}
	return "return"
}

func (s *DropStmt) String() string {
	return fmt.Sprintf("_ = %s", exprString(s.Value))
}

func (s *IfStmt) String() string {
	return "if { ... }"
}

func (s *LoopStmt) String() string {
	return fmt.Sprintf("loop L%d { ... }", s.Label)
}

func (s *BlockStmt) String() string {
	return fmt.Sprintf("block L%d { ... }", s.Label)
}

func (s *BreakStmt) String() string {
	if s.Cond != nil {
		return fmt.Sprintf("if %s break L%d", exprString(s.Cond), s.Label)
	}
	return fmt.Sprintf("break L%d", s.Label)
}

func (s *SwitchStmt) String() string {
	return fmt.Sprintf("switch %s { ... }", exprString(s.Value))
}

func (s *WhileStmt) String() string {
	return fmt.Sprintf("while %s { ... }", exprString(s.Cond))
}

func (s *ContinueStmt) String() string {
	return "continue"
}

func stmtString(s Stmt) string {
	switch v := s.(type) {
	case *AssignStmt:
		return v.String()
	case *StoreStmt:
		return v.String()
	case *CallStmt:
		return v.String()
	case *ReturnStmt:
		return v.String()
	case *DropStmt:
		return v.String()
	case *IfStmt:
		return v.String()
	case *LoopStmt:
		return v.String()
	case *BlockStmt:
		return v.String()
	case *BreakStmt:
		return v.String()
	case *SwitchStmt:
		return v.String()
	case *WhileStmt:
		return v.String()
	case *ContinueStmt:
		return v.String()
	}
	return "?"
}

func exprString(e Expr) string {
	if e == nil {
		return "?"
	}
	switch v := e.(type) {
	case *LocalExpr:
		return v.String()
	case *ParamExpr:
		return v.String()
	case *GlobalExpr:
		return v.String()
	case *ConstExpr:
		return v.String()
	case *BinaryExpr:
		return v.String()
	case *UnaryExpr:
		return v.String()
	case *CallExpr:
		return v.String()
	case *LoadExpr:
		return v.String()
	case *TernaryExpr:
		return v.String()
	case *NegExpr:
		return v.String()
	}
	return "?"
}

func opSymbol(op wasm.Opcode) string {
	switch op {
	case wasm.OpI32Add, wasm.OpI64Add:
		return "+"
	case wasm.OpI32Sub, wasm.OpI64Sub:
		return "-"
	case wasm.OpI32Mul, wasm.OpI64Mul:
		return "*"
	case wasm.OpI32DivS, wasm.OpI32DivU, wasm.OpI64DivS, wasm.OpI64DivU:
		return "/"
	case wasm.OpI32RemS, wasm.OpI32RemU, wasm.OpI64RemS, wasm.OpI64RemU:
		return "%"
	case wasm.OpI32And, wasm.OpI64And:
		return "&"
	case wasm.OpI32Or, wasm.OpI64Or:
		return "|"
	case wasm.OpI32Xor, wasm.OpI64Xor:
		return "^"
	case wasm.OpI32Shl, wasm.OpI64Shl:
		return "<<"
	case wasm.OpI32ShrS, wasm.OpI32ShrU, wasm.OpI64ShrS, wasm.OpI64ShrU:
		return ">>"
	case wasm.OpI32Eq, wasm.OpI64Eq, wasm.OpF64Eq:
		return "=="
	case wasm.OpI32Ne, wasm.OpI64Ne, wasm.OpF64Ne:
		return "!="
	case wasm.OpI32LtS, wasm.OpI32LtU, wasm.OpI64LtS, wasm.OpI64LtU, wasm.OpF64Lt:
		return "<"
	case wasm.OpI32GtS, wasm.OpI32GtU, wasm.OpI64GtS, wasm.OpI64GtU, wasm.OpF64Gt:
		return ">"
	case wasm.OpI32LeS, wasm.OpI32LeU, wasm.OpI64LeS, wasm.OpI64LeU, wasm.OpF64Le:
		return "<="
	case wasm.OpI32GeS, wasm.OpI32GeU, wasm.OpI64GeS, wasm.OpI64GeU, wasm.OpF64Ge:
		return ">="
	}
	return wasm.OpcodeNames[op]
}

func opName(op wasm.Opcode) string {
	switch op {
	case wasm.OpI32Eqz, wasm.OpI64Eqz:
		return "!"
	case wasm.OpI32Clz, wasm.OpI64Clz:
		return "clz"
	case wasm.OpI32Ctz, wasm.OpI64Ctz:
		return "ctz"
	case wasm.OpI32Popcnt, wasm.OpI64Popcnt:
		return "popcnt"
	case wasm.OpI32WrapI64:
		return "i32"
	case wasm.OpI64ExtendI32S, wasm.OpI64ExtendI32U:
		return "i64"
	case wasm.OpI32TruncSatF32S, wasm.OpI32TruncSatF32U,
		wasm.OpI32TruncSatF64S, wasm.OpI32TruncSatF64U:
		return "i32_trunc_sat"
	case wasm.OpI64TruncSatF32S, wasm.OpI64TruncSatF32U,
		wasm.OpI64TruncSatF64S, wasm.OpI64TruncSatF64U:
		return "i64_trunc_sat"
	case wasm.OpI64ReinterpretF64:
		return "i64_reinterpret"
	case wasm.OpF64ReinterpretI64:
		return "f64_reinterpret"
	}
	return wasm.OpcodeNames[op]
}
