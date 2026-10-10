package cmd

import (
	"context"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"x-foot/internal/controller/auth"
	"x-foot/internal/controller/company"
	"x-foot/internal/controller/hello"
	"x-foot/internal/controller/route"
	"x-foot/internal/controller/system"
	_ "x-foot/internal/logic/accounting"
	_ "x-foot/internal/logic/auth"
	_ "x-foot/internal/logic/company"
	_ "x-foot/internal/logic/route"
	_ "x-foot/internal/logic/system"
	"x-foot/internal/service"
	"x-foot/internal/utility/middleware"
	"x-foot/internal/utility/response"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 系统无任何用户时初始化默认账号
			if err = service.Auth().SeedDefaultUsers(ctx); err != nil {
				g.Log().Error(ctx, "初始化默认账号失败:", err)
				return err
			}

			// 补齐缺失的默认菜单，保证菜单权限有可授权的数据
			if err = service.System().SeedDefaultMenus(ctx); err != nil {
				g.Log().Error(ctx, "初始化默认菜单失败:", err)
				return err
			}

			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.Middleware(response.Middleware, middleware.Auth)
				group.Bind(
					hello.NewV1(),
					auth.NewV1(),
					company.NewV1(),
					route.NewV1(),
					system.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)
