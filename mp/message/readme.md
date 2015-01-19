# 普通消息和事件的接收和回复

### 简单示例
```Go
package main

import (
	"fmt"
	"github.com/chanxuehong/wechatv2/mp"
	"github.com/chanxuehong/wechatv2/mp/message"
	"github.com/chanxuehong/wechatv2/util"
	"log"
	"net/http"
)

// 处理普通文本消息, 原样返回
func TextMessageHandler(w http.ResponseWriter, r *mp.Request) {
	textReq := message.GetTextRequest(r.Msg)
	textResp := message.NewTextResponse(textReq.FromUserName, textReq.ToUserName,
		textReq.Content, textReq.CreateTime)

	if err := mp.WriteAESResponse(w, r, textResp); err != nil {
		log.Println(err)
	}
}

// 上报地理位置事件处理
func LocationEventHandler(w http.ResponseWriter, r *mp.Request) {
	event := message.GetLocationEvent(r.Msg)
	fmt.Println(event) // 处理事件
}

func main() {
	aesKey, err := util.AESKeyDecode("encodedAESKey")
	if err != nil {
		panic(err)
	}

	messageServeMux := mp.NewMessageServeMux()
	messageServeMux.MessageHandleFunc(message.MsgRequestTypeText, TextMessageHandler)
	messageServeMux.EventHandleFunc(message.EventTypeLocation, LocationEventHandler)

	wechatServer := mp.NewDefaultWechatServer("id", "token", "appid", messageServeMux, aesKey)

	wechatServerFrontend := mp.NewWechatServerFrontend(wechatServer, nil)

	http.Handle("/wechat", wechatServerFrontend)
	http.ListenAndServe(":80", nil)
}
```