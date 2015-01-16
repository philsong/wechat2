// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechatv2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechatv2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"

	"github.com/chanxuehong/wechatv2/util"
)

// 回复消息给微信服务器.
// 要求 msg 是有效的消息数据结构(经过 encoding/xml marshal 后符合消息的格式).
func WriteRawResponse(w http.ResponseWriter, r *Request, msg interface{}) (err error) {
	if w == nil {
		return errors.New("nil http.ResponseWriter")
	}
	if msg == nil {
		return errors.New("nil message")
	}
	return xml.NewEncoder(w).Encode(msg)
}

// 安全模式 和 兼容模式, 回复微信请求的 http body
type ResponseHttpBody struct {
	XMLName      struct{} `xml:"xml" json:"-"`
	EncryptedMsg string   `xml:"Encrypt"`
	MsgSignature string   `xml:"MsgSignature"`
	TimeStamp    int64    `xml:"TimeStamp"`
	Nonce        string   `xml:"Nonce"`
}

// 回复消息给微信服务器.
// 要求 msg 是有效的消息数据结构(经过 encoding/xml marshal 后符合消息的格式).
func WriteAESResponse(w http.ResponseWriter, r *Request, msg interface{}) (err error) {
	if w == nil {
		return errors.New("nil http.ResponseWriter")
	}
	if r == nil {
		return errors.New("nil Request")
	}
	if msg == nil {
		return errors.New("nil message")
	}

	MsgRawXML, err := xml.Marshal(msg)
	if err != nil {
		return
	}

	EncryptedMsg := util.AESEncryptMsg(r.Random, MsgRawXML, r.WechatAppId, r.AESKey)
	base64EncryptedMsg := base64.StdEncoding.EncodeToString(EncryptedMsg)

	responseHttpBody := ResponseHttpBody{
		EncryptedMsg: base64EncryptedMsg,
		TimeStamp:    r.TimeStamp,
		Nonce:        r.Nonce,
	}

	timestampStr := strconv.FormatInt(responseHttpBody.TimeStamp, 10)
	responseHttpBody.MsgSignature = util.MsgSign(r.WechatToken, timestampStr,
		responseHttpBody.Nonce, responseHttpBody.EncryptedMsg)

	return xml.NewEncoder(w).Encode(&responseHttpBody)
}
