### 发送客服消息示例
```Go
package main

import (
	"fmt"

	"github.com/chanxuehong/wechat2/mp"
	"github.com/chanxuehong/wechat2/mp/message/custom"
)

var TokenService = mp.NewDefaultTokenService("appid", "appsecret", nil)

func main() {
	text := custom.NewText("touser", "content", "")

	clt := custom.NewClient(TokenService, nil)
	if err := clt.SendText(text); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ok")
}
```
