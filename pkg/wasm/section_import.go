package wasm

func ParseImportSection(content []byte, baseOffset int) ([]Import, error) {
	p := &parser{data: content}

	count, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read import count")
	}

	imports := make([]Import, 0, count)

	for i := 0; i < int(count); i++ {
		modLen, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read module name length")
		}
		modBytes, err := p.readBytes(int(modLen))
		if err != nil {
			return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read module name")
		}

		nameLen, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read import name length")
		}
		nameBytes, err := p.readBytes(int(nameLen))
		if err != nil {
			return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read import name")
		}

		kind, err := p.readByte()
		if err != nil {
			return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read import kind")
		}

		imp := Import{
			Module: string(modBytes),
			Name:   string(nameBytes),
			Kind:   ImportKind(kind),
		}

		switch ImportKind(kind) {
		case ImportFunc:
			typeIdx, err := p.readU32()
			if err != nil {
				return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read func type index")
			}
			imp.TypeIdx = typeIdx

		case ImportTable:
			_, err := p.readByte()
			if err != nil {
				return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read table elemtype")
			}
			limits, err := p.readLimits(baseOffset)
			if err != nil {
				return nil, err
			}
			imp.Table = limits

		case ImportMemory:
			limits, err := p.readLimits(baseOffset)
			if err != nil {
				return nil, err
			}
			imp.Memory = limits

		case ImportGlobal:
			valType, err := p.readByte()
			if err != nil {
				return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read global type")
			}
			mut, err := p.readByte()
			if err != nil {
				return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read global mutability")
			}
			imp.Global = &GlobalType{
				Type:    ValType(valType),
				Mutable: mut == 1,
			}

		default:
			return nil, newError(ErrInvalidSection, int64(baseOffset+p.offset), "unknown import kind 0x%02x", kind)
		}

		imports = append(imports, imp)
	}

	return imports, nil
}

func (p *parser) readLimits(baseOffset int) (*Limits, error) {
	flags, err := p.readByte()
	if err != nil {
		return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read limits flags")
	}

	min, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read limits min")
	}

	limits := &Limits{Min: min}

	if flags&0x01 != 0 {
		max, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read limits max")
		}
		limits.Max = max
		limits.HasMax = true
	}

	return limits, nil
}
