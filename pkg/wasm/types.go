package wasm

type SectionID byte

const (
	SectionCustom   SectionID = 0
	SectionType     SectionID = 1
	SectionImport   SectionID = 2
	SectionFunction SectionID = 3
	SectionTable    SectionID = 4
	SectionMemory   SectionID = 5
	SectionGlobal   SectionID = 6
	SectionExport   SectionID = 7
	SectionStart    SectionID = 8
	SectionElement  SectionID = 9
	SectionCode     SectionID = 10
	SectionData     SectionID = 11
)

type Module struct {
	Version  uint32
	Sections []Section
}

type Section struct {
	ID      SectionID
	Offset  uint64
	Size    uint32
	Content []byte
}

type Instruction struct {
	Offset     uint64
	Opcode     Opcode
	Name       string
	Immediates []any
}

type ValType byte

const (
	ValI32 ValType = 0x7F
	ValI64 ValType = 0x7E
	ValF32 ValType = 0x7D
	ValF64 ValType = 0x7C
)

func (v ValType) String() string {
	switch v {
	case ValI32:
		return "i32"
	case ValI64:
		return "i64"
	case ValF32:
		return "f32"
	case ValF64:
		return "f64"
	default:
		return "unknown"
	}
}

type FuncType struct {
	Params  []ValType
	Results []ValType
}

func (f *FuncType) String() string {
	s := "(func"
	if len(f.Params) > 0 {
		s += " (param"
		for _, p := range f.Params {
			s += " " + p.String()
		}
		s += ")"
	}
	if len(f.Results) > 0 {
		s += " (result"
		for _, r := range f.Results {
			s += " " + r.String()
		}
		s += ")"
	}
	s += ")"
	return s
}

type ExportKind byte

const (
	ExportFunc   ExportKind = 0x00
	ExportTable  ExportKind = 0x01
	ExportMemory ExportKind = 0x02
	ExportGlobal ExportKind = 0x03
)

type Export struct {
	Name  string
	Kind  ExportKind
	Index uint32
}

type ImportKind byte

const (
	ImportFunc   ImportKind = 0x00
	ImportTable  ImportKind = 0x01
	ImportMemory ImportKind = 0x02
	ImportGlobal ImportKind = 0x03
)

type Limits struct {
	Min    uint32
	Max    uint32
	HasMax bool
}

type GlobalType struct {
	Type    ValType
	Mutable bool
}

type Import struct {
	Module string
	Name   string
	Kind   ImportKind
	TypeIdx   uint32
	Table     *Limits
	Memory    *Limits
	Global    *GlobalType
}
