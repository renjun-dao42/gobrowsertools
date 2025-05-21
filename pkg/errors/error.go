package errors

import (
	pkgerr "github.com/pkg/errors"
)

var (
	New          = pkgerr.New
	Wrap         = pkgerr.Wrap
	Wrapf        = pkgerr.Wrapf
	Errorf       = pkgerr.Errorf
	WithStack    = pkgerr.WithStack
	WithMessage  = pkgerr.WithMessage
	WithMessagef = pkgerr.WithMessagef
	Cause        = pkgerr.Cause
	Is           = pkgerr.Is
	As           = pkgerr.As
	Unwrap       = pkgerr.Unwrap
)

type CodeError interface {
	error
	Code() int
}

const (
	InternalErrorCode = 500 // Internal error
	RequestTimeout    = 408 // Request timeout
)

var InternalError = &SvrError{code: InternalErrorCode, info: "internal error"}

type SvrError struct {
	code int
	info string
}

func (e *SvrError) Error() string {
	return e.info
}

func (e *SvrError) Code() int {
	return e.code
}

func NewWithCode(code int, err error) CodeError {
	var codeError CodeError
	if As(err, &codeError) {
		return codeError
	}

	return &SvrError{
		code: code,
		info: err.Error(),
	}
}

func NewWithInfo(code int, info string) CodeError {
	return &SvrError{
		code: code,
		info: info,
	}
}
