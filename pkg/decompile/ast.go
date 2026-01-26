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

type NotExpr struct {
	Arg Expr
}

type AssignStmt struct {
	Target    Expr
	Value     Expr
	SrcOffset uint64
	Offsets   []uint64
}

type StoreStmt struct {
	Op        wasm.Opcode
	Addr      Expr
	Value     Expr
	Offset    uint32
	SrcOffset uint64
	Offsets   []uint64
}

type CallStmt struct {
	Call      *CallExpr
	SrcOffset uint64
	Offsets   []uint64
}

type ReturnStmt struct {
	Value     Expr
	SrcOffset uint64
	Offsets   []uint64
}

type DropStmt struct {
	Value     Expr
	SrcOffset uint64
	Offsets   []uint64
}

type SwitchStmt struct {
	Value   Expr
	Cases   []int
	Default int
	Offsets []uint64
}

type SwitchCase struct {
	Value int
	Body  []Stmt
}

type FlatSwitchStmt struct {
	Value   Expr
	Cases   []SwitchCase
	Default []Stmt
	Offsets []uint64
}

type WhileStmt struct {
	Cond    Expr
	Body    []Stmt
	Offsets []uint64
}

type ContinueStmt struct {
	Offsets []uint64
}

type ErrorExpr struct {
	Message string
	Offset  uint64
	Opcode  string
}

type ErrorStmt struct {
	Message string
	Offset  uint64
	Opcode  string
}

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
func (*NotExpr) node()     {}
func (*ErrorExpr) node()   {}

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
func (*NotExpr) expr()     {}
func (*ErrorExpr) expr()   {}

func (*AssignStmt) node()  {}
func (*StoreStmt) node()   {}
func (*CallStmt) node()    {}
func (*ReturnStmt) node()  {}
func (*DropStmt) node()    {}
func (*SwitchStmt) node()     {}
func (*FlatSwitchStmt) node() {}
func (*WhileStmt) node()      {}
func (*ContinueStmt) node()   {}
func (*ErrorStmt) node()      {}

func (*AssignStmt) stmt()  {}
func (*StoreStmt) stmt()   {}
func (*CallStmt) stmt()    {}
func (*ReturnStmt) stmt()  {}
func (*DropStmt) stmt()    {}
func (*SwitchStmt) stmt()     {}
func (*FlatSwitchStmt) stmt() {}
func (*WhileStmt) stmt()      {}
func (*ContinueStmt) stmt()   {}
func (*ErrorStmt) stmt()      {}
