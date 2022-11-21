package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"smartest-go/pkg/e"
	util "smartest-go/pkg/util/const"
)

type Response struct {
	Code         int         `json:"code"`
	Msg          string      `json:"msg"`
	ErrMsg       string      `json:"errMsg"`
	ResponseData interface{} `json:"data"`
}

func ErrorResp(c *gin.Context, code int, errMsg string, data interface{}) {
	resp(c, http.StatusOK, code, errMsg, data)
}

func UnauthorizedResp(c *gin.Context, code int, errMsg string) {
	resp(c, http.StatusUnauthorized, code, errMsg, nil)
}

func SuccessRespByCode(c *gin.Context, code int, data interface{}) {
	resp(c, http.StatusOK, code, "", data)
}

func SuccessResponseData(c *gin.Context, count, data interface{}) {
	SuccessResp(c, struct {
		Count interface{} `json:"count"`
		Data  interface{} `json:"data"`
	}{
		Count: count,
		Data:  data,
	})
}

func SuccessResp(c *gin.Context, data interface{}) {
	resp(c, http.StatusOK, e.SUCCESS, "", data)
}

func SuccessPureResp(c *gin.Context, data interface{}) {
	pureResp(c, http.StatusOK, e.SUCCESS, "", data)
}

func resp(c *gin.Context, httpCode, code int, errMsg string, data interface{}) {
	resp := Response{
		Code:         code,
		Msg:          e.GetMsg(code),
		ErrMsg:       errMsg,
		ResponseData: data,
	}
	c.Set(util.LogResponse, &resp)
	c.JSON(httpCode, resp)
}

func pureResp(c *gin.Context, httpCode, code int, errMsg string, data interface{}) {
	resp := Response{
		Code:         code,
		Msg:          e.GetMsg(code),
		ErrMsg:       errMsg,
		ResponseData: data,
	}
	c.Set(util.LogResponse, &resp)
	c.PureJSON(httpCode, resp)
}
