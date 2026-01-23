package tests

import (
	"path/filepath"
	"testing"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func getSection(mod *wasm.Module, id wasm.SectionID) *wasm.Section {
	for i := range mod.Sections {
		if mod.Sections[i].ID == id {
			return &mod.Sections[i]
		}
	}
	return nil
}

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

func TestResolve(t *testing.T) {
	path := filepath.Join("testdata", "add2.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	if len(rm.Types) != 1 {
		t.Errorf("expected 1 type, got %d", len(rm.Types))
	}

	if len(rm.Functions) != 1 {
		t.Errorf("expected 1 function, got %d", len(rm.Functions))
	}

	fn := rm.GetFunctionByName("add")
	if fn == nil {
		t.Fatal("function 'add' not found")
	}

	t.Logf("Function: %s", fn.Name)
	t.Logf("  Index: %d", fn.Index)
	t.Logf("  Type: %s", fn.Type.String())
	t.Logf("  Imported: %v", fn.Imported)
	t.Logf("  Instructions: %d", len(fn.Body.Instructions))

	for _, instr := range fn.Body.Instructions {
		t.Logf("    %s %v", instr.Name, instr.Immediates)
	}

	if fn.Type.String() != "(func (param i32 i32) (result i32))" {
		t.Errorf("unexpected type: %s", fn.Type.String())
	}
}

func TestResolveWithImports(t *testing.T) {
	path := filepath.Join("testdata", "with_import.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	if len(rm.Functions) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(rm.Functions))
	}

	t.Logf("Functions:")
	for _, fn := range rm.Functions {
		t.Logf("  [%d] %s - %s (imported=%v)", fn.Index, fn.Name, fn.Type.String(), fn.Imported)
	}

	if rm.Functions[0].Index != 0 {
		t.Errorf("func 0 index: got %d", rm.Functions[0].Index)
	}
	if !rm.Functions[0].Imported {
		t.Error("func 0 should be imported")
	}
	if rm.Functions[0].Name != "env.log" {
		t.Errorf("func 0 name: got %s", rm.Functions[0].Name)
	}

	if rm.Functions[1].Index != 1 {
		t.Errorf("func 1 index: got %d", rm.Functions[1].Index)
	}
	if rm.Functions[1].Imported {
		t.Error("func 1 should not be imported")
	}
	if rm.Functions[1].Name != "main" {
		t.Errorf("func 1 name: got %s, want 'main'", rm.Functions[1].Name)
	}
}

func TestWATOutput(t *testing.T) {
	path := filepath.Join("testdata", "add2.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	wat := rm.ToWAT()
	t.Logf("WAT output:\n%s", wat)
}

func TestWATOutputWithImports(t *testing.T) {
	path := filepath.Join("testdata", "with_import.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	wat := rm.ToWAT()
	t.Logf("WAT output:\n%s", wat)
}

func TestArithmeticOpcodes(t *testing.T) {
	path := filepath.Join("testdata", "arithmetic.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	if len(rm.Functions) != 3 {
		t.Fatalf("expected 3 functions, got %d", len(rm.Functions))
	}

	multiply := rm.GetFunctionByName("multiply")
	if multiply == nil {
		t.Fatal("function 'multiply' not found")
	}

	hasOp := func(fn *wasm.ResolvedFunction, name string) bool {
		for _, instr := range fn.Body.Instructions {
			if instr.Name == name {
				return true
			}
		}
		return false
	}

	if !hasOp(multiply, "i32.mul") {
		t.Error("multiply should have i32.mul")
	}

	isZero := rm.GetFunctionByName("is_zero")
	if isZero == nil {
		t.Fatal("function 'is_zero' not found")
	}
	if !hasOp(isZero, "i32.eqz") {
		t.Error("is_zero should have i32.eqz")
	}

	bitwise := rm.GetFunctionByName("bitwise")
	if bitwise == nil {
		t.Fatal("function 'bitwise' not found")
	}
	if !hasOp(bitwise, "i32.and") {
		t.Error("bitwise should have i32.and")
	}
	if !hasOp(bitwise, "i32.or") {
		t.Error("bitwise should have i32.or")
	}
	if !hasOp(bitwise, "i32.xor") {
		t.Error("bitwise should have i32.xor")
	}

	wat := rm.ToWAT()
	t.Logf("WAT output:\n%s", wat)
}

func TestTableDataStart(t *testing.T) {
	path := filepath.Join("testdata", "with_table_data.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	if len(rm.Tables) != 1 {
		t.Errorf("expected 1 table, got %d", len(rm.Tables))
	}

	if len(rm.Memories) != 1 {
		t.Errorf("expected 1 memory, got %d", len(rm.Memories))
	}

	if rm.Start == nil {
		t.Error("expected start section")
	} else if *rm.Start != 0 {
		t.Errorf("expected start func 0, got %d", *rm.Start)
	}

	if len(rm.Data) != 1 {
		t.Errorf("expected 1 data segment, got %d", len(rm.Data))
	} else {
		if string(rm.Data[0].Data) != "hello" {
			t.Errorf("expected data 'hello', got %q", string(rm.Data[0].Data))
		}
	}

	wat := rm.ToWAT()
	t.Logf("WAT output:\n%s", wat)
}

func TestElementAndNames(t *testing.T) {
	path := filepath.Join("testdata", "with_elem.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	if len(rm.Tables) != 1 {
		t.Errorf("expected 1 table, got %d", len(rm.Tables))
	}

	if len(rm.Elements) != 1 {
		t.Errorf("expected 1 element segment, got %d", len(rm.Elements))
	} else {
		elem := rm.Elements[0]
		if len(elem.FuncIdxs) != 2 {
			t.Errorf("expected 2 func indices, got %d", len(elem.FuncIdxs))
		}
	}

	if rm.Names == nil {
		t.Error("expected names section")
	} else {
		if len(rm.Names.FunctionNames) < 2 {
			t.Errorf("expected at least 2 function names, got %d", len(rm.Names.FunctionNames))
		}
		if rm.Names.FunctionNames[0] != "double" {
			t.Errorf("expected func 0 name 'double', got %q", rm.Names.FunctionNames[0])
		}
		if rm.Names.FunctionNames[1] != "triple" {
			t.Errorf("expected func 1 name 'triple', got %q", rm.Names.FunctionNames[1])
		}
	}

	double := rm.GetFunctionByName("double")
	if double == nil {
		t.Error("function 'double' not found")
	}

	triple := rm.GetFunctionByName("triple")
	if triple == nil {
		t.Error("function 'triple' not found")
	}

	wat := rm.ToWAT()
	t.Logf("WAT output:\n%s", wat)
}
