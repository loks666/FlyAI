package applog

import (
	"io"
	"log"
	"path/filepath"
	"sync"
)

var (
	httpMonitorLoggerOnce sync.Once
	httpMonitorLogger     *log.Logger
)

func getHTTPMonitorLogger() *log.Logger {
	httpMonitorLoggerOnce.Do(func() {
		path := filepath.Join("logs", "http-monitor.log")
		logger, _, err := newFileLogger(path)
		if err != nil {
			log.Printf("打开监控接口访问日志失败: %v", err)
			httpMonitorLogger = log.New(io.Discard, "", log.LstdFlags)
			return
		}
		httpMonitorLogger = logger
	})
	return httpMonitorLogger
}

func HTTPMonitorPrintf(format string, values ...any) {
	getHTTPMonitorLogger().Printf(format, values...)
}
