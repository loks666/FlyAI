package friends

import "go-wechat/service"

// ClearAiUsedToday
// @description: 清空AI日使用次数
func ClearAiUsedToday() {
	service.ClearAiUsedToday()
}
