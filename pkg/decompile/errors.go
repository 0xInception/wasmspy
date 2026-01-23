package decompile

import "fmt"

type ErrorCode int

const (
	ErrStackUnderflow ErrorCode = iota + 1
	ErrInvalidIndex
	ErrTypeMismatch
	ErrUnknownOpcode
)

var errorMessages = map[ErrorCode]string{
	ErrStackUnderflow: "stack underflow",
	ErrInvalidIndex:   "invalid index",
	ErrTypeMismatch:   "type mismatch",
	ErrUnknownOpcode:  "unknown opcode",
}

type AnalysisError struct {
	Code    ErrorCode
	Offset  uint64
	Opcode  string
	Details string
}

func (e *AnalysisError) Error() string {
	msg := errorMessages[e.Code]
	if e.Details != "" {
		msg = e.Details
	}
	return fmt.Sprintf("%s at offset 0x%x (%s)", msg, e.Offset, e.Opcode)
}

func newError(code ErrorCode, offset uint64, opcode string, format string, args ...any) *AnalysisError {
	return &AnalysisError{
		Code:    code,
		Offset:  offset,
		Opcode:  opcode,
		Details: fmt.Sprintf(format, args...),
	}
}
