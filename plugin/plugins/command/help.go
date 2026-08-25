package command

import (
	"github.com/loks666/FlyAI/plugin"
	"github.com/loks666/FlyAI/utils"
)

// HelpCmd
// @description: 帮助指令
func HelpCmd(m *plugin.MessageContext) {
	str := `帮助菜单:
指令消息必须以'/'开头，比如: '/帮助'。
支持的指令:

#1. 雷神加速器
/ls option args
option: 指令选项，可选值: 
  绑定账户：'绑定'、'b'，参数: 账户名 密码 [-f]，-f表示强制绑定，非必传项
  详情: '详情'、'i'
  暂停: '暂停'、'p'
示例: 绑定:
/ls 绑定 123456 123456 或者 /ls b 123456 123456

#2. 肯德基疯狂星期四文案
/kfc、/肯德基

#3. AI助手
/ai option
option: 指令选项，可选值:
  启用: '启用'、'打开'、'enable'
  停用: '停用'、'禁用'、'关闭'、'disable'

`
	utils.SendMessage(m.FromUser, m.GroupUser, str, 0)

}
