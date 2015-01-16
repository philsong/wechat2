// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechatv2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechatv2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	wechatjson "github.com/chanxuehong/wechatv2/json"
)

type WechatClient struct {
	accessToken  string // 缓存当前的 access token
	TokenService TokenService
	HttpClient   *http.Client
}

// 获取 access token.
func (clt *WechatClient) Token() (token string, err error) {
	if clt.accessToken != "" {
		token = clt.accessToken
		return
	}
	token, err = clt.TokenService.Token()
	if err != nil {
		return
	}
	clt.accessToken = token
	return
}

// 请求微信服务器更新 access token.
//  NOTE:
//  1. 一般情况下无需调用该函数, 请使用 Token() 获取 access token.
//  2. 即使 access token 失效(错误代码 40001, 正常情况下不会出现),
//     也请谨慎调用 TokenRefresh, 建议直接返回错误! 因为很有可能高并发情况下造成雪崩效应!
//  3. 再次强调, 调用这个函数你应该知道发生了什么!!!
func (clt *WechatClient) TokenRefresh() (token string, err error) {
	token, err = clt.TokenService.TokenRefresh()
	if err != nil {
		return
	}
	clt.accessToken = token
	return
}

// 当 WechatClient.Token() 返回的 token 失效时获取新的 token.
func (clt *WechatClient) GetNewToken() (token string, err error) {
	// 当 TokenService 伺服程序更新了 access token, 但是缓存里还没有更新的时候就会产生失效,
	// 这个时候需要去 TokenService 获取新的 access token.
	for i := 0; ; {
		token, err = clt.TokenService.Token()
		if err != nil {
			return
		}
		if clt.accessToken != token {
			clt.accessToken = token
			return
		}

		if i++; i < 10 {
			time.Sleep(50 * time.Millisecond) // 等待 50ms 再次获取
			continue
		}
		break
	}

	err = errors.New("WechatClient.GetNewToken failed")
	return
}

// 把 request marshal 为 JSON, 放入 http 请求的 body 中, POST 到微信服务器 url 上,
// 然后把微信服务器返回的 JSON 解析到 response.
func (clt *WechatClient) PostJSON(url string, request interface{}, response interface{}) (err error) {
	buf := textBufferPool.Get().(*bytes.Buffer) // io.ReadWriter
	buf.Reset()                                 // important
	defer textBufferPool.Put(buf)               // important

	if err = wechatjson.NewEncoder(buf).Encode(request); err != nil {
		return
	}

	httpResp, err := clt.HttpClient.Post(url, "application/json; charset=utf-8", buf)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("http.Status: %s", httpResp.Status)
	}
	if err = json.NewDecoder(httpResp.Body).Decode(response); err != nil {
		return
	}
	return
}

// GET 微信资源 url, 把微信服务器返回的 JSON 解析到 response.
func (clt *WechatClient) GetJSON(url string, response interface{}) (err error) {
	httpResp, err := clt.HttpClient.Get(url)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("http.Status: %s", httpResp.Status)
	}
	if err = json.NewDecoder(httpResp.Body).Decode(response); err != nil {
		return
	}
	return
}
