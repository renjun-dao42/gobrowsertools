package timeout

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	bufPool        = &BufferPool{}
	defaultTimeout = &Timeout{
		timeout: 5 * time.Second,
		response: func(c *gin.Context) {
			c.String(http.StatusRequestTimeout, http.StatusText(http.StatusRequestTimeout))
		},
		callback: nil,
		local:    false,
	}
)

// New wraps a handler and aborts the process of the handler if the timeout is reached
func New(opts ...Option) gin.HandlerFunc {
	for _, opt := range opts {
		if opt == nil {
			panic("timeout Option not be nil")
		}

		opt(defaultTimeout)
	}

	return newTimeoutHandler(defaultTimeout)
}

func newTimeoutHandler(t *Timeout) gin.HandlerFunc {
	return func(c *gin.Context) {
		if t == nil || (!t.local && skip(c)) {
			c.Next()

			return
		}

		finish := make(chan struct{}, 1)
		panicChan := make(chan interface{}, 1)

		w := c.Writer
		buffer := bufPool.Get()
		tw := NewWriter(w, buffer)
		c.Writer = tw

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			c.Next()
			finish <- struct{}{}
		}()

		select {
		case p := <-panicChan:
			tw.FreeBuffer()

			c.Writer = w

			panic(p)

		case <-finish:
			tw.mu.Lock()
			defer tw.mu.Unlock()

			dst := tw.ResponseWriter.Header()

			for k, vv := range tw.Header() {
				dst[k] = vv
			}

			tw.ResponseWriter.WriteHeader(tw.code)

			if _, err := tw.ResponseWriter.Write(buffer.Bytes()); err != nil {
				panic(err)
			}

			tw.FreeBuffer()
			bufPool.Put(buffer)

		case <-time.After(t.timeout):
			c.Abort()
			tw.mu.Lock()
			defer tw.mu.Unlock()
			tw.timeout = true
			tw.FreeBuffer()
			bufPool.Put(buffer)

			c.Writer = w
			t.response(c)

			if t.callback != nil {
				t.callback(c)
			}

			c.Writer = tw
		}
	}
}

func Skip() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func Reset(opts ...Option) gin.HandlerFunc {
	t := *defaultTimeout

	for _, opt := range opts {
		if opt == nil {
			panic("timeout Option not be nil")
		}

		opt(&t)
	}

	t.local = true

	return func(c *gin.Context) {
		newTimeoutHandler(&t)(c)
	}
}

func skip(c *gin.Context) bool {
	skipFn := runtime.FuncForPC(reflect.ValueOf(Skip).Pointer()).Name()
	resetFn := runtime.FuncForPC(reflect.ValueOf(Reset).Pointer()).Name()

	for _, f := range c.HandlerNames() {
		if strings.HasPrefix(f, skipFn) || strings.HasPrefix(f, resetFn) {
			return true
		}
	}

	return false
}
