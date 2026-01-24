package tests

import (
	"path/filepath"
	"testing"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

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
