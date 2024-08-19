package utils

import (
	"encoding/base64"
	"fmt"
	"go-wechat/config"
)

// GetManagerUrl
// @description: 获取管理页面链接
// @param wxId
// @return newUrl
func GetManagerUrl(wxId string) (url string) {
	// 生成管理页面链接
	url = fmt.Sprintf("%s/manager.html?id=%s", config.Conf.System.Domain, wxId)
	// base64一下
	encodeString := base64.StdEncoding.EncodeToString([]byte(url))
	// 拼接新链接（这个是一个已备案的域名）
	url = "https://redirect.wjg95.cn/?s=" + encodeString
	return
}
