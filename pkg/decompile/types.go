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
	Instr  *wasm.Instruction
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

func CollectValueOffsets(v *Value) []uint64 {
	if v == nil {
		return nil
	}

	seen := make(map[uint64]bool)
	collectValueOffsetsImpl(v, seen)

	offsets := make([]uint64, 0, len(seen))
	for off := range seen {
		offsets = append(offsets, off)
	}
	return offsets
}

func collectValueOffsetsImpl(v *Value, seen map[uint64]bool) {
	if v == nil {
		return
	}

	if v.Instr != nil {
		seen[v.Instr.Offset] = true
	}

	if v.Op != nil {
		if v.Op.Instr != nil {
			seen[v.Op.Instr.Offset] = true
		}
		for _, input := range v.Op.Inputs {
			collectValueOffsetsImpl(input, seen)
		}
	}

	if v.Error != nil && v.Error.Offset != 0 {
		seen[v.Error.Offset] = true
	}
}
