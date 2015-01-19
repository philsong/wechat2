### 创建菜单的示例

```Go
package main

import (
	"fmt"

	"github.com/chanxuehong/wechatv2/mp"
	"github.com/chanxuehong/wechatv2/mp/menu"
)

var TokenService = mp.NewDefaultTokenService("appid", "appsecret", nil)

func main() {
	var subButtons = make([]menu.Button, 2)
	subButtons[0].SetAsViewButton("搜索", "http://www.soso.com/")
	subButtons[1].SetAsClickButton("赞一下我们", "V1001_GOOD")

	var mn menu.Menu
	mn.Buttons = make([]menu.Button, 3)
	mn.Buttons[0].SetAsClickButton("今日歌曲", "V1001_TODAY_MUSIC")
	mn.Buttons[1].SetAsViewButton("视频", "http://v.qq.com/")
	mn.Buttons[2].SetAsSubMenuButton("菜单", subButtons)

	clt := menu.NewClient(TokenService, nil)
	if err := clt.MenuCreate(mn); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("ok")
}
```
