package utils

import (
	"fmt"
	neturl "net/url"
	"strings"

	"github.com/loks666/FlyAI/config"
)

// GetManagerUrl
// @description: 获取管理页面链接
// @param wxId
// @return newUrl
func GetManagerUrl(wxId string) string {
	return fmt.Sprintf("%s/manager.html?id=%s", strings.TrimRight(config.Conf.System.Domain, "/"), neturl.QueryEscape(wxId))
}
