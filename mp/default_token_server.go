// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/philsong/wechat2 for the canonical source repository
// @license     https://github.com/philsong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const defaultTickDuration = time.Minute // 设置 44 秒以上就不会超过限制(2000次/日)

var _ TokenServer = new(DefaultTokenServer)

// TokenServer 的简单实现.
//  NOTE:
//  一般用于单进程环境, 因为 DefaultTokenServer 同时也实现了一个简单的中控服务器, 而不是简单的
//  实现了 TokenServer 接口, 所以整个系统只能存在一个 DefaultTokenServer 实例!!!
type DefaultTokenServer struct {
	appid, appsecret string
	httpClient       *http.Client

	// 缓存最后一次从微信服务器获取的 access_token 的结果.
	// goroutine tokenAutoUpdate() 里有个定时器, 每次触发都会更新 currentToken.
	currentToken struct {
		rwmutex sync.RWMutex
		token   string
		err     error
	}

	// goroutine tokenAutoUpdate() 监听 resetTokenRefreshTickChan,
	// 如果有新的数据, 则重置定时器, 定时时间为 resetTokenRefreshTickChan 传过来的数据.
	resetTokenRefreshTickChan chan time.Duration

	tokenRefresh struct {
		mutex            sync.Mutex
		lastGetTimestamp int64 // 最后一次从服务器获取 access_token 的时间戳
	}
}

// 创建一个新的 DefaultTokenServer.
//  如果 httpClient == nil 则默认使用 http.DefaultClient.
func NewDefaultTokenServer(appid, appsecret string,
	httpClient *http.Client) (srv *DefaultTokenServer) {

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	srv = &DefaultTokenServer{
		appid:                     appid,
		appsecret:                 appsecret,
		httpClient:                httpClient,
		resetTokenRefreshTickChan: make(chan time.Duration),
	}

	// 获取 access_token 并启动 goroutine tokenAutoUpdate
	resp, err := srv.getToken()
	if err != nil {
		srv.currentToken.token = ""
		srv.currentToken.err = err
		go srv.tokenAutoUpdate(defaultTickDuration)
	} else {
		srv.currentToken.token = resp.Token
		srv.currentToken.err = nil
		go srv.tokenAutoUpdate(time.Duration(resp.ExpiresIn) * time.Second)
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

		srv.resetTokenRefreshTickChan <- defaultTickDuration
	} else {
		srv.currentToken.rwmutex.Lock()
		srv.currentToken.token = resp.Token
		srv.currentToken.err = nil
		srv.currentToken.rwmutex.Unlock()

		token = resp.Token
		srv.resetTokenRefreshTickChan <- time.Duration(resp.ExpiresIn) * time.Second
	}

	srv.tokenRefresh.lastGetTimestamp = timeNow ////
	return
}

// 单独一个 goroutine 来定时获取 access_token.
//  tickDuration: 启动后初始 tickDuration.
func (srv *DefaultTokenServer) tokenAutoUpdate(tickDuration time.Duration) {
	var ticker *time.Ticker

NEW_TICK_DURATION:
	ticker = time.NewTicker(tickDuration)
	for {
		select {
		case tickDuration = <-srv.resetTokenRefreshTickChan:
			ticker.Stop()
			goto NEW_TICK_DURATION

		case <-ticker.C:
			resp, err := srv.getToken()
			if err != nil {
				srv.currentToken.rwmutex.Lock()
				srv.currentToken.token = ""
				srv.currentToken.err = err
				srv.currentToken.rwmutex.Unlock()

				// 出错则重置到 defaultTickDuration
				if tickDuration != defaultTickDuration {
					ticker.Stop()
					tickDuration = defaultTickDuration
					goto NEW_TICK_DURATION
				}
			} else {
				srv.currentToken.rwmutex.Lock()
				srv.currentToken.token = resp.Token
				srv.currentToken.err = nil
				srv.currentToken.rwmutex.Unlock()

				newTickDuration := time.Duration(resp.ExpiresIn) * time.Second
				if tickDuration != newTickDuration {
					ticker.Stop()
					tickDuration = newTickDuration
					goto NEW_TICK_DURATION
				}
			}
		}
	}
}

type tokenResponse struct {
	Token     string `json:"access_token"` // 获取到的凭证
	ExpiresIn int64  `json:"expires_in"`   // 凭证有效时间，单位：秒
}

// 从微信服务器获取 access_token.
func (srv *DefaultTokenServer) getToken() (resp *tokenResponse, err error) {
	url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" +
		srv.appid + "&secret=" + srv.appsecret

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

	// 由于网络的延时, access_token 过期时间留了一个缓冲区;
	// 正常情况下微信服务器会返回 7200, 则缓冲区的大小为 10 分钟.
	switch {
	case result.ExpiresIn > 60*60:
		result.ExpiresIn -= 60 * 10
		resp = &result.tokenResponse
		return
	case result.ExpiresIn > 60*30:
		result.ExpiresIn -= 60 * 5
		resp = &result.tokenResponse
		return
	case result.ExpiresIn > 60*5:
		result.ExpiresIn -= 60
		resp = &result.tokenResponse
		return
	case result.ExpiresIn > 60:
		result.ExpiresIn -= 10
		resp = &result.tokenResponse
		return
	case result.ExpiresIn > 0:
		resp = &result.tokenResponse
		return
	default:
		err = errors.New("invalid expires_in: " + strconv.FormatInt(result.ExpiresIn, 10))
		return
	}
}
