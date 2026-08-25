package tcpserver

import (
	"github.com/loks666/FlyAI/config"
	"log"
	"net"
)

// forward
// @description: 转发消息
func forward(msg []byte) {
	// 使用socket转发消息
	for _, s := range config.Conf.Wechat.Forward {
		conn, err := net.Dial("tcp", s)
		if err != nil {
			log.Printf("转发消息失败，错误信息: %v", err)
			continue
		}
		_, err = conn.Write(msg)
		if err != nil {
			log.Printf("转发消息失败，错误信息: %v", err)
		}
		_ = conn.Close()
	}
}
