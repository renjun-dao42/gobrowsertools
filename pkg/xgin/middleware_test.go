package xgin

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoggerWriter(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", bytes.NewBuffer(nil))

	LoggerWriter()(c)
}

func TestSkipHandler(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	SkipHandler(c)
}

func TestRecoveryWriter(t *testing.T) {
	c, e := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", bytes.NewBuffer(nil))

	e.Use(func(context *gin.Context) {
		panic(1)
	})

	RecoveryWriter()(c)
}
