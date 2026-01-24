package wasm

import "fmt"

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

	if sec := sections[SectionTable]; sec != nil {
		tables, err := ParseTableSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("table section: %w", err)
		}
		rm.Tables = tables
	}

	if sec := sections[SectionMemory]; sec != nil {
		memories, err := ParseMemorySection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("memory section: %w", err)
		}
		rm.Memories = memories
	}

	if sec := sections[SectionGlobal]; sec != nil {
		globals, err := ParseGlobalSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("global section: %w", err)
		}
		rm.Globals = globals
	}

	if sec := sections[SectionExport]; sec != nil {
		exports, err := ParseExportSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("export section: %w", err)
		}
		rm.Exports = exports
	}

	if sec := sections[SectionStart]; sec != nil {
		startIdx, err := ParseStartSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("start section: %w", err)
		}
		rm.Start = &startIdx
	}

	if sec := sections[SectionElement]; sec != nil {
		elements, err := ParseElementSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("element section: %w", err)
		}
		rm.Elements = elements
	}

	if sec := sections[SectionData]; sec != nil {
		data, err := ParseDataSection(sec.Content, int(sec.Offset))
		if err != nil {
			return nil, fmt.Errorf("data section: %w", err)
		}
		rm.Data = data
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

	for i := range mod.Sections {
		if mod.Sections[i].ID != SectionCustom {
			continue
		}
		content := mod.Sections[i].Content
		if len(content) < 5 {
			continue
		}
		nameLen, n, err := ReadLEB128U32FromSlice(content)
		if err != nil || int(nameLen)+n > len(content) {
			continue
		}
		secName := string(content[n : n+int(nameLen)])
		if secName == "name" {
			payload := content[n+int(nameLen):]
			names, err := ParseNameSection(payload, int(mod.Sections[i].Offset)+n+int(nameLen))
			if err == nil && names != nil {
				rm.Names = names
				for idx, name := range names.FunctionNames {
					if _, hasExport := exportNames[idx]; !hasExport {
						for j := range rm.Functions {
							if rm.Functions[j].Index == idx {
								rm.Functions[j].Name = name
								break
							}
						}
					}
				}
			}
			break
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

func (rm *ResolvedModule) BuildMemory() []byte {
	if len(rm.Data) == 0 {
		return nil
	}
	var maxEnd uint32
	for _, seg := range rm.Data {
		offset := getDataOffset(seg.Offset)
		end := offset + uint32(len(seg.Data))
		if end > maxEnd {
			maxEnd = end
		}
	}
	mem := make([]byte, maxEnd)
	for _, seg := range rm.Data {
		offset := getDataOffset(seg.Offset)
		copy(mem[offset:], seg.Data)
	}
	return mem
}

func getDataOffset(instrs []Instruction) uint32 {
	for _, instr := range instrs {
		if instr.Opcode == OpI32Const && len(instr.Immediates) > 0 {
			if v, ok := instr.Immediates[0].(int32); ok {
				return uint32(v)
			}
			if v, ok := instr.Immediates[0].(uint32); ok {
				return v
			}
		}
	}
	return 0
}

func (rm *ResolvedModule) ReadString(addr, length uint32) string {
	mem := rm.BuildMemory()
	if mem == nil || int(addr+length) > len(mem) {
		return ""
	}
	return string(mem[addr : addr+length])
}
