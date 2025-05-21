package timeout

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTimeout(t *testing.T) {
	WithTime(time.Second)
}

func testResponse(c *gin.Context) {
	c.String(http.StatusRequestTimeout, "test response")
}

func TestCustomResponse(t *testing.T) {
	WithResponse(func(context *gin.Context) {
		// nothing to do
	})
}

func TestWithCallback(t *testing.T) {
	WithCallback(func(context *gin.Context) {
		// nothing to do
	})
}

func TestSuccess(t *testing.T) {
	r := gin.New()
	r.GET("/", New(
		WithTime(1*time.Second),
		WithResponse(testResponse),
	))

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "", w.Body.String())
}

func TestReset(t *testing.T) {
	reset := Reset(WithTime(time.Microsecond))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	reset(c)
}

func TestWithResponse(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	defaultTimeout.response(c)
}
