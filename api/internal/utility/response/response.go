// Package response 统一 HTTP 响应格式，与前端约定一致：{ code, msg, data }
package response

import (
	"github.com/gogf/gf/v2/net/ghttp"
)

// 响应码
const (
	CodeSuccess = "0000" // 成功
	CodeError   = "5000" // 业务/系统异常
	CodeLogout  = "8888" // 登录失效，前端需要退出到登录页
)

// Response 统一响应结构
type Response struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Middleware 统一响应格式中间件
func Middleware(r *ghttp.Request) {
	r.Middleware.Next()

	// 已有输出（例如直接写入的 HTML、文件流）不再包装
	if r.Response.BufferLength() > 0 {
		return
	}

	var (
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = CodeSuccess
		msg  = ""
	)
	if err != nil {
		code = CodeError
		msg = err.Error()
	}

	r.Response.WriteJson(Response{
		Code: code,
		Msg:  msg,
		Data: res,
	})
}
