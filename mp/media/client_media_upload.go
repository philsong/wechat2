// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechatv2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechatv2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package media

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// 上传多媒体图片
func (clt *Client) MediaUploadImage(_filepath string) (info *MediaInfo, err error) {
	return clt.mediaUpload(MEDIA_TYPE_IMAGE, _filepath)
}

// 上传多媒体语音
func (clt *Client) MediaUploadVoice(_filepath string) (info *MediaInfo, err error) {
	return clt.mediaUpload(MEDIA_TYPE_VOICE, _filepath)
}

// 上传多媒体视频
func (clt *Client) MediaUploadVideo(_filepath string) (info *MediaInfo, err error) {
	return clt.mediaUpload(MEDIA_TYPE_VIDEO, _filepath)
}

// 上传多媒体缩略图
func (clt *Client) MediaUploadThumb(_filepath string) (info *MediaInfo, err error) {
	return clt.mediaUpload(MEDIA_TYPE_THUMB, _filepath)
}

// 上传多媒体
func (clt *Client) mediaUpload(mediaType, _filepath string) (info *MediaInfo, err error) {
	file, err := os.Open(_filepath)
	if err != nil {
		return
	}
	defer file.Close()

	return clt.mediaUploadFromReader(mediaType, filepath.Base(_filepath), file)
}

// 上传多媒体图片
//  NOTE: 参数 filename 不是文件路径, 是指定 multipart form 里面文件名称
func (clt *Client) MediaUploadImageFromReader(filename string, mediaReader io.Reader) (info *MediaInfo, err error) {
	if filename == "" {
		err = errors.New("empty filename")
		return
	}
	if mediaReader == nil {
		err = errors.New("nil mediaReader")
		return
	}
	return clt.mediaUploadFromReader(MEDIA_TYPE_IMAGE, filename, mediaReader)
}

// 上传多媒体语音
//  NOTE: 参数 filename 不是文件路径, 是指定 multipart form 里面文件名称
func (clt *Client) MediaUploadVoiceFromReader(filename string, mediaReader io.Reader) (info *MediaInfo, err error) {
	if filename == "" {
		err = errors.New("empty filename")
		return
	}
	if mediaReader == nil {
		err = errors.New("nil mediaReader")
		return
	}
	return clt.mediaUploadFromReader(MEDIA_TYPE_VOICE, filename, mediaReader)
}

// 上传多媒体视频
//  NOTE: 参数 filename 不是文件路径, 是指定 multipart form 里面文件名称
func (clt *Client) MediaUploadVideoFromReader(filename string, mediaReader io.Reader) (info *MediaInfo, err error) {
	if filename == "" {
		err = errors.New("empty filename")
		return
	}
	if mediaReader == nil {
		err = errors.New("nil mediaReader")
		return
	}
	return clt.mediaUploadFromReader(MEDIA_TYPE_VIDEO, filename, mediaReader)
}

// 上传多媒体缩略图
//  NOTE: 参数 filename 不是文件路径, 是指定 multipart form 里面文件名称
func (clt *Client) MediaUploadThumbFromReader(filename string, mediaReader io.Reader) (info *MediaInfo, err error) {
	if filename == "" {
		err = errors.New("empty filename")
		return
	}
	if mediaReader == nil {
		err = errors.New("nil mediaReader")
		return
	}
	return clt.mediaUploadFromReader(MEDIA_TYPE_THUMB, filename, mediaReader)
}
