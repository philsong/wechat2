// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechatv2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechatv2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package message

import (
	"github.com/chanxuehong/wechatv2/mp"
)

const (
	// 微信服务器推送过来的消息类型
	MsgRequestTypeText     mp.MessageType = "text"     // 文本消息
	MsgRequestTypeImage                   = "image"    // 图片消息
	MsgRequestTypeVoice                   = "voice"    // 语音消息
	MsgRequestTypeVideo                   = "video"    // 视频消息
	MsgRequestTypeLocation                = "location" // 地理位置消息
	MsgRequestTypeLink                    = "link"     // 链接消息
	MsgRequestTypeEvent                   = "event"    // 事件推送
)

// 文本消息
type TextRequest struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	MsgId   int64  `xml:"MsgId"   json:"MsgId"`   // 消息id, 64位整型
	Content string `xml:"Content" json:"Content"` // 文本消息内容
}

func GetTextRequest(msg *mp.MixedMessage) *TextRequest {
	return &TextRequest{
		CommonMessageHeader: msg.CommonMessageHeader,
		MsgId:               msg.MsgId,
		Content:             msg.Content,
	}
}

// 图片消息
type ImageRequest struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	MsgId   int64  `xml:"MsgId"   json:"MsgId"`   // 消息id, 64位整型
	MediaId string `xml:"MediaId" json:"MediaId"` // 图片消息媒体id，可以调用多媒体文件下载接口拉取数据。
	PicURL  string `xml:"PicUrl"  json:"PicUrl"`  // 图片链接
}

func GetImageRequest(msg *mp.MixedMessage) *ImageRequest {
	return &ImageRequest{
		CommonMessageHeader: msg.CommonMessageHeader,
		MsgId:               msg.MsgId,
		MediaId:             msg.MediaId,
		PicURL:              msg.PicURL,
	}
}

// 语音消息
type VoiceRequest struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	MsgId   int64  `xml:"MsgId"   json:"MsgId"`   // 消息id, 64位整型
	MediaId string `xml:"MediaId" json:"MediaId"` // 语音消息媒体id，可以调用多媒体文件下载接口拉取该媒体
	Format  string `xml:"Format"  json:"Format"`  // 语音格式，如amr，speex等

	// 语音识别结果，UTF8编码，
	// NOTE: 需要开通语音识别功能，否则该字段为空，即使开通了语音识别该字段还是有可能为空
	Recognition string `xml:"Recognition,omitempty" json:"Recognition,omitempty"`
}

func GetVoiceRequest(msg *mp.MixedMessage) *VoiceRequest {
	return &VoiceRequest{
		CommonMessageHeader: msg.CommonMessageHeader,
		MsgId:               msg.MsgId,
		MediaId:             msg.MediaId,
		Format:              msg.Format,
		Recognition:         msg.Recognition,
	}
}

// 视频消息
type VideoRequest struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	MsgId        int64  `xml:"MsgId"        json:"MsgId"`        // 消息id, 64位整型
	MediaId      string `xml:"MediaId"      json:"MediaId"`      // 视频消息媒体id，可以调用多媒体文件下载接口拉取数据。
	ThumbMediaId string `xml:"ThumbMediaId" json:"ThumbMediaId"` // 视频消息缩略图的媒体id，可以调用多媒体文件下载接口拉取数据。
}

func GetVideoRequest(msg *mp.MixedMessage) *VideoRequest {
	return &VideoRequest{
		CommonMessageHeader: msg.CommonMessageHeader,
		MsgId:               msg.MsgId,
		MediaId:             msg.MediaId,
		ThumbMediaId:        msg.ThumbMediaId,
	}
}

// 地理位置消息
type LocationRequest struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	MsgId     int64   `xml:"MsgId"      json:"MsgId"`      // 消息id, 64位整型
	LocationX float64 `xml:"Location_X" json:"Location_X"` // 地理位置纬度
	LocationY float64 `xml:"Location_Y" json:"Location_Y"` // 地理位置经度
	Scale     int     `xml:"Scale"      json:"Scale"`      // 地图缩放大小
	Label     string  `xml:"Label"      json:"Label"`      // 地理位置信息
}

func GetLocationRequest(msg *mp.MixedMessage) *LocationRequest {
	return &LocationRequest{
		CommonMessageHeader: msg.CommonMessageHeader,
		MsgId:               msg.MsgId,
		LocationX:           msg.LocationX,
		LocationY:           msg.LocationY,
		Scale:               msg.Scale,
		Label:               msg.Label,
	}
}

// 链接消息
type LinkRequest struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	MsgId       int64  `xml:"MsgId"       json:"MsgId"`       // 消息id, 64位整型
	Title       string `xml:"Title"       json:"Title"`       // 消息标题
	Description string `xml:"Description" json:"Description"` // 消息描述
	URL         string `xml:"Url"         json:"Url"`         // 消息链接
}

func GetLinkRequest(msg *mp.MixedMessage) *LinkRequest {
	return &LinkRequest{
		CommonMessageHeader: msg.CommonMessageHeader,
		MsgId:               msg.MsgId,
		Title:               msg.Title,
		Description:         msg.Description,
		URL:                 msg.URL,
	}
}
