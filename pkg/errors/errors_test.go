package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewErrors(t *testing.T) {
	errs := NewErrors(New("err1"))
	errs.Add(New("err2"))
	str := errs.Error()
	assert.Equal(t, str, "err1, err2")

	err := errs.Err()
	assert.Error(t, err)
}

func TestNewSafeErrors(t *testing.T) {
	errs := NewSafeErrors(New("err1"))
	errs.Add(New("err2"))
	str := errs.Error()
	assert.Equal(t, str, "err1, err2")

	err := errs.Err()
	assert.Error(t, err)
}
