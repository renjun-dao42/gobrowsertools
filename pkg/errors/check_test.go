package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqual(t *testing.T) {
	var err error

	ce1 := NewWithInfo(6001, "code error 1")
	ce2 := NewWithInfo(6002, "code error 2")

	err = New("native err")
	assert.False(t, EqualCodeError(err, ce1))

	err = WithMessage(ce1, "with error message")
	err = WithStack(err)

	assert.True(t, EqualCodeError(err, ce1))
	assert.False(t, EqualCodeError(err, ce2))

	assert.False(t, EqualCodeError(nil, ce1))
}

func add(a int, b int) (int, error) {
	if b == 0 {
		return 0, New("b == 0")
	}

	return a + b, nil
}
func TestIgnore(t *testing.T) {
	Ignore1(add(1, 0))
	Ignore2(1, 0, nil)
	Ignore3(1, 1, 0, nil)
}

func TestUnreachable(t *testing.T) {
	assert.Panics(t, Unreachable)
}

func TestAssert(t *testing.T) {
	assert.Panics(t, func() {
		Assert(false)
	})

	assert.NotPanics(t, func() {
		Assert(true)
	})
}

func TestUnimplemented(t *testing.T) {
	assert.Panics(t, Unimplemented)
}

func TestCheck(t *testing.T) {
	assert.Panics(t, func() {
		Check(New("panic"))
	})

	assert.Panics(t, func() {
		Check(NewWithInfo(500, "unauth"), "!auth")
	})
}

func TestThrow(t *testing.T) {
	assert.Panics(t, func() {
		Throw(New("fsf"))
	})
}
