package wasm

func ParseExportSection(content []byte, baseOffset int) ([]Export, error) {
	p := &parser{data: content}

	count, err := p.readU32()
	if err != nil {
		return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read export count")
	}

	exports := make([]Export, 0, count)

	for i := 0; i < int(count); i++ {
		nameLen, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read export name length")
		}

		nameBytes, err := p.readBytes(int(nameLen))
		if err != nil {
			return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read export name")
		}

		kind, err := p.readByte()
		if err != nil {
			return nil, wrapError(ErrTruncated, int64(baseOffset+p.offset), err, "failed to read export kind")
		}

		index, err := p.readU32()
		if err != nil {
			return nil, wrapError(ErrInvalidSection, int64(baseOffset+p.offset), err, "failed to read export index")
		}

		exports = append(exports, Export{
			Name:  string(nameBytes),
			Kind:  ExportKind(kind),
			Index: index,
		})
	}

	return exports, nil
}
