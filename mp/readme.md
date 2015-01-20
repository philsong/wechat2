# 微信公众平台 订阅号, 服务号 golang SDK

### 一个 URL 监听一个公众号的消息
```Go
package main

import (
	"fmt"
	"github.com/chanxuehong/wechat2/mp"
	"github.com/chanxuehong/wechat2/mp/menu"
	"github.com/chanxuehong/wechat2/util"
	"net/http"
)

func MenuClickEventHandler(w http.ResponseWriter, r *mp.Request) {
	event := menu.GetClickEvent(r.MixedMsg)
	fmt.Println(event.EventKey)
	return
}

func main() {
	aesKey, err := util.AESKeyDecode("encodedAESKey")
	if err != nil {
		panic(err)
	}

	messageServeMux := mp.NewMessageServeMux()
	messageServeMux.EventHandleFunc(menu.EventTypeClick, MenuClickEventHandler)

	wechatServer := mp.NewDefaultWechatServer("id", "token", "appid", messageServeMux, aesKey)

	wechatServerFrontend := mp.NewWechatServerFrontend(wechatServer, nil)

	http.Handle("/wechat", wechatServerFrontend)
	http.ListenAndServe(":80", nil)
}
```

### 一个 URL 监听多个公众号的消息
```Go
package main

import (
	"fmt"
	"github.com/chanxuehong/wechat2/mp"
	"github.com/chanxuehong/wechat2/mp/menu"
	"github.com/chanxuehong/wechat2/util"
	"net/http"
)

func MenuClickEventHandler(w http.ResponseWriter, r *mp.Request) {
	event := menu.GetClickEvent(r.MixedMsg)
	fmt.Println(event.EventKey)
	return
}

func main() {
	aesKey1, err := util.AESKeyDecode("encodedAESKey1")
	if err != nil {
		panic(err)
	}

	messageServeMux1 := mp.NewMessageServeMux()
	messageServeMux1.EventHandleFunc(menu.EventTypeClick, MenuClickEventHandler)

	wechatServer1 := mp.NewDefaultWechatServer("id1", "token1", "appid1", messageServeMux1, aesKey1)

	aesKey2, err := util.AESKeyDecode("encodedAESKey2")
	if err != nil {
		panic(err)
	}

	messageServeMux2 := mp.NewMessageServeMux()
	messageServeMux2.EventHandleFunc(menu.EventTypeClick, MenuClickEventHandler)

	wechatServer2 := mp.NewDefaultWechatServer("id2", "token2", "appid2", messageServeMux2, aesKey2)

	var multiWechatServerFrontend mp.MultiWechatServerFrontend
	multiWechatServerFrontend.SetWechatServer("wechat1", wechatServer1)
	multiWechatServerFrontend.SetWechatServer("wechat2", wechatServer2)

	http.Handle("/wechat", &multiWechatServerFrontend)
	http.ListenAndServe(":80", nil)
}
```