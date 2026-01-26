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

func TestAllLinesHaveMappings(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "testdata", "arithmetic.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	for i, fn := range rm.Functions {
		if fn.Imported {
			continue
		}

		result := DecompileWithMappings(&rm.Functions[i], rm)
		lines := countNonEmptyLines(result.Code)

		mappedLines := make(map[int]bool)
		for _, m := range result.Mappings {
			mappedLines[m.Line] = true
		}

		var unmapped []int
		for line := 1; line <= lines; line++ {
			if !mappedLines[line] {
				unmapped = append(unmapped, line)
			}
		}

		if len(unmapped) > 0 {
			t.Errorf("Function %s (index %d): %d/%d lines have no mapping: %v\nCode:\n%s",
				fn.Name, fn.Index, len(unmapped), lines, unmapped, result.Code)
		}
	}
}

func countNonEmptyLines(code string) int {
	count := 0
	for _, line := range splitLines(code) {
		if len(line) > 0 {
			count++
		}
	}
	return count
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
