package gftoken

import (
	"fmt"
	"github.com/gogf/gf/container/gvar"
	"time"
)

// addGfToken 添加 token 信息
func (t *Token) addGfToken(uuid, token string) error {
	if !t.config.Multiple {
		if err := t.delGfUuidToken(uuid); err != nil {
			return err
		}
	}

	nowTime := time.Now()
	tokenInfo := &TokenInfo{
		Uuid:           uuid,
		Token:          token,
		CreatedTime:    nowTime,
		RefreshTime:    nowTime,
		ExpirationTime: nowTime.Add(t.config.Timeout),
	}

	gfTokenToData := fmt.Sprintf(VarsTokenToData, token)
	if t.config.Mode == CacheRedis {
		if _, err := t.redis().Do("HSET", VarsUuidToToken, uuid, token); err != nil {
			return err
		}

		if _, err := t.redis().Do("HSET", gfTokenToData, "uuid", tokenInfo.Uuid); err != nil {
			return err
		}
		t.redis().Do("HSET", gfTokenToData, "token", tokenInfo.Token)
		t.redis().Do("HSET", gfTokenToData, "created_time", tokenInfo.CreatedTime)
		t.redis().Do("HSET", gfTokenToData, "refresh_time", tokenInfo.RefreshTime)
		t.redis().Do("HSET", gfTokenToData, "expiration_time", tokenInfo.ExpirationTime)

	} else {
		gfUuidToToken := VarsUuidToToken + "_" + uuid
		if err := t.cache().Set(gfUuidToToken, token, 0); err != nil {
			return err
		}

		if err := t.cache().Set(gfTokenToData, tokenInfo, 0); err != nil {
			return err
		}
	}

	return nil
}

// delGfUuidToken 根据 UUID 删除 TOKEN
func (t *Token) delGfUuidToken(uuid string) error {
	var (
		err         error
		oldToken    string
		oldTokenRes *gvar.Var
	)
	if t.config.Mode == CacheRedis {
		oldTokenRes, err = t.redis().DoVar("HGET", VarsUuidToToken, uuid)

	} else {
		gfUuidToToken := VarsUuidToToken + "_" + uuid
		oldTokenRes, err = t.cache().GetVar(gfUuidToToken)
	}

	if err != nil {
		return err
	}

	oldToken = oldTokenRes.String()
	if oldToken == "" {
		return nil
	}

	return t.delGfToken(oldToken)
}

// delGfToken 删除 TOKEN
func (t *Token) delGfToken(token string) error {
	gfTokenToData := fmt.Sprintf(VarsTokenToData, token)
	if t.config.Mode == CacheRedis {
		if _, err := t.redis().Do("DEL", gfTokenToData); err != nil {
			return err
		}

	} else {
		if _, err := t.cache().Remove(gfTokenToData); err != nil {
			return err
		}
	}

	return nil
}

// getGfToken 获取 TOKEN 信息
func (t *Token) getGfToken(token string) (*TokenInfo, error) {
	if token == "" {
		return nil, ErrTokenKeyNil
	}

	var (
		err       error
		tokenInfo *TokenInfo = nil
		res       *gvar.Var
	)

	gfTokenToData := fmt.Sprintf(VarsTokenToData, token)
	if t.config.Mode == CacheRedis {
		res, err = t.redis().DoVar("HGETALL", gfTokenToData)
	} else {
		res, err = t.cache().GetVar(gfTokenToData)
	}

	if err != nil {
		return nil, err
	} else if res.IsEmpty() {
		return nil, ErrTokenInfoNil
	}

	if e := res.Struct(&tokenInfo); e != nil {
		return nil, e
	}

	return tokenInfo, nil
}

func (t *Token) clearGfToken() error {
	if t.config.Mode == CacheRedis {
		keyVar, err := t.redis().DoVar("KEYS", fmt.Sprintf("%s*", VarsRedisPrefix))
		if err != nil {
			return err
		}

		for _, key := range keyVar.Strings() {
			if _, err := t.redis().Do("DEL", key); err != nil {
				return err
			}
		}

	} else {
		if err := t.cache().Clear(); err != nil {
			return err
		}
	}

	return nil
}
