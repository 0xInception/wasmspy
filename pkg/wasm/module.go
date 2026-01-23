package wasm

import "fmt"

type ResolvedModule struct {
	Version   uint32
	Types     []FuncType
	Imports   []Import
	Functions []ResolvedFunction
	Memories  []Limits
	Exports   []Export
}

type ResolvedFunction struct {
	Index    uint32
	Name     string
	Type     *FuncType
	Imported bool
	Import   *Import
	Body     *FunctionBody
}

func Resolve(mod *Module) (*ResolvedModule, error) {
	rm := &ResolvedModule{
		Version: mod.Version,
	}

	sections := make(map[SectionID]*Section)
	for i := range mod.Sections {
		sections[mod.Sections[i].ID] = &mod.Sections[i]
	}

	if sec := sections[SectionType]; sec != nil {
		types, err := ParseTypeSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("type section: %w", err)
		}
		rm.Types = types
	}

	if sec := sections[SectionImport]; sec != nil {
		imports, err := ParseImportSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("import section: %w", err)
		}
		rm.Imports = imports
	}

	var funcTypeIndices []uint32
	if sec := sections[SectionFunction]; sec != nil {
		indices, err := ParseFunctionSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("function section: %w", err)
		}
		funcTypeIndices = indices
	}

	if sec := sections[SectionMemory]; sec != nil {
		memories, err := ParseMemorySection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("memory section: %w", err)
		}
		rm.Memories = memories
	}

	if sec := sections[SectionExport]; sec != nil {
		exports, err := ParseExportSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("export section: %w", err)
		}
		rm.Exports = exports
	}

	var bodies []FunctionBody
	if sec := sections[SectionCode]; sec != nil {
		b, err := ParseCodeSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("code section: %w", err)
		}
		bodies = b
	}

	funcIndex := uint32(0)

	for i := range rm.Imports {
		if rm.Imports[i].Kind != ImportFunc {
			continue
		}

		fn := ResolvedFunction{
			Index:    funcIndex,
			Name:     fmt.Sprintf("%s.%s", rm.Imports[i].Module, rm.Imports[i].Name),
			Imported: true,
			Import:   &rm.Imports[i],
		}

		if rm.Imports[i].TypeIdx < uint32(len(rm.Types)) {
			fn.Type = &rm.Types[rm.Imports[i].TypeIdx]
		}

		rm.Functions = append(rm.Functions, fn)
		funcIndex++
	}

	for i, typeIdx := range funcTypeIndices {
		fn := ResolvedFunction{
			Index:    funcIndex,
			Name:     fmt.Sprintf("func_%d", funcIndex),
			Imported: false,
		}

		if typeIdx < uint32(len(rm.Types)) {
			fn.Type = &rm.Types[typeIdx]
		}

		if i < len(bodies) {
			fn.Body = &bodies[i]
		}

		rm.Functions = append(rm.Functions, fn)
		funcIndex++
	}

	exportNames := make(map[uint32]string)
	for _, exp := range rm.Exports {
		if exp.Kind == ExportFunc {
			exportNames[exp.Index] = exp.Name
		}
	}

	for i := range rm.Functions {
		if name, ok := exportNames[rm.Functions[i].Index]; ok {
			rm.Functions[i].Name = name
		}
	}

	return rm, nil
}

func (rm *ResolvedModule) GetFunction(index uint32) *ResolvedFunction {
	for i := range rm.Functions {
		if rm.Functions[i].Index == index {
			return &rm.Functions[i]
		}
	}
	return nil
}

func (rm *ResolvedModule) GetFunctionByName(name string) *ResolvedFunction {
	for i := range rm.Functions {
		if rm.Functions[i].Name == name {
			return &rm.Functions[i]
		}
	}
	return nil
}
