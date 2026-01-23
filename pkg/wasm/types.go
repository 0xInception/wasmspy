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
