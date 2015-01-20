### 上传图片示例

```Go
package main

import (
	"fmt"

	"github.com/chanxuehong/wechat2/mp"
	"github.com/chanxuehong/wechat2/mp/media"
)

var TokenService = mp.NewDefaultTokenService("appid", "appsecret", nil)

func main() {
	clt := media.NewClient(TokenService, nil)
	info, err := clt.UploadImage("d:\\img.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(info)
}
```
