package applog

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	contactSyncLoggerOnce sync.Once
	contactSyncLogger     *log.Logger
)

func getContactSyncLogger() *log.Logger {
	contactSyncLoggerOnce.Do(func() {
		path := filepath.Join("logs", "contact-sync.log")
		logger, _, err := newFileLogger(path)
		if err != nil {
			log.Printf("打开通讯录同步日志失败: %v", err)
			contactSyncLogger = log.New(io.Discard, "", log.LstdFlags)
			return
		}
		contactSyncLogger = logger
	})
	return contactSyncLogger
}

func newFileLogger(path string) (*log.Logger, *os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, err
	}
	return log.New(file, "", log.LstdFlags), file, nil
}

func ContactSyncPrintf(format string, values ...any) {
	getContactSyncLogger().Printf(format, values...)
}

func ContactSyncPrintln(values ...any) {
	getContactSyncLogger().Println(values...)
}
