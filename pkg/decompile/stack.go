package decompile

import (
	"fmt"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func Analyze(fn *wasm.ResolvedFunction, module *wasm.ResolvedModule) *Analysis {
	a := &Analysis{Func: fn}

	if fn.Body == nil {
		return a
	}

	locals := buildLocals(fn)
	var stack []*Value

	for i := range fn.Body.Instructions {
		instr := &fn.Body.Instructions[i]
		before := copyStack(stack)

		var err error
		stack, err = simulateInstr(instr, stack, locals, fn, module)
		if err != nil {
			a.Errors = append(a.Errors, fmt.Errorf("offset 0x%x: %w", instr.Offset, err))
		}

		a.Frames = append(a.Frames, Frame{
			Instr:  instr,
			Stack:  before,
			Locals: copyStack(locals),
		})
	}

	return a
}

func buildLocals(fn *wasm.ResolvedFunction) []*Value {
	var locals []*Value

	if fn.Type != nil {
		for i, t := range fn.Type.Params {
			locals = append(locals, &Value{
				Type:   t,
				Source: SourceParam,
				Index:  uint32(i),
			})
		}
	}

	if fn.Body != nil {
		idx := uint32(len(locals))
		for _, entry := range fn.Body.Locals {
			for j := uint32(0); j < entry.Count; j++ {
				locals = append(locals, &Value{
					Type:   wasm.ValType(entry.Type),
					Source: SourceLocal,
					Index:  idx,
				})
				idx++
			}
		}
	}

	return locals
}

func simulateInstr(instr *wasm.Instruction, stack []*Value, locals []*Value, fn *wasm.ResolvedFunction, module *wasm.ResolvedModule) ([]*Value, error) {
	switch instr.Opcode {
	case wasm.OpLocalGet:
		idx := getU32(instr.Immediates, 0)
		if int(idx) >= len(locals) {
			return stack, fmt.Errorf("local index %d out of bounds", idx)
		}
		stack = append(stack, &Value{
			Type:   locals[idx].Type,
			Source: SourceLocal,
			Index:  idx,
		})

	case wasm.OpLocalSet:
		idx := getU32(instr.Immediates, 0)
		if len(stack) < 1 {
			return stack, fmt.Errorf("stack underflow")
		}
		val := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if int(idx) < len(locals) {
			locals[idx] = val
		}

	case wasm.OpLocalTee:
		idx := getU32(instr.Immediates, 0)
		if len(stack) < 1 {
			return stack, fmt.Errorf("stack underflow")
		}
		val := stack[len(stack)-1]
		if int(idx) < len(locals) {
			locals[idx] = val
		}

	case wasm.OpGlobalGet:
		idx := getU32(instr.Immediates, 0)
		var t wasm.ValType = wasm.ValI32
		if module != nil && int(idx) < len(module.Globals) {
			t = module.Globals[idx].Type.Type
		}
		stack = append(stack, &Value{
			Type:   t,
			Source: SourceGlobal,
			Index:  idx,
		})

	case wasm.OpGlobalSet:
		if len(stack) < 1 {
			return stack, fmt.Errorf("stack underflow")
		}
		stack = stack[:len(stack)-1]

	case wasm.OpI32Const:
		stack = append(stack, &Value{
			Type:   wasm.ValI32,
			Source: SourceConst,
			Const:  getImmediate(instr.Immediates, 0),
		})

	case wasm.OpI64Const:
		stack = append(stack, &Value{
			Type:   wasm.ValI64,
			Source: SourceConst,
			Const:  getImmediate(instr.Immediates, 0),
		})

	case wasm.OpF32Const:
		stack = append(stack, &Value{
			Type:   wasm.ValF32,
			Source: SourceConst,
			Const:  getImmediate(instr.Immediates, 0),
		})

	case wasm.OpF64Const:
		stack = append(stack, &Value{
			Type:   wasm.ValF64,
			Source: SourceConst,
			Const:  getImmediate(instr.Immediates, 0),
		})

	case wasm.OpCall:
		idx := getU32(instr.Immediates, 0)
		var sig *wasm.FuncType
		if module != nil {
			if f := module.GetFunction(idx); f != nil && f.Type != nil {
				sig = f.Type
			}
		}
		if sig != nil {
			if len(stack) < len(sig.Params) {
				return stack, fmt.Errorf("stack underflow for call")
			}
			inputs := stack[len(stack)-len(sig.Params):]
			stack = stack[:len(stack)-len(sig.Params)]
			for _, t := range sig.Results {
				stack = append(stack, &Value{
					Type:   t,
					Source: SourceOp,
					Op:     &OpValue{Instr: instr, Inputs: copyStack(inputs)},
				})
			}
		}

	case wasm.OpCallIndirect:
		typeIdx := getU32(instr.Immediates, 0)
		var sig *wasm.FuncType
		if module != nil && int(typeIdx) < len(module.Types) {
			sig = &module.Types[typeIdx]
		}
		if len(stack) < 1 {
			return stack, fmt.Errorf("stack underflow for call_indirect")
		}
		stack = stack[:len(stack)-1]
		if sig != nil {
			if len(stack) < len(sig.Params) {
				return stack, fmt.Errorf("stack underflow for call_indirect")
			}
			inputs := stack[len(stack)-len(sig.Params):]
			stack = stack[:len(stack)-len(sig.Params)]
			for _, t := range sig.Results {
				stack = append(stack, &Value{
					Type:   t,
					Source: SourceOp,
					Op:     &OpValue{Instr: instr, Inputs: copyStack(inputs)},
				})
			}
		}

	default:
		sig, ok := OpSignatures[instr.Opcode]
		if ok {
			if len(stack) < len(sig.Inputs) {
				return stack, fmt.Errorf("stack underflow: need %d, have %d", len(sig.Inputs), len(stack))
			}
			inputs := stack[len(stack)-len(sig.Inputs):]
			stack = stack[:len(stack)-len(sig.Inputs)]
			for _, t := range sig.Outputs {
				stack = append(stack, &Value{
					Type:   t,
					Source: SourceOp,
					Op:     &OpValue{Instr: instr, Inputs: copyStack(inputs)},
				})
			}
		}
	}

	return stack, nil
}

func copyStack(stack []*Value) []*Value {
	if stack == nil {
		return nil
	}
	cp := make([]*Value, len(stack))
	copy(cp, stack)
	return cp
}

func getU32(imm []any, idx int) uint32 {
	if idx >= len(imm) {
		return 0
	}
	switch v := imm[idx].(type) {
	case uint32:
		return v
	case int32:
		return uint32(v)
	case byte:
		return uint32(v)
	}
	return 0
}

func getImmediate(imm []any, idx int) any {
	if idx >= len(imm) {
		return nil
	}
	return imm[idx]
}
