// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechatv2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechatv2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

import (
	"net/http"
	"net/url"
)

type HttpHandler struct {
	wechatServer          WechatServer
	invalidRequestHandler InvalidRequestHandler
}

func NewHttpHandler(wechatServer WechatServer, invalidRequestHandler InvalidRequestHandler) *HttpHandler {
	if wechatServer == nil {
		panic("mp: nil wechatServer")
	}
	if invalidRequestHandler == nil {
		invalidRequestHandler = DefaultInvalidRequestHandler
	}

	return &HttpHandler{
		wechatServer:          wechatServer,
		invalidRequestHandler: invalidRequestHandler,
	}
}

func (handler *HttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wechatServer := handler.wechatServer
	invalidRequestHandler := handler.invalidRequestHandler

	urlValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		invalidRequestHandler.ServeInvalidRequest(w, r, err)
		return
	}

	ServeHTTP(w, r, urlValues, wechatServer, invalidRequestHandler)
}
