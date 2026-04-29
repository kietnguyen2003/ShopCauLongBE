package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type apiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func successResponse(c *gin.Context, status int, msg string, data interface{}) {
	c.JSON(status, apiResponse{
		Code: status,
		Msg:  msg,
		Data: data,
	})
}

func errorResponse(c *gin.Context, status int, msg string) {
	if msg == "" {
		msg = http.StatusText(status)
	}

	c.JSON(status, apiResponse{
		Code: status,
		Msg:  msg,
		Data: nil,
	})
}
