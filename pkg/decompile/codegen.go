package decompile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

type LineMapping struct {
	Line    int      `json:"line"`
	Offsets []uint64 `json:"offsets"`
}

type DecompileResult struct {
	Code     string        `json:"code"`
	Mappings []LineMapping `json:"mappings"`
}

type codegenCtx struct {
	names     *NameResolver
	numParams int
}

func Decompile(fn *wasm.ResolvedFunction, module *wasm.ResolvedModule) string {
	result := DecompileWithMappings(fn, module)
	return result.Code
}

func DecompileWithMappings(fn *wasm.ResolvedFunction, module *wasm.ResolvedModule) *DecompileResult {
	numParams := 0
	if fn.Type != nil {
		numParams = len(fn.Type.Params)
	}

	ctx := &codegenCtx{
		names:     NewNameResolver(module, uint32(fn.Index)),
		numParams: numParams,
	}

	mc := &mappingCodegen{
		ctx:      ctx,
		line:     1,
		mappings: make(map[int][]uint64),
	}

	var funcStartOffset uint64
	var funcEndOffset uint64
	if fn.Body != nil {
		funcStartOffset = uint64(fn.Body.Offset)
		if len(fn.Body.Instructions) > 0 {
			funcEndOffset = fn.Body.Instructions[len(fn.Body.Instructions)-1].Offset
		}
	}

	mc.writeLineWithOffsets(formatSignature(fn, ctx)+" {", []uint64{funcStartOffset})

	body := BuildStatements(fn, module)
	SimplifyBody(body)
	RecoverLoops(body)
	RecoverIfElse(body)
	CollapseSwitchBlocks(body)
	mc.writeBodyMapped(body, 1)

	mc.writeLineWithOffsets("}", []uint64{funcEndOffset})

	result := &DecompileResult{
		Code:     mc.b.String(),
		Mappings: mc.buildMappings(),
	}
	return result
}

type mappingCodegen struct {
	ctx      *codegenCtx
	b        strings.Builder
	line     int
	mappings map[int][]uint64
}

func (mc *mappingCodegen) writeLine(s string) {
	mc.b.WriteString(s)
	mc.b.WriteString("\n")
	mc.line++
}

func (mc *mappingCodegen) writeLineWithOffsets(s string, offsets []uint64) {
	mc.b.WriteString(s)
	mc.b.WriteString("\n")
	if len(offsets) > 0 {
		mc.mappings[mc.line] = offsets
	}
	mc.line++
}

func (mc *mappingCodegen) buildMappings() []LineMapping {
	var result []LineMapping
	for line, offsets := range mc.mappings {
		sort.Slice(offsets, func(i, j int) bool { return offsets[i] < offsets[j] })
		result = append(result, LineMapping{Line: line, Offsets: offsets})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Line < result[j].Line })
	return result
}

func (mc *mappingCodegen) writeBodyMapped(body *FuncBody, indent int) {
	for _, stmt := range body.Stmts {
		mc.writeStmtMapped(stmt, indent)
	}

	if body.Return != nil {
		prefix := strings.Repeat("  ", indent)
		mc.writeLineWithOffsets(fmt.Sprintf("%sreturn %s", prefix, exprStr(body.Return, mc.ctx)), body.ReturnOffsets)
	}
}

func (mc *mappingCodegen) writeStmtMapped(stmt Stmt, indent int) {
	prefix := strings.Repeat("  ", indent)

	switch s := stmt.(type) {
	case *AssignStmt:
		mc.writeLineWithOffsets(fmt.Sprintf("%s%s = %s", prefix, exprStr(s.Target, mc.ctx), exprStr(s.Value, mc.ctx)), s.Offsets)

	case *StoreStmt:
		addr := exprStr(s.Addr, mc.ctx)
		if s.Offset > 0 {
			mc.writeLineWithOffsets(fmt.Sprintf("%smem[%s + %d] = %s", prefix, addr, s.Offset, exprStr(s.Value, mc.ctx)), s.Offsets)
		} else {
			mc.writeLineWithOffsets(fmt.Sprintf("%smem[%s] = %s", prefix, addr, exprStr(s.Value, mc.ctx)), s.Offsets)
		}

	case *CallStmt:
		mc.writeLineWithOffsets(fmt.Sprintf("%s%s", prefix, callStr(s.Call, mc.ctx)), s.Offsets)

	case *ReturnStmt:
		if s.Value != nil {
			mc.writeLineWithOffsets(fmt.Sprintf("%sreturn %s", prefix, exprStr(s.Value, mc.ctx)), s.Offsets)
		} else {
			mc.writeLineWithOffsets(fmt.Sprintf("%sreturn", prefix), s.Offsets)
		}

	case *DropStmt:
		mc.writeLineWithOffsets(fmt.Sprintf("%s_ = %s", prefix, exprStr(s.Value, mc.ctx)), s.Offsets)

	case *IfStmt:
		cond := "..."
		if s.Cond != nil {
			cond = exprStr(s.Cond, mc.ctx)
		}
		mc.writeLineWithOffsets(fmt.Sprintf("%sif %s {", prefix, cond), []uint64{s.SrcOffset})
		for _, inner := range s.Then {
			mc.writeStmtMapped(inner, indent+1)
		}
		if len(s.Else) > 0 {
			mc.writeLineWithOffsets(fmt.Sprintf("%s} else {", prefix), []uint64{s.SrcOffset})
			for _, inner := range s.Else {
				mc.writeStmtMapped(inner, indent+1)
			}
		}
		mc.writeLineWithOffsets(fmt.Sprintf("%s}", prefix), []uint64{s.EndOffset})

	case *LoopStmt:
		mc.writeLineWithOffsets(fmt.Sprintf("%sloop L%d {", prefix, s.Label), []uint64{s.SrcOffset})
		for _, inner := range s.Body {
			mc.writeStmtMapped(inner, indent+1)
		}
		mc.writeLineWithOffsets(fmt.Sprintf("%s}", prefix), []uint64{s.EndOffset})

	case *BlockStmt:
		mc.writeLineWithOffsets(fmt.Sprintf("%sblock L%d {", prefix, s.Label), []uint64{s.SrcOffset})
		for _, inner := range s.Body {
			mc.writeStmtMapped(inner, indent+1)
		}
		mc.writeLineWithOffsets(fmt.Sprintf("%s}", prefix), []uint64{s.EndOffset})

	case *BreakStmt:
		if s.Cond != nil {
			mc.writeLineWithOffsets(fmt.Sprintf("%sif %s break L%d", prefix, exprStr(s.Cond, mc.ctx), s.Label), s.Offsets)
		} else {
			mc.writeLineWithOffsets(fmt.Sprintf("%sbreak L%d", prefix, s.Label), s.Offsets)
		}

	case *SwitchStmt:
		mc.writeLineWithOffsets(fmt.Sprintf("%sswitch %s {", prefix, exprStr(s.Value, mc.ctx)), s.Offsets)
		for i, label := range s.Cases {
			mc.writeLineWithOffsets(fmt.Sprintf("%s  case %d: break L%d", prefix, i, label), offsetsWithSubIndex(s.Offsets, i+1))
		}
		mc.writeLineWithOffsets(fmt.Sprintf("%s  default: break L%d", prefix, s.Default), offsetsWithSubIndex(s.Offsets, len(s.Cases)+1))
		mc.writeLineWithOffsets(fmt.Sprintf("%s}", prefix), offsetsWithSubIndex(s.Offsets, len(s.Cases)+2))

	case *FlatSwitchStmt:
		mc.writeLineWithOffsets(fmt.Sprintf("%sswitch %s {", prefix, exprStr(s.Value, mc.ctx)), s.Offsets)
		subIdx := 1
		for _, c := range s.Cases {
			mc.writeLineWithOffsets(fmt.Sprintf("%scase %d:", prefix, c.Value), offsetsWithSubIndex(s.Offsets, subIdx))
			subIdx++
			for _, inner := range c.Body {
				mc.writeStmtMapped(inner, indent+1)
			}
		}
		if len(s.Default) > 0 {
			mc.writeLineWithOffsets(fmt.Sprintf("%sdefault:", prefix), offsetsWithSubIndex(s.Offsets, subIdx))
			subIdx++
			for _, inner := range s.Default {
				mc.writeStmtMapped(inner, indent+1)
			}
		}
		mc.writeLineWithOffsets(fmt.Sprintf("%s}", prefix), offsetsWithSubIndex(s.Offsets, subIdx))

	case *WhileStmt:
		mc.writeLineWithOffsets(fmt.Sprintf("%swhile %s {", prefix, exprStr(s.Cond, mc.ctx)), s.Offsets)
		for _, inner := range s.Body {
			mc.writeStmtMapped(inner, indent+1)
		}
		mc.writeLineWithOffsets(fmt.Sprintf("%s}", prefix), s.Offsets)

	case *ContinueStmt:
		mc.writeLineWithOffsets(fmt.Sprintf("%scontinue", prefix), s.Offsets)

	case *ErrorStmt:
		mc.writeLine(fmt.Sprintf("%s// ERROR at 0x%x: %s", prefix, s.Offset, s.Message))
	}
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

type FunctionError struct {
	FuncIndex uint32
	FuncName  string
	Errors    []DecompileError
}

type ModuleErrors struct {
	Functions    []FunctionError
	TotalErrors  int
	UniqueErrors map[string]int // opcode -> count
}

func CollectErrors(module *wasm.ResolvedModule) *ModuleErrors {
	result := &ModuleErrors{
		UniqueErrors: make(map[string]int),
	}

	for i, fn := range module.Functions {
		if fn.Imported {
			continue
		}
		body := BuildStatements(&module.Functions[i], module)
		if len(body.Errors) > 0 {
			funcErr := FunctionError{
				FuncIndex: fn.Index,
				FuncName:  fn.Name,
				Errors:    body.Errors,
			}
			result.Functions = append(result.Functions, funcErr)
			result.TotalErrors += len(body.Errors)
			for _, e := range body.Errors {
				key := e.Opcode
				if key == "" {
					key = e.Message
				}
				result.UniqueErrors[key]++
			}
		}
	}

	return result
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

	case *ErrorStmt:
		b.WriteString(fmt.Sprintf("%s// ERROR at 0x%x: %s\n", prefix, s.Offset, s.Message))
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
	case *ErrorExpr:
		return fmt.Sprintf("/* ERROR at 0x%x: %s */", v.Offset, v.Message)
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

func offsetsWithSubIndex(offsets []uint64, subIndex int) []uint64 {
	if len(offsets) == 0 {
		return nil
	}
	result := make([]uint64, len(offsets))
	for i, offset := range offsets {
		// Encode sub-index in upper 16 bits (bits 48-63)
		// This preserves the original offset in the lower 48 bits
		result[i] = offset | (uint64(subIndex) << 48)
	}
	return result
}
