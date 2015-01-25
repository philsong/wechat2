// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package jssdk

// jsapi_ticket 中控服务器接口.
type TicketServer interface {
	// 从中控服务器获取 jsapi_ticket, 该 jsapi_ticket 一般缓存在某个地方.
	Ticket() (ticket string, err error)

	// 请求 jsapi_ticket 中控服务器到微信服务器刷新 jsapi_ticket.
	TicketRefresh() (ticket string, err error)
}
