package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/app"
	"github.com/loks666/FlyAI/common/applog"
	"github.com/loks666/FlyAI/initialization"
	"github.com/loks666/FlyAI/mq"
	"github.com/loks666/FlyAI/orchestration"
	"github.com/loks666/FlyAI/router"
	"github.com/loks666/FlyAI/tasks"
	"github.com/loks666/FlyAI/tcpserver"
	"github.com/loks666/FlyAI/webui"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func initializeApplication() {
	initialization.InitConfig() // 初始化配置
	initialization.Plugin()     // 注册插件
	tasks.InitTasks()           // 初始化定时任务
	mq.Init()                   // 初始化MQ

	if flag, _ := strconv.ParseBool(os.Getenv("DONT_SEND")); flag {
		log.Printf("已设置环境变量DONT_SEND，不发送消息")
	}
	if flag, _ := strconv.ParseBool(os.Getenv("DONT_SAVE")); flag {
		log.Printf("已设置环境变量DONT_SAVE，不保存消息")
	}
}

func monitorAccessFileLogger(writeLog func(string, ...any)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path != "/api/monitor/groups" {
			ctx.Next()
			return
		}
		startedAt := time.Now()
		ctx.Next()
		writeLog(
			"status=%d latency=%s client=%s method=%s path=%q",
			ctx.Writer.Status(),
			time.Since(startedAt),
			ctx.ClientIP(),
			ctx.Request.Method,
			ctx.Request.URL.RequestURI(),
		)
	}
}

func managementURL(httpPort string) string {
	if configured := strings.TrimSpace(os.Getenv("SYSTEM_DOMAIN")); configured != "" {
		return strings.TrimRight(configured, "/")
	}
	return "http://127.0.0.1:" + httpPort
}

func autoOpenBrowserEnabled() bool {
	configured := strings.TrimSpace(os.Getenv("OPEN_BROWSER"))
	if configured == "" {
		return true
	}
	enabled, err := strconv.ParseBool(configured)
	return err == nil && enabled
}

// bridgeAutoStartEnabled 控制 FlyAI 启动后是否按已保存的机器人微信号自动拉起桥接器，
// 默认开启；设为 false 时回到「在管理页面手动点启动」的行为。
func bridgeAutoStartEnabled() bool {
	configured := strings.TrimSpace(os.Getenv("BRIDGE_AUTOSTART"))
	if configured == "" {
		return true
	}
	enabled, err := strconv.ParseBool(configured)
	return err == nil && enabled
}

func openBrowser(url string) error {
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Start()
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dependencies, err := orchestration.Start(ctx)
	if err != nil {
		log.Printf("FlyAI 依赖启动失败: %v", err)
		return
	}
	defer dependencies.Close()
	app.SetBridgeRuntime(dependencies)
	initializeApplication()

	// 启动TCP服务
	go tcpserver.Start()

	// 启动HTTP服务
	engine := gin.New()
	engine.Use(monitorAccessFileLogger(applog.HTTPMonitorPrintf))
	engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/api/monitor/groups"}}))
	engine.Use(gin.Recovery())

	webAssets, err := webui.Load()
	if err != nil {
		log.Printf("新版管理页面资源加载失败: %v", err)
		return
	}
	log.Printf("前端静态资源目录: %s", webAssets.Root())
	engine.StaticFS("/_next", webAssets.NextAssets())
	// 404返回数据
	engine.NoRoute(router.WebFallback(webAssets))
	engine.NoMethod(func(ctx *gin.Context) {
		ctx.String(http.StatusMethodNotAllowed, "不支持的请求方式")
	})
	// 初始化路由
	router.Init(engine, webAssets)
	httpPort := os.Getenv("HUB_HTTP_PORT")
	if _, err := strconv.Atoi(httpPort); err != nil {
		httpPort = "9528"
	}
	server := &http.Server{Addr: "127.0.0.1:" + httpPort, Handler: engine}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Panicf("服务启动失败：%v", err)
	}
	defer listener.Close()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	// 监听已就绪，桥接器可以回调 /api/monitor/groups 和消息接收端口了。
	// StartBridge 最长会等 90 秒，放到协程里跑，不阻塞管理页面。
	if bridgeAutoStartEnabled() {
		go app.AutoStartBridge(ctx)
	} else {
		log.Printf("已设置 BRIDGE_AUTOSTART=false，微信桥接器需在管理页面手动启动")
	}
	if autoOpenBrowserEnabled() {
		url := managementURL(httpPort)
		if err := openBrowser(url); err != nil {
			log.Printf("自动打开管理页面失败: %v；请手动访问 %s", err, url)
		} else {
			log.Printf("已打开管理页面: %s", url)
		}
	}
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Panicf("服务启动失败：%v", err)
	}
}
