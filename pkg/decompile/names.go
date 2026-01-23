package decompile

import (
	"fmt"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

type NameResolver struct {
	module  *wasm.ResolvedModule
	funcIdx uint32
}

func NewNameResolver(module *wasm.ResolvedModule, funcIdx uint32) *NameResolver {
	return &NameResolver{module: module, funcIdx: funcIdx}
}

func (r *NameResolver) Local(idx uint32, numParams int) string {
	if r.module != nil && r.module.Names != nil {
		if locals, ok := r.module.Names.LocalNames[r.funcIdx]; ok {
			if name, ok := locals[idx]; ok {
				return name
			}
		}
	}

	if int(idx) < numParams {
		return fmt.Sprintf("p%d", idx)
	}
	return fmt.Sprintf("v%d", idx)
}

func (r *NameResolver) Global(idx uint32) string {
	return fmt.Sprintf("global%d", idx)
}

func (r *NameResolver) Func(idx uint32) string {
	if r.module != nil {
		if fn := r.module.GetFunction(idx); fn != nil && fn.Name != "" {
			return fn.Name
		}
	}
	return fmt.Sprintf("func%d", idx)
}
