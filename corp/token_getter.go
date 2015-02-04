// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package corp

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 从微信服务器获取 access_token 的接口.
type TokenGetter interface {
	GetToken() (token string, err error)
}

var _ TokenGetter = new(DefaultTokenGetter)

type DefaultTokenGetter struct {
	corpId     string
	corpSecret string
	httpClient *http.Client
}

// NewDefaultTokenGetter 创建一个新的 DefaultTokenGetter.
//  如果 httpClient == nil 则默认使用 http.DefaultClient
func NewDefaultTokenGetter(corpId, corpSecret string, httpClient *http.Client) *DefaultTokenGetter {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &DefaultTokenGetter{
		corpId:     corpId,
		corpSecret: corpSecret,
		httpClient: httpClient,
	}
}

func (getter *DefaultTokenGetter) GetToken() (token string, err error) {
	url := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + getter.corpId +
		"&corpsecret=" + getter.corpSecret

	httpResp, err := getter.httpClient.Get(url)
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
		Token     string `json:"access_token"`
		ExpiresIn int64  `json:"expires_in"`
	}
	if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return
	}

	if result.ErrCode != ErrCodeOK {
		err = &result.Error
		return
	}

	token = result.Token
	return
}
