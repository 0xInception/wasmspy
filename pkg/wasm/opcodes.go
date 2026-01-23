package wasm

type Opcode byte

const (
	OpUnreachable Opcode = 0x00
	OpNop         Opcode = 0x01
	OpBlock       Opcode = 0x02
	OpLoop        Opcode = 0x03
	OpEnd         Opcode = 0x0b
	OpBr          Opcode = 0x0c
	OpBrIf        Opcode = 0x0d
	OpReturn      Opcode = 0x0f
	OpCall        Opcode = 0x10
	OpLocalGet    Opcode = 0x20
	OpLocalSet    Opcode = 0x21
	OpLocalTee    Opcode = 0x22
	OpI32Const    Opcode = 0x41
	OpI64Const    Opcode = 0x42
	OpI32Add      Opcode = 0x6a
	OpI32Sub      Opcode = 0x6b
)

var OpcodeNames = map[Opcode]string{
	0x00: "unreachable",
	0x01: "nop",
	0x02: "block",
	0x03: "loop",
	0x0b: "end",
	0x0c: "br",
	0x0d: "br_if",
	0x0f: "return",
	0x10: "call",
	0x20: "local.get",
	0x21: "local.set",
	0x22: "local.tee",
	0x41: "i32.const",
	0x42: "i64.const",
	0x6a: "i32.add",
	0x6b: "i32.sub",
}
