# 微信公众平台 订阅号, 服务号 golang SDK

```Go
// 基本的消息监听程序
package main

import (
	"fmt"
	"github.com/chanxuehong/wechatv2/mp"
	"github.com/chanxuehong/wechatv2/mp/menu"
	"github.com/chanxuehong/wechatv2/util"
	"net/http"
)

func MenuClickEventHandler(w http.ResponseWriter, r *mp.Request) {
	event := menu.GetClickEvent(r.Msg)
	fmt.Println(event.EventKey)
	return
}

func main() {
	aesKey, err := util.AESKeyDecode("encodedAESKey")
	if err != nil {
		panic(err)
	}

	messageServeMux := mp.NewMessageServeMux()
	messageServeMux.EventHandleFunc(menu.EVENT_TYPE_CLICK, MenuClickEventHandler)

	wechatServer := mp.NewDefaultWechatServer("id", "token", "appid", messageServeMux, aesKey)

	httpHandler := mp.NewHttpHandler(wechatServer, nil)

	http.Handle("/wechat", httpHandler)
	http.ListenAndServe(":80", nil)
}
```
