package app

import (
	"context"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/common/current"
	"github.com/loks666/FlyAI/initialization"
	"github.com/loks666/FlyAI/model/dto"
	"github.com/loks666/FlyAI/orchestration"
	"github.com/loks666/FlyAI/service"
)

var (
	bridgeRuntimeMu sync.RWMutex
	bridgeRuntime   *orchestration.Runtime
)

func SetBridgeRuntime(runtime *orchestration.Runtime) {
	bridgeRuntimeMu.Lock()
	bridgeRuntime = runtime
	bridgeRuntimeMu.Unlock()
}

func currentBridgeRuntime() *orchestration.Runtime {
	bridgeRuntimeMu.RLock()
	defer bridgeRuntimeMu.RUnlock()
	return bridgeRuntime
}

type bridgeSettingParam struct {
	WechatAccount string `json:"wechatAccount" binding:"required"`
}

func GetBridgeStatus(ctx *gin.Context) {
	setting, err := service.GetBridgeSetting()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取桥接器配置失败"})
		return
	}
	runtime := currentBridgeRuntime()
	if runtime == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"message": "桥接器管理器尚未初始化"})
		return
	}
	ctx.JSON(http.StatusOK, runtime.Status(ctx.Request.Context(), setting.WechatAccount))
}

func UpdateBridgeSetting(ctx *gin.Context) {
	var input bridgeSettingParam
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "请填写机器人微信号"})
		return
	}
	runtime := currentBridgeRuntime()
	if runtime == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"message": "桥接器管理器尚未初始化"})
		return
	}
	currentSetting, err := service.GetBridgeSetting()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取桥接器配置失败"})
		return
	}
	if currentSetting.WechatAccount != "" && !strings.EqualFold(currentSetting.WechatAccount, strings.TrimSpace(input.WechatAccount)) {
		runtime.StopBridge(ctx.Request.Context(), currentSetting.WechatAccount)
		current.SetRobotInfo(dto.RobotUserInfo{})
	}
	setting, err := service.SaveBridgeWechatAccount(input.WechatAccount)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "机器人微信号已保存，请启动桥接器完成账号与窗口绑定", "setting": setting})
}

func StartBridge(ctx *gin.Context) {
	setting, err := service.GetBridgeSetting()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取桥接器配置失败"})
		return
	}
	runtime := currentBridgeRuntime()
	if runtime == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"message": "桥接器管理器尚未初始化"})
		return
	}
	status, err := runtime.StartBridge(ctx.Request.Context(), setting.WechatAccount)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error(), "status": runtime.Status(ctx.Request.Context(), setting.WechatAccount)})
		return
	}
	if err := finishBridgeStart(status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "桥接器已启动，但保存账号解析结果失败", "status": status})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "微信桥接器已启动并绑定目标账号", "status": status})
}

// finishBridgeStart 收尾工作：持久化桥接器解析出的账号信息、刷新机器人身份并拉一次通讯录。
// 手动启动和服务启动时的自动启动共用这段逻辑。
func finishBridgeStart(status orchestration.BridgeStatus) error {
	if err := service.SaveBridgeResolution(status.ResolvedWxid, status.AccountDirectory); err != nil {
		return err
	}
	initialization.InitWechatRobotInfo()
	go service.SyncFriendCatalog()
	return nil
}

// AutoStartBridge 在 FlyAI 启动时按数据库中已保存的机器人微信号自动拉起桥接器。
// Runtime 是进程内状态，FlyAI 一退出桥接器就被回收，没有这一步每次重启都要再去管理页面点一次「启动」。
func AutoStartBridge(ctx context.Context) {
	setting, err := service.GetBridgeSetting()
	if err != nil {
		log.Printf("自动启动微信桥接器失败：读取桥接器配置失败: %v", err)
		return
	}
	if strings.TrimSpace(setting.WechatAccount) == "" {
		log.Printf("尚未配置机器人微信号，微信桥接器等待在管理页面配置后启动")
		return
	}
	runtime := currentBridgeRuntime()
	if runtime == nil {
		log.Printf("桥接器管理器尚未初始化，跳过自动启动微信桥接器")
		return
	}
	log.Printf("正在自动启动微信桥接器: %s", setting.WechatAccount)
	status, err := runtime.StartBridge(ctx, setting.WechatAccount)
	if err != nil {
		log.Printf("自动启动微信桥接器失败: %v；请在管理页面重试", err)
		return
	}
	if err := finishBridgeStart(status); err != nil {
		log.Printf("微信桥接器已启动，但保存账号解析结果失败: %v", err)
		return
	}
	log.Printf("微信桥接器已就绪并绑定账号: %s (%s)", status.WechatAccount, status.ResolvedWxid)
}

func StopBridge(ctx *gin.Context) {
	setting, err := service.GetBridgeSetting()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取桥接器配置失败"})
		return
	}
	runtime := currentBridgeRuntime()
	if runtime == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"message": "桥接器管理器尚未初始化"})
		return
	}
	status := runtime.StopBridge(ctx.Request.Context(), setting.WechatAccount)
	current.SetRobotInfo(dto.RobotUserInfo{})
	ctx.JSON(http.StatusOK, gin.H{"message": "微信桥接器已停止", "status": status})
}
