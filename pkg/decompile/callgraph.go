package decompile

import (
	"fmt"
	"strings"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

type CallGraph struct {
	Callers map[uint32][]uint32
	Callees map[uint32][]uint32
}

func BuildCallGraph(module *wasm.ResolvedModule) *CallGraph {
	cg := &CallGraph{
		Callers: make(map[uint32][]uint32),
		Callees: make(map[uint32][]uint32),
	}

	for i := range module.Functions {
		fn := &module.Functions[i]
		if fn.Body == nil {
			continue
		}

		for _, instr := range fn.Body.Instructions {
			if instr.Opcode == wasm.OpCall {
				callee := getU32(instr.Immediates, 0)
				cg.addEdge(uint32(fn.Index), callee)
			}
		}
	}

	return cg
}

func (cg *CallGraph) addEdge(caller, callee uint32) {
	for _, c := range cg.Callees[caller] {
		if c == callee {
			return
		}
	}
	cg.Callees[caller] = append(cg.Callees[caller], callee)
	cg.Callers[callee] = append(cg.Callers[callee], caller)
}

func (cg *CallGraph) Roots(module *wasm.ResolvedModule) []uint32 {
	var roots []uint32
	for i := range module.Functions {
		fn := &module.Functions[i]
		if fn.Imported {
			continue
		}
		if len(cg.Callers[uint32(fn.Index)]) == 0 {
			roots = append(roots, uint32(fn.Index))
		}
	}
	return roots
}

func (cg *CallGraph) String(module *wasm.ResolvedModule) string {
	var b strings.Builder

	for i := range module.Functions {
		fn := &module.Functions[i]
		if fn.Imported || fn.Body == nil {
			continue
		}

		callees := cg.Callees[uint32(fn.Index)]
		if len(callees) == 0 {
			continue
		}

		name := fn.Name
		if name == "" {
			name = fmt.Sprintf("func_%d", fn.Index)
		}

		b.WriteString(name)
		b.WriteString(" -> ")

		for j, callee := range callees {
			if j > 0 {
				b.WriteString(", ")
			}
			if f := module.GetFunction(callee); f != nil && f.Name != "" {
				b.WriteString(f.Name)
			} else {
				b.WriteString(fmt.Sprintf("func_%d", callee))
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}
