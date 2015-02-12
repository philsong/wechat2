// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/philsong/wechat2 for the canonical source repository
// @license     https://github.com/philsong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

import (
	"fmt"
	"net/http"
	"net/url"
)

// 实现了 http.Handler, 处理一个公众号的消息(事件)请求.
type WechatServerFrontend struct {
	wechatServer          WechatServer
	invalidRequestHandler InvalidRequestHandler
}

func NewWechatServerFrontend(server WechatServer, handler InvalidRequestHandler) *WechatServerFrontend {
	if server == nil {
		panic("mp: nil WechatServer")
	}
	if handler == nil {
		handler = DefaultInvalidRequestHandler
	}

	return &WechatServerFrontend{
		wechatServer:          server,
		invalidRequestHandler: handler,
	}
}

// 实现 http.Handler.
func (frontend *WechatServerFrontend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wechatServer := frontend.wechatServer
	invalidRequestHandler := frontend.invalidRequestHandler

	urlValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		fmt.Println("err", err)
		invalidRequestHandler.ServeInvalidRequest(w, r, err)
		return
	}

	ServeHTTP(w, r, urlValues, wechatServer, invalidRequestHandler)
}
