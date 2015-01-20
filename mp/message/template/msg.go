// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechatv2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechatv2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package template

type TemplateMessage struct {
	ToUser     string `json:"touser"`             // 必须, 接受者OpenID
	TemplateId string `json:"template_id"`        // 必须, 模版ID
	URL        string `json:"url,omitempty"`      // 可选, 用户点击后跳转的URL，该URL必须处于开发者在公众平台网站中设置的域中
	TopColor   string `json:"topcolor,omitempty"` // 可选, 整个消息的颜色, 可以不设置

	// 必须, 模版调用时需要的参数赋值内容，具体有哪些参数视模版内容而定;
	// Data 一般是一个 *struct 或者 map, 要求被 encoding/json 格式化后满足特定的模板需求.
	Data interface{} `json:"data"`
}
