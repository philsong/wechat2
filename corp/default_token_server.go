// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/philsong/wechat2 for the canonical source repository
// @license     https://github.com/philsong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package corp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var _ TokenServer = new(DefaultTokenServer)

// TokenServer 的简单实现.
//  NOTE:
//  一般用于单进程环境, 因为 DefaultTokenServer 同时也实现了一个简单的中控服务器, 而不是简单的
//  实现了 TokenServer 接口, 所以整个系统只能存在一个 DefaultTokenServer 实例!!!
type DefaultTokenServer struct {
	corpId, corpSecret string
	httpClient         *http.Client

	// 缓存最后一次从微信服务器获取的 access_token 的结果.
	currentToken struct {
		rwmutex sync.RWMutex
		token   string
		err     error
	}

	tokenRefresh struct {
		mutex            sync.Mutex
		lastGetTimestamp int64 // 最后一次从服务器获取 access_token 的时间戳
	}
}

// 创建一个新的 DefaultTokenServer.
//  如果 httpClient == nil 则默认使用 http.DefaultClient.
func NewDefaultTokenServer(corpId, corpSecret string,
	httpClient *http.Client) (srv *DefaultTokenServer) {

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	srv = &DefaultTokenServer{
		corpId:     corpId,
		corpSecret: corpSecret,
		httpClient: httpClient,
	}

	// 获取 access_token
	resp, err := srv.getToken()
	if err != nil {
		srv.currentToken.token = ""
		srv.currentToken.err = err
	} else {
		srv.currentToken.token = resp.Token
		srv.currentToken.err = nil
	}
	return
}

func (srv *DefaultTokenServer) Token() (token string, err error) {
	srv.currentToken.rwmutex.RLock()
	token = srv.currentToken.token
	err = srv.currentToken.err
	srv.currentToken.rwmutex.RUnlock()
	return
}

func (srv *DefaultTokenServer) TokenRefresh() (token string, err error) {
	srv.tokenRefresh.mutex.Lock()
	defer srv.tokenRefresh.mutex.Unlock()

	timeNow := time.Now().Unix()

	// 如果 5 秒内调用过, 则直接返回原来调用的结果.
	if timeNow < srv.tokenRefresh.lastGetTimestamp+5 {
		return srv.Token()
	}

	resp, err := srv.getToken()
	if err != nil {
		srv.currentToken.rwmutex.Lock()
		srv.currentToken.token = ""
		srv.currentToken.err = err
		srv.currentToken.rwmutex.Unlock()
	} else {
		srv.currentToken.rwmutex.Lock()
		srv.currentToken.token = resp.Token
		srv.currentToken.err = nil
		srv.currentToken.rwmutex.Unlock()

		token = resp.Token
	}

	srv.tokenRefresh.lastGetTimestamp = timeNow ////
	return
}

type tokenResponse struct {
	Token     string `json:"access_token"` // 获取到的凭证
	ExpiresIn int64  `json:"expires_in"`   // 凭证有效时间，单位：秒
}

// 从微信服务器获取 access_token.
func (srv *DefaultTokenServer) getToken() (resp *tokenResponse, err error) {
	url := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + srv.corpId +
		"&corpsecret=" + srv.corpSecret

	httpResp, err := srv.httpClient.Get(url)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", httpResp.Status)
		return
	}

	var result struct {
		Error
		tokenResponse
	}
	if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return
	}

	if result.ErrCode != ErrCodeOK {
		err = &result.Error
		return
	}
	resp = &result.tokenResponse
	return
}
