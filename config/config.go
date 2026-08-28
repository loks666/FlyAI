package config

// Conf 配置
var Conf conf

// Config
// @description: 配置
type conf struct {
	System   system                  `json:"system" yaml:"system"`     // 系统配置
	Task     task                    `json:"task" yaml:"task"`         // 定时任务配置
	MySQL    mysql                   `json:"mysql" yaml:"mysql"`       // MySQL 配置
	Redis    redisConfig             `json:"redis" yaml:"redis"`       // Redis 配置
	Wechat   wechat                  `json:"wechat" yaml:"wechat"`     // 微信助手
	Mq       mq                      `json:"mq" yaml:"mq"`             // MQ 配置
	Ai       ai                      `json:"ai" yaml:"ai"`             // AI配置
	Resource map[string]resourceItem `json:"resource" yaml:"resource"` // 资源配置
}
