package gftoken

import (
	"fmt"
	"github.com/gogf/gf/frame/gins"
	"github.com/gogf/gf/util/gutil"
)

// initToken 获取 token 初始配置
func initToken(name ...string) *Token {
	tokenName := defaultTokenName
	if len(name) > 0 && name[0] != "" {
		tokenName = name[0]
	}
	if k := tokenMapping.Get(tokenName); k != nil {
		return k.(*Token)
	}
	k := &Token{
		name: tokenName,
	}
	if err := k.SetConfig(NewConfig()); err != nil {
		panic(err)
	}
	k.SetAuthDestroy(authDestroy)
	tokenMapping.Set(tokenName, k)
	return k
}

// Get 获取 Token
func Get(name ...string) *Token {
	instanceKey := fmt.Sprintf("gf.core.component.token.%v", name)
	return instances.GetOrSetFuncLock(instanceKey, func() interface{} {
		k := initToken(name...)
		if gins.Config().Available() {
			var m map[string]interface{}
			nodeKey, _ := gutil.MapPossibleItemByKey(gins.Config().GetMap("."), configName)
			if nodeKey == "" {
				nodeKey = configName
			}
			m = gins.Config().GetMap(fmt.Sprintf(`%s.%s`, nodeKey, k.GetName()))
			if len(m) == 0 {
				m = gins.Config().GetMap(nodeKey)
			}
			if len(m) > 0 {
				if err := k.SetConfigWithMap(m); err != nil {
					panic(err)
				}
			}
		}
		return k
	}).(*Token)
}
