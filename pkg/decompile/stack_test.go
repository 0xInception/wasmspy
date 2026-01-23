package decompile

import (
	"path/filepath"
	"testing"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func TestAnalyzeAdd(t *testing.T) {
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

	analysis := Analyze(fn, rm)

	if len(analysis.Errors) > 0 {
		for _, e := range analysis.Errors {
			t.Errorf("analysis error: %v", e)
		}
	}

	if len(analysis.Frames) != len(fn.Body.Instructions) {
		t.Errorf("expected %d frames, got %d", len(fn.Body.Instructions), len(analysis.Frames))
	}

	for i, frame := range analysis.Frames {
		t.Logf("frame %d: %s stack=%d", i, frame.Instr.Name, len(frame.Stack))
		for j, v := range frame.Stack {
			t.Logf("  stack[%d]: %s source=%d", j, v.Type.String(), v.Source)
		}
	}
}

func TestAnalyzeArithmetic(t *testing.T) {
	path := filepath.Join("..", "..", "tests", "testdata", "arithmetic.wasm")

	mod, err := wasm.ParseFile(path)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	rm, err := wasm.Resolve(mod)
	if err != nil {
		t.Fatalf("resolve error: %v", err)
	}

	fn := rm.GetFunctionByName("multiply")
	if fn == nil {
		t.Fatal("function 'multiply' not found")
	}

	analysis := Analyze(fn, rm)

	if len(analysis.Errors) > 0 {
		for _, e := range analysis.Errors {
			t.Errorf("analysis error: %v", e)
		}
	}

	lastFrame := analysis.Frames[len(analysis.Frames)-1]
	if lastFrame.Instr.Opcode != wasm.OpEnd {
		t.Logf("last instruction: %s", lastFrame.Instr.Name)
	}

	mulFrame := findFrame(analysis.Frames, wasm.OpI32Mul)
	if mulFrame == nil {
		t.Fatal("i32.mul frame not found")
	}

	if len(mulFrame.Stack) != 2 {
		t.Errorf("expected 2 values on stack before i32.mul, got %d", len(mulFrame.Stack))
	}

	for i, v := range mulFrame.Stack {
		if v.Source != SourceParam {
			t.Errorf("stack[%d] expected SourceParam, got %d", i, v.Source)
		}
	}
}

func findFrame(frames []Frame, op wasm.Opcode) *Frame {
	for i := range frames {
		if frames[i].Instr.Opcode == op {
			return &frames[i]
		}
	}
	return nil
}
