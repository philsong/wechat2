// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

import (
	"net/http"
	"net/url"
)

// 实现了 http.Handler, 处理一个公众号的消息
type WechatServerFrontend struct {
	wechatServer          WechatServer
	invalidRequestHandler InvalidRequestHandler
}

func NewWechatServerFrontend(wechatServer WechatServer, invalidRequestHandler InvalidRequestHandler) *WechatServerFrontend {
	if wechatServer == nil {
		panic("mp: nil wechatServer")
	}
	if invalidRequestHandler == nil {
		invalidRequestHandler = DefaultInvalidRequestHandler
	}

	return &WechatServerFrontend{
		wechatServer:          wechatServer,
		invalidRequestHandler: invalidRequestHandler,
	}
}

func (front *WechatServerFrontend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wechatServer := front.wechatServer
	invalidRequestHandler := front.invalidRequestHandler

	urlValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		invalidRequestHandler.ServeInvalidRequest(w, r, err)
		return
	}

	ServeHTTP(w, r, urlValues, wechatServer, invalidRequestHandler)
}
