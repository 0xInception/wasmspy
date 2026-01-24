package decompile

import (
	"fmt"
	"strings"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

type codegenCtx struct {
	names     *NameResolver
	numParams int
}

func Decompile(fn *wasm.ResolvedFunction, module *wasm.ResolvedModule) string {
	var b strings.Builder

	numParams := 0
	if fn.Type != nil {
		numParams = len(fn.Type.Params)
	}

	ctx := &codegenCtx{
		names:     NewNameResolver(module, uint32(fn.Index)),
		numParams: numParams,
	}

	b.WriteString(formatSignature(fn, ctx))
	b.WriteString(" {\n")

	body := BuildStatements(fn, module)
	SimplifyBody(body)
	RecoverLoops(body)
	CollapseSwitchBlocks(body)
	writeBody(&b, body, 1, ctx)

	b.WriteString("}\n")

	return b.String()
}

func DecompileModule(module *wasm.ResolvedModule) string {
	var b strings.Builder

	for i, fn := range module.Functions {
		if fn.Imported {
			continue
		}
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(Decompile(&module.Functions[i], module))
	}

	return b.String()
}

func formatSignature(fn *wasm.ResolvedFunction, ctx *codegenCtx) string {
	name := fn.Name
	if name == "" {
		name = fmt.Sprintf("func_%d", fn.Index)
	}

	params := ""
	if fn.Type != nil {
		for i, p := range fn.Type.Params {
			if i > 0 {
				params += ", "
			}
			pname := ctx.names.Local(uint32(i), ctx.numParams)
			params += fmt.Sprintf("%s %s", p.String(), pname)
		}
	}

	ret := ""
	if fn.Type != nil && len(fn.Type.Results) > 0 {
		types := make([]string, len(fn.Type.Results))
		for i, r := range fn.Type.Results {
			types[i] = r.String()
		}
		ret = " -> " + strings.Join(types, ", ")
	}

	return fmt.Sprintf("func %s(%s)%s", name, params, ret)
}

func writeBody(b *strings.Builder, body *FuncBody, indent int, ctx *codegenCtx) {
	prefix := strings.Repeat("  ", indent)

	for _, stmt := range body.Stmts {
		writeStmt(b, stmt, indent, ctx)
	}

	if body.Return != nil {
		b.WriteString(fmt.Sprintf("%sreturn %s\n", prefix, exprStr(body.Return, ctx)))
	}
}

func writeStmt(b *strings.Builder, stmt Stmt, indent int, ctx *codegenCtx) {
	prefix := strings.Repeat("  ", indent)

	switch s := stmt.(type) {
	case *AssignStmt:
		b.WriteString(fmt.Sprintf("%s%s = %s\n", prefix, exprStr(s.Target, ctx), exprStr(s.Value, ctx)))

	case *StoreStmt:
		addr := exprStr(s.Addr, ctx)
		if s.Offset > 0 {
			b.WriteString(fmt.Sprintf("%smem[%s + %d] = %s\n", prefix, addr, s.Offset, exprStr(s.Value, ctx)))
		} else {
			b.WriteString(fmt.Sprintf("%smem[%s] = %s\n", prefix, addr, exprStr(s.Value, ctx)))
		}

	case *CallStmt:
		b.WriteString(fmt.Sprintf("%s%s\n", prefix, callStr(s.Call, ctx)))

	case *ReturnStmt:
		if s.Value != nil {
			b.WriteString(fmt.Sprintf("%sreturn %s\n", prefix, exprStr(s.Value, ctx)))
		} else {
			b.WriteString(fmt.Sprintf("%sreturn\n", prefix))
		}

	case *DropStmt:
		b.WriteString(fmt.Sprintf("%s_ = %s\n", prefix, exprStr(s.Value, ctx)))

	case *IfStmt:
		cond := "..."
		if s.Cond != nil {
			cond = exprStr(s.Cond, ctx)
		}
		b.WriteString(fmt.Sprintf("%sif %s {\n", prefix, cond))
		for _, inner := range s.Then {
			writeStmt(b, inner, indent+1, ctx)
		}
		if len(s.Else) > 0 {
			b.WriteString(fmt.Sprintf("%s} else {\n", prefix))
			for _, inner := range s.Else {
				writeStmt(b, inner, indent+1, ctx)
			}
		}
		b.WriteString(fmt.Sprintf("%s}\n", prefix))

	case *LoopStmt:
		b.WriteString(fmt.Sprintf("%sloop L%d {\n", prefix, s.Label))
		for _, inner := range s.Body {
			writeStmt(b, inner, indent+1, ctx)
		}
		b.WriteString(fmt.Sprintf("%s}\n", prefix))

	case *BlockStmt:
		b.WriteString(fmt.Sprintf("%sblock L%d {\n", prefix, s.Label))
		for _, inner := range s.Body {
			writeStmt(b, inner, indent+1, ctx)
		}
		b.WriteString(fmt.Sprintf("%s}\n", prefix))

	case *BreakStmt:
		if s.Cond != nil {
			b.WriteString(fmt.Sprintf("%sif %s break L%d\n", prefix, exprStr(s.Cond, ctx), s.Label))
		} else {
			b.WriteString(fmt.Sprintf("%sbreak L%d\n", prefix, s.Label))
		}

	case *SwitchStmt:
		b.WriteString(fmt.Sprintf("%sswitch %s {\n", prefix, exprStr(s.Value, ctx)))
		for i, label := range s.Cases {
			b.WriteString(fmt.Sprintf("%s  case %d: break L%d\n", prefix, i, label))
		}
		b.WriteString(fmt.Sprintf("%s  default: break L%d\n", prefix, s.Default))
		b.WriteString(fmt.Sprintf("%s}\n", prefix))

	case *FlatSwitchStmt:
		b.WriteString(fmt.Sprintf("%sswitch %s {\n", prefix, exprStr(s.Value, ctx)))
		for _, c := range s.Cases {
			b.WriteString(fmt.Sprintf("%scase %d:\n", prefix, c.Value))
			for _, inner := range c.Body {
				writeStmt(b, inner, indent+1, ctx)
			}
		}
		if len(s.Default) > 0 {
			b.WriteString(fmt.Sprintf("%sdefault:\n", prefix))
			for _, inner := range s.Default {
				writeStmt(b, inner, indent+1, ctx)
			}
		}
		b.WriteString(fmt.Sprintf("%s}\n", prefix))

	case *WhileStmt:
		b.WriteString(fmt.Sprintf("%swhile %s {\n", prefix, exprStr(s.Cond, ctx)))
		for _, inner := range s.Body {
			writeStmt(b, inner, indent+1, ctx)
		}
		b.WriteString(fmt.Sprintf("%s}\n", prefix))

	case *ContinueStmt:
		b.WriteString(fmt.Sprintf("%scontinue\n", prefix))
	}
}

func exprStr(e Expr, ctx *codegenCtx) string {
	if e == nil {
		return "?"
	}
	switch v := e.(type) {
	case *LocalExpr:
		return ctx.names.Local(v.Index, ctx.numParams)
	case *ParamExpr:
		return ctx.names.Local(v.Index, ctx.numParams)
	case *GlobalExpr:
		return ctx.names.Global(v.Index)
	case *ConstExpr:
		return fmt.Sprintf("%v", v.Value)
	case *BinaryExpr:
		return fmt.Sprintf("(%s %s %s)", exprStr(v.Left, ctx), opSymbol(v.Op), exprStr(v.Right, ctx))
	case *UnaryExpr:
		return fmt.Sprintf("%s(%s)", opName(v.Op), exprStr(v.Arg, ctx))
	case *CallExpr:
		return callStr(v, ctx)
	case *LoadExpr:
		addr := exprStr(v.Addr, ctx)
		if v.Offset > 0 {
			return fmt.Sprintf("mem[%s + %d]", addr, v.Offset)
		}
		return fmt.Sprintf("mem[%s]", addr)
	case *TernaryExpr:
		return fmt.Sprintf("(%s ? %s : %s)", exprStr(v.Cond, ctx), exprStr(v.ThenResult, ctx), exprStr(v.ElseResult, ctx))
	case *NegExpr:
		return fmt.Sprintf("-%s", exprStr(v.Arg, ctx))
	case *NotExpr:
		return fmt.Sprintf("!(%s)", exprStr(v.Arg, ctx))
	}
	return "?"
}

func callStr(c *CallExpr, ctx *codegenCtx) string {
	name := ctx.names.Func(c.FuncIndex)
	if c.FuncIndex == 0xFFFFFFFF {
		name = "indirect"
	}
	args := ""
	for i, arg := range c.Args {
		if i > 0 {
			args += ", "
		}
		args += exprStr(arg, ctx)
	}
	return fmt.Sprintf("%s(%s)", name, args)
}
