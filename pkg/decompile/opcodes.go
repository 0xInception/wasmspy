package decompile

import "github.com/0xInception/wasmspy/pkg/wasm"

type Signature struct {
	Inputs  []wasm.ValType
	Outputs []wasm.ValType
}

var i32 = wasm.ValI32
var i64 = wasm.ValI64
var f32 = wasm.ValF32
var f64 = wasm.ValF64

var OpSignatures = map[wasm.Opcode]Signature{
	wasm.OpUnreachable: {},
	wasm.OpNop:         {},
	wasm.OpReturn:      {},
	wasm.OpEnd:         {},

	wasm.OpDrop:   {Inputs: []wasm.ValType{i32}},
	wasm.OpSelect: {Inputs: []wasm.ValType{i32, i32, i32}, Outputs: []wasm.ValType{i32}},

	wasm.OpI32Const: {Outputs: []wasm.ValType{i32}},
	wasm.OpI64Const: {Outputs: []wasm.ValType{i64}},
	wasm.OpF32Const: {Outputs: []wasm.ValType{f32}},
	wasm.OpF64Const: {Outputs: []wasm.ValType{f64}},

	wasm.OpI32Eqz:  {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Eq:   {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Ne:   {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32LtS:  {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32LtU:  {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32GtS:  {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32GtU:  {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32LeS:  {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32LeU:  {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32GeS:  {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32GeU:  {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},

	wasm.OpI64Eqz:  {Inputs: []wasm.ValType{i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64Eq:   {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64Ne:   {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64LtS:  {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64LtU:  {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64GtS:  {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64GtU:  {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64LeS:  {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64LeU:  {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64GeS:  {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64GeU:  {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i32}},

	wasm.OpI32Clz:    {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Ctz:    {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Popcnt: {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Add:    {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Sub:    {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Mul:    {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32DivS:   {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32DivU:   {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32RemS:   {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32RemU:   {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32And:    {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Or:     {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Xor:    {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Shl:    {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32ShrS:   {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32ShrU:   {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Rotl:   {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Rotr:   {Inputs: []wasm.ValType{i32, i32}, Outputs: []wasm.ValType{i32}},

	wasm.OpI64Clz:    {Inputs: []wasm.ValType{i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Ctz:    {Inputs: []wasm.ValType{i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Popcnt: {Inputs: []wasm.ValType{i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Add:    {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Sub:    {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Mul:    {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64DivS:   {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64DivU:   {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64RemS:   {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64RemU:   {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64And:    {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Or:     {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Xor:    {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Shl:    {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64ShrS:   {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64ShrU:   {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Rotl:   {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Rotr:   {Inputs: []wasm.ValType{i64, i64}, Outputs: []wasm.ValType{i64}},

	wasm.OpI32WrapI64:    {Inputs: []wasm.ValType{i64}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64ExtendI32S: {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64ExtendI32U: {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i64}},

	wasm.OpI32Load:    {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64Load:    {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i64}},
	wasm.OpF32Load:    {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{f32}},
	wasm.OpF64Load:    {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{f64}},
	wasm.OpI32Load8S:  {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Load8U:  {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Load16S: {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI32Load16U: {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i32}},
	wasm.OpI64Load8S:  {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Load8U:  {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Load16S: {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Load16U: {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Load32S: {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i64}},
	wasm.OpI64Load32U: {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i64}},

	wasm.OpI32Store:   {Inputs: []wasm.ValType{i32, i32}},
	wasm.OpI64Store:   {Inputs: []wasm.ValType{i32, i64}},
	wasm.OpF32Store:   {Inputs: []wasm.ValType{i32, f32}},
	wasm.OpF64Store:   {Inputs: []wasm.ValType{i32, f64}},
	wasm.OpI32Store8:  {Inputs: []wasm.ValType{i32, i32}},
	wasm.OpI32Store16: {Inputs: []wasm.ValType{i32, i32}},
	wasm.OpI64Store8:  {Inputs: []wasm.ValType{i32, i64}},
	wasm.OpI64Store16: {Inputs: []wasm.ValType{i32, i64}},
	wasm.OpI64Store32: {Inputs: []wasm.ValType{i32, i64}},

	wasm.OpMemorySize: {Outputs: []wasm.ValType{i32}},
	wasm.OpMemoryGrow: {Inputs: []wasm.ValType{i32}, Outputs: []wasm.ValType{i32}},

	wasm.OpBr:   {},
	wasm.OpBrIf: {Inputs: []wasm.ValType{i32}},
}
