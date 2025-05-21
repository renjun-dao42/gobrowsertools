package timeout

import (
	"browsertools/pkg/errors"
	"bytes"
	"sync"
)

// BufferPool is Pool of *bytes.Buffer
type BufferPool struct {
	pool sync.Pool
}

// Get a bytes.Buffer pointer
func (p *BufferPool) Get() *bytes.Buffer {
	v := p.pool.Get()
	if v == nil {
		return &bytes.Buffer{}
	}

	b, ok := v.(*bytes.Buffer)
	errors.Assert(ok)

	return b
}

// Put a bytes.Buffer pointer to BufferPool
func (p *BufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	p.pool.Put(buf)
}
