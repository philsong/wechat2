### 获取 jsapi_ticket 示例
```Go
package main

import (
	"fmt"

	"github.com/chanxuehong/wechat2/mp"
	"github.com/chanxuehong/wechat2/mp/jssdk"
)

var TokenServer = mp.NewDefaultTokenServer("appid", "appsecret", nil)
var TicketServer = jssdk.NewDefaultTicketServer(TokenServer, nil)

func main() {
	fmt.Println(TicketServer.Ticket())
}
```