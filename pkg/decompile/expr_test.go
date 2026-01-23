package decompile

import (
	"path/filepath"
	"testing"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func TestExpressionRecovery(t *testing.T) {
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

	endFrame := analysis.Frames[len(analysis.Frames)-1]
	if len(endFrame.Stack) != 1 {
		t.Fatalf("expected 1 value on stack before end, got %d", len(endFrame.Stack))
	}

	result := endFrame.Stack[0]
	expr := ValueToExpr(result)

	str := exprString(expr)
	t.Logf("Expression: %s", str)

	if str != "(p0 + p1)" {
		t.Errorf("unexpected expression: %s", str)
	}
}

func TestExpressionArithmetic(t *testing.T) {
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

	endFrame := analysis.Frames[len(analysis.Frames)-1]
	result := endFrame.Stack[0]
	expr := ValueToExpr(result)

	str := exprString(expr)
	t.Logf("multiply: %s", str)

	if str != "(p0 * p1)" {
		t.Errorf("unexpected: %s", str)
	}

	fn = rm.GetFunctionByName("is_zero")
	if fn == nil {
		t.Fatal("function 'is_zero' not found")
	}

	analysis = Analyze(fn, rm)
	endFrame = analysis.Frames[len(analysis.Frames)-1]
	result = endFrame.Stack[0]
	expr = ValueToExpr(result)

	str = exprString(expr)
	t.Logf("is_zero: %s", str)

	if str != "!(p0)" {
		t.Errorf("unexpected: %s", str)
	}
}
