package gftoken

import (
	"errors"
	"github.com/gogf/gf/container/gmap"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcache"
	"time"
)

type (
	Token struct {
		name   string
		config TokenConfig

		allowList   AllowList
		authDestroy func(*ghttp.Request, error) // 登录检测函数
	}

	AllowList struct {
		All,
		Post,
		Delete,
		Put,
		Get g.ArrayStr
	}

	CacheMode uint

	TokenInfo struct {
		Uuid           string    `json:"uuid"`
		Token          string    `json:"token"`
		CreatedTime    time.Time `json:"created_time"`
		RefreshTime    time.Time `json:"refresh_time"`
		ExpirationTime time.Time `json:"expiration_time"`
	}
)

const (
	CacheMemory CacheMode = iota
	CacheRedis
)

const (
	defaultTokenName = "token"

	// VarsRedisPrefix token 前缀
	VarsRedisPrefix = "::TOKEN::"
	// VarsTokenToData token -> data
	VarsTokenToData = "::TOKEN::TOKEN@DATA_%s"
	// VarsUuidToToken uuid -> token
	VarsUuidToToken = "::TOKEN::UUID@TOKEN"
)

var (
	tokenMapping = gmap.NewStrAnyMap(true)

	tokenCache = gcache.New()

	configName = "token"
	instances  = gmap.NewStrAnyMap(true)

	ErrTokenKeyNil    = errors.New("token key is nil")
	ErrTokenInfoNil   = errors.New("token info is nil")
	ErrTokenExpired   = errors.New("token is expired")
	ErrEmptyClearCron = errors.New("gcron data is empty")
)

var authDestroy = func(r *ghttp.Request, e error) {
	_ = r.Response.WriteJson(g.Map{
		"code":    9999,
		"message": e.Error(),
	})
}
