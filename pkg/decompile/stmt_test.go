package decompile

import (
	"path/filepath"
	"testing"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func TestDecompileControlFlow(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "testdata", "control_flow.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	result := DecompileModule(rm)
	t.Logf("Decompiled:\n%s", result)
}
