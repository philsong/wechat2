// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/philsong/wechat2 for the canonical source repository
// @license     https://github.com/philsong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package pay

import (
	"fmt"
)

type Error struct {
	XMLName    struct{} `xml:"xml"                  json:"-"`
	ReturnCode string   `xml:"return_code"          json:"return_code"`
	ReturnMsg  string   `xml:"return_msg,omitempty" json:"return_msg,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("return_code: %q, return_msg: %q", e.ReturnCode, e.ReturnMsg)
}
