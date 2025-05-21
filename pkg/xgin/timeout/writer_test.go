package timeout

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewWriter(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	writer := NewWriter(c.Writer, bytes.NewBuffer(nil))

	_, err := writer.Write([]byte("abc"))
	assert.NoError(t, err)

	writer.WriteHeader(200)

	_, err = writer.WriteString("bcd")
	assert.NoError(t, err)

	header := writer.Header()
	assert.NotNil(t, header)

	writer.FreeBuffer()
}
