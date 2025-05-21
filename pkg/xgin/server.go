package xgin

import (
	"browsertools/pkg/errors"
	"time"

	"github.com/gin-gonic/gin"
)

func New(middleware ...gin.HandlerFunc) *gin.Engine {
	middleware = append(middleware,
		LoggerWriter(),
		Timeout(time.Second*30, errors.ErrRequestTimeout),
		RecoveryWriter())

	router := gin.New()
	router.Use(middleware...)
	router.NoRoute(HandleNotFound)
	router.NoMethod(HandleNotFound)

	return router
}
