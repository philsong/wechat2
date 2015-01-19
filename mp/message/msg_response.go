// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechatv2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechatv2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package message

import (
	"errors"
	"fmt"

	"github.com/chanxuehong/wechatv2/mp"
)

const (
	MsgResponseTypeText                    = "text"                      // 文本消息
	MsgResponseTypeImage                   = "image"                     // 图片消息
	MsgResponseTypeVoice                   = "voice"                     // 语音消息
	MsgResponseTypeVideo                   = "video"                     // 视频消息
	MsgResponseTypeMusic                   = "music"                     // 音乐消息
	MsgResponseTypeNews                    = "news"                      // 图文消息
	MsgResponseTypeTransferCustomerService = "transfer_customer_service" // 将消息转发到多客服
)

// 文本消息
type TextResponse struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	Content string `xml:"Content" json:"Content"` // 回复的消息内容, 支持换行符
}

// 新建文本消息
//  NOTE: content 支持换行符
func NewTextResponse(to, from, content string, timestamp int64) (text *TextResponse) {
	return &TextResponse{
		CommonMessageHeader: mp.CommonMessageHeader{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   timestamp,
			MsgType:      MsgResponseTypeText,
		},
		Content: content,
	}
}

// 图片消息
type ImageResponse struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	Image struct {
		MediaId string `xml:"MediaId" json:"MediaId"` // MediaId 通过上传多媒体文件得到
	} `xml:"Image" json:"Image"`
}

// 新建图片消息
//  MediaId 通过上传多媒体文件得到
func NewImageResponse(to, from, mediaId string, timestamp int64) (image *ImageResponse) {
	image = &ImageResponse{
		CommonMessageHeader: mp.CommonMessageHeader{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   timestamp,
			MsgType:      MsgResponseTypeImage,
		},
	}
	image.Image.MediaId = mediaId
	return
}

// 语音消息
type VoiceResponse struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	Voice struct {
		MediaId string `xml:"MediaId" json:"MediaId"` // MediaId 通过上传多媒体文件得到
	} `xml:"Voice" json:"Voice"`
}

// 新建语音消息
//  MediaId 通过上传多媒体文件得到
func NewVoiceResponse(to, from, mediaId string, timestamp int64) (voice *VoiceResponse) {
	voice = &VoiceResponse{
		CommonMessageHeader: mp.CommonMessageHeader{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   timestamp,
			MsgType:      MsgResponseTypeVoice,
		},
	}
	voice.Voice.MediaId = mediaId
	return
}

// 视频消息
type VideoResponse struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	Video struct {
		MediaId     string `xml:"MediaId"               json:"MediaId"`               // MediaId 通过上传多媒体文件得到
		Title       string `xml:"Title,omitempty"       json:"Title,omitempty"`       // 视频消息的标题
		Description string `xml:"Description,omitempty" json:"Description,omitempty"` // 视频消息的描述
	} `xml:"Video" json:"Video"`
}

// 新建视频消息
//  MediaId 通过上传多媒体文件得到
//  title, description 可以为 ""
func NewVideoResponse(to, from, mediaId, title, description string, timestamp int64) (video *VideoResponse) {
	video = &VideoResponse{
		CommonMessageHeader: mp.CommonMessageHeader{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   timestamp,
			MsgType:      MsgResponseTypeVideo,
		},
	}
	video.Video.MediaId = mediaId
	video.Video.Title = title
	video.Video.Description = description
	return
}

// 音乐消息
type MusicResponse struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	Music struct {
		Title        string `xml:"Title,omitempty"       json:"Title,omitempty"`       // 音乐标题
		Description  string `xml:"Description,omitempty" json:"Description,omitempty"` // 音乐描述
		MusicURL     string `xml:"MusicUrl"              json:"MusicUrl"`              // 音乐链接
		HQMusicURL   string `xml:"HQMusicUrl"            json:"HQMusicUrl"`            // 高质量音乐链接, WIFI环境优先使用该链接播放音乐
		ThumbMediaId string `xml:"ThumbMediaId"          json:"ThumbMediaId"`          // 缩略图的媒体id, 通过上传多媒体文件得到
	} `xml:"Music" json:"Music"`
}

// 新建音乐消息
//  thumbMediaId 通过上传多媒体文件得到
//  title, description 可以为 ""
func NewMusicResponse(to, from, thumbMediaId, musicURL,
	HQMusicURL, title, description string, timestamp int64) (music *MusicResponse) {

	music = &MusicResponse{
		CommonMessageHeader: mp.CommonMessageHeader{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   timestamp,
			MsgType:      MsgResponseTypeMusic,
		},
	}
	music.Music.Title = title
	music.Music.Description = description
	music.Music.MusicURL = musicURL
	music.Music.HQMusicURL = HQMusicURL
	music.Music.ThumbMediaId = thumbMediaId
	return
}

// 图文消息里的 Article
type NewsArticle struct {
	Title       string `xml:"Title,omitempty"       json:"Title,omitempty"`       // 图文消息标题
	Description string `xml:"Description,omitempty" json:"Description,omitempty"` // 图文消息描述
	PicURL      string `xml:"PicUrl,omitempty"      json:"PicUrl,omitempty"`      // 图片链接, 支持JPG, PNG格式, 较好的效果为大图360*200, 小图200*200
	URL         string `xml:"Url,omitempty"         json:"Url,omitempty"`         // 点击图文消息跳转链接
}

const (
	NewsArticleCountLimit = 10 // 被动回复图文消息的文章数据最大数
)

// 图文消息.
//  NOTE: Articles 赋值的同时也要更改 ArticleCount 字段, 建议用 NewNews() 和 News.AppendArticle()
type NewsResponse struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	ArticleCount int           `xml:"ArticleCount"            json:"ArticleCount"`       // 图文消息个数, 限制为10条以内
	Articles     []NewsArticle `xml:"Articles>item,omitempty" json:"Articles,omitempty"` // 多条图文消息信息, 默认第一个item为大图, 注意, 如果图文数超过10, 则将会无响应
}

// NOTE: articles 的长度不能超过 NewsArticleCountLimit
func NewNewsResponse(to, from string, articles []NewsArticle, timestamp int64) (news *NewsResponse) {
	news = &NewsResponse{
		CommonMessageHeader: mp.CommonMessageHeader{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   timestamp,
			MsgType:      MsgResponseTypeNews,
		},
	}
	news.Articles = articles
	news.ArticleCount = len(articles)
	return
}

// 更新 news.ArticleCount 字段, 使其等于 len(news.Articles)
func (news *NewsResponse) UpdateArticleCount() {
	news.ArticleCount = len(news.Articles)
}

// 增加文章到图文消息中, 该方法会自动更新 News.ArticleCount 字段
func (news *NewsResponse) AppendArticle(article ...NewsArticle) {
	news.Articles = append(news.Articles, article...)
	news.ArticleCount = len(news.Articles)
}

// 检查 News 是否有效，有效返回 nil，否则返回错误信息
func (news *NewsResponse) CheckValid() (err error) {
	n := len(news.Articles)

	if n != news.ArticleCount {
		err = fmt.Errorf("图文消息的 ArticleCount == %d, 实际文章个数为 %d", news.ArticleCount, n)
		return
	}
	if n <= 0 {
		err = errors.New("图文消息里没有文章")
		return
	}
	if n > NewsArticleCountLimit {
		err = fmt.Errorf("图文消息的文章个数不能超过 %d, 现在为 %d", NewsArticleCountLimit, n)
		return
	}
	return
}

// 将消息转发到多客服
type TransferToCustomerServiceResponse struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader
}

func NewTransferToCustomerServiceResponse(to, from string, timestamp int64) *TransferToCustomerServiceResponse {
	return &TransferToCustomerServiceResponse{
		CommonMessageHeader: mp.CommonMessageHeader{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   timestamp,
			MsgType:      MsgResponseTypeTransferCustomerService,
		},
	}
}

// 将消息转发到指定客服
type TransferToSpecialCustomerServiceResponse struct {
	XMLName struct{} `xml:"xml" json:"-"`
	mp.CommonMessageHeader

	TransInfo struct {
		KfAccount string `xml:"KfAccount"         json:"KfAccount"`
	} `xml:"TransInfo"         json:"TransInfo"`
}

func NewTransferToSpecialCustomerServiceResponse(to, from, KfAccount string, timestamp int64) (msg *TransferToSpecialCustomerServiceResponse) {
	msg = &TransferToSpecialCustomerServiceResponse{
		CommonMessageHeader: mp.CommonMessageHeader{
			ToUserName:   to,
			FromUserName: from,
			CreateTime:   timestamp,
			MsgType:      MsgResponseTypeTransferCustomerService,
		},
	}
	msg.TransInfo.KfAccount = KfAccount
	return
}
