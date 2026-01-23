package wasm

const (
	NameSubsectionModule   = 0
	NameSubsectionFunction = 1
	NameSubsectionLocal    = 2
)

func ParseNameSection(content []byte, baseOffset int) (*NameMap, error) {
	p := &parser{data: content}

	nm := &NameMap{
		FunctionNames: make(map[uint32]string),
		LocalNames:    make(map[uint32]map[uint32]string),
	}

	for p.offset < len(p.data) {
		subsectionID, err := p.readByte()
		if err != nil {
			break
		}

		size, err := p.readU32()
		if err != nil {
			return nm, nil
		}

		subsectionEnd := p.offset + int(size)
		if subsectionEnd > len(p.data) {
			return nm, nil
		}

		switch subsectionID {
		case NameSubsectionFunction:
			count, err := p.readU32()
			if err != nil {
				p.offset = subsectionEnd
				continue
			}

			for i := uint32(0); i < count && p.offset < subsectionEnd; i++ {
				idx, err := p.readU32()
				if err != nil {
					break
				}
				name, err := p.readString()
				if err != nil {
					break
				}
				nm.FunctionNames[idx] = name
			}

		case NameSubsectionLocal:
			count, err := p.readU32()
			if err != nil {
				p.offset = subsectionEnd
				continue
			}

			for i := uint32(0); i < count && p.offset < subsectionEnd; i++ {
				funcIdx, err := p.readU32()
				if err != nil {
					break
				}

				localCount, err := p.readU32()
				if err != nil {
					break
				}

				locals := make(map[uint32]string)
				for j := uint32(0); j < localCount && p.offset < subsectionEnd; j++ {
					localIdx, err := p.readU32()
					if err != nil {
						break
					}
					name, err := p.readString()
					if err != nil {
						break
					}
					locals[localIdx] = name
				}
				nm.LocalNames[funcIdx] = locals
			}
		}

		p.offset = subsectionEnd
	}

	return nm, nil
}
