package config

// ai
// @description: AI配置
type ai struct {
	Enable       bool      `json:"enable" yaml:"enable"`             // 是否启用AI
	Model        string    `json:"dto" yaml:"dto"`                   // 模型
	SummaryModel string    `json:"summaryModel" yaml:"summaryModel"` // 总结模型
	ApiKey       string    `json:"apiKey" yaml:"apiKey"`             // API Key
	BaseUrl      string    `json:"baseUrl" yaml:"baseUrl"`           // API地址
	Personality  string    `json:"personality" yaml:"personality"`   // 人设
	Models       []aiModel `json:"models" yaml:"models"`             // 模型列表
}

// aiModel
// @description: AI模型
type aiModel struct {
	Name  string `json:"name" yaml:"name"` // 模型名称
	Model string `json:"dto" yaml:"dto"`   // 模型代码
}
