package wasm

func ParseStartSection(data []byte, baseOffset int) (uint32, error) {
	if len(data) == 0 {
		return 0, newError(ErrTruncated, int64(baseOffset), "empty start section")
	}

	funcIdx, _, err := ReadLEB128U32FromSlice(data)
	if err != nil {
		return 0, wrapError(ErrInvalidLEB128, int64(baseOffset), err, "start function index")
	}

	return funcIdx, nil
}
