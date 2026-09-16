package tcpserver

import (
	"bytes"
	"github.com/loks666/FlyAI/common/current"
	"github.com/loks666/FlyAI/model/dto"
	"log"
	"net"
	"strings"
	"testing"
)

func TestMessageLogKeepsContentAndSenderBeforeSystemFields(t *testing.T) {
	var output bytes.Buffer
	previousWriter := log.Writer()
	previousFlags := log.Flags()
	previousHandler := current.GetRobotMessageHandler()
	log.SetOutput(&output)
	log.SetFlags(0)
	current.SetRobotMessageHandler(nil)
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
		current.SetRobotMessageHandler(previousHandler)
	})

	parse(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}, []byte(`{
		"msgId": 9,
		"content": "wxid_sender:\nhello\nworld",
		"fromUser": "group@chatroom",
		"type": 1
	}`))
	line := strings.TrimSpace(output.String())
	fields := []string{
		`"content":"hello world"`,
		`"senderNickname":"wxid_sender"`,
		`"senderGroupNickname":""`,
		`"senderWechatAccount":""`,
		`"groupNickname":"group@chatroom"`,
		`"groupUser":"wxid_sender"`,
		`"fromUser":"group@chatroom"`,
		`"msgId":9`,
		`"remote":"127.0.0.1:12345"`,
		`"event":"wechat_message"`,
	}
	lastIndex := -1
	for _, field := range fields {
		index := strings.Index(line, field)
		if index < 0 {
			t.Fatalf("message log missing %s: %s", field, line)
		}
		if index <= lastIndex {
			t.Fatalf("message log field %s is out of order: %s", field, line)
		}
		lastIndex = index
	}
	if strings.Contains(line, `\n`) {
		t.Fatalf("message log should not contain escaped newlines: %s", line)
	}
}

func TestSingleLineLogContentRemovesRedundantSenderEnvelope(t *testing.T) {
	message := &dto.Message{
		Content:   "wxid_sender:\nfirst line\r\nsecond line",
		GroupUser: "wxid_sender",
	}
	if got := singleLineLogContent(message); got != "first line second line" {
		t.Fatalf("singleLineLogContent() = %q", got)
	}
}

func TestHistoricalMessageDoesNotEmitPerMessageLog(t *testing.T) {
	var output bytes.Buffer
	previousWriter := log.Writer()
	previousFlags := log.Flags()
	log.SetOutput(&output)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
		log.SetFlags(previousFlags)
	})

	logMessage(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}, &dto.Message{
		MsgId:      10,
		Content:    "old message",
		IsBackfill: true,
	}, 100)
	if output.Len() != 0 {
		t.Fatalf("historical message emitted a per-message log: %s", output.String())
	}
}
