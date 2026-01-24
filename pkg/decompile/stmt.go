package decompile

import (
	"fmt"
	"strings"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

type Block struct {
	Kind       BlockKind
	Label      int
	Stmts      []Stmt
	Else       []Stmt
	Cond       Expr
	Result     wasm.ValType
	ThenResult Expr
	StackDepth int
}

type BlockKind int

const (
	BlockPlain BlockKind = iota
	BlockLoop
	BlockIf
)

type IfStmt struct {
	Cond Expr
	Then []Stmt
	Else []Stmt
}

type LoopStmt struct {
	Label int
	Body  []Stmt
}

type BlockStmt struct {
	Label int
	Body  []Stmt
}

type BreakStmt struct {
	Label int
	Cond  Expr
}

func (*IfStmt) node()    {}
func (*LoopStmt) node()  {}
func (*BlockStmt) node() {}
func (*BreakStmt) node() {}

func (*IfStmt) stmt()    {}
func (*LoopStmt) stmt()  {}
func (*BlockStmt) stmt() {}
func (*BreakStmt) stmt() {}

type FuncBody struct {
	Stmts  []Stmt
	Return Expr
	Errors []DecompileError
}

type DecompileError struct {
	Offset  uint64
	Opcode  string
	Message string
}

func BuildStatements(fn *wasm.ResolvedFunction, module *wasm.ResolvedModule) *FuncBody {
	if fn.Body == nil {
		return &FuncBody{}
	}

	b := &stmtBuilder{
		fn:     fn,
		module: module,
		locals: buildLocals(fn),
	}

	return b.build()
}

type stmtBuilder struct {
	fn          *wasm.ResolvedFunction
	module      *wasm.ResolvedModule
	locals      []*Value
	stack       []*Value
	blocks      []*Block
	stmts       []Stmt
	labelID     int
	currInstr   *wasm.Instruction
	unreachable bool
	errors      []DecompileError
}

func (b *stmtBuilder) build() *FuncBody {
	for i := range b.fn.Body.Instructions {
		instr := &b.fn.Body.Instructions[i]
		b.currInstr = instr
		b.processInstr(instr)
	}

	var ret Expr
	if len(b.stack) > 0 {
		ret = ValueToExpr(b.stack[0])
	}

	return &FuncBody{
		Stmts:  b.stmts,
		Return: ret,
		Errors: b.errors,
	}
}

func (b *stmtBuilder) recordError(offset uint64, opcode, message string) {
	b.errors = append(b.errors, DecompileError{
		Offset:  offset,
		Opcode:  opcode,
		Message: message,
	})
}

func (b *stmtBuilder) processInstr(instr *wasm.Instruction) {
	switch instr.Opcode {
	case wasm.OpBlock:
		b.labelID++
		result := parseBlockType(instr)
		b.blocks = append(b.blocks, &Block{
			Kind:       BlockPlain,
			Label:      b.labelID,
			Result:     result,
			StackDepth: len(b.stack),
		})
		b.unreachable = false

	case wasm.OpLoop:
		b.labelID++
		result := parseBlockType(instr)
		b.blocks = append(b.blocks, &Block{
			Kind:       BlockLoop,
			Label:      b.labelID,
			Result:     result,
			StackDepth: len(b.stack),
		})
		b.unreachable = false

	case wasm.OpIf:
		b.labelID++
		result := parseBlockType(instr)
		cond := b.pop()
		b.blocks = append(b.blocks, &Block{
			Kind:       BlockIf,
			Label:      b.labelID,
			Cond:       ValueToExpr(cond),
			Stmts:      []Stmt{},
			Result:     result,
			StackDepth: len(b.stack),
		})
		b.unreachable = false

	case wasm.OpElse:
		if len(b.blocks) > 0 {
			block := b.blocks[len(b.blocks)-1]
			if block.Result != 0 && len(b.stack) > block.StackDepth {
				block.ThenResult = ValueToExpr(b.pop())
			}
			block.Else = block.Stmts
			block.Stmts = []Stmt{}
			b.unreachable = false
			b.stack = b.stack[:block.StackDepth]
		}

	case wasm.OpEnd:
		if len(b.blocks) > 0 {
			block := b.blocks[len(b.blocks)-1]
			b.blocks = b.blocks[:len(b.blocks)-1]

			b.stack = b.stack[:block.StackDepth]

			if block.Kind == BlockIf && block.Result != 0 {
				b.push(&Value{
					Type:   block.Result,
					Source: SourceOp,
					Op: &OpValue{
						Instr:   nil,
						Inputs:  nil,
						Ternary: &TernaryValue{
							Cond:       block.Cond,
							ThenResult: block.ThenResult,
							ElseResult: nil,
						},
					},
				})
			} else if block.Result != 0 {
				b.push(&Value{
					Type:   block.Result,
					Source: SourceConst,
					Const:  int32(0),
				})
			}

			b.unreachable = false

			var stmt Stmt
			switch block.Kind {
			case BlockPlain:
				if len(block.Stmts) > 0 {
					stmt = &BlockStmt{Label: block.Label, Body: block.Stmts}
				}
			case BlockLoop:
				stmt = &LoopStmt{Label: block.Label, Body: block.Stmts}
			case BlockIf:
				stmt = &IfStmt{Cond: block.Cond, Then: block.Else, Else: block.Stmts}
			}

			if stmt != nil {
				b.emit(stmt)
			}
		}

	case wasm.OpUnreachable:
		b.unreachable = true

	case wasm.OpBr:
		label := getU32(instr.Immediates, 0)
		target := b.getBlockLabel(int(label))
		b.emit(&BreakStmt{Label: target})
		b.unreachable = true

	case wasm.OpBrIf:
		label := getU32(instr.Immediates, 0)
		cond := b.pop()
		target := b.getBlockLabel(int(label))
		b.emit(&BreakStmt{Label: target, Cond: ValueToExpr(cond)})

	case wasm.OpBrTable:
		idx := b.pop()
		labels, ok := instr.Immediates[0].([]uint32)
		if ok && len(labels) > 0 {
			cases := make([]int, len(labels)-1)
			for i := 0; i < len(labels)-1; i++ {
				cases[i] = b.getBlockLabel(int(labels[i]))
			}
			def := b.getBlockLabel(int(labels[len(labels)-1]))
			b.emit(&SwitchStmt{Value: ValueToExpr(idx), Cases: cases, Default: def})
		}
		b.unreachable = true

	case wasm.OpLocalSet:
		idx := getU32(instr.Immediates, 0)
		val := b.pop()
		b.locals[idx] = val
		b.emit(&AssignStmt{
			Target: &LocalExpr{Index: idx, Type: val.Type},
			Value:  ValueToExpr(val),
		})

	case wasm.OpLocalTee:
		idx := getU32(instr.Immediates, 0)
		if len(b.stack) > 0 {
			val := b.stack[len(b.stack)-1]
			b.locals[idx] = val
			b.emit(&AssignStmt{
				Target: &LocalExpr{Index: idx, Type: val.Type},
				Value:  ValueToExpr(val),
			})
		}

	case wasm.OpGlobalSet:
		idx := getU32(instr.Immediates, 0)
		val := b.pop()
		var t wasm.ValType = wasm.ValI32
		if val != nil {
			t = val.Type
		}
		b.emit(&AssignStmt{
			Target: &GlobalExpr{Index: idx, Type: t},
			Value:  ValueToExpr(val),
		})

	case wasm.OpI32Store, wasm.OpI64Store, wasm.OpF32Store, wasm.OpF64Store,
		wasm.OpI32Store8, wasm.OpI32Store16, wasm.OpI64Store8, wasm.OpI64Store16, wasm.OpI64Store32:
		val := b.pop()
		addr := b.pop()
		var offset uint32
		if len(instr.Immediates) >= 2 {
			offset = getU32(instr.Immediates, 1)
		}
		b.emit(&StoreStmt{
			Op:     instr.Opcode,
			Addr:   ValueToExpr(addr),
			Value:  ValueToExpr(val),
			Offset: offset,
		})

	case wasm.OpDrop:
		val := b.pop()
		b.emit(&DropStmt{Value: ValueToExpr(val)})

	case wasm.OpSelect:
		cond := b.pop()
		val2 := b.pop()
		val1 := b.pop()
		resultType := wasm.ValI32
		if val1 != nil {
			resultType = val1.Type
		} else if val2 != nil {
			resultType = val2.Type
		}
		b.push(&Value{
			Type:   resultType,
			Source: SourceOp,
			Op: &OpValue{
				Instr:  instr,
				Inputs: []*Value{val1, val2, cond},
				Ternary: &TernaryValue{
					Cond:       ValueToExpr(cond),
					ThenResult: ValueToExpr(val1),
					ElseResult: ValueToExpr(val2),
				},
			},
		})

	case wasm.OpReturn:
		var val Expr
		if len(b.stack) > 0 {
			val = ValueToExpr(b.pop())
		}
		b.emit(&ReturnStmt{Value: val})
		b.unreachable = true

	case wasm.OpCall:
		idx := getU32(instr.Immediates, 0)
		var sig *wasm.FuncType
		var name string
		if b.module != nil {
			if f := b.module.GetFunction(idx); f != nil {
				if f.Type != nil {
					sig = f.Type
				}
				name = f.Name
			}
		}
		if sig != nil {
			args := make([]Expr, len(sig.Params))
			for i := len(sig.Params) - 1; i >= 0; i-- {
				args[i] = ValueToExpr(b.pop())
			}
			call := &CallExpr{FuncIndex: idx, FuncName: name, Args: args}
			if len(sig.Results) > 0 {
				b.push(&Value{
					Type:   sig.Results[0],
					Source: SourceOp,
					Op:     &OpValue{Instr: instr, Inputs: nil},
				})
			} else {
				b.emit(&CallStmt{Call: call})
			}
		}

	case wasm.OpCallIndirect:
		typeIdx := getU32(instr.Immediates, 0)
		funcIdx := b.pop()
		var sig *wasm.FuncType
		if b.module != nil && int(typeIdx) < len(b.module.Types) {
			sig = &b.module.Types[typeIdx]
		}
		if sig != nil {
			args := make([]Expr, len(sig.Params)+1)
			args[0] = ValueToExpr(funcIdx)
			for i := len(sig.Params) - 1; i >= 0; i-- {
				args[i+1] = ValueToExpr(b.pop())
			}
			call := &CallExpr{FuncIndex: 0xFFFFFFFF, Args: args}
			if len(sig.Results) > 0 {
				b.push(&Value{
					Type:   sig.Results[0],
					Source: SourceOp,
					Op:     &OpValue{Instr: instr, Inputs: nil},
				})
			} else {
				b.emit(&CallStmt{Call: call})
			}
		} else {
			b.emit(&CallStmt{Call: &CallExpr{FuncIndex: 0xFFFFFFFF, Args: []Expr{ValueToExpr(funcIdx)}}})
		}

	default:
		b.simulateOp(instr)
	}
}

func (b *stmtBuilder) simulateOp(instr *wasm.Instruction) {
	switch instr.Opcode {
	case wasm.OpLocalGet:
		idx := getU32(instr.Immediates, 0)
		if int(idx) < len(b.locals) {
			var src SourceKind = SourceLocal
			numParams := 0
			if b.fn.Type != nil {
				numParams = len(b.fn.Type.Params)
			}
			if int(idx) < numParams {
				src = SourceParam
			}
			b.push(&Value{
				Type:   b.locals[idx].Type,
				Source: src,
				Index:  idx,
			})
		}

	case wasm.OpGlobalGet:
		idx := getU32(instr.Immediates, 0)
		var t wasm.ValType = wasm.ValI32
		if b.module != nil && int(idx) < len(b.module.Globals) {
			t = b.module.Globals[idx].Type.Type
		}
		b.push(&Value{Type: t, Source: SourceGlobal, Index: idx})

	case wasm.OpI32Const:
		b.push(&Value{Type: wasm.ValI32, Source: SourceConst, Const: getImmediate(instr.Immediates, 0)})
	case wasm.OpI64Const:
		b.push(&Value{Type: wasm.ValI64, Source: SourceConst, Const: getImmediate(instr.Immediates, 0)})
	case wasm.OpF32Const:
		b.push(&Value{Type: wasm.ValF32, Source: SourceConst, Const: getImmediate(instr.Immediates, 0)})
	case wasm.OpF64Const:
		b.push(&Value{Type: wasm.ValF64, Source: SourceConst, Const: getImmediate(instr.Immediates, 0)})

	default:
		sig, ok := OpSignatures[instr.Opcode]
		if ok {
			inputs := make([]*Value, len(sig.Inputs))
			for i := len(sig.Inputs) - 1; i >= 0; i-- {
				inputs[i] = b.pop()
			}
			for _, t := range sig.Outputs {
				b.push(&Value{
					Type:   t,
					Source: SourceOp,
					Op:     &OpValue{Instr: instr, Inputs: inputs},
				})
			}
		} else {
			name := instr.Name
			if !strings.Contains(name, "0x") {
				name = fmt.Sprintf("%s (0x%x)", name, instr.Opcode)
			}
			msg := fmt.Sprintf("unsupported: %s", name)
			b.emit(&ErrorStmt{
				Message: msg,
				Offset:  instr.Offset,
				Opcode:  instr.Name,
			})
			b.recordError(instr.Offset, instr.Name, msg)
		}
	}
}

func (b *stmtBuilder) push(v *Value) {
	b.stack = append(b.stack, v)
}

func (b *stmtBuilder) pop() *Value {
	if len(b.stack) == 0 {
		if b.unreachable {
			return &Value{
				Type:   wasm.ValI32,
				Source: SourceConst,
				Const:  int32(0),
			}
		}
		var offset uint64
		var opcode string
		var msg string
		if b.currInstr != nil {
			offset = b.currInstr.Offset
			opcode = b.currInstr.Name
			msg = fmt.Sprintf("stack underflow at %s (0x%x)", b.currInstr.Name, b.currInstr.Opcode)
		} else {
			msg = "stack underflow"
		}
		b.recordError(offset, opcode, msg)
		return &Value{
			Type:   wasm.ValI32,
			Source: SourceError,
			Error: &ErrorValue{
				Message: msg,
				Offset:  offset,
			},
		}
	}
	v := b.stack[len(b.stack)-1]
	b.stack = b.stack[:len(b.stack)-1]
	return v
}

func (b *stmtBuilder) emit(s Stmt) {
	if len(b.blocks) > 0 {
		block := b.blocks[len(b.blocks)-1]
		block.Stmts = append(block.Stmts, s)
	} else {
		b.stmts = append(b.stmts, s)
	}
}

func (b *stmtBuilder) getBlockLabel(depth int) int {
	idx := len(b.blocks) - 1 - depth
	if idx >= 0 && idx < len(b.blocks) {
		return b.blocks[idx].Label
	}
	return 0
}

func parseBlockType(instr *wasm.Instruction) wasm.ValType {
	if len(instr.Immediates) == 0 {
		return 0
	}
	bt, ok := instr.Immediates[0].(byte)
	if !ok {
		return 0
	}
	if bt == 0x40 {
		return 0
	}
	return wasm.ValType(bt)
}
