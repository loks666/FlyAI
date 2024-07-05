package plugins

import (
	"context"
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/sashabaranov/go-openai"
	"go-wechat/client"
	"go-wechat/common/current"
	"go-wechat/config"
	"go-wechat/model/entity"
	"go-wechat/plugin"
	"go-wechat/service"
	"go-wechat/types"
	"go-wechat/utils"
	"log"
	"regexp"
	"strings"
	"time"
)

// AI
// @description: AI消息
// @param m
func AI(m *plugin.MessageContext) {
	if !config.Conf.Ai.Enable {
		return
	}

	// 取出所有启用了AI的好友或群组
	var friendInfo entity.Friend
	client.MySQL.Where("wxid = ?", m.FromUser).First(&friendInfo)
	if friendInfo.Wxid == "" {
		return
	}
	// 判断有没有启用AI
	if !friendInfo.EnableAi {
		return
	}

	// 预处理一下发送的消息，用正则去掉@机器人的内容
	re := regexp.MustCompile(`@([^ | ]+)`)
	matches := re.FindStringSubmatch(m.Content)
	if len(matches) > 0 {
		// 过滤掉第一个匹配到的
		m.Content = strings.Replace(m.Content, matches[0], "", 1)
	}

	// 处理预设角色，默认是配置文件里的，如果数据库配置不为空，则使用数据库配置
	prompt := config.Conf.Ai.Personality
	var dbPrompt entity.AiAssistant
	if friendInfo.Prompt != "" {
		// 取出配置的角色
		client.MySQL.First(&dbPrompt, "id = ?", friendInfo.Prompt)
		if dbPrompt.Id != "" {
			prompt = dbPrompt.Personality
		}
	}

	// 配置模型
	chatModel := openai.GPT3Dot5Turbo0613
	if friendInfo.AiModel != "" {
		chatModel = friendInfo.AiModel
	} else if dbPrompt.Model != "" {
		chatModel = dbPrompt.Model
	} else if config.Conf.Ai.Model != "" {
		chatModel = config.Conf.Ai.Model
	}

	// 组装消息体
	messages := make([]openai.ChatCompletionMessage, 0)
	if prompt != "" {
		// 填充人设
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		})
	}

	// 查询发信人前面几条文字信息，组装进来
	var oldMessages []entity.Message
	if m.GroupUser == "" {
		// 私聊
		oldMessages = getUserPrivateMessages(m.FromUser)
	} else {
		// 群聊
		oldMessages = getGroupUserMessages(m.MsgId, m.FromUser, m.GroupUser)
	}

	// 翻转数组
	slice.Reverse(oldMessages)
	// 循环填充消息
	for _, message := range oldMessages {
		// 剔除@机器人的内容
		msgStr := message.Content
		matches = re.FindStringSubmatch(msgStr)
		if len(matches) > 0 {
			// 过滤掉第一个匹配到的
			msgStr = strings.Replace(msgStr, matches[0], "", 1)
		}
		// 填充消息
		role := openai.ChatMessageRoleUser
		if message.FromUser == current.GetRobotInfo().WxId {
			// 如果收信人不是机器人，表示这条消息是 AI 发的
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msgStr,
		})
	}

	// 填充用户消息
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: m.Content,
	})

	// 默认使用AI回复
	conf := openai.DefaultConfig(config.Conf.Ai.ApiKey)
	if config.Conf.Ai.BaseUrl != "" {
		conf.BaseURL = fmt.Sprintf("%s/v1", config.Conf.Ai.BaseUrl)
	}
	ai := openai.NewClientWithConfig(conf)
	resp, err := ai.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    chatModel,
			Messages: messages,
		},
	)

	if err != nil {
		log.Printf("OpenAI聊天发起失败: %v", err.Error())
		utils.SendMessage(m.FromUser, m.GroupUser, "AI炸啦~", 0)
		return
	}

	// 返回消息为空
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		utils.SendMessage(m.FromUser, m.GroupUser, "AI似乎抽风了，没有告诉我你需要的回答~", 0)
		return
	}

	// 异步更新一下已使用的AI次数
	go service.UpdateUsedAiTokens(m.FromUser, resp.Usage.TotalTokens)

	// 保存一下AI 返回的消息，消息 Id 使用传入 Id 的负数
	var replyMessage entity.Message
	replyMessage.MsgId = -m.MsgId
	replyMessage.CreateTime = int(time.Now().Local().Unix())
	replyMessage.CreateAt = time.Now().Local()
	replyMessage.Content = resp.Choices[0].Message.Content
	replyMessage.FromUser = current.GetRobotInfo().WxId // 发信人是机器人
	replyMessage.GroupUser = m.GroupUser                // 群成员
	replyMessage.ToUser = m.FromUser                    // 收信人是发信人
	replyMessage.Type = types.MsgTypeText
	service.SaveMessage(replyMessage) // 保存消息

	// 发送消息
	replyMsg := resp.Choices[0].Message.Content
	if m.GroupUser != "" {
		replyMsg = "\n" + resp.Choices[0].Message.Content
	}
	utils.SendMessage(m.FromUser, m.GroupUser, replyMsg, 0)
}

// getGroupUserMessages
// @description: 获取群成员消息
// @return records
func getGroupUserMessages(msgId int64, groupId, groupUserId string) (records []entity.Message) {
	subQuery := client.MySQL.
		Where("from_user = ? AND group_user = ? AND display_full_content LIKE ?", groupId, groupUserId, "%在群聊中@了你").
		Or("to_user = ? AND group_user = ?", groupId, groupUserId)

	client.MySQL.Model(&entity.Message{}).
		Where("msg_id != ?", msgId).
		Where("type = ?", types.MsgTypeText).
		Where("create_at >= DATE_SUB(NOW(),INTERVAL 30 MINUTE)").
		Where(subQuery).
		Order("create_at desc").
		Limit(4).Find(&records)
	return
}

// getUserPrivateMessages
// @description: 获取用户私聊消息
// @return records
func getUserPrivateMessages(userId string) (records []entity.Message) {
	subQuery := client.MySQL.
		Where("from_user = ?", userId).Or("to_user = ?", userId)

	client.MySQL.Model(&entity.Message{}).
		Where("type = ?", types.MsgTypeText).
		Where("create_at >= DATE_SUB(NOW(),INTERVAL 30 MINUTE)").
		Where(subQuery).
		Order("create_at desc").
		Limit(4).Find(&records)
	return
}
