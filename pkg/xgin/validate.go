package xgin

import (
	"browsertools/log"
	"browsertools/pkg/errors"
	"context"
	"encoding/json"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var _validate *validator.Validate

func init() {
	_validate = validator.New()
}

func MustBindContext(ctx *gin.Context, req interface{}) {
	err := ContextBindWithValid(ctx, req)
	if err != nil {
		log.Errorf("check argument error:%v", err)
		panic(errors.ErrArgument)
	}
}

func MustBindQuery(ctx *gin.Context, req interface{}) {
	err := ContextBindQueryWithValid(ctx, req)
	if err != nil {
		log.Errorf("check argument error:%v", err)
		panic(errors.ErrArgument)
	}
}

func ContextBindWithValid(ctx *gin.Context, obj interface{}) (err error) {
	err = ctx.ShouldBind(obj)
	if err != nil {
		return err
	}

	PrintInterface(ctx, obj)

	if _validate != nil {
		err = _validate.Struct(obj)
	}

	return err
}

func ContextBindQueryWithValid(ctx *gin.Context, obj interface{}) (err error) {
	err = ctx.ShouldBindQuery(obj)
	if err != nil {
		return err
	}

	PrintInterface(ctx, obj)

	if _validate != nil {
		err = _validate.Struct(obj)
	}

	return err
}

func TelephoneValid(phone string) bool {
	// reg := `^1([387][0-9]|14[579]|5[^4]|16[6]|7[1-35-8]|9[189])\d{8}$`
	reg := `^1\d{10}$`

	rgx := regexp.MustCompile(reg)

	return rgx.MatchString(phone)
}

func PrintInterface(ctx context.Context, i interface{}) {
	buf, err := json.Marshal(i)
	if err == nil {
		log.Println(string(buf))
	}
}
