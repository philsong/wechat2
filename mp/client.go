// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	wechatjson "github.com/chanxuehong/wechat2/json"
)

// 微信公众号"主动"请求功能的基本封装.
type WechatClient struct {
	accessToken string // 缓存当前的 access_token

	TokenServer TokenServer
	HttpClient  *http.Client
}

// 获取 access_token.
func (clt *WechatClient) Token() (token string, err error) {
	token, err = clt.TokenServer.Token()
	if err != nil {
		clt.accessToken = ""
		return
	}
	clt.accessToken = token
	return
}

// 请求微信服务器更新 access_token.
//  NOTE:
//  1. 一般情况下无需调用该函数, 请使用 Token() 获取 access_token.
//  2. 即使 access_token 失效(错误代码 40001, 正常情况下不会出现),
//     也请谨慎调用 TokenRefresh, 建议直接返回错误! 因为很有可能高并发情况下造成雪崩效应!
//  3. 再次强调, 调用这个函数你应该知道发生了什么!!!
func (clt *WechatClient) TokenRefresh() (token string, err error) {
	token, err = clt.TokenServer.TokenRefresh()
	if err != nil {
		clt.accessToken = ""
		return
	}
	clt.accessToken = token
	return
}

// 当 WechatClient.Token() 返回的 access_token 失效时获取新的 access_token.
func (clt *WechatClient) GetNewToken() (token string, err error) {
	// 失效有两种可能:
	// 1. 中控服务器更新了 access_token, 但是没有及时更新到缓存, 导致此次 WechatClient.Token()
	//    获取到的不是有效的 access_token;
	//    (NOTE: 目前这种情况基本不会出现, 微信服务器兼容更新时刻多个 access_token 都有效)
	// 2. 就是微信服务器主动失效了 access_token, 但是中控服务器不知道这个情况而没有及时更新
	//    access_token, 所以这个时候就需要主动刷新 access_token.

	// 策略:
	//     先到中控服务器去查询是否有新的 access_token, 如果没有新的 access_token 则请求调用
	// WechatClient.TokenRefresh() 返回 access_token.
	token, err = clt.TokenServer.Token()
	if err != nil {
		clt.accessToken = ""
		return
	}
	if clt.accessToken != token {
		clt.accessToken = token
		return
	}
	return clt.TokenRefresh()
}

// 用 encoding/json 把 request marshal 为 JSON, 放入 http 请求的 body 中,
// POST 到微信服务器, 然后将微信服务器返回的 JSON 用 encoding/json 解析到 response.
//
//  NOTE:
//  1. 一般不用调用这个方法, 请直接调用高层次的封装方法;
//  2. 最终的 URL == incompleteURL + access_token;
//  3. response 要求是 struct 的指针, 并且该 struct 拥有属性:
//     ErrCode int `json:"errcode"` (可以是直接属性, 也可以是匿名属性里的属性)
func (clt *WechatClient) PostJSON(incompleteURL string, request interface{}, response interface{}) (err error) {
	buf := textBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer textBufferPool.Put(buf)

	if err = wechatjson.NewEncoder(buf).Encode(request); err != nil {
		return
	}
	requestBytes := buf.Bytes()

	token, err := clt.Token()
	if err != nil {
		return
	}

	hasRetried := false
RETRY:
	finalURL := incompleteURL + token

	httpResp, err := clt.HttpClient.Post(finalURL, "application/json; charset=utf-8", bytes.NewReader(requestBytes))
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

	// 请注意:
	// 下面获取 ErrCode 的代码不具备通用性!!!
	//
	// 因为本 SDK 的 response 都是
	//  struct {
	//    Error
	//    XXX
	//  }
	// 的结构, 所以用下面简单的方法得到 ErrCode.
	//
	// 如果你是直接调用这个函数, 那么要根据你的 response 数据结构修改下面的代码.
	ErrCode := reflect.ValueOf(response).Elem().FieldByName("ErrCode").Int()

	switch ErrCode {
	case ErrCodeOK:
		return
	case ErrCodeInvalidCredential, ErrCodeTimeout:
		if !hasRetried {
			hasRetried = true

			if token, err = clt.GetNewToken(); err != nil {
				return
			}
			goto RETRY
		}
		fallthrough
	default:
		return
	}
}

// GET 微信资源, 然后将微信服务器返回的 JSON 用 encoding/json 解析到 response.
//
//  NOTE:
//  1. 一般不用调用这个方法, 请直接调用高层次的封装方法;
//  2. 最终的 URL == incompleteURL + access_token;
//  3. response 要求是 struct 的指针, 并且该 struct 拥有属性:
//     ErrCode int `json:"errcode"` (可以是直接属性, 也可以是匿名属性里的属性)
func (clt *WechatClient) GetJSON(incompleteURL string, response interface{}) (err error) {
	token, err := clt.Token()
	if err != nil {
		return
	}

	hasRetried := false
RETRY:
	finalURL := incompleteURL + token

	httpResp, err := clt.HttpClient.Get(finalURL)
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

	// 请注意:
	// 下面获取 ErrCode 的代码不具备通用性!!!
	//
	// 因为本 SDK 的 response 都是
	//  struct {
	//    Error
	//    XXX
	//  }
	// 的结构, 所以用下面简单的方法得到 ErrCode.
	//
	// 如果你是直接调用这个函数, 那么要根据你的 response 数据结构修改下面的代码.
	ErrCode := reflect.ValueOf(response).Elem().FieldByName("ErrCode").Int()

	switch ErrCode {
	case ErrCodeOK:
		return
	case ErrCodeInvalidCredential, ErrCodeTimeout:
		if !hasRetried {
			hasRetried = true

			if token, err = clt.GetNewToken(); err != nil {
				return
			}
			goto RETRY
		}
		fallthrough
	default:
		return
	}
}
