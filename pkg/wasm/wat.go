package wasm

import (
	"fmt"
	"strings"
)

func (rm *ResolvedModule) ToWAT() string {
	var b strings.Builder

	funcExports := make(map[uint32]string)
	memExports := make(map[uint32]string)
	globalExports := make(map[uint32]string)
	tableExports := make(map[uint32]string)

	for _, exp := range rm.Exports {
		switch exp.Kind {
		case ExportFunc:
			funcExports[exp.Index] = exp.Name
		case ExportMemory:
			memExports[exp.Index] = exp.Name
		case ExportGlobal:
			globalExports[exp.Index] = exp.Name
		case ExportTable:
			tableExports[exp.Index] = exp.Name
		}
	}

	b.WriteString("(module\n")

	for _, imp := range rm.Imports {
		b.WriteString(fmt.Sprintf("  %s\n", formatImport(&imp, rm.Types)))
	}

	for _, fn := range rm.Functions {
		if fn.Imported {
			continue
		}
		exportName := funcExports[fn.Index]
		b.WriteString(formatFunction(&fn, exportName))
	}

	for i, mem := range rm.Memories {
		exp := memExports[uint32(i)]
		if exp != "" {
			b.WriteString(fmt.Sprintf("  (memory (export %q) %s)\n", exp, formatLimits(&mem)))
		} else {
			b.WriteString(fmt.Sprintf("  (memory %s)\n", formatLimits(&mem)))
		}
	}

	for i, glob := range rm.Globals {
		b.WriteString(formatGlobal(&glob, globalExports[uint32(i)]))
	}

	b.WriteString(")")

	return b.String()
}

func formatImport(imp *Import, types []FuncType) string {
	var desc string

	switch imp.Kind {
	case ImportFunc:
		if int(imp.TypeIdx) < len(types) {
			desc = formatImportFunc(&types[imp.TypeIdx])
		} else {
			desc = "(func)"
		}
	case ImportMemory:
		desc = fmt.Sprintf("(memory %s)", formatLimits(imp.Memory))
	case ImportGlobal:
		if imp.Global.Mutable {
			desc = fmt.Sprintf("(global (mut %s))", imp.Global.Type.String())
		} else {
			desc = fmt.Sprintf("(global %s)", imp.Global.Type.String())
		}
	case ImportTable:
		desc = fmt.Sprintf("(table %s funcref)", formatLimits(imp.Table))
	}

	return fmt.Sprintf("(import %q %q %s)", imp.Module, imp.Name, desc)
}

func formatImportFunc(typ *FuncType) string {
	var parts []string

	if len(typ.Params) > 0 {
		params := make([]string, len(typ.Params))
		for i, p := range typ.Params {
			params[i] = p.String()
		}
		parts = append(parts, fmt.Sprintf("(param %s)", strings.Join(params, " ")))
	}

	if len(typ.Results) > 0 {
		results := make([]string, len(typ.Results))
		for i, r := range typ.Results {
			results[i] = r.String()
		}
		parts = append(parts, fmt.Sprintf("(result %s)", strings.Join(results, " ")))
	}

	if len(parts) == 0 {
		return "(func)"
	}
	return fmt.Sprintf("(func %s)", strings.Join(parts, " "))
}

func formatFunction(fn *ResolvedFunction, exportName string) string {
	var b strings.Builder

	b.WriteString("  (func")

	if exportName != "" {
		b.WriteString(fmt.Sprintf(" (export %q)", exportName))
	}

	if fn.Type != nil {
		if len(fn.Type.Params) > 0 {
			b.WriteString(" (param")
			for _, p := range fn.Type.Params {
				b.WriteString(" ")
				b.WriteString(p.String())
			}
			b.WriteString(")")
		}
		if len(fn.Type.Results) > 0 {
			b.WriteString(" (result")
			for _, r := range fn.Type.Results {
				b.WriteString(" ")
				b.WriteString(r.String())
			}
			b.WriteString(")")
		}
	}

	b.WriteString("\n")

	if fn.Body != nil {
		for _, loc := range fn.Body.Locals {
			for i := uint32(0); i < loc.Count; i++ {
				b.WriteString(fmt.Sprintf("    (local %s)\n", ValType(loc.Type).String()))
			}
		}

		for _, instr := range fn.Body.Instructions {
			if instr.Opcode == OpEnd {
				continue
			}
			b.WriteString("    ")
			b.WriteString(formatInstruction(&instr))
			b.WriteString("\n")
		}
	}

	b.WriteString("  )\n")

	return b.String()
}

func formatInstruction(instr *Instruction) string {
	if len(instr.Immediates) == 0 {
		return instr.Name
	}

	var args []string
	for _, imm := range instr.Immediates {
		switch v := imm.(type) {
		case uint32:
			args = append(args, fmt.Sprintf("%d", v))
		case int32:
			args = append(args, fmt.Sprintf("%d", v))
		case int64:
			args = append(args, fmt.Sprintf("%d", v))
		case byte:
			args = append(args, fmt.Sprintf("%d", v))
		default:
			args = append(args, fmt.Sprintf("%v", v))
		}
	}

	return fmt.Sprintf("%s %s", instr.Name, strings.Join(args, " "))
}

func formatGlobal(glob *Global, exportName string) string {
	var b strings.Builder

	b.WriteString("  (global")

	if exportName != "" {
		b.WriteString(fmt.Sprintf(" (export %q)", exportName))
	}

	if glob.Type.Mutable {
		b.WriteString(fmt.Sprintf(" (mut %s)", glob.Type.Type.String()))
	} else {
		b.WriteString(fmt.Sprintf(" %s", glob.Type.Type.String()))
	}

	for _, instr := range glob.Init {
		if instr.Opcode == OpEnd {
			continue
		}
		b.WriteString(fmt.Sprintf(" (%s)", formatInstruction(&instr)))
	}

	b.WriteString(")\n")

	return b.String()
}

func formatLimits(lim *Limits) string {
	if lim.HasMax {
		return fmt.Sprintf("%d %d", lim.Min, lim.Max)
	}
	return fmt.Sprintf("%d", lim.Min)
}

func formatTypes(types []ValType) string {
	if len(types) == 0 {
		return ""
	}
	parts := make([]string, len(types))
	for i, t := range types {
		parts[i] = t.String()
	}
	return strings.Join(parts, " ")
}
