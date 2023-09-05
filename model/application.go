package model

import "github.com/rhine-tech/scene/errcode"

type AppResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func NewOkResponse() AppResponse {
	return AppResponse{Code: 0, Msg: "ok", Data: nil}
}

func NewDataResponse(data interface{}) AppResponse {
	return AppResponse{Code: 0, Msg: "ok", Data: data}
}

func NewErrorResponse(code int, err error) AppResponse {
	return AppResponse{Code: code, Msg: err.Error(), Data: nil}
}

func NewErrorCodeResponse(err *errcode.Error) AppResponse {
	return AppResponse{Code: err.Code, Msg: err.Error(), Data: nil}
}

func TryErrorCodeResponse(err error) AppResponse {
	ec, ok := err.(*errcode.Error)
	if ok {
		return NewErrorCodeResponse(ec)
	}
	return NewErrorResponse(-1, err)
}
