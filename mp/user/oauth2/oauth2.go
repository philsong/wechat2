// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package oauth2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chanxuehong/wechat2/mp"
)

// 用户相关的 oauth2 token 信息
type OAuth2Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64 // 过期时间, unixtime, 分布式系统要求时间同步, 建议使用 NTP

	OpenId string
	Scopes []string // 用户授权的作用域
}

// 判断授权的 access token 是否过期, 过期返回 true, 否则返回 false
func (token *OAuth2Token) accessTokenExpired() bool {
	return time.Now().Unix() > token.ExpiresAt
}

type Client struct {
	*OAuth2Config
	*OAuth2Token // 程序会自动更新最新的 OAuth2Token 到这个字段, 如有必要该字段可以保存起来

	HttpClient *http.Client // 如果 httpClient == nil 则默认用 http.DefaultClient
}

func (clt *Client) httpClient() *http.Client {
	if clt.HttpClient != nil {
		return clt.HttpClient
	}
	return http.DefaultClient
}

// 通过code换取网页授权 access_token.
//  NOTE:
//  1. Client 需要指定 OAuth2Config
//  2. 如果指定了 OAuth2Token, 则会更新这个 OAuth2Token, 同时返回的也是指定的 OAuth2Token;
//     否则会重新分配一个 OAuth2Token.
func (clt *Client) Exchange(code string) (token *OAuth2Token, err error) {
	if clt.OAuth2Config == nil {
		err = errors.New("没有提供 OAuth2Config")
		return
	}

	tk := clt.OAuth2Token
	if tk == nil {
		tk = new(OAuth2Token)
	}

	if err = clt.updateToken(tk, oauth2ExchangeTokenURL(clt.AppId, clt.AppSecret, code)); err != nil {
		return
	}

	clt.OAuth2Token = tk
	token = tk
	return
}

// 刷新access_token（如果需要）.
//  NOTE: Client 需要指定 OAuth2Config, OAuth2Token
func (clt *Client) TokenRefresh() (token *OAuth2Token, err error) {
	if clt.OAuth2Config == nil {
		err = errors.New("没有提供 OAuth2Config")
		return
	}
	if clt.OAuth2Token == nil {
		err = errors.New("没有提供 OAuth2Token")
		return
	}
	if len(clt.RefreshToken) == 0 {
		err = errors.New("没有有效的 RefreshToken")
		return
	}

	if err = clt.updateToken(clt.OAuth2Token, oauth2RefreshTokenURL(clt.AppId, clt.RefreshToken)); err != nil {
		return
	}

	token = clt.OAuth2Token
	return
}

// 检查 access_token 是否有效.
//  NOTE:
//  1. Client 需要指定 OAuth2Token
//  2. 先判断 err 然后再判断 valid
func (clt *Client) CheckAccessTokenValid() (valid bool, err error) {
	if clt.OAuth2Token == nil {
		err = errors.New("没有提供 OAuth2Token")
		return
	}
	if len(clt.AccessToken) == 0 {
		err = errors.New("没有有效的 AccessToken")
		return
	}
	if len(clt.OpenId) == 0 {
		err = errors.New("没有有效的 OpenId")
		return
	}

	httpResp, err := clt.httpClient().Get(checkAccessTokenValidURL(clt.AccessToken, clt.OpenId))
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", httpResp.Status)
		return
	}

	var result mp.Error
	if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return
	}

	switch result.ErrCode {
	case mp.ErrCodeOK:
		valid = true
		return
	case 40001:
		return
	default:
		err = &result
		return
	}
}

// 从服务器获取新的 token 更新 tk
func (clt *Client) updateToken(tk *OAuth2Token, url string) (err error) {
	httpResp, err := clt.httpClient().Get(url)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("http.Status: %s", httpResp.Status)
	}

	var result struct {
		mp.Error
		AccessToken  string `json:"access_token"`  // 网页授权接口调用凭证,注意：此access_token与基础支持的access_token不同
		RefreshToken string `json:"refresh_token"` // 用户刷新access_token
		ExpiresIn    int64  `json:"expires_in"`    // access_token接口调用凭证超时时间，单位（秒）
		OpenId       string `json:"openid"`        // 用户唯一标识，请注意，在未关注公众号时，用户访问公众号的网页，也会产生一个用户和公众号唯一的OpenID
		Scope        string `json:"scope"`         // 用户授权的作用域，使用逗号（,）分隔
	}

	if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return
	}

	if result.ErrCode != mp.ErrCodeOK {
		return &result.Error
	}

	// 由于网络的延时 以及 分布式服务器之间的时间可能不是绝对同步, access token 过期时间留了一个缓冲区;
	// 正常情况下微信服务器会返回 7200, 则缓冲区的大小为 20 分钟, 这样分布式服务器之间的时间差
	// 在 20 分钟内基本不会出现问题!
	switch {
	case result.ExpiresIn > 60*60:
		result.ExpiresIn -= 60 * 20

	case result.ExpiresIn > 60*30:
		result.ExpiresIn -= 60 * 10

	case result.ExpiresIn > 60*15:
		result.ExpiresIn -= 60 * 5

	case result.ExpiresIn > 60*5:
		result.ExpiresIn -= 60

	case result.ExpiresIn > 60:
		result.ExpiresIn -= 20

	case result.ExpiresIn > 0:

	default:
		err = fmt.Errorf("错误的 expires_in 参数: %d", result.ExpiresIn)
		return
	}

	tk.AccessToken = result.AccessToken
	if len(result.RefreshToken) > 0 {
		tk.RefreshToken = result.RefreshToken
	}
	tk.ExpiresAt = time.Now().Unix() + result.ExpiresIn

	tk.OpenId = result.OpenId
	tk.Scopes = strings.Split(result.Scope, ",")
	return
}
