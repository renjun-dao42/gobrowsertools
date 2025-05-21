package xgin

import (
	"agent/pkg/errors"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func initGinMetrics() (*gin.Engine, *httptest.ResponseRecorder) {
	r := New()

	// RecordMetric interface
	r.GET("/testRecordMetrics", func(c *gin.Context) {
		// c.String(backend.StatusOK, )
	})

	w := httptest.NewRecorder()

	return r, w
}

func TestRecordMetrics(t *testing.T) {
	tests := []struct {
		name     string
		wantCode int
		wantMsg  string
	}{
		{
			"unitTestRecordMetrics",
			200,
			"",
		},
	}

	r, w := initGinMetrics()
	req, _ := http.NewRequestWithContext(context.TODO(), "GET", "/testRecordMetrics", nil)
	r.ServeHTTP(w, req)

	for _, tt := range tests {
		assert.Equal(t, tt.wantCode, w.Code)
		assert.Equal(t, tt.wantMsg, w.Body.String())
	}
}

func TestTriggerCode500(t *testing.T) {
	TriggerErrorCode("/test", errors.InternalError)
}
