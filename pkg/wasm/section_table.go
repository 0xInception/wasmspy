package wasm

func ParseTableSection(content []byte, baseOffset int) ([]Table, error) {
	p := &parser{data: content}

	count, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidLEB128, int64(baseOffset+p.offset), err, "table count")
	}

	tables := make([]Table, 0, count)
	for i := uint32(0); i < count; i++ {
		elemType, err := p.readByte()
		if err != nil {
			return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "element type")
		}

		lim, err := p.readLimits(baseOffset)
		if err != nil {
			return nil, err
		}

		tables = append(tables, Table{
			Type:   ElemType(elemType),
			Limits: *lim,
		})
	}

	return tables, nil
}
