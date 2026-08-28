package tcpserver

import (
	"encoding/json"
	"github.com/loks666/FlyAI/common/current"
	"github.com/loks666/FlyAI/model/dto"
	"github.com/loks666/FlyAI/plugin/plugins"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/types"
	"log"
	"net"
	"strings"
)

// parse
// @description: 解析消息
// @param msg
func parse(remoteAddr net.Addr, msg []byte) {
	m, err := parseMessage(msg)
	if err != nil {
		logParseError(remoteAddr, msg, err)
		return
	}
	logMessage(remoteAddr, m, len(msg))

	// 插件不为空，开始执行
	if p := current.GetRobotMessageHandler(); p != nil {
		p(m)
	}
}

func parseMessage(msg []byte) (*dto.Message, error) {
	var m dto.Message
	if err := json.Unmarshal(msg, &m); err != nil {
		return nil, err
	}
	// 记录原始数据
	m.Raw = string(msg)

	// 提取出群成员信息
	// Sys类型的消息正文不包含微信 Id，所以不需要处理
	if m.IsGroup() && m.Type != types.MsgTypeSys {
		// 群消息，处理一下消息和发信人
		groupUser := strings.Split(m.Content, "\n")[0]
		groupUser = strings.ReplaceAll(groupUser, ":", "")
		// 如果两个id一致，说明是系统发的
		if m.FromUser != groupUser {
			m.GroupUser = groupUser
		}
		// 用户的操作单独提出来处理一下
		m.Content = strings.Join(strings.Split(m.Content, "\n")[1:], "\n")
	}
	service.PopulateMessageIdentity(&m)
	return &m, nil
}

type batchResult struct {
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

func parseBatch(remoteAddr net.Addr, payload []byte) (batchResult, error) {
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(payload, &rawMessages); err != nil {
		return batchResult{}, err
	}
	messages := make([]*dto.Message, 0, len(rawMessages))
	result := batchResult{}
	for _, raw := range rawMessages {
		message, err := parseMessage(raw)
		if err != nil || !message.IsBackfill {
			result.Failed++
			continue
		}
		messages = append(messages, message)
	}
	if err := plugins.SaveDTOBatchToDb(messages); err != nil {
		result.Failed += len(messages)
		logBatchResult(remoteAddr, result, err)
		return result, err
	}
	result.Success = len(messages)
	logBatchResult(remoteAddr, result, nil)
	return result, nil
}

func logParseError(remoteAddr net.Addr, msg []byte, parseErr error) {
	entry, _ := json.Marshal(struct {
		Raw    string `json:"raw"`
		Error  string `json:"error"`
		Bytes  int    `json:"bytes"`
		Remote string `json:"remote"`
		Event  string `json:"event"`
	}{string(msg), parseErr.Error(), len(msg), remoteAddr.String(), "wechat_message_parse_failed"})
	log.Print(string(entry))
}

func logBatchResult(remoteAddr net.Addr, result batchResult, batchErr error) {
	entry := struct {
		Success int    `json:"success"`
		Failed  int    `json:"failed"`
		Error   string `json:"error,omitempty"`
		Remote  string `json:"remote"`
		Event   string `json:"event"`
	}{Success: result.Success, Failed: result.Failed, Remote: remoteAddr.String(), Event: "wechat_backfill_batch"}
	if batchErr != nil {
		entry.Error = batchErr.Error()
	}
	encoded, _ := json.Marshal(entry)
	log.Print(string(encoded))
}

func logMessage(remoteAddr net.Addr, m *dto.Message, size int) {
	// Historical replay is logged once per batch by parseBatch.
	if m.IsBackfill {
		return
	}
	entry, err := json.Marshal(struct {
		Content             string            `json:"content"`
		SenderNickname      string            `json:"senderNickname"`
		SenderGroupNickname string            `json:"senderGroupNickname"`
		SenderWechatAccount string            `json:"senderWechatAccount"`
		GroupNickname       string            `json:"groupNickname"`
		GroupUser           string            `json:"groupUser,omitempty"`
		FromUser            string            `json:"fromUser"`
		MsgId               int64             `json:"msgId"`
		Type                types.MessageType `json:"type"`
		Backfill            bool              `json:"backfill"`
		Bytes               int               `json:"bytes"`
		Remote              string            `json:"remote"`
		Event               string            `json:"event"`
	}{
		Content:             m.Content,
		SenderNickname:      m.SenderNickname,
		SenderGroupNickname: m.SenderGroupNickname,
		SenderWechatAccount: m.SenderWechatAccount,
		GroupNickname:       m.GroupNickname,
		GroupUser:           m.GroupUser,
		FromUser:            m.FromUser,
		MsgId:               m.MsgId,
		Type:                m.Type,
		Backfill:            m.IsBackfill,
		Bytes:               size,
		Remote:              remoteAddr.String(),
		Event:               "wechat_message",
	})
	if err != nil {
		log.Printf("消息日志序列化失败: %v", err)
	} else {
		log.Print(string(entry))
	}
}
