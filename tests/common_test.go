package tests

import "github.com/0xInception/wasmspy/pkg/wasm"

func getSection(mod *wasm.Module, id wasm.SectionID) *wasm.Section {
	for i := range mod.Sections {
		if mod.Sections[i].ID == id {
			return &mod.Sections[i]
		}
	}
	return nil
}
