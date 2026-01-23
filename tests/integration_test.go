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
