package errors

import "fmt"

func Check(err error, msg ...interface{}) {
	if err != nil {
		throw(err, msg...)
	}
}

func Throw(err error, msg ...interface{}) {
	Assert(err != nil)

	throw(err, msg...)
}

func throw(err error, msg ...interface{}) {
	var (
		info      string
		code      int
		codeError CodeError
	)

	if As(err, &codeError) {
		info = codeError.Error()
		code = codeError.Code()
	} else {
		info = err.Error()
		code = InternalErrorCode
	}

	if len(msg) > 0 {
		info = fmt.Sprintf("%s, %v", info, msg)
	}

	panic(NewWithInfo(code, info))
}

func EqualCodeError(err error, codeError CodeError) bool {
	var err1 CodeError

	if As(err, &err1) {
		return err1.Code() == codeError.Code()
	}

	return false
}

func Ignore(_ error) {
	// ignore error
}

func Ignore1(_ interface{}, err error) {
	// ignore error
}

func Ignore2(_ interface{}, _ interface{}, err error) {
	// ignore error
}

func Ignore3(_ interface{}, _ interface{}, _ interface{}, err error) {
	// ignore error
}

func Unreachable() {
	panic("unreachable!")
}

func Assert(exp bool) { // nolint
	if !exp {
		Unreachable()
	}
}

func Unimplemented() {
	panic("unimplemented")
}
