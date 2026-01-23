package decompile

import (
	"path/filepath"
	"testing"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func TestDecompileAdd(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "testdata", "add2.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	fn := rm.GetFunctionByName("add")
	if fn == nil {
		t.Fatal("function 'add' not found")
	}

	result := Decompile(fn, rm)
	t.Logf("Decompiled:\n%s", result)

	if result == "" {
		t.Error("empty output")
	}
}

func TestDecompileModule(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "testdata", "arithmetic.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	result := DecompileModule(rm)
	t.Logf("Decompiled module:\n%s", result)
}
