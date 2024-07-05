package entity

// PluginData
// @description: 插件数据
type PluginData struct {
	UserId     string `json:"userId"`     // 用户Id
	PluginCode string `json:"pluginCode"` // 插件编码
	Data       string `json:"data"`       // 数据
}

func (PluginData) TableName() string {
	return "t_plugin_data"
}
