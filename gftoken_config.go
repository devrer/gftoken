package gftoken

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/gf/util/gutil"
	"strings"
	"time"
)

// TokenConfig token config 结构
type TokenConfig struct {
	Timeout  time.Duration // 超时时间
	Multiple bool          // 单账号多终端登录
	Refresh  int64         //  每次请求 token 续期时长, 单位: 秒. 默认 0 则为不续期
	Mode     CacheMode     // 缓存模式: 0 内存, 1 Redis, 默认 0
	Redis    string        // 缓存模式为 Redis 时, Redis 的配置选项名

	ClearCron string // 定时清理过期 token 的任务

	LoginPath   string     // 登录地址
	AllowPrefix string     // 过滤前缀
	AllowList   g.SliceStr // 过滤路径
	Method      g.SliceStr // token 参数的方式: header,get,post
	Key         string     // token 参数键名
}

// NewConfig new config
func NewConfig() TokenConfig {
	return TokenConfig{
		Timeout:   3600 * 24 * 7 * time.Second, // 7 天
		ClearCron: "0 0 2 * * *",               // 每天凌晨2点执行
		LoginPath: "/login",
		Method:    []string{"header"},
		Key:       "Authorization",
	}
}

// SetConfigWithMap map config 转换
func (t *Token) SetConfigWithMap(m map[string]interface{}) error {
	var config TokenConfig
	m = gutil.MapCopy(m)
	if err := gconv.Struct(m, &config); err != nil {
		return err
	}
	return t.SetConfig(config)
}

// SetConfig 设置 config
func (t *Token) SetConfig(c TokenConfig) error {
	if c.Timeout > 0 {
		t.SetTimeout(c.Timeout)
	}

	t.SetMultiple(c.Multiple)

	if c.LoginPath != "" {
		t.SetLoginPath(c.LoginPath)
	}

	t.SetAllowPrefix(c.AllowPrefix)

	if len(c.AllowList) > 0 {
		t.SetAllowList(c.AllowList)
	}

	if len(c.Method) > 0 {
		t.SetMethod(c.Method)
	}

	if c.Key != "" {
		t.SetKey(c.Key)
	}

	t.SetMode(c.Mode)

	t.SetRedis(c.Redis)

	if c.ClearCron != "" {
		t.SetClearCron(c.ClearCron)
	}
	return nil
}

// GetName 获取 token 名
func (t *Token) GetName() string {
	return t.name
}

// SetTimeout 设置超时时间
func (t *Token) SetTimeout(value time.Duration) {
	t.config.Timeout = value * time.Second
}

// SetMultiple 设置单点唯一登录
func (t *Token) SetMultiple(value bool) {
	t.config.Multiple = value
}

// SetLoginPath 设置登录的 API 的 URL
func (t *Token) SetLoginPath(value string) {
	t.config.LoginPath = value
}

// SetAllowPrefix 设置忽略的 API 前缀
func (t *Token) SetAllowPrefix(value string) {
	t.config.AllowPrefix = value
}

// SetAllowList 设置忽略的 URL
func (t *Token) SetAllowList(value g.ArrayStr) {
	methodArr := g.MapStrBool{
		"get":    true,
		"post":   true,
		"put":    true,
		"delete": true,
		"all":    true,
	}
	for _, val := range value {
		strArr := strings.Split(strings.ReplaceAll(val, " ", ""), ":")
		if len(strArr) == 0 || strArr[0] == "" {
			continue
		}

		var methodStr, valueStr string
		if len(strArr) == 1 {
			methodStr = "all"
			valueStr = strArr[0]
		} else if len(strArr) == 2 {
			methodLower := strings.ToLower(strArr[0])
			if ok, has := methodArr[methodLower]; has && ok {
				methodStr = methodLower
			} else {
				methodStr = "all"
			}
			valueStr = strArr[1]
		}

		urls := strings.Split(valueStr, ",")
		switch methodStr {
		case "get":
			t.allowList.Get = append(t.allowList.Get, urls...)
		case "post":
			t.allowList.Post = append(t.allowList.Post, urls...)
		case "put":
			t.allowList.Put = append(t.allowList.Put, urls...)
		case "delete":
			t.allowList.Delete = append(t.allowList.Delete, urls...)
		default:
			t.allowList.All = append(t.allowList.All, urls...)
		}
	}
}

// SetMethod 设置 token 的获取位置
func (t *Token) SetMethod(value g.SliceStr) {
	t.config.Method = value
}

// SetKey 设置 token 的键名
func (t *Token) SetKey(value string) {
	t.config.Key = value
}

// SetMode　设置缓存模式
func (t *Token) SetMode(value CacheMode) {
	if value != 0 && value != 1 {
		value = 0
	}
	t.config.Mode = value
}

// SetRedis　结合 goframe 的 redis 配置项的名称
func (t *Token) SetRedis(value string) {
	defer func() {
		if err := recover(); err != nil {
			t.config.Mode = CacheMemory
		}
	}()

	t.config.Redis = value

	if t.config.Mode == CacheRedis {
		t.redis().Conn()
	}
}

// SetClearCron 设置删除过期 token 的计划任务
func (t *Token) SetClearCron(value string) {
	t.config.ClearCron = value
}

// SetAuthDestroy 设置校验不成功的 HOOK
func (t *Token) SetAuthDestroy(fn func(*ghttp.Request, error)) {
	if fn != nil {
		t.authDestroy = fn
	}
}
