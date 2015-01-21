// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package oauth2

import (
	"net/url"
)

// 请求用户授权获取code的跳转的地址.
//  appid:       公众号的唯一标识
//  redirectURL: 授权后重定向的回调链接地址，请使用urlencode对链接进行处理
//               用户授权后跳转到 redirectURL?code=CODE&state=STATE
//               用户禁止授权跳转到 redirectURL?state=STATE
//  scope:       应用授权作用域，snsapi_base （不弹出授权页面，直接跳转，只能获取用户openid），
//               snsapi_userinfo （弹出授权页面，可通过openid拿到昵称、性别、所在地。
//               并且，即使在未关注的情况下，只要用户授权，也能获取其信息）
//  state:       重定向后会带上state参数，开发者可以填写a-zA-Z0-9的参数值
func OAuth2AuthCodeURL(appid, redirectURL, scope, state string) string {
	// https://open.weixin.qq.com/connect/oauth2/authorize?appid=APPID
	// &redirect_uri=REDIRECT_URI&response_type=code&scope=SCOPE&state=STATE#wechat_redirect
	return "https://open.weixin.qq.com/connect/oauth2/authorize?appid=" + appid +
		"&redirect_uri=" + url.QueryEscape(redirectURL) +
		"&response_type=code&scope=" + url.QueryEscape(scope) +
		"&state=" + url.QueryEscape(state) + "#wechat_redirect"
}

// 通过code换取网页授权access_token
func oauth2ExchangeTokenURL(appid, appsecret, code string) string {
	// https://api.weixin.qq.com/sns/oauth2/access_token?appid=APPID&secret=SECRET
	// &code=CODE&grant_type=authorization_code
	return "https://api.weixin.qq.com/sns/oauth2/access_token?appid=" + appid +
		"&secret=" + appsecret +
		"&code=" + url.QueryEscape(code) +
		"&grant_type=authorization_code"
}

// 刷新access_token（如果需要）
func oauth2RefreshTokenURL(appid, refreshToken string) string {
	// https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=APPID
	// &grant_type=refresh_token&refresh_token=REFRESH_TOKEN
	return "https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=" + appid +
		"&grant_type=refresh_token&refresh_token=" + refreshToken
}

// 检验授权凭证（access_token）是否有效
func checkAccessTokenValidURL(token, openid string) string {
	// https://api.weixin.qq.com/sns/auth?access_token=ACCESS_TOKEN&openid=OPENID
	return "https://api.weixin.qq.com/sns/auth?access_token=" + token +
		"&openid=" + openid
}

// 拉取用户信息(需scope为 snsapi_userinfo)
func userInfoURL(token, openid, lang string) string {
	// https://api.weixin.qq.com/sns/userinfo?access_token=ACCESS_TOKEN
	// &openid=OPENID&lang=zh_CN
	return "https://api.weixin.qq.com/sns/userinfo?access_token=" + token +
		"&openid=" + openid +
		"&lang=" + lang
}
