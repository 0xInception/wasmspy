package decompile

import (
	"path/filepath"
	"testing"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func TestCallGraph(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "testdata", "add.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	cg := BuildCallGraph(rm)
	t.Logf("Call graph:\n%s", cg.String(rm))
	t.Logf("Roots: %v", cg.Roots(rm))
}
