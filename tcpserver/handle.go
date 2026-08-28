package tcpserver

import (
	"bytes"
	"encoding/json"
	"github.com/loks666/FlyAI/config"
	"io"
	"log"
	"net"
)

// process
// @description: 处理连接
// @param conn
func process(conn net.Conn) {
	// 处理完关闭连接
	defer conn.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, conn); err != nil {
		log.Printf("[%s]返回数据失败，错误信息: %v", conn.RemoteAddr(), err)
	}
	// Each accepted connection already runs in its own goroutine. Parse before
	// acknowledging so historical bridge replay is naturally backpressured and
	// cannot flood MySQL with thousands of concurrent writes.
	payload := bytes.TrimSpace(buf.Bytes())
	response := "200 OK"
	if len(payload) > 0 && payload[0] == '[' {
		result, err := parseBatch(conn.RemoteAddr(), payload)
		encoded, _ := json.Marshal(result)
		if err != nil {
			response = "500 ERROR " + string(encoded)
		} else {
			response = "200 OK " + string(encoded)
		}
	} else {
		parse(conn.RemoteAddr(), payload)
	}

	// 转发到其他地方去
	if len(config.Conf.Wechat.Forward) > 0 {
		go forward(buf.Bytes())
	}
	// 将接受到的数据返回给客户端
	if _, err := conn.Write([]byte(response)); err != nil {
		log.Printf("[%s]返回数据失败，错误信息: %v", conn.RemoteAddr(), err)
	}
}
