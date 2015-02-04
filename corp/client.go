// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package corp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"time"

	wechatjson "github.com/chanxuehong/wechat2/json"
)

// CorpClient 封装了主动请求功能.
type CorpClient struct {
	accessToken string // 缓存当前的 access_token

	// 使用 TokenCache 而不是和公众平台一样使用 TokenServer 是因为企业号如果使用 TokenServer
	// 则 access_token 会一直不变, 这样对于安全是不利的!
	TokenCache  TokenCache
	TokenGetter TokenGetter
	HttpClient  *http.Client
}

// 从缓存中获取 access_token, 如果缓存中没有则从微信服务器获取.
func (clt *CorpClient) Token() (token string, err error) {
	token, err = clt.TokenCache.Token()
	switch {
	case err == nil:
		clt.accessToken = token
		return
	case err == ErrCacheMiss:
		return clt.TokenRefresh()
	default:
		clt.accessToken = ""
		return
	}
}

// 从微信服务器获取 access_token 并更新到 TokenCache.
func (clt *CorpClient) TokenRefresh() (token string, err error) {
	if token, err = clt.TokenGetter.GetToken(); err != nil {
		clt.accessToken = ""
		return
	}
	if err = clt.TokenCache.PutToken(token); err != nil {
		clt.accessToken = ""
		return
	}
	clt.accessToken = token
	return
}

var mathRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// 返回一个 [2, 7) 的随机数.
func getRetryNum() int {
	return mathRand.Intn(5) + 2
}

// 当 CorpClient.Token() 返回的 access_token 过期时获取新的 access_token.
func (clt *CorpClient) GetNewToken() (token string, err error) {
	retryNum := getRetryNum()
	for i := 0; ; {
		token, err = clt.TokenCache.Token()
		if err != nil {
			if err == ErrCacheMiss {
				return clt.TokenRefresh() // 刷新 access_token
			}
			clt.accessToken = ""
			return
		}
		if clt.accessToken != token {
			clt.accessToken = token
			return
		}

		if i++; i < retryNum { // 这样写是避免最后一次还要等 50ms
			time.Sleep(50 * time.Millisecond) // 50ms 后再次尝试获取 access_token
			continue
		}
		return clt.TokenRefresh() // 刷新 access_token
	}
}

// 用 encoding/json 把 request marshal 为 JSON, 放入 http 请求的 body 中,
// POST 到微信服务器, 然后将微信服务器返回的 JSON 用 encoding/json 解析到 response.
//
//  NOTE:
//  1. 一般不用调用这个方法, 请直接调用高层次的封装方法;
//  2. 最终的 URL == incompleteURL + access_token;
//  3. response 要求是 struct 的指针, 并且该 struct 拥有属性:
//     ErrCode int `json:"errcode"` (可以是直接属性, 也可以是匿名属性里的属性)
func (clt *CorpClient) PostJSON(incompleteURL string, request interface{}, response interface{}) (err error) {
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
	case ErrCodeTimeout, ErrCodeInvalidCredential:
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
func (clt *CorpClient) GetJSON(incompleteURL string, response interface{}) (err error) {
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
	case ErrCodeTimeout, ErrCodeInvalidCredential:
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
