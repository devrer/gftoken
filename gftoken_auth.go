package gftoken

import (
	"github.com/gogf/gf/net/ghttp"
	"strings"
	"time"
)

// Auth token 校验
func (t *Token) Auth(r *ghttp.Request) {
	rURL := r.URL.String()

	defer func() {
		if err := recover(); err != nil {
			t.authDestroy(r, err.(error))
		}
	}()

	// 过滤登录功能
	if t.config.LoginPath == rURL {
		r.Middleware.Next()
		return
	}

	// 过滤白名单 - All
	for _, value := range t.allowList.All {
		if strings.Compare(rURL, t.config.AllowPrefix+value) == 0 {
			r.Middleware.Next()
			return
		}
	}

	// 过滤其它方法
	var allowList []string
	switch strings.ToLower(r.Method) {
	case "get":
		allowList = t.allowList.Get
	case "post":
		allowList = t.allowList.Post
	case "put":
		allowList = t.allowList.Put
	case "delete":
		allowList = t.allowList.Delete
	}
	for _, value := range allowList {
		if strings.Compare(rURL, t.config.AllowPrefix+value) == 0 {
			r.Middleware.Next()
			return
		}
	}

	// token 键名不存在
	tokenKey := t.Key(r)
	if tokenKey == "" {
		panic(ErrTokenKeyNil)
	}

	// token 数据错误
	tokenInfo, err := t.getGfToken(tokenKey)
	if err != nil {
		panic(err)
	}

	// token 数据不存在
	if tokenInfo == nil {
		panic(ErrTokenInfoNil)
	}

	// token 已过期
	if time.Now().Sub(tokenInfo.ExpirationTime) > 0 {
		panic(ErrTokenExpired)
	}

	r.Middleware.Next()
}
