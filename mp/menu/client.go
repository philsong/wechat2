// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechatv2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechatv2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package menu

import (
	"net/http"

	"github.com/chanxuehong/wechatv2/mp"
)

type Client struct {
	mp.WechatClient
}

// 创建一个新的 Client.
//  如果 HttpClient == nil 则默认用 http.DefaultClient
func NewClient(TokenService mp.TokenService, HttpClient *http.Client) *Client {
	if TokenService == nil {
		panic("TokenService == nil")
	}
	if HttpClient == nil {
		HttpClient = http.DefaultClient
	}

	return &Client{
		WechatClient: mp.WechatClient{
			TokenService: TokenService,
			HttpClient:   HttpClient,
		},
	}
}

// 创建自定义菜单.
func (clt *Client) CreateMenu(menu Menu) (err error) {
	var result mp.Error

	token, err := clt.Token()
	if err != nil {
		return
	}

	hasRetried := false
RETRY:
	url := "https://api.weixin.qq.com/cgi-bin/menu/create?access_token=" + token

	if err = clt.PostJSON(url, menu, &result); err != nil {
		return
	}

	switch result.ErrCode {
	case mp.ErrCodeOK:
		return
	case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
		if !hasRetried {
			hasRetried = true

			if token, err = clt.GetNewToken(); err != nil {
				return
			}
			goto RETRY
		}
		fallthrough
	default:
		err = &result
		return
	}
}

// 删除自定义菜单
func (clt *Client) DeleteMenu() (err error) {
	var result mp.Error

	token, err := clt.Token()
	if err != nil {
		return
	}

	hasRetried := false
RETRY:
	url := "https://api.weixin.qq.com/cgi-bin/menu/delete?access_token=" + token

	if err = clt.GetJSON(url, &result); err != nil {
		return
	}

	switch result.ErrCode {
	case mp.ErrCodeOK:
		return
	case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
		if !hasRetried {
			hasRetried = true

			if token, err = clt.GetNewToken(); err != nil {
				return
			}
			goto RETRY
		}
		fallthrough
	default:
		err = &result
		return
	}
}

// 获取自定义菜单
func (clt *Client) GetMenu() (menu Menu, err error) {
	var result struct {
		mp.Error
		Menu Menu `json:"menu"`
	}

	token, err := clt.Token()
	if err != nil {
		return
	}

	hasRetried := false
RETRY:
	url := "https://api.weixin.qq.com/cgi-bin/menu/get?access_token=" + token

	if err = clt.GetJSON(url, &result); err != nil {
		return
	}

	switch result.ErrCode {
	case mp.ErrCodeOK:
		menu = result.Menu
		return
	case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
		if !hasRetried {
			hasRetried = true

			if token, err = clt.GetNewToken(); err != nil {
				return
			}
			goto RETRY
		}
		fallthrough
	default:
		err = &result.Error
		return
	}
}
