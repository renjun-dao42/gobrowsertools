package response

import "testbrowser/pkg/errors"

const SuccessCode = 200

const SuccessMsg = "Succeed"

type Response struct {
	ErrorCode   int         `json:"code"`
	Description string      `json:"msg"`
	Data        interface{} `json:"data,omitempty"`
}

func New(data interface{}) *Response {
	resp := Response{
		ErrorCode:   SuccessCode,
		Description: SuccessMsg,
		Data:        data,
	}

	return &resp
}

func Fail(code int, desc string) *Response {
	return &Response{
		ErrorCode:   code,
		Description: desc,
		Data:        nil,
	}
}

func Err(err error) *Response {
	if err == nil {
		return &Response{
			ErrorCode:   errors.InternalErrorCode,
			Description: "Empty error message!",
		}
	}

	var codeError errors.CodeError
	if errors.As(err, &codeError) {
		return &Response{
			ErrorCode:   codeError.Code(),
			Description: codeError.Error(),
			Data:        nil,
		}
	}

	return &Response{
		ErrorCode:   errors.InternalErrorCode,
		Description: err.Error(),
	}
}

func (resp *Response) GetError() error {
	if resp.ErrorCode == SuccessCode {
		return nil
	}

	return errors.NewWithInfo(resp.ErrorCode, resp.Description)
}
