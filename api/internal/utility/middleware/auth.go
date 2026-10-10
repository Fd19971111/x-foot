// Package middleware 提供 HTTP 中间件
package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"

	"x-foot/internal/utility/response"
	"x-foot/internal/utility/token"
)

// whiteList 无需登录即可访问的接口
//
// 注意：/route/getConstantRoutes 必须放行。
// 它就是「登录前可访问的内置路由」（登录页、403/404/500、外链页），
// 前端路由守卫在每次导航（包括未登录访问登录页）都会先取一次；
// 若返回 8888，前端会按「登录失效」处理并触发登出重置，导致退出后无法再登录。
var whiteList = []string{
	"/auth/login",
	"/auth/refreshToken",
	"/route/getConstantRoutes",
	"/hello",
}

// Auth 登录鉴权中间件：校验 Authorization 头中的令牌，并把 userId 注入上下文
func Auth(r *ghttp.Request) {
	if gstr.InArray(whiteList, r.URL.Path) {
		r.Middleware.Next()
		return
	}

	tokenStr := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if tokenStr == "" {
		unauthorized(r, "未登录或登录状态已失效，请先登录")
		return
	}

	userId, err := token.Parse(tokenStr)
	if err != nil {
		unauthorized(r, "登录状态已过期，请重新登录")
		return
	}

	r.SetCtxVar("userId", userId)

	r.Middleware.Next()
}

func unauthorized(r *ghttp.Request, msg string) {
	r.Response.WriteJson(response.Response{
		Code: response.CodeLogout,
		Msg:  msg,
		Data: nil,
	})
	r.Exit()
}
