package gftoken

import (
	"encoding/base64"
	"fmt"
	"github.com/gogf/gf/database/gredis"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcache"
	"github.com/gogf/gf/os/gcron"
	"github.com/gogf/guuid"
	"strings"
	"time"
)

// Add 添加 token 数据
func (t *Token) Add(uuid string) (string, error) {
	randData, err := guuid.New().MarshalText()
	if err != nil {
		return "", err
	}
	userToken := base64.StdEncoding.EncodeToString(randData)
	return userToken, t.addGfToken(uuid, userToken)
}

// Get 获取 token 数据
func (t *Token) Get(token string) (*TokenInfo, error) {
	return t.getGfToken(token)
}

// Key 获取 token 值
func (t *Token) Key(r *ghttp.Request) string {
	var tokenKey string
	for _, method := range t.config.Method {
		if method == "header" {
			tokenKey = r.Header.Get(t.config.Key)
		} else if method == "get" {
			tokenKey = r.GetQueryString(t.config.Key)
		} else if method == "post" {
			tokenKey = r.FormValue(t.config.Key)
		}

		if tokenKey != "" {
			break
		}
	}
	return tokenKey
}

// Refresh 刷新 token 过期时间
func (t *Token) Refresh(token string) (*TokenInfo, error) {
	tokenInfo, err := t.getGfToken(token)
	if err != nil {
		return nil, err
	}

	nowTime := time.Now()
	tokenInfo.RefreshTime = nowTime
	tokenInfo.ExpirationTime = nowTime.Add(t.config.Timeout)

	gfTokenToData := fmt.Sprintf(VarsTokenToData, token)
	if t.config.Mode == CacheRedis {
		t.redis().Do("HSET", gfTokenToData, "refresh_time", tokenInfo.RefreshTime)
		t.redis().Do("HSET", gfTokenToData, "expiration_time", tokenInfo.ExpirationTime)

	} else {
		if e2 := t.cache().Set(gfTokenToData, tokenInfo, 0); e2 != nil {
			return nil, e2
		}
	}

	return tokenInfo, nil
}

// Delete 删除 token
func (t *Token) Delete(token string) error {
	return t.delGfToken(token)
}

// Clear 清空 token 相关数据
func (t *Token) Clear() error {
	return t.clearGfToken()
}

// ClearCron 定时清除过期的 token
func (t *Token) ClearCron() error {
	if t.config.ClearCron == "" {
		return ErrEmptyClearCron
	}

	var err error
	var entry *gcron.Entry
	redisTokenKey := fmt.Sprintf(VarsTokenToData, "*")
	cacheTokenKey := fmt.Sprintf(VarsTokenToData, "")

	if t.config.Mode == CacheRedis {
		entry, err = gcron.Add(t.config.ClearCron, func() {
			keyVar, e0 := t.redis().DoVar("KEYS", redisTokenKey)
			if e0 != nil {
				entry.Close()
				return
			}
			for _, key := range keyVar.Strings() {
				val, e1 := t.redis().DoVar("HGET", key, "expiration_time")
				if e1 != nil {
					continue
				}
				if time.Now().Sub(val.Time()) > 0 {
					if _, e2 := t.redis().Do("DEL", key); e2 != nil {
						continue
					}
				}
			}
		})
		if err != nil {
			return err
		}
	} else {
		entry, err = gcron.Add(t.config.ClearCron, func() {
			cacheVar, e0 := t.cache().Data()
			if e0 != nil {
				entry.Close()
				return
			}
			for key, value := range cacheVar {
				if strings.HasPrefix(key.(string), cacheTokenKey) {
					tokenInfo := value.(*TokenInfo)
					if time.Now().Sub(tokenInfo.ExpirationTime) > 0 {
						if _, e2 := t.cache().Remove(key); e2 != nil {
							continue
						}
					}
				}
			}
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Token) redis() *gredis.Redis {
	return g.Redis(t.config.Redis)
}

func (t *Token) cache() *gcache.Cache {
	return tokenCache
}
