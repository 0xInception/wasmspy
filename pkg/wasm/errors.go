package wasm

import "fmt"

type ErrorCode int

const (
	ErrInvalidMagic ErrorCode = iota + 1
	ErrInvalidVersion
	ErrTruncated
	ErrInvalidLEB128
	ErrInvalidOpcode
	ErrInvalidSection
	ErrInvalidIndex
	ErrSectionOverflow
)

type ParseError struct {
	Code    ErrorCode
	Msg     string
	Offset  int64
	Section SectionID
	Cause   error
}

func (e *ParseError) Error() string {
	if e.Offset >= 0 {
		return fmt.Sprintf("%s at offset 0x%x", e.Msg, e.Offset)
	}
	return e.Msg
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

func newError(code ErrorCode, offset int64, format string, args ...any) *ParseError {
	return &ParseError{
		Code:   code,
		Msg:    fmt.Sprintf(format, args...),
		Offset: offset,
	}
}

func wrapError(code ErrorCode, offset int64, cause error, format string, args ...any) *ParseError {
	return &ParseError{
		Code:   code,
		Msg:    fmt.Sprintf(format, args...),
		Offset: offset,
		Cause:  cause,
	}
}
