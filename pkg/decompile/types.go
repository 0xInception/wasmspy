package decompile

import "github.com/0xInception/wasmspy/pkg/wasm"

type SourceKind int

const (
	SourceLocal SourceKind = iota
	SourceGlobal
	SourceConst
	SourceOp
	SourceParam
	SourceMemory
	SourceError
)

type Value struct {
	Type   wasm.ValType
	Source SourceKind
	Index  uint32
	Const  any
	Op     *OpValue
	Error  *ErrorValue
}

type ErrorValue struct {
	Message string
	Offset  uint64
	Opcode  string
}

type OpValue struct {
	Instr   *wasm.Instruction
	Inputs  []*Value
	Ternary *TernaryValue
}

type TernaryValue struct {
	Cond       Expr
	ThenResult Expr
	ElseResult Expr
}

type Frame struct {
	Instr  *wasm.Instruction
	Stack  []*Value
	Locals []*Value
}

type Analysis struct {
	Func   *wasm.ResolvedFunction
	Frames []Frame
	Errors []error
}
