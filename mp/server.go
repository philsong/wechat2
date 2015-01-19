// @description wechat 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/chanxuehong/wechatv2 for the canonical source repository
// @license     https://github.com/chanxuehong/wechatv2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package mp

import (
	"errors"
	"sync"
)

type WechatServer interface {
	Id() string    // 获取公众号的原始ID, 等于后台中的 公众号设置-->帐号详情-->原始ID
	Token() string // 获取公众号的Token, 和后台中的设置相等

	AppId() string
	CurrentAESKey() [32]byte // 获取当前有效的 AES 加密 Key
	LastAESKey() [32]byte    // 获取最后一个有效的 AES 加密 Key

	MessageHandler() MessageHandler // 获取 MessageHandler
}

var _ WechatServer = new(DefaultWechatServer)

type DefaultWechatServer struct {
	id    string
	token string
	appId string

	messageHandler MessageHandler

	rwmutex           sync.RWMutex
	currentAESKey     [32]byte // 当前的 AES Key
	lastAESKey        [32]byte // 最后一个 AES Key
	isLastAESKeyValid bool     // lastAESKey 是否有效, 如果是 lastAESKey 是 zero 则无效
}

// 初始化 DefaultAgent
//  如果不知道自己的 AppId 是多少, 可以先随便填入一个字符串,
//  这样正常情况下会出现 AppId mismatch 错误, 错误的 have 后面的就是正确的 AppId
func NewDefaultWechatServer(id, token, appId string, messageHandler MessageHandler,
	AESKey []byte) (srv *DefaultWechatServer) {

	if messageHandler == nil {
		panic("mp: nil messageHandler")
	}
	if len(AESKey) != 32 {
		panic("mp: the length of AESKey must equal to 32")
	}

	srv = &DefaultWechatServer{
		id:             id,
		token:          token,
		appId:          appId,
		messageHandler: messageHandler,
	}
	copy(srv.currentAESKey[:], AESKey)
	return
}

func (srv *DefaultWechatServer) Id() string {
	return srv.id
}
func (srv *DefaultWechatServer) Token() string {
	return srv.token
}
func (srv *DefaultWechatServer) AppId() string {
	return srv.appId
}
func (srv *DefaultWechatServer) MessageHandler() MessageHandler {
	return srv.messageHandler
}
func (srv *DefaultWechatServer) CurrentAESKey() (key [32]byte) {
	srv.rwmutex.RLock()
	key = srv.currentAESKey
	srv.rwmutex.RUnlock()
	return
}
func (srv *DefaultWechatServer) LastAESKey() (key [32]byte) {
	srv.rwmutex.RLock()
	if srv.isLastAESKeyValid {
		key = srv.lastAESKey
	} else {
		key = srv.currentAESKey
	}
	srv.rwmutex.RUnlock()
	return
}
func (srv *DefaultWechatServer) UpdateAESKey(AESKey []byte) (err error) {
	if len(AESKey) != 32 {
		return errors.New("the length of AESKey must equal to 32")
	}

	srv.rwmutex.Lock()
	srv.lastAESKey = srv.currentAESKey
	srv.isLastAESKeyValid = true
	copy(srv.currentAESKey[:], AESKey)
	srv.rwmutex.Unlock()
	return
}
