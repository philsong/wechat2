// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechatv2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechatv2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/chanxuehong/wechatv2/mp"
)

const (
	multipartBoundary    = "--------wvm6LNx=y4rEq?BUD(k_:0Pj2V.M'J)t957K-Sh/Q1ZA+ceWFunTRdfGaXgY"
	multipartContentType = "multipart/form-data; boundary=" + multipartBoundary

	// ----------wvm6LNx=y4rEq?BUD(k_:0Pj2V.M'J)t957K-Sh/Q1ZA+ceWFunTRdfGaXgY
	// Content-Disposition: form-data; name="file"; filename="filename"
	// Content-Type: application/octet-stream
	//
	// mediaReader
	// ----------wvm6LNx=y4rEq?BUD(k_:0Pj2V.M'J)t957K-Sh/Q1ZA+ceWFunTRdfGaXgY--
	//
	multipartFormDataFront = "--" + multipartBoundary +
		"\r\nContent-Disposition: form-data; name=\"file\"; filename=\""
	multipartFormDataMiddle = "\"\r\nContent-Type: application/octet-stream\r\n\r\n"
	multipartFormDataEnd    = "\r\n--" + multipartBoundary + "--\r\n"

	multipartConstPartLen = len(multipartFormDataFront) +
		len(multipartFormDataMiddle) + len(multipartFormDataEnd)
)

// copy from mime/multipart/writer.go
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

// copy from mime/multipart/writer.go
func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func mediaUploadURL(mediatype, accesstoken string) string {
	return "http://file.api.weixin.qq.com/cgi-bin/media/upload?type=" + mediatype +
		"&access_token=" + accesstoken
}

// 上传多媒体
func (clt *Client) mediaUploadFromReader(mediaType, filename string, reader io.Reader) (info *MediaInfo, err error) {
	filename = escapeQuotes(filename)

	switch v := reader.(type) {
	case *os.File:
		return clt.mediaUploadFromOSFile(mediaType, filename, v)
	case *bytes.Buffer:
		return clt.mediaUploadFromBytesBuffer(mediaType, filename, v)
	case *bytes.Reader:
		return clt.mediaUploadFromBytesReader(mediaType, filename, v)
	case *strings.Reader:
		return clt.mediaUploadFromStringsReader(mediaType, filename, v)
	default:
		return clt.mediaUploadFromIOReader(mediaType, filename, v)
	}
}

func (clt *Client) mediaUploadFromOSFile(mediaType, filename string, file *os.File) (info *MediaInfo, err error) {
	fi, err := file.Stat()
	if err != nil {
		return
	}

	// 非常规文件, FileInfo.Size() 不一定准确
	if !fi.Mode().IsRegular() {
		return clt.mediaUploadFromIOReader(mediaType, filename, file)
	}

	originalOffset, err := file.Seek(0, 1)
	if err != nil {
		return
	}
	ContentLength := int64(multipartConstPartLen+len(filename)) +
		fi.Size() - originalOffset

	token, err := clt.Token()
	if err != nil {
		return
	}

	hasRetried := false
RETRY:
	finalURL := mediaUploadURL(mediaType, token)

	if hasRetried {
		if _, err = file.Seek(originalOffset, 0); err != nil {
			return
		}
	}
	mr := io.MultiReader(
		strings.NewReader(multipartFormDataFront),
		strings.NewReader(filename),
		strings.NewReader(multipartFormDataMiddle),
		file,
		strings.NewReader(multipartFormDataEnd),
	)

	httpReq, err := http.NewRequest("POST", finalURL, mr)
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", multipartContentType)
	httpReq.ContentLength = ContentLength

	httpResp, err := clt.HttpClient.Do(httpReq)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", httpResp.Status)
		return
	}

	switch mediaType {
	case MEDIA_TYPE_THUMB: // 返回的是 thumb_media_id 而不是 media_id
		var result struct {
			mp.Error
			MediaType string `json:"type"`
			MediaId   string `json:"thumb_media_id"`
			CreatedAt int64  `json:"created_at"`
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return
		}

		switch result.ErrCode {
		case mp.ErrCodeOK:
			info = &MediaInfo{
				MediaType: result.MediaType,
				MediaId:   result.MediaId,
				CreatedAt: result.CreatedAt,
			}
			return
		case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
			if !hasRetried {
				hasRetried = true

				if token, err = clt.GetNewToken(); err != nil {
					return
				}
				goto RETRY
			}
			fallthrough
		default:
			err = &result.Error
			return
		}

	default:
		var result struct {
			mp.Error
			MediaInfo
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return
		}

		switch result.ErrCode {
		case mp.ErrCodeOK:
			info = &MediaInfo{
				MediaType: result.MediaType,
				MediaId:   result.MediaId,
				CreatedAt: result.CreatedAt,
			}
			return
		case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
			if !hasRetried {
				hasRetried = true

				if token, err = clt.GetNewToken(); err != nil {
					return
				}
				goto RETRY
			}
			fallthrough
		default:
			err = &result.Error
			return
		}
	}
}

func (clt *Client) mediaUploadFromBytesBuffer(mediaType, filename string, buffer *bytes.Buffer) (info *MediaInfo, err error) {
	fileBytes := buffer.Bytes()
	ContentLength := int64(multipartConstPartLen + len(filename) + len(fileBytes))

	token, err := clt.Token()
	if err != nil {
		return
	}

	hasRetried := false
RETRY:
	finalURL := mediaUploadURL(mediaType, token)

	mr := io.MultiReader(
		strings.NewReader(multipartFormDataFront),
		strings.NewReader(filename),
		strings.NewReader(multipartFormDataMiddle),
		bytes.NewReader(fileBytes),
		strings.NewReader(multipartFormDataEnd),
	)

	httpReq, err := http.NewRequest("POST", finalURL, mr)
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", multipartContentType)
	httpReq.ContentLength = ContentLength

	httpResp, err := clt.HttpClient.Do(httpReq)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", httpResp.Status)
		return
	}

	switch mediaType {
	case MEDIA_TYPE_THUMB: // 返回的是 thumb_media_id 而不是 media_id
		var result struct {
			mp.Error
			MediaType string `json:"type"`
			MediaId   string `json:"thumb_media_id"`
			CreatedAt int64  `json:"created_at"`
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return
		}

		switch result.ErrCode {
		case mp.ErrCodeOK:
			info = &MediaInfo{
				MediaType: result.MediaType,
				MediaId:   result.MediaId,
				CreatedAt: result.CreatedAt,
			}
			return
		case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
			if !hasRetried {
				hasRetried = true

				if token, err = clt.GetNewToken(); err != nil {
					return
				}
				goto RETRY
			}
			fallthrough
		default:
			err = &result.Error
			return
		}

	default:
		var result struct {
			mp.Error
			MediaInfo
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return
		}

		switch result.ErrCode {
		case mp.ErrCodeOK:
			info = &MediaInfo{
				MediaType: result.MediaType,
				MediaId:   result.MediaId,
				CreatedAt: result.CreatedAt,
			}
			return
		case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
			if !hasRetried {
				hasRetried = true

				if token, err = clt.GetNewToken(); err != nil {
					return
				}
				goto RETRY
			}
			fallthrough
		default:
			err = &result.Error
			return
		}
	}
}

func (clt *Client) mediaUploadFromBytesReader(mediaType, filename string, reader *bytes.Reader) (info *MediaInfo, err error) {
	originalOffset, err := reader.Seek(0, 1)
	if err != nil {
		return
	}
	ContentLength := int64(multipartConstPartLen + len(filename) + reader.Len())

	token, err := clt.Token()
	if err != nil {
		return
	}

	hasRetried := false
RETRY:
	finalURL := mediaUploadURL(mediaType, token)

	if hasRetried {
		if _, err = reader.Seek(originalOffset, 0); err != nil {
			return
		}
	}
	mr := io.MultiReader(
		strings.NewReader(multipartFormDataFront),
		strings.NewReader(filename),
		strings.NewReader(multipartFormDataMiddle),
		reader,
		strings.NewReader(multipartFormDataEnd),
	)

	httpReq, err := http.NewRequest("POST", finalURL, mr)
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", multipartContentType)
	httpReq.ContentLength = ContentLength

	httpResp, err := clt.HttpClient.Do(httpReq)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", httpResp.Status)
		return
	}

	switch mediaType {
	case MEDIA_TYPE_THUMB: // 返回的是 thumb_media_id 而不是 media_id
		var result struct {
			mp.Error
			MediaType string `json:"type"`
			MediaId   string `json:"thumb_media_id"`
			CreatedAt int64  `json:"created_at"`
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return
		}

		switch result.ErrCode {
		case mp.ErrCodeOK:
			info = &MediaInfo{
				MediaType: result.MediaType,
				MediaId:   result.MediaId,
				CreatedAt: result.CreatedAt,
			}
			return
		case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
			if !hasRetried {
				hasRetried = true

				if token, err = clt.GetNewToken(); err != nil {
					return
				}
				goto RETRY
			}
			fallthrough
		default:
			err = &result.Error
			return
		}

	default:
		var result struct {
			mp.Error
			MediaInfo
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return
		}

		switch result.ErrCode {
		case mp.ErrCodeOK:
			info = &MediaInfo{
				MediaType: result.MediaType,
				MediaId:   result.MediaId,
				CreatedAt: result.CreatedAt,
			}
			return
		case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
			if !hasRetried {
				hasRetried = true

				if token, err = clt.GetNewToken(); err != nil {
					return
				}
				goto RETRY
			}
			fallthrough
		default:
			err = &result.Error
			return
		}
	}
}

func (clt *Client) mediaUploadFromStringsReader(mediaType, filename string, reader *strings.Reader) (info *MediaInfo, err error) {
	originalOffset, err := reader.Seek(0, 1)
	if err != nil {
		return
	}
	ContentLength := int64(multipartConstPartLen + len(filename) + reader.Len())

	token, err := clt.Token()
	if err != nil {
		return
	}

	hasRetried := false
RETRY:
	finalURL := mediaUploadURL(mediaType, token)

	if hasRetried {
		if _, err = reader.Seek(originalOffset, 0); err != nil {
			return
		}
	}
	mr := io.MultiReader(
		strings.NewReader(multipartFormDataFront),
		strings.NewReader(filename),
		strings.NewReader(multipartFormDataMiddle),
		reader,
		strings.NewReader(multipartFormDataEnd),
	)

	httpReq, err := http.NewRequest("POST", finalURL, mr)
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", multipartContentType)
	httpReq.ContentLength = ContentLength

	httpResp, err := clt.HttpClient.Do(httpReq)
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", httpResp.Status)
		return
	}

	switch mediaType {
	case MEDIA_TYPE_THUMB: // 返回的是 thumb_media_id 而不是 media_id
		var result struct {
			mp.Error
			MediaType string `json:"type"`
			MediaId   string `json:"thumb_media_id"`
			CreatedAt int64  `json:"created_at"`
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return
		}

		switch result.ErrCode {
		case mp.ErrCodeOK:
			info = &MediaInfo{
				MediaType: result.MediaType,
				MediaId:   result.MediaId,
				CreatedAt: result.CreatedAt,
			}
			return
		case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
			if !hasRetried {
				hasRetried = true

				if token, err = clt.GetNewToken(); err != nil {
					return
				}
				goto RETRY
			}
			fallthrough
		default:
			err = &result.Error
			return
		}

	default:
		var result struct {
			mp.Error
			MediaInfo
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return
		}

		switch result.ErrCode {
		case mp.ErrCodeOK:
			info = &MediaInfo{
				MediaType: result.MediaType,
				MediaId:   result.MediaId,
				CreatedAt: result.CreatedAt,
			}
			return
		case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
			if !hasRetried {
				hasRetried = true

				if token, err = clt.GetNewToken(); err != nil {
					return
				}
				goto RETRY
			}
			fallthrough
		default:
			err = &result.Error
			return
		}
	}
}

func (clt *Client) mediaUploadFromIOReader(mediaType, filename string, reader io.Reader) (info *MediaInfo, err error) {
	bodyBuf := mp.MediaBufferPool.Get().(*bytes.Buffer) // io.ReadWriter
	bodyBuf.Reset()                                     // important
	defer mp.MediaBufferPool.Put(bodyBuf)               // important

	bodyBuf.WriteString(multipartFormDataFront)
	bodyBuf.WriteString(filename)
	bodyBuf.WriteString(multipartFormDataMiddle)
	if _, err = io.Copy(bodyBuf, reader); err != nil {
		return
	}
	bodyBuf.WriteString(multipartFormDataEnd)

	bodyBytes := bodyBuf.Bytes()

	token, err := clt.Token()
	if err != nil {
		return
	}

	hasRetried := false
RETRY:
	finalURL := mediaUploadURL(mediaType, token)

	httpResp, err := clt.HttpClient.Post(finalURL, multipartContentType, bytes.NewReader(bodyBytes))
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http.Status: %s", httpResp.Status)
		return
	}

	switch mediaType {
	case MEDIA_TYPE_THUMB: // 返回的是 thumb_media_id 而不是 media_id
		var result struct {
			mp.Error
			MediaType string `json:"type"`
			MediaId   string `json:"thumb_media_id"`
			CreatedAt int64  `json:"created_at"`
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return
		}

		switch result.ErrCode {
		case mp.ErrCodeOK:
			info = &MediaInfo{
				MediaType: result.MediaType,
				MediaId:   result.MediaId,
				CreatedAt: result.CreatedAt,
			}
			return
		case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
			if !hasRetried {
				hasRetried = true

				if token, err = clt.GetNewToken(); err != nil {
					return
				}
				goto RETRY
			}
			fallthrough
		default:
			err = &result.Error
			return
		}

	default:
		var result struct {
			mp.Error
			MediaInfo
		}
		if err = json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
			return
		}

		switch result.ErrCode {
		case mp.ErrCodeOK:
			info = &MediaInfo{
				MediaType: result.MediaType,
				MediaId:   result.MediaId,
				CreatedAt: result.CreatedAt,
			}
			return
		case mp.ErrCodeInvalidCredential, mp.ErrCodeTimeout:
			if !hasRetried {
				hasRetried = true

				if token, err = clt.GetNewToken(); err != nil {
					return
				}
				goto RETRY
			}
			fallthrough
		default:
			err = &result.Error
			return
		}
	}
}
