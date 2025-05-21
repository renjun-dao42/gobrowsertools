package httpserver

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"testbrowser/pkg/response"
)

type Server struct {
	addr   string
	router *gin.Engine
}

func New(addr string) *Server {
	router := gin.New()
	ctrl := NewController()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.New("ok"))
	})

	browser := router.Group("/browser")
	{
		browser.POST("/screenshot", ctrl.Screenshot)
		browser.POST("/openTab", ctrl.OpenTab)
		browser.POST("/getConsoleLogs", ctrl.GetConsoleLogs)
	}

	return &Server{addr: addr, router: router}
}

func (s *Server) Start() {
	if err := s.router.Run(s.addr); err != nil {
		panic(fmt.Sprintf("start Http server [%s] error:%v", s.addr, err))
	}
}
