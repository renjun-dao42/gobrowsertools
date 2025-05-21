package timeout

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Option for timeout
type Option func(*Timeout)

// WithTime set timeout
func WithTime(timeout time.Duration) Option {
	return func(t *Timeout) {
		t.timeout = timeout
	}
}

// WithResponse add gin handler
func WithResponse(h gin.HandlerFunc) Option {
	return func(t *Timeout) {
		t.response = h
	}
}

// WithCallback add gin handler
func WithCallback(h gin.HandlerFunc) Option {
	return func(t *Timeout) {
		t.callback = h
	}
}

type Timeout struct {
	timeout  time.Duration
	response gin.HandlerFunc
	callback gin.HandlerFunc
	local    bool
}
