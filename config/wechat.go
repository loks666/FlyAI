package config

import "strings"

// wxHelper
// @description: 微信助手
type wechat struct {
	Host    string   `json:"host" yaml:"host"`       // 本地桥接 HTTP 地址
	Forward []string `json:"forward" yaml:"forward"` // 可选的原始事件转发地址
}

// Check
// @description: 检查配置是否可用
// @receiver w
// @return bool
func (w wechat) Check() bool {
	if w.Host == "" {
		return false
	}
	return true
}

func (w wechat) GetURL(uri string) string {
	host := w.Host
	if !strings.HasPrefix(w.Host, "http://") {
		host = "http://" + w.Host
	}
	return host + uri
}
