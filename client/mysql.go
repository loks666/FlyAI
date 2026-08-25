package client

import (
	"github.com/loks666/FlyAI/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strconv"
)

// MySQL MySQL客户端
var MySQL *gorm.DB

func InitMySQLClient() {
	// 创建连接对象
	mysqlConfig := mysql.Config{
		DSN:                     config.Conf.MySQL.GetDSN(),
		DontSupportRenameIndex:  true, // 重命名索引时采用删除并新建的方式
		DontSupportRenameColumn: true, // 用 `change` 重命名列
	}

	// gorm 配置
	gormConfig := gorm.Config{}
	// 是否开启调试模式
	if flag, _ := strconv.ParseBool(os.Getenv("GORM_DEBUG")); flag {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}

	conn, err := gorm.Open(mysql.New(mysqlConfig), &gormConfig)
	if err != nil {
		log.Panicf("初始化MySQL连接失败, 错误信息: %v", err)
	} else {
		log.Println("MySQL连接成功")
	}
	MySQL = conn
}
