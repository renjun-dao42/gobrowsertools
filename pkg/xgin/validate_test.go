package xgin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func initGin(url string, value interface{}) *gin.Engine {
	r := gin.Default()
	r.GET(url, func(c *gin.Context) {
		err := ContextBindWithValid(c, value)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}
		c.String(http.StatusOK, "")
	})

	return r
}

func TestContextBindWithValid(t *testing.T) {
	// 前置条件
	tests := []struct {
		url      string
		value    interface{}
		wantCode int
		wantMsg  string
	}{
		{
			"/testFirstReturn",
			make(chan int),
			http.StatusBadRequest,
			"validator: (nil chan int)",
		},
		{
			"/testSecondReturn",
			&struct {
				key   string
				value string
			}{},
			http.StatusOK, ""},
	}

	for _, tt := range tests {
		r := initGin(tt.url, tt.value)
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.TODO(), "GET", tt.url, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, tt.wantCode, w.Code)
		assert.Equal(t, tt.wantMsg, w.Body.String())
	}
}

func TestTelephoneValid(t *testing.T) {
	type args struct {
		phone string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"unitText--true",
			args{phone: "13531184972"},
			true,
		},
		{"unitText--false",
			args{
				phone: "1111111",
			},
			false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			if got := TelephoneValid(tt.args.phone); got != tt.want {
				t.Errorf("TelephoneValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintInterface(t *testing.T) {
	type args struct {
		ctx context.Context
		i   interface{}
	}

	tests := []struct {
		name string
		args args
	}{
		{
			"unitTestError",
			args{
				ctx: context.Background(),
				i:   make(chan int),
			},
		},
		{
			"unitTestNoError",
			args{ctx: context.Background(),
				i: 1,
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			PrintInterface(tt.args.ctx, tt.args.i)
		})
	}
}

func TestContextBindQueryWithValid(t *testing.T) {
	// 前置条件
	tests := []struct {
		url      string
		value    interface{}
		wantCode int
		wantMsg  string
	}{
		{
			"/testFirstReturn",
			make(chan int),
			http.StatusBadRequest,
			"validator: (nil chan int)",
		},
		{
			"/testSecondReturn",
			&struct {
				key   string
				value string
			}{
				key:   "testKey",
				value: "testValue",
			},
			http.StatusOK,
			"",
		},
	}

	for _, tt := range tests {
		r := initGin(tt.url, tt.value)
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.TODO(), "GET", tt.url, nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, tt.wantCode, w.Code)
		assert.Equal(t, tt.wantMsg, w.Body.String())
	}
}
