package utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"go-wechat/config"
	"go-wechat/model/vo"
	"log"
	"strings"
)

// GetAiSummary
// @description: AI总结群聊记录
// @param groupName
// @param records
// @return replyMsg
// @return err
func GetAiSummary(groupName string, records []vo.TextMessageItem) (replyMsg string, err error) {
	// 组装对话记录为字符串
	var content []string
	for _, record := range records {
		content = append(content, fmt.Sprintf(`{"%s": "%s"}--end--`, record.Nickname, strings.ReplaceAll(record.Message, "\n", "。。")))
	}

	prompt := `# Role: 数据分析师和信息整理专家
        # Background: 用户需要从群聊记录中提取有价值的信息，包括技术话题讨论、工具分享、网站或资源链接。
        # Profile: 你是一位专业的数据分析师，擅长从大量文本中提取关键信息，并能够高效地整理和总结内容。
        # Skills: 数据分析、信息提取、内容总结、网络资源评估。
        # Goals: 设计一个流程，帮助用户识别群聊中的技术话题、共享的工具和资源，并高效地总结每个话题的要点。
        # Constrains: 该流程需要简洁明了，易于用户理解和操作，同时确保不遗漏任何重要信息。
        # OutputFormat: 结构化的文本总结，包含关键讨论点、结论等。
        # Workflow:
        ## 1. 快速浏览群聊记录
         - 通过快速浏览聊天记录，初步识别出主要话题。
        ## 2. 识别和分类主要话题
         - 识别群聊中的主要话题，分类讨论内容。
         - 每个话题独立提取，避免混淆。
		 - 话题识别尽可能准确，确保不遗漏重要信息。
        ## 3. 编写总结概括
         - 对每个话题进行总结，包含以下内容：
           - 关键讨论点：主要的讨论内容和观点。
           - 参与成员：参与讨论的成员的昵称。
           - 结论：得出的结论或共识，内容要尽可能详细。
        ## 4. 组织信息
         - 以逻辑清晰、易于理解的方式组织信息。
         - 使用结构化文本，使总结内容易于阅读。
        ## 5. 生成最终总结
         - 根据上述步骤，生成每个话题的详细总结。
         - 确保每个总结完整且有条理。
         - 返回数据格式为结构化文本，方便用户查阅。
        ## 6. 群聊记录消息格式
        格式: {"{nickname}": "{content}"}--end--
        # Examples:
        - 话题：技术讨论 - AI在医疗领域的应用
          - 关键讨论点：讨论了AI在医疗诊断中的应用前景和现有挑战。
          - 成员：张三、李四、王五。
          - 结论：AI技术有望提高医疗诊断的准确性和效率，但数据隐私和模型解释性仍需解决。

        - 话题：美食讨论 - 杭州到底有没有美食
          - 关键讨论点：讨论了杭州的美食有哪些、分布地区、评价如何。
          - 成员：张三、李四、王五。
          - 结论：杭州好像没有美食。
`

	msg := fmt.Sprintf("群名称: %s\n聊天记录如下:\n%s", groupName, strings.Join(content, "\n"))

	// AI总结
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: msg,
		},
	}

	// 默认使用AI回复
	conf := openai.DefaultConfig(config.Conf.Ai.ApiKey)
	if config.Conf.Ai.BaseUrl != "" {
		conf.BaseURL = fmt.Sprintf("%s/v1", config.Conf.Ai.BaseUrl)
	}
	ai := openai.NewClientWithConfig(conf)
	var resp openai.ChatCompletionResponse
	resp, err = ai.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    config.Conf.Ai.SummaryModel,
			Messages: messages,
		},
	)

	if err != nil {
		log.Printf("群聊记录总结失败: %v", err.Error())
		return
	}

	// 返回消息为空
	if resp.Choices[0].Message.Content == "" {
		err = errors.New("AI返回结果为空")
		return
	}
	replyMsg = resp.Choices[0].Message.Content
	return
}
