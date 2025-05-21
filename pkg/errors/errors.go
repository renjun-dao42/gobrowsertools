package errors

import (
	"bytes"
	"sync"
)

type Errors struct {
	errs []error
}

func NewErrors(inErrors ...error) *Errors {
	var errs []error

	for _, err := range inErrors {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return &Errors{errs: errs}
}

func (errs *Errors) Add(err error) {
	if err != nil {
		errs.errs = append(errs.errs, err)
	}
}

func (errs *Errors) Err() error {
	if len(errs.errs) == 0 {
		return nil
	}

	return errs
}

func (errs *Errors) Error() string {
	buf := bytes.NewBuffer(nil)

	for _, err := range errs.errs {
		if buf.Len() > 0 {
			buf.WriteString(", ")
		}

		buf.WriteString(err.Error())
	}

	return buf.String()
}

type SafeErrors struct {
	mu   *sync.Mutex
	errs *Errors
}

func NewSafeErrors(inErrors ...error) *SafeErrors {
	return &SafeErrors{errs: NewErrors(inErrors...), mu: &sync.Mutex{}}
}

func (e *SafeErrors) Add(err error) {
	if err != nil {
		e.mu.Lock()
		defer e.mu.Unlock()

		e.errs.Add(err)
	}
}

func (e *SafeErrors) Error() string {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.errs.Error()
}

func (e *SafeErrors) Err() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.errs.Err()
}
