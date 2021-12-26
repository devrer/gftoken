gftoken
---
goframe token

### 功能
- 自定义 `token` 过期时间
- `token` 过期时间自动续期
- 支持单点多端登录
- 数据支持内存模式和 `Redis` 缓存模式
- 自定义删除过期数据
- 支持接口白名单
- 支持 `header` 方式, 支持 `GET`, `POST` 参数方式

### `token` 配置信息
```toml
[token]
# token 有效时长(秒)
timeout = 604800
# 是否只允许账号唯一登录
multiple = true
# 自动刷新有效时长
refresh = 0
# 保存模式。 0: memory, 1: redis
mode = 0
# 若保存模式为 redis， 则使用的 redis 名
redis = ""
# 删除过期 token 的计划任务, 默认每天凌晨 2 点执行: 0 0 2 * * *
clearCron = "0 0 2 * * *"

# 登录接口
loginPath = "/api/user/login"
# 过滤的接口前缀(allowList 列表的接口前面添加此前缀)
allowPrefix = "/api"
# 过滤的接口地址。
allowList = [
    "get: /home/index",
    "post: /user/logout",
    "put:",
    "delete:",
    "all:",
]
# 获取 token 的位置。允许 header, get, post 方式，如 ["header", "get", "post"]
method = ["header", "get", "post"]
# 获取 token 的键名
key = "Authorization"
```

|参数名|参数类型|默认值|参数说明|
|:---|:---|:---|:---|
timeout|int|604800|秒, 默认 `7` 天有效期
multiple|bool|false|是否只允许账号唯一登录(只允许一个终端同时在线)。 默认 `false` 则为不允许
refresh|int|0|秒, 每次请求 token 续期时长。 默认 `0` 则为不续期
mode|int|0|数据缓存模式。若为内存模式，则程序关闭数据就丢失。 `0` 内存, `1` Redis。 默认 `0` 则为内存模式
redis|string||缓存模式(`mode`)为 Redis 时，Redis 的配置选项名。 只有 `mode` 为 Redis 模式时启用
clearCron|string|0 0 2 * * * |删除过期 token 的计划任务。 默认每天凌晨 `2` 点执行: `0 0 2 * * *`。 更多配置请查看 **[GF官方文档](https://goframe.org/pages/viewpage.action?pageId=1114187)**
loginPath|string|/login|登录接口网址，此“网址”为白名单，不参与校验。 默认为 `/login`。
allowPrefix|string||过滤的接口前缀(`allowList` 列表的接口前面添加此前缀)。
allowList|[]string||过滤的接口地址。 格式为：`["请求方法1: 接口地址一,接口地址二","请求方法2: 接口地址一,接口地址二"]`。 请求方法如 `get`/`post`/`delete`/`put`/`all` (`all` 为全部请求方法)
method|string|header|获取 `token` 的位置。 允许 `header`, `get`, `post` 方式，若为 `["header", "get", "post"]` 则依次获取。默认为 `header`
key|string|Authorization|获取 `token` 的键名。 默认为 `Authorization`


### 方法
```go
func Get(name ...string) *Token
func (t *Token) Add(uuid string) (string, error)
func (t *Token) Auth(r *ghttp.Request)
func (t *Token) Clear() error
func (t *Token) ClearCron() error
func (t *Token) Delete(token string) error
func (t *Token) Get(token string) (*TokenInfo, error)
func (t *Token) GetName() string
func (t *Token) Key(r *ghttp.Request) string
func (t *Token) Refresh(token string) (*TokenInfo, error)
func (t *Token) SetAllowList(value g.ArrayStr)
func (t *Token) SetAllowPrefix(value string)
func (t *Token) SetAuthDestroy(fn func(*ghttp.Request, error))
func (t *Token) SetClearCron(value string)
func (t *Token) SetConfig(c TokenConfig) error
func (t *Token) SetConfigWithMap(m map[string]interface{}) error
func (t *Token) SetKey(value string)
func (t *Token) SetLoginPath(value string)
func (t *Token) SetMethod(value g.SliceStr)
func (t *Token) SetMode(value CacheMode)
func (t *Token) SetMultiple(value bool)
func (t *Token) SetRedis(value string)
func (t *Token) SetTimeout(value time.Duration)
```
**更多说明文档请查阅:** [https://pkg.go.dev/github.com/skiy/gftoken](https://pkg.go.dev/github.com/skiy/gftoken)


### 使用说明
> 可查看 `example` 例子。

`config.toml` 配置文件
```
# 登录接口
loginPath = "/api/login"
# 过滤的接口前缀(allowList 列表的接口前面添加此前缀)
allowPrefix = "/api"
# 过滤的接口地址。
allowList = [
    "get: /index",
    "post: /logout",
    "put:",
    "delete:",
    "all:",
]
```

`router.go` 路由文件
```go
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

// 全局校验，对所有接口生效
//s.Use(tk.Auth) 

s.Group("/", func(sg *ghttp.RouterGroup) {
	sg.GET("/", func(r *ghttp.Request) {
		r.Response.Write("Index Page")
	})

	sg.Group("/api", func(wg *ghttp.RouterGroup) {

		// 指定入口校验，仅对 `/api` 下的接口生效
		wg.Middleware(tk.Auth) 

		wg.GET("/login", func(r *ghttp.Request) {
			userId := 1

			// 设置 token
			tkVal, err := gftoken.Get().Add(gconv.String(userId))

			if err != nil {
				_ = r.Response.WriteJson(g.Map{
					"code":    -1,
					"message": err.Error(),
				})
				return
			}

			result := g.Map{
				"id":       userId,
				"username": "admin",
				"token":    tkVal,
			}
			resp := g.Map{
				"message": "登录成功",
				"result":  result,
			}
			_ = r.Response.WriteJson(resp)
		})
	})
})

```

- 添加 token: `gftoken.Get().Add(userId)`
- 删除 token: `gftoken.Get().Delete(tokenString)`
- 获取 token 信息: `gftoken.Get().Get(tokenString)`

