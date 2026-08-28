package tcpserver

import (
	"log"
	"net"
	"os"
	"strconv"
)

// Start
// @description: 启动服务
func Start() {
	// 建立 tcp 服务
	port := os.Getenv("HUB_INGEST_PORT")
	if _, err := strconv.Atoi(port); err != nil {
		port = "19099"
	}
	listen, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", port))
	if err != nil {
		log.Printf("listen failed, err:%v", err)
		return
	}
	workerCount := positiveIntEnv("RECEIVE_WORKERS", 4)
	queueSize := positiveIntEnv("RECEIVE_QUEUE_SIZE", 1000)
	connections := make(chan net.Conn, queueSize)
	for i := 0; i < workerCount; i++ {
		go func() {
			for conn := range connections {
				func() {
					defer func() {
						if recovered := recover(); recovered != nil {
							log.Printf("消息接收工作器发生异常: %v", recovered)
						}
					}()
					process(conn)
				}()
			}
		}()
	}
	log.Printf("消息接收工作池已启动: workers=%d queue=%d", workerCount, queueSize)

	for {
		// 等待客户端建立连接
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("accept failed, err:%v", err)
			continue
		}
		// The bounded queue applies backpressure instead of creating an
		// unbounded goroutine for every live or historical message.
		connections <- conn
	}
}

func positiveIntEnv(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
