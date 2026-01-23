package decompile

import (
	"testing"

	"github.com/0xInception/wasmspy/pkg/wasm"
)

func TestSimplifyNegation(t *testing.T) {
	expr := &BinaryExpr{
		Op:    wasm.OpI32Sub,
		Left:  &ConstExpr{Value: int32(0), Type: wasm.ValI32},
		Right: &ParamExpr{Index: 0, Type: wasm.ValI32},
		Type:  wasm.ValI32,
	}

	result := Simplify(expr)
	neg, ok := result.(*NegExpr)
	if !ok {
		t.Fatalf("expected NegExpr, got %T", result)
	}

	if _, ok := neg.Arg.(*ParamExpr); !ok {
		t.Fatalf("expected ParamExpr arg, got %T", neg.Arg)
	}

	t.Logf("simplified: %s", exprString(result))
}

func TestSimplifyAddZero(t *testing.T) {
	expr := &BinaryExpr{
		Op:    wasm.OpI32Add,
		Left:  &ParamExpr{Index: 0, Type: wasm.ValI32},
		Right: &ConstExpr{Value: int32(0), Type: wasm.ValI32},
		Type:  wasm.ValI32,
	}

	result := Simplify(expr)
	if _, ok := result.(*ParamExpr); !ok {
		t.Fatalf("expected ParamExpr, got %T", result)
	}
}

func TestSimplifyConstantFolding(t *testing.T) {
	expr := &BinaryExpr{
		Op:    wasm.OpI32Add,
		Left:  &ConstExpr{Value: int32(2), Type: wasm.ValI32},
		Right: &ConstExpr{Value: int32(3), Type: wasm.ValI32},
		Type:  wasm.ValI32,
	}

	result := Simplify(expr)
	c, ok := result.(*ConstExpr)
	if !ok {
		t.Fatalf("expected ConstExpr, got %T", result)
	}

	if c.Value != int32(5) {
		t.Fatalf("expected 5, got %v", c.Value)
	}
}
