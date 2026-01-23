package decompile

import "github.com/0xInception/wasmspy/pkg/wasm"

type Node interface {
	node()
}

type Expr interface {
	Node
	expr()
}

type Stmt interface {
	Node
	stmt()
}

type LocalExpr struct {
	Index uint32
	Type  wasm.ValType
}

type GlobalExpr struct {
	Index uint32
	Type  wasm.ValType
}

type ConstExpr struct {
	Value any
	Type  wasm.ValType
}

type ParamExpr struct {
	Index uint32
	Type  wasm.ValType
}

type BinaryExpr struct {
	Op    wasm.Opcode
	Left  Expr
	Right Expr
	Type  wasm.ValType
}

type UnaryExpr struct {
	Op   wasm.Opcode
	Arg  Expr
	Type wasm.ValType
}

type CallExpr struct {
	FuncIndex uint32
	FuncName  string
	Args      []Expr
	Type      wasm.ValType
}

type LoadExpr struct {
	Op     wasm.Opcode
	Addr   Expr
	Offset uint32
	Type   wasm.ValType
}

type TernaryExpr struct {
	Cond       Expr
	ThenResult Expr
	ElseResult Expr
	Type       wasm.ValType
}

type NegExpr struct {
	Arg  Expr
	Type wasm.ValType
}

type AssignStmt struct {
	Target Expr
	Value  Expr
}

type StoreStmt struct {
	Op     wasm.Opcode
	Addr   Expr
	Value  Expr
	Offset uint32
}

type CallStmt struct {
	Call *CallExpr
}

type ReturnStmt struct {
	Value Expr
}

type DropStmt struct {
	Value Expr
}

type SwitchStmt struct {
	Value   Expr
	Cases   []int
	Default int
}

type WhileStmt struct {
	Cond Expr
	Body []Stmt
}

type ContinueStmt struct{}

func (*LocalExpr) node()   {}
func (*GlobalExpr) node()  {}
func (*ConstExpr) node()   {}
func (*ParamExpr) node()   {}
func (*BinaryExpr) node()  {}
func (*UnaryExpr) node()   {}
func (*CallExpr) node()    {}
func (*LoadExpr) node()    {}
func (*TernaryExpr) node() {}
func (*NegExpr) node()     {}

func (*LocalExpr) expr()   {}
func (*GlobalExpr) expr()  {}
func (*ConstExpr) expr()   {}
func (*ParamExpr) expr()   {}
func (*BinaryExpr) expr()  {}
func (*UnaryExpr) expr()   {}
func (*CallExpr) expr()    {}
func (*LoadExpr) expr()    {}
func (*TernaryExpr) expr() {}
func (*NegExpr) expr()     {}

func (*AssignStmt) node()  {}
func (*StoreStmt) node()   {}
func (*CallStmt) node()    {}
func (*ReturnStmt) node()  {}
func (*DropStmt) node()    {}
func (*SwitchStmt) node()    {}
func (*WhileStmt) node()     {}
func (*ContinueStmt) node()  {}

func (*AssignStmt) stmt()  {}
func (*StoreStmt) stmt()   {}
func (*CallStmt) stmt()    {}
func (*ReturnStmt) stmt()  {}
func (*DropStmt) stmt()    {}
func (*SwitchStmt) stmt()    {}
func (*WhileStmt) stmt()     {}
func (*ContinueStmt) stmt()  {}
