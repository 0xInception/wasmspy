package tests

import (
	"path/filepath"
	"testing"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

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
