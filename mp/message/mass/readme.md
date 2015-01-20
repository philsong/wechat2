# 群发消息接口

### 根据分组进行群发消息示例
```Go
package main

import (
	"fmt"

	"github.com/chanxuehong/wechat2/mp"
	"github.com/chanxuehong/wechat2/mp/message/mass/masstogroup"
)

var TokenService = mp.NewDefaultTokenService("appid", "appsecret", nil)

func main() {
	text := masstogroup.NewText(1 /* groupid */, "content")

	clt := masstogroup.NewClient(TokenService, nil)

	msgId, err := clt.SendText(text)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("msgId:", msgId)
}
```
