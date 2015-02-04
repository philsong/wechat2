// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package corp

import (
	"errors"
	"sync"
)

var ErrCacheMiss = errors.New("corp: cache miss")

// access_token 缓存接口, see token_cache.png
type TokenCache interface {
	// 从缓存里获取 access_token, 如果没有找到返回 ErrCacheMiss!!!
	Token() (token string, err error)

	// 添加(设置) access_token
	PutToken(token string) (err error)
}

var _ TokenCache = new(DefaultTokenCache)

type DefaultTokenCache struct {
	rwmutex sync.RWMutex
	token   string
}

func (cache *DefaultTokenCache) Token() (token string, err error) {
	cache.rwmutex.RLock()
	if cache.token == "" {
		err = ErrCacheMiss
	} else {
		token = cache.token
	}
	cache.rwmutex.RUnlock()
	return
}

func (cache *DefaultTokenCache) PutToken(token string) (err error) {
	if cache.token == "" {
		return errors.New("token is empty")
	}
	cache.rwmutex.Lock()
	cache.token = token
	cache.rwmutex.Unlock()
	return
}
