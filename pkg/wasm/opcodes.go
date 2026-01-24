package wasm

type Opcode uint16

const (
	OpUnreachable Opcode = 0x00
	OpNop         Opcode = 0x01
	OpBlock       Opcode = 0x02
	OpLoop        Opcode = 0x03
	OpIf          Opcode = 0x04
	OpElse        Opcode = 0x05
	OpEnd         Opcode = 0x0b
	OpBr          Opcode = 0x0c
	OpBrIf        Opcode = 0x0d
	OpBrTable     Opcode = 0x0e
	OpReturn      Opcode = 0x0f
	OpCall        Opcode = 0x10
	OpCallIndirect Opcode = 0x11

	OpDrop   Opcode = 0x1a
	OpSelect Opcode = 0x1b

	OpLocalGet  Opcode = 0x20
	OpLocalSet  Opcode = 0x21
	OpLocalTee  Opcode = 0x22
	OpGlobalGet Opcode = 0x23
	OpGlobalSet Opcode = 0x24

	OpI32Load    Opcode = 0x28
	OpI64Load    Opcode = 0x29
	OpF32Load    Opcode = 0x2a
	OpF64Load    Opcode = 0x2b
	OpI32Load8S  Opcode = 0x2c
	OpI32Load8U  Opcode = 0x2d
	OpI32Load16S Opcode = 0x2e
	OpI32Load16U Opcode = 0x2f
	OpI64Load8S  Opcode = 0x30
	OpI64Load8U  Opcode = 0x31
	OpI64Load16S Opcode = 0x32
	OpI64Load16U Opcode = 0x33
	OpI64Load32S Opcode = 0x34
	OpI64Load32U Opcode = 0x35
	OpI32Store   Opcode = 0x36
	OpI64Store   Opcode = 0x37
	OpF32Store   Opcode = 0x38
	OpF64Store   Opcode = 0x39
	OpI32Store8  Opcode = 0x3a
	OpI32Store16 Opcode = 0x3b
	OpI64Store8  Opcode = 0x3c
	OpI64Store16 Opcode = 0x3d
	OpI64Store32 Opcode = 0x3e
	OpMemorySize Opcode = 0x3f
	OpMemoryGrow Opcode = 0x40

	OpI32Const Opcode = 0x41
	OpI64Const Opcode = 0x42
	OpF32Const Opcode = 0x43
	OpF64Const Opcode = 0x44

	OpI32Eqz Opcode = 0x45
	OpI32Eq  Opcode = 0x46
	OpI32Ne  Opcode = 0x47
	OpI32LtS Opcode = 0x48
	OpI32LtU Opcode = 0x49
	OpI32GtS Opcode = 0x4a
	OpI32GtU Opcode = 0x4b
	OpI32LeS Opcode = 0x4c
	OpI32LeU Opcode = 0x4d
	OpI32GeS Opcode = 0x4e
	OpI32GeU Opcode = 0x4f

	OpI64Eqz Opcode = 0x50
	OpI64Eq  Opcode = 0x51
	OpI64Ne  Opcode = 0x52
	OpI64LtS Opcode = 0x53
	OpI64LtU Opcode = 0x54
	OpI64GtS Opcode = 0x55
	OpI64GtU Opcode = 0x56
	OpI64LeS Opcode = 0x57
	OpI64LeU Opcode = 0x58
	OpI64GeS Opcode = 0x59
	OpI64GeU Opcode = 0x5a

	OpI32Clz    Opcode = 0x67
	OpI32Ctz    Opcode = 0x68
	OpI32Popcnt Opcode = 0x69
	OpI32Add    Opcode = 0x6a
	OpI32Sub    Opcode = 0x6b
	OpI32Mul    Opcode = 0x6c
	OpI32DivS   Opcode = 0x6d
	OpI32DivU   Opcode = 0x6e
	OpI32RemS   Opcode = 0x6f
	OpI32RemU   Opcode = 0x70
	OpI32And    Opcode = 0x71
	OpI32Or     Opcode = 0x72
	OpI32Xor    Opcode = 0x73
	OpI32Shl    Opcode = 0x74
	OpI32ShrS   Opcode = 0x75
	OpI32ShrU   Opcode = 0x76
	OpI32Rotl   Opcode = 0x77
	OpI32Rotr   Opcode = 0x78

	OpI64Clz    Opcode = 0x79
	OpI64Ctz    Opcode = 0x7a
	OpI64Popcnt Opcode = 0x7b
	OpI64Add    Opcode = 0x7c
	OpI64Sub    Opcode = 0x7d
	OpI64Mul    Opcode = 0x7e
	OpI64DivS   Opcode = 0x7f
	OpI64DivU   Opcode = 0x80
	OpI64RemS   Opcode = 0x81
	OpI64RemU   Opcode = 0x82
	OpI64And    Opcode = 0x83
	OpI64Or     Opcode = 0x84
	OpI64Xor    Opcode = 0x85
	OpI64Shl    Opcode = 0x86
	OpI64ShrS   Opcode = 0x87
	OpI64ShrU   Opcode = 0x88
	OpI64Rotl   Opcode = 0x89
	OpI64Rotr   Opcode = 0x8a

	OpF64Eq Opcode = 0x60
	OpF64Ne Opcode = 0x61
	OpF64Lt Opcode = 0x62
	OpF64Gt Opcode = 0x63
	OpF64Le Opcode = 0x64
	OpF64Ge Opcode = 0x65

	OpI32WrapI64      Opcode = 0xa7
	OpI64ExtendI32S   Opcode = 0xac
	OpI64ExtendI32U   Opcode = 0xad
	OpI64ReinterpretF64 Opcode = 0xbd
	OpF64ReinterpretI64 Opcode = 0xbf

	OpMiscPrefix Opcode = 0xfc

	OpI32TruncSatF32S Opcode = 0xfc00
	OpI32TruncSatF32U Opcode = 0xfc01
	OpI32TruncSatF64S Opcode = 0xfc02
	OpI32TruncSatF64U Opcode = 0xfc03
	OpI64TruncSatF32S Opcode = 0xfc04
	OpI64TruncSatF32U Opcode = 0xfc05
	OpI64TruncSatF64S Opcode = 0xfc06
	OpI64TruncSatF64U Opcode = 0xfc07

	OpMemoryInit Opcode = 0xfc08
	OpDataDrop   Opcode = 0xfc09
	OpMemoryCopy Opcode = 0xfc0a
	OpMemoryFill Opcode = 0xfc0b
	OpTableInit  Opcode = 0xfc0c
	OpElemDrop   Opcode = 0xfc0d
	OpTableCopy  Opcode = 0xfc0e
	OpTableGrow  Opcode = 0xfc0f
	OpTableSize  Opcode = 0xfc10
	OpTableFill  Opcode = 0xfc11
)

var OpcodeNames = map[Opcode]string{
	0x00: "unreachable",
	0x01: "nop",
	0x02: "block",
	0x03: "loop",
	0x04: "if",
	0x05: "else",
	0x0b: "end",
	0x0c: "br",
	0x0d: "br_if",
	0x0e: "br_table",
	0x0f: "return",
	0x10: "call",
	0x11: "call_indirect",

	0x1a: "drop",
	0x1b: "select",

	0x20: "local.get",
	0x21: "local.set",
	0x22: "local.tee",
	0x23: "global.get",
	0x24: "global.set",

	0x28: "i32.load",
	0x29: "i64.load",
	0x2a: "f32.load",
	0x2b: "f64.load",
	0x2c: "i32.load8_s",
	0x2d: "i32.load8_u",
	0x2e: "i32.load16_s",
	0x2f: "i32.load16_u",
	0x30: "i64.load8_s",
	0x31: "i64.load8_u",
	0x32: "i64.load16_s",
	0x33: "i64.load16_u",
	0x34: "i64.load32_s",
	0x35: "i64.load32_u",
	0x36: "i32.store",
	0x37: "i64.store",
	0x38: "f32.store",
	0x39: "f64.store",
	0x3a: "i32.store8",
	0x3b: "i32.store16",
	0x3c: "i64.store8",
	0x3d: "i64.store16",
	0x3e: "i64.store32",
	0x3f: "memory.size",
	0x40: "memory.grow",

	0x41: "i32.const",
	0x42: "i64.const",
	0x43: "f32.const",
	0x44: "f64.const",

	0x45: "i32.eqz",
	0x46: "i32.eq",
	0x47: "i32.ne",
	0x48: "i32.lt_s",
	0x49: "i32.lt_u",
	0x4a: "i32.gt_s",
	0x4b: "i32.gt_u",
	0x4c: "i32.le_s",
	0x4d: "i32.le_u",
	0x4e: "i32.ge_s",
	0x4f: "i32.ge_u",

	0x50: "i64.eqz",
	0x51: "i64.eq",
	0x52: "i64.ne",
	0x53: "i64.lt_s",
	0x54: "i64.lt_u",
	0x55: "i64.gt_s",
	0x56: "i64.gt_u",
	0x57: "i64.le_s",
	0x58: "i64.le_u",
	0x59: "i64.ge_s",
	0x5a: "i64.ge_u",

	0x67: "i32.clz",
	0x68: "i32.ctz",
	0x69: "i32.popcnt",
	0x6a: "i32.add",
	0x6b: "i32.sub",
	0x6c: "i32.mul",
	0x6d: "i32.div_s",
	0x6e: "i32.div_u",
	0x6f: "i32.rem_s",
	0x70: "i32.rem_u",
	0x71: "i32.and",
	0x72: "i32.or",
	0x73: "i32.xor",
	0x74: "i32.shl",
	0x75: "i32.shr_s",
	0x76: "i32.shr_u",
	0x77: "i32.rotl",
	0x78: "i32.rotr",

	0x79: "i64.clz",
	0x7a: "i64.ctz",
	0x7b: "i64.popcnt",
	0x7c: "i64.add",
	0x7d: "i64.sub",
	0x7e: "i64.mul",
	0x7f: "i64.div_s",
	0x80: "i64.div_u",
	0x81: "i64.rem_s",
	0x82: "i64.rem_u",
	0x83: "i64.and",
	0x84: "i64.or",
	0x85: "i64.xor",
	0x86: "i64.shl",
	0x87: "i64.shr_s",
	0x88: "i64.shr_u",
	0x89: "i64.rotl",
	0x8a: "i64.rotr",

	0x60: "f64.eq",
	0x61: "f64.ne",
	0x62: "f64.lt",
	0x63: "f64.gt",
	0x64: "f64.le",
	0x65: "f64.ge",

	0xa7: "i32.wrap_i64",
	0xac: "i64.extend_i32_s",
	0xad: "i64.extend_i32_u",
	0xbd: "i64.reinterpret_f64",
	0xbf: "f64.reinterpret_i64",

	0xfc00: "i32.trunc_sat_f32_s",
	0xfc01: "i32.trunc_sat_f32_u",
	0xfc02: "i32.trunc_sat_f64_s",
	0xfc03: "i32.trunc_sat_f64_u",
	0xfc04: "i64.trunc_sat_f32_s",
	0xfc05: "i64.trunc_sat_f32_u",
	0xfc06: "i64.trunc_sat_f64_s",
	0xfc07: "i64.trunc_sat_f64_u",

	0xfc08: "memory.init",
	0xfc09: "data.drop",
	0xfc0a: "memory.copy",
	0xfc0b: "memory.fill",
	0xfc0c: "table.init",
	0xfc0d: "elem.drop",
	0xfc0e: "table.copy",
	0xfc0f: "table.grow",
	0xfc10: "table.size",
	0xfc11: "table.fill",
}
