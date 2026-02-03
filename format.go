package main

import (
	"fmt"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func formatFunctionDisasm(fn *wasm.ResolvedFunction, rm *wasm.ResolvedModule, indented bool, ann *Annotations) string {
	if fn.Imported {
		return fmt.Sprintf("; imported: %s", fn.Name)
	}

	name := fn.Name
	funcComment := ""
	if ann != nil {
		if fa := ann.Functions[fmt.Sprintf("%d", fn.Index)]; fa != nil {
			if fa.Name != "" {
				name = fa.Name
			}
			if fa.Comment != "" {
				funcComment = fa.Comment
			}
		}
	}

	var result string
	result = fmt.Sprintf("; Function %d: %s\n", fn.Index, name)
	if funcComment != "" {
		result += fmt.Sprintf("; %s\n", funcComment)
	}

	if fn.Type != nil {
		result += fmt.Sprintf("; Params: %d, Results: %d\n", len(fn.Type.Params), len(fn.Type.Results))
	}

	if fn.Body != nil {
		localIdx := 0
		if fn.Type != nil {
			localIdx = len(fn.Type.Params)
		}
		for _, loc := range fn.Body.Locals {
			for i := uint32(0); i < loc.Count; i++ {
				result += fmt.Sprintf(";   local[%d]: %s\n", localIdx, wasm.ValType(loc.Type).String())
				localIdx++
			}
		}
		result += "\n"

		indent := 0
		for _, instr := range fn.Body.Instructions {
			comment := ""
			if ann != nil {
				if c := ann.Comments[fmt.Sprintf("0x%x", instr.Offset)]; c != "" {
					comment = " ; " + c
				}
			}

			if indented {
				if instr.Opcode == wasm.OpEnd || instr.Opcode == wasm.OpElse {
					if indent > 0 {
						indent--
					}
				}

				prefix := ""
				for j := 0; j < indent; j++ {
					prefix += "  "
				}
				result += fmt.Sprintf("%08x: %s%s%s\n", instr.Offset, prefix, formatInstrWithImm(&instr), comment)

				if instr.Opcode == wasm.OpBlock || instr.Opcode == wasm.OpLoop ||
					instr.Opcode == wasm.OpIf || instr.Opcode == wasm.OpElse {
					indent++
				}
			} else {
				result += fmt.Sprintf("%08x: %s%s\n", instr.Offset, formatInstrWithImm(&instr), comment)
			}
		}
	}

	return result
}

func formatInstrWithImm(instr *wasm.Instruction) string {
	if len(instr.Immediates) == 0 {
		return instr.Name
	}
	result := instr.Name
	for _, imm := range instr.Immediates {
		switch v := imm.(type) {
		case []uint32:
			result += " ["
			for i, n := range v {
				if i > 0 {
					result += ", "
				}
				result += fmt.Sprintf("%d", n)
			}
			result += "]"
		default:
			result += fmt.Sprintf(" %v", imm)
		}
	}
	return result
}

func formatFunctionWAT(fn *wasm.ResolvedFunction, rm *wasm.ResolvedModule) string {
	if fn.Imported {
		for _, imp := range rm.Imports {
			if imp.Kind == wasm.ImportFunc {
				return fmt.Sprintf("(import %q %q (func))", imp.Module, imp.Name)
			}
		}
		return "(import)"
	}

	var result string
	result = fmt.Sprintf(";; Function %d: %s\n", fn.Index, fn.Name)
	result += "(func"

	if fn.Type != nil {
		if len(fn.Type.Params) > 0 {
			result += " (param"
			for _, p := range fn.Type.Params {
				result += " " + p.String()
			}
			result += ")"
		}
		if len(fn.Type.Results) > 0 {
			result += " (result"
			for _, r := range fn.Type.Results {
				result += " " + r.String()
			}
			result += ")"
		}
	}
	result += "\n"

	if fn.Body != nil {
		for _, loc := range fn.Body.Locals {
			for i := uint32(0); i < loc.Count; i++ {
				result += fmt.Sprintf("  (local %s)\n", wasm.ValType(loc.Type).String())
			}
		}
		for _, instr := range fn.Body.Instructions {
			if instr.Opcode == wasm.OpEnd {
				continue
			}
			result += "  " + formatInstr(&instr) + "\n"
		}
	}

	result += ")"
	return result
}

func formatInstr(instr *wasm.Instruction) string {
	if len(instr.Immediates) == 0 {
		return instr.Name
	}
	result := instr.Name
	for _, imm := range instr.Immediates {
		result += fmt.Sprintf(" %v", imm)
	}
	return result
}
