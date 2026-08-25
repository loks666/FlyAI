package main

import (
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/initialization"
	"github.com/loks666/FlyAI/mq"
	"github.com/loks666/FlyAI/router"
	"github.com/loks666/FlyAI/tasks"
	"github.com/loks666/FlyAI/tcpserver"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func init() {
	initialization.InitConfig()          // 初始化配置
	initialization.InitWechatRobotInfo() // 初始化机器人信息
	initialization.Plugin()              // 注册插件
	tasks.InitTasks()                    // 初始化定时任务
	mq.Init()                            // 初始化MQ

	if flag, _ := strconv.ParseBool(os.Getenv("DONT_SEND")); flag {
		log.Printf("已设置环境变量DONT_SEND，不发送消息")
	}
	if flag, _ := strconv.ParseBool(os.Getenv("DONT_SAVE")); flag {
		log.Printf("已设置环境变量DONT_SAVE，不保存消息")
	}
}

func main() {
	// 启动TCP服务
	go tcpserver.Start()

	// 启动HTTP服务
	app := gin.Default()

	// 自定义模板引擎函数
	app.SetFuncMap(template.FuncMap{
		"codeToChinese": func(code string) string {
			switch code {
			case "friend":
				return "通讯录"
			case "group":
				return "群组列表"
			case "index":
				return "首页"
			case "assistant":
				return "AI角色"
			case "settings":
				return "功能配置"
			default:
				return "其他页面"
			}
		},
		"boolToChinese": func(flag bool) string {
			if flag {
				return "是"
			}
			return "否"
		},
		"checkIsGroup": func(str string) bool {
			// 如果id包含@chatroom就是群组
			return strings.HasSuffix(str, "@chatroom")
		},
	})

	app.LoadHTMLGlob("views/*.html")
	app.Static("/assets", "./views/static")
	app.StaticFile("/favicon.ico", "./views/wechat.ico")
	// 404返回数据
	app.NoRoute(func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.URL.Path, "/api") {
			ctx.String(404, "接口不存在")
			return
		}
		// 404直接跳转到首页
		ctx.Redirect(302, "/404.html")
	})
	app.NoMethod(func(ctx *gin.Context) {
		ctx.String(http.StatusMethodNotAllowed, "不支持的请求方式")
	})
	// 初始化路由
	router.Init(app)
	httpPort := os.Getenv("HUB_HTTP_PORT")
	if _, err := strconv.Atoi(httpPort); err != nil {
		httpPort = "8080"
	}
	if err := app.Run("127.0.0.1:" + httpPort); err != nil {
		log.Panicf("服务启动失败：%v", err)
	}
}
