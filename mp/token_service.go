// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

// access_token 伺服接口, 用于集中式获取 access_token 场景, see token_service.png.
type TokenService interface {
	// 获取 access_token, 该 access_token 一般缓存在某个地方.
	Token() (token string, err error)

	// 从微信服务器获取新的 access_token.
	TokenRefresh() (token string, err error)
}
