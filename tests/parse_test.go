package tests

import (
	"path/filepath"
	"testing"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func TestParseAndDisassembleRealFile(t *testing.T) {
	path := filepath.Join("testdata", "add.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	var codeSection *wasm.Section
	for i := range mod.Sections {
		if mod.Sections[i].ID == wasm.SectionCode {
			codeSection = &mod.Sections[i]
			break
		}
	}
	if codeSection == nil {
		t.Fatal("Could not find Code Section")
	}

	bodies, err := wasm.ParseCodeSection(codeSection.Content, int(codeSection.Offset))
	if err != nil {
		t.Fatalf("Failed to parse code section: %v", err)
	}

	if len(bodies) == 0 {
		t.Fatal("Expected at least one function body, found 0")
	}

	instrs := bodies[0].Instructions

	expectedOps := []string{"i32.const", "i32.const", "i32.add", "end"}

	if len(instrs) != len(expectedOps) {
		t.Fatalf("Expected %d instructions, got %d", len(expectedOps), len(instrs))
	}

	for i, want := range expectedOps {
		got := instrs[i].Name
		if got != want {
			t.Errorf("Instruction %d: expected %s, got %s", i, want, got)
		}
	}
}

func TestParseAllSections(t *testing.T) {
	path := filepath.Join("testdata", "add.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	typeSec := getSection(mod, wasm.SectionType)
	if typeSec == nil {
		t.Fatal("missing type section")
	}
	types, err := wasm.ParseTypeSection(typeSec.Content, int(typeSec.Offset))
	if err != nil {
		t.Fatalf("failed to parse type section: %v", err)
	}

	funcSec := getSection(mod, wasm.SectionFunction)
	if funcSec == nil {
		t.Fatal("missing function section")
	}
	funcIndices, err := wasm.ParseFunctionSection(funcSec.Content, int(funcSec.Offset))
	if err != nil {
		t.Fatalf("failed to parse function section: %v", err)
	}

	var exports []wasm.Export
	exportSec := getSection(mod, wasm.SectionExport)
	if exportSec != nil {
		exports, err = wasm.ParseExportSection(exportSec.Content, int(exportSec.Offset))
		if err != nil {
			t.Fatalf("failed to parse export section: %v", err)
		}
	}

	codeSec := getSection(mod, wasm.SectionCode)
	if codeSec == nil {
		t.Fatal("missing code section")
	}
	bodies, err := wasm.ParseCodeSection(codeSec.Content, int(codeSec.Offset))
	if err != nil {
		t.Fatalf("failed to parse code section: %v", err)
	}

	t.Logf("Types: %d", len(types))
	for i, typ := range types {
		t.Logf("  type[%d] = %s", i, typ.String())
	}

	t.Logf("Functions: %d", len(funcIndices))
	for i, typeIdx := range funcIndices {
		t.Logf("  func[%d] -> type[%d] = %s", i, typeIdx, types[typeIdx].String())
	}

	t.Logf("Exports: %d", len(exports))
	for _, exp := range exports {
		if exp.Kind == wasm.ExportFunc {
			typeIdx := funcIndices[exp.Index]
			t.Logf("  export %q = func[%d] %s", exp.Name, exp.Index, types[typeIdx].String())
		}
	}

	t.Logf("Code bodies: %d", len(bodies))
	for i, body := range bodies {
		t.Logf("  func[%d]: %d instructions", i, len(body.Instructions))
	}
}
