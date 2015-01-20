// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechat2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package media

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// 上传多媒体图片
func (clt *Client) UploadImage(_filepath string) (info *MediaInfo, err error) {
	return clt.uploadMedia(MediaTypeImage, _filepath)
}

// 上传多媒体语音
func (clt *Client) UploadVoice(_filepath string) (info *MediaInfo, err error) {
	return clt.uploadMedia(MediaTypeVoice, _filepath)
}

// 上传多媒体视频
func (clt *Client) UploadVideo(_filepath string) (info *MediaInfo, err error) {
	return clt.uploadMedia(MediaTypeVideo, _filepath)
}

// 上传多媒体缩略图
func (clt *Client) UploadThumb(_filepath string) (info *MediaInfo, err error) {
	return clt.uploadMedia(MediaTypeThumb, _filepath)
}

// 上传多媒体
func (clt *Client) uploadMedia(mediaType, _filepath string) (info *MediaInfo, err error) {
	file, err := os.Open(_filepath)
	if err != nil {
		return
	}
	defer file.Close()

	return clt.uploadMediaFromReader(mediaType, filepath.Base(_filepath), file)
}

// 上传多媒体图片
//  NOTE: 参数 filename 不是文件路径, 是指定 multipart form 里面文件名称
func (clt *Client) UploadImageFromReader(filename string, mediaReader io.Reader) (info *MediaInfo, err error) {
	if filename == "" {
		err = errors.New("empty filename")
		return
	}
	if mediaReader == nil {
		err = errors.New("nil mediaReader")
		return
	}
	return clt.uploadMediaFromReader(MediaTypeImage, filename, mediaReader)
}

// 上传多媒体语音
//  NOTE: 参数 filename 不是文件路径, 是指定 multipart form 里面文件名称
func (clt *Client) UploadVoiceFromReader(filename string, mediaReader io.Reader) (info *MediaInfo, err error) {
	if filename == "" {
		err = errors.New("empty filename")
		return
	}
	if mediaReader == nil {
		err = errors.New("nil mediaReader")
		return
	}
	return clt.uploadMediaFromReader(MediaTypeVoice, filename, mediaReader)
}

// 上传多媒体视频
//  NOTE: 参数 filename 不是文件路径, 是指定 multipart form 里面文件名称
func (clt *Client) UploadVideoFromReader(filename string, mediaReader io.Reader) (info *MediaInfo, err error) {
	if filename == "" {
		err = errors.New("empty filename")
		return
	}
	if mediaReader == nil {
		err = errors.New("nil mediaReader")
		return
	}
	return clt.uploadMediaFromReader(MediaTypeVideo, filename, mediaReader)
}

// 上传多媒体缩略图
//  NOTE: 参数 filename 不是文件路径, 是指定 multipart form 里面文件名称
func (clt *Client) UploadThumbFromReader(filename string, mediaReader io.Reader) (info *MediaInfo, err error) {
	if filename == "" {
		err = errors.New("empty filename")
		return
	}
	if mediaReader == nil {
		err = errors.New("nil mediaReader")
		return
	}
	return clt.uploadMediaFromReader(MediaTypeThumb, filename, mediaReader)
}
