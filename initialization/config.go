package initialization

import (
	"github.com/fsnotify/fsnotify"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/config"
	"github.com/loks666/FlyAI/model/entity"
	"github.com/loks666/FlyAI/service"
	"github.com/spf13/viper"
	"log"
)

func ensureSchema() {
	migrator := client.MySQL.Migrator()
	// Older deployments created these tables outside GORM. A fresh Docker
	// database has no schema yet, so create only missing base tables here while
	// leaving existing installations and their data untouched.
	baseModels := []interface{}{
		&entity.Friend{},
		&entity.GroupUser{},
		&entity.Message{},
		&entity.AiAssistant{},
		&entity.HotTop{},
		&entity.PluginData{},
	}
	for _, model := range baseModels {
		if migrator.HasTable(model) {
			continue
		}
		if err := migrator.CreateTable(model); err != nil {
			log.Panicf("创建基础数据表失败: %v", err)
		}
	}
	if !migrator.HasColumn(&entity.Friend{}, "EnableMonitor") {
		if err := migrator.AddColumn(&entity.Friend{}, "EnableMonitor"); err != nil {
			log.Panicf("新增群监控开关字段失败: %v", err)
		}
		if err := client.MySQL.Model(&entity.Friend{}).
			Where("wxid LIKE '%@chatroom'").
			Where("enable_ai = 1 OR enable_chat_rank = 1 OR enable_welcome = 1 OR enable_command = 1 OR enable_summary = 1 OR enable_news = 1 OR enable_hot_top = 1 OR clear_member > 0").
			Update("enable_monitor", true).Error; err != nil {
			log.Panicf("迁移已有群功能开关失败: %v", err)
		}
		log.Printf("数据库升级完成: 已新增群监控开关")
	}
	if !migrator.HasColumn(&entity.Friend{}, "EnableDormant") {
		if err := migrator.AddColumn(&entity.Friend{}, "EnableDormant"); err != nil {
			log.Panicf("新增潜水统计开关字段失败: %v", err)
		}
		log.Printf("数据库升级完成: 已新增潜水统计开关")
	}
	if !migrator.HasColumn(&entity.Friend{}, "EnableCommand") {
		if err := migrator.AddColumn(&entity.Friend{}, "EnableCommand"); err != nil {
			log.Panicf("新增私聊指令开关字段失败: %v", err)
		}
		log.Printf("数据库升级完成: 已新增私聊指令开关")
	}
	rankColumns := []struct {
		field       string
		column      string
		label       string
		inheritRank bool
	}{
		{field: "EnableWeeklyRank", column: "enable_weekly_rank", label: "周排行榜", inheritRank: true},
		{field: "EnableYearlyRank", column: "enable_yearly_rank", label: "年排行榜"},
	}
	for _, column := range rankColumns {
		if migrator.HasColumn(&entity.Friend{}, column.field) {
			continue
		}
		if err := migrator.AddColumn(&entity.Friend{}, column.field); err != nil {
			log.Panicf("新增%s开关字段失败: %v", column.label, err)
		}
		if column.inheritRank {
			if err := client.MySQL.Model(&entity.Friend{}).
				Where("wxid LIKE '%@chatroom'").
				Where("enable_chat_rank = 1").
				Update(column.column, true).Error; err != nil {
				log.Panicf("迁移%s开关失败: %v", column.label, err)
			}
		}
		log.Printf("数据库升级完成: 已新增%s开关", column.label)
	}
	if !migrator.HasColumn(&entity.GroupUser{}, "WechatNickname") {
		if err := migrator.AddColumn(&entity.GroupUser{}, "WechatNickname"); err != nil {
			log.Panicf("新增群成员微信昵称字段失败: %v", err)
		}
		log.Printf("数据库升级完成: 已新增群成员微信昵称字段")
	}
	messageIdentityColumns := []struct {
		field string
		label string
	}{
		{field: "SenderNickname", label: "消息发送者微信昵称"},
		{field: "SenderGroupNickname", label: "消息发送者群内昵称"},
		{field: "SenderWechatAccount", label: "消息发送者微信号"},
		{field: "GroupNickname", label: "消息群聊名称"},
	}
	for _, column := range messageIdentityColumns {
		if migrator.HasColumn(&entity.Message{}, column.field) {
			continue
		}
		if err := migrator.AddColumn(&entity.Message{}, column.field); err != nil {
			log.Panicf("新增%s字段失败: %v", column.label, err)
		}
		log.Printf("数据库升级完成: 已新增%s字段", column.label)
	}
	hadFeatureSchedule := migrator.HasColumn(&entity.FeatureSetting{}, "Schedule")
	hadMessageThreshold := migrator.HasColumn(&entity.FeatureSetting{}, "MessageThreshold")
	if err := client.MySQL.AutoMigrate(&entity.FeatureSetting{}); err != nil {
		log.Panicf("创建功能设置表失败: %v", err)
	}
	if err := client.MySQL.AutoMigrate(&entity.Owner{}); err != nil {
		log.Panicf("创建主人配置表失败: %v", err)
	}
	if err := client.MySQL.AutoMigrate(&entity.BridgeSetting{}); err != nil {
		log.Panicf("创建桥接器配置表失败: %v", err)
	}
	if _, err := service.GetBridgeSetting(); err != nil {
		log.Panicf("初始化桥接器配置失败: %v", err)
	}
	defaults := []entity.FeatureSetting{
		{Feature: "news", SendTime: "08:00", Schedule: "daily"},
		{Feature: "weekly", SendTime: "09:30", Schedule: "monday"},
		{Feature: "yearly", SendTime: "00:00", Schedule: "yearly"},
		{Feature: "dormant", SendTime: "09:00", Schedule: "monday", MessageThreshold: 1},
	}
	if err := client.MySQL.Model(&entity.FeatureSetting{}).
		Where("feature = ?", "yearly").
		Updates(map[string]any{"send_time": "00:00", "schedule": "yearly"}).Error; err != nil {
		log.Panicf("固定年排行榜推送时间失败: %v", err)
	}
	if err := client.MySQL.Model(&entity.Friend{}).
		Where("enable_yesterday_rank = 1 OR enable_monthly_rank = 1").
		Updates(map[string]any{"enable_yesterday_rank": false, "enable_monthly_rank": false}).Error; err != nil {
		log.Panicf("停用已移除排行榜功能失败: %v", err)
	}
	for _, setting := range defaults {
		if err := client.MySQL.FirstOrCreate(&setting, entity.FeatureSetting{Feature: setting.Feature}).Error; err != nil {
			log.Panicf("初始化功能设置[%s]失败: %v", setting.Feature, err)
		}
	}
	if !hadFeatureSchedule {
		if err := client.MySQL.Model(&entity.FeatureSetting{}).
			Where("feature IN ?", []string{"weekly", "dormant"}).
			Update("schedule", "monday").Error; err != nil {
			log.Panicf("迁移每周功能推送周期失败: %v", err)
		}
	}
	if !hadMessageThreshold {
		if err := client.MySQL.Model(&entity.FeatureSetting{}).
			Where("feature = ?", "dormant").
			Update("message_threshold", 1).Error; err != nil {
			log.Panicf("迁移潜水消息阈值失败: %v", err)
		}
	}
}

// 配置管理工具
var vp *viper.Viper

// InitConfig
// @description: 初始化配置
func InitConfig() {
	vp = viper.New()
	vp.AddConfigPath(".")      // 设置配置文件路径
	vp.SetConfigName("config") // 设置配置文件名
	vp.SetConfigType("yaml")   // 设置配置文件类型
	// Docker 和桥接器由 Go 编排层依据 .env 启动。这些连接参数必须与
	// compose/bridge 实际端口保持一致；网页可编辑的业务配置仍以
	// config.yaml 为准，不会被 .env 默认值覆盖。
	environmentBindings := map[string]string{
		"system.domain":  "SYSTEM_DOMAIN",
		"wechat.host":    "WECHAT_HOST",
		"mysql.host":     "MYSQL_HOST",
		"mysql.port":     "MYSQL_PORT",
		"mysql.user":     "MYSQL_USER",
		"mysql.password": "MYSQL_PASSWORD",
		"mysql.db":       "MYSQL_DATABASE",
		"redis.enable":   "REDIS_ENABLE",
		"redis.host":     "REDIS_HOST",
		"redis.port":     "REDIS_PORT",
		"redis.password": "REDIS_PASSWORD",
		"redis.db":       "REDIS_DB",
	}
	for key, environmentName := range environmentBindings {
		if err := vp.BindEnv(key, environmentName); err != nil {
			log.Panicf("绑定环境变量 %s 失败: %v", environmentName, err)
		}
	}
	// 读取配置文件
	if err := vp.ReadInConfig(); err != nil {
		log.Panicf("读取配置文件失败: %v", err)
	}
	// 绑定配置文件
	if err := vp.Unmarshal(&config.Conf); err != nil {
		log.Panicf("配置文件解析失败: %v", err)
	}
	log.Printf("配置文件解析完成")
	if !config.Conf.Wechat.Check() {
		log.Panicf("微信本地桥接配置缺失")
	}
	// 初始化数据库连接
	client.InitMySQLClient()
	ensureSchema()
	client.InitRedisClient()
	go service.WarmContactCache()

	// 下面的代码是配置变动之后自动刷新的
	vp.WatchConfig()
	vp.OnConfigChange(func(e fsnotify.Event) {
		// 绑定配置文件
		if err := vp.Unmarshal(&config.Conf); err != nil {
			log.Printf("配置文件更新失败: %v", err)
		} else {
			if !config.Conf.Wechat.Check() {
				log.Panicf("微信本地桥接配置缺失")
			}
			// 初始化数据库连接
			client.InitMySQLClient()
			ensureSchema()
			client.InitRedisClient()
			go service.WarmContactCache()
		}
	})
}
