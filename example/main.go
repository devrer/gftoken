package main

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/skiy/gftoken"
	"time"
)

var (
	User    g.MapStrAny
	UserMap = g.MapStrStr{
		"admin": "admin888",
		"user":  "user888",
	}
	UidMap = g.MapStrStr{
		"admin": "1",
		"user":  "2",
	}
)

func init() {
	s := g.Server()

	tk := gftoken.Get()

	tk.SetAuthDestroy(func(r *ghttp.Request, e error) {
		var code = 999
		var message = e.Error()
		if e == gftoken.ErrTokenExpired {
			message = "TOKEN 已过期，请重新登录"
		} else {
			message = "TOKEN 无效，请重新登录"
		}
		resp := g.Map{
			"code":    code,
			"message": message,
			"result":  "",
		}
		_ = r.Response.WriteJson(resp)
	})
	_ = tk.ClearCron()
	//s.Use(tk.Auth) // 全局校验

	s.Group("/", func(sg *ghttp.RouterGroup) {
		sg.GET("/", func(r *ghttp.Request) {
			r.Response.Write("Index Page")
		})

		sg.Group("/api", func(wg *ghttp.RouterGroup) {
			wg.Middleware(tk.Auth) // 指定入口校验

			// 登录页面
			wg.POST("/login", func(r *ghttp.Request) {
				account := r.Request.PostFormValue("account")
				password := r.Request.PostFormValue("password")

				pwd := UserMap[account]
				if pwd == "" {
					_ = r.Response.WriteJson(g.Map{"code": -1, "message": "用户不存在"})
					return
				}

				if pwd != password {
					_ = r.Response.WriteJson(g.Map{"code": -1, "message": "密码不正确"})
					return
				}

				userId := UidMap[account]

				// 生成 token
				tkVal, err := gftoken.Get().Add(userId)

				if err != nil {
					_ = r.Response.WriteJson(g.Map{
						"code":    -1,
						"message": err.Error(),
					})
					return
				}

				info := g.Map{
					"user": account,
					"pwd":  password,
					"time": time.Now(),
				}

				User = g.MapStrAny{
					tkVal: info,
				}

				result := g.Map{
					"id":    userId,
					"user":  info,
					"token": tkVal,
				}
				resp := g.Map{
					"message": "登录成功",
					"result":  result,
				}
				_ = r.Response.WriteJson(resp)
			})

			// 登录首页
			wg.GET("/index", func(r *ghttp.Request) {
				r.Response.Write("Api Index Page")
			})

			// 用户信息
			wg.GET("/info", func(r *ghttp.Request) {
				authorization := r.GetHeader("Authorization")
				_ = r.Response.WriteJson(User[authorization])
			})

			// 退出登录
			wg.POST("/logout", func(r *ghttp.Request) {
				authorization := r.GetHeader("Authorization")
				// 删除 token
				if err := gftoken.Get().Delete(authorization); err != nil {
					_ = r.Response.WriteJson(g.Map{"message": err.Error()})
					return
				}
				User[authorization] = nil
				_ = r.Response.WriteJson(g.Map{"message": "退出成功"})
			})
		})
	})
}

func main() {
	g.Server().Run()
}
