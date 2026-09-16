package service

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/loks666/FlyAI/config"
)

// pointWechatHostAt 让 config.Conf.Wechat 指向测试用的地址，并在用例结束后还原。
func pointWechatHostAt(t *testing.T, host string) {
	t.Helper()
	original := config.Conf.Wechat.Host
	t.Cleanup(func() { config.Conf.Wechat.Host = original })
	config.Conf.Wechat.Host = host
}

// closedLoopbackAddress 返回一个刚刚释放、确定没有进程监听的本机地址，
// 用来模拟「桥接器没有启动」。
func closedLoopbackAddress(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("分配本机端口失败: %v", err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("释放本机端口失败: %v", err)
	}
	return address
}

func TestSyncGroupMembersReportsBridgeNotRunning(t *testing.T) {
	pointWechatHostAt(t, closedLoopbackAddress(t))

	err := SyncGroupMembers("00000000000@chatroom")
	if !errors.Is(err, ErrBridgeNotRunning) {
		t.Fatalf("SyncGroupMembers() = %v, 期望 ErrBridgeNotRunning", err)
	}
	if !strings.Contains(err.Error(), "微信桥接器未启动") {
		t.Fatalf("错误信息对用户不够明确: %q", err.Error())
	}
}

func TestBridgeReadyFalseWhenBridgeIsDown(t *testing.T) {
	pointWechatHostAt(t, closedLoopbackAddress(t))

	if BridgeReady() {
		t.Fatal("桥接器未启动时 BridgeReady() 应为 false")
	}
}

// 桥接器 /health 的返回契约：code=1 且 data.sendReady=true 才算真正可用。
// 少了 sendReady 这个字段，所有同步都会被守卫挡下，所以这里把契约钉住。
func TestBridgeReadyRequiresSendReadyFromHealth(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{name: "账号与窗口都已绑定", body: `{"code":1,"msg":"success","data":{"sendReady":true}}`, want: true},
		{name: "窗口尚未绑定", body: `{"code":1,"msg":"success","data":{"sendReady":false}}`, want: false},
		{name: "缺少 sendReady 字段", body: `{"code":1,"msg":"success","data":{}}`, want: false},
		{name: "桥接器返回失败码", body: `{"code":0,"msg":"starting","data":{"sendReady":true}}`, want: false},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/health" {
					writer.WriteHeader(http.StatusNotFound)
					return
				}
				writer.Header().Set("Content-Type", "application/json; charset=utf-8")
				_, _ = writer.Write([]byte(item.body))
			}))
			defer server.Close()
			pointWechatHostAt(t, strings.TrimPrefix(server.URL, "http://"))

			if got := BridgeReady(); got != item.want {
				t.Fatalf("BridgeReady() = %v, 期望 %v", got, item.want)
			}
		})
	}
}
