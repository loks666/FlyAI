package applog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContactSyncFileLoggerCreatesAndAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "contact-sync.log")
	logger, file, err := newFileLogger(path)
	if err != nil {
		t.Fatal(err)
	}
	logger.Println("同步好友列表完成")
	if err = file.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "同步好友列表完成") {
		t.Fatalf("unexpected log content: %s", data)
	}
}
