package app

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/tasks/watergroup"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
	"sync"
)

type runWeeklyRankParam struct {
	WxId   string `json:"wxId" binding:"required"`
	DryRun bool   `json:"dryRun"`
}

const weeklyRankAllKey = "__all__"

var weeklyRankState = struct {
	sync.Mutex
	running map[string]bool
}{running: make(map[string]bool)}

func acquireWeeklyRank(key string, all bool) bool {
	weeklyRankState.Lock()
	defer weeklyRankState.Unlock()
	if all {
		if len(weeklyRankState.running) > 0 {
			return false
		}
		weeklyRankState.running[weeklyRankAllKey] = true
		return true
	}
	if weeklyRankState.running[weeklyRankAllKey] || weeklyRankState.running[key] {
		return false
	}
	weeklyRankState.running[key] = true
	return true
}

func releaseWeeklyRank(key string) {
	weeklyRankState.Lock()
	delete(weeklyRankState.running, key)
	weeklyRankState.Unlock()
}

// RunWeeklyRank manually sends one enabled group's rolling seven-day ranking.
func RunWeeklyRank(ctx *gin.Context) {
	var p runWeeklyRankParam
	if err := ctx.ShouldBindJSON(&p); err != nil || !strings.HasSuffix(p.WxId, "@chatroom") {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "群参数错误"})
		return
	}

	group, err := service.GetFriendInfoById(p.WxId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "群不存在"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取群配置失败"})
		return
	}
	if !group.IsOk || !group.EnableMonitor || !group.EnableChatRank {
		ctx.JSON(http.StatusConflict, gin.H{"message": "请先开启该群的监控和每周排行榜"})
		return
	}

	if !acquireWeeklyRank(p.WxId, false) {
		ctx.JSON(http.StatusConflict, gin.H{"message": "该群排行榜正在生成，请勿重复点击"})
		return
	}
	defer releaseWeeklyRank(p.WxId)

	if p.DryRun {
		preview, err := watergroup.BuildWeekRank(p.WxId)
		if errors.Is(err, watergroup.ErrNoWeekMessages) {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"message": "最近7天没有可排行的消息"})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "预览生成成功，未发送", "preview": preview})
		return
	}

	if _, err := watergroup.SendWeekToGroup(p.WxId); err != nil {
		if errors.Is(err, watergroup.ErrNoWeekMessages) {
			ctx.JSON(http.StatusUnprocessableEntity, gin.H{"message": "最近7天没有可排行的消息"})
			return
		}
		log.Printf("手动发送群[%s]最近7天排行失败: %v", p.WxId, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "发送失败，请查看服务日志"})
		return
	}

	log.Printf("已手动发送群[%s]最近7天排行", p.WxId)
	ctx.JSON(http.StatusOK, gin.H{"message": "最近7天排行榜已发送"})
}

type weeklyRankResult struct {
	WxId     string `json:"wxId"`
	Nickname string `json:"nickname"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}

// RunAllWeeklyRanks sends the rolling seven-day ranking to every group that has
// both monitoring and weekly ranking enabled.
func RunAllWeeklyRanks(ctx *gin.Context) {
	var p struct {
		DryRun bool `json:"dryRun"`
	}
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	groups, err := service.GetAllEnableChatRank()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取群配置失败"})
		return
	}
	if len(groups) == 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"message": "没有同时开启监控和周报的群"})
		return
	}
	if !acquireWeeklyRank(weeklyRankAllKey, true) {
		ctx.JSON(http.StatusConflict, gin.H{"message": "已有排行榜发送任务正在执行"})
		return
	}
	defer releaseWeeklyRank(weeklyRankAllKey)

	results := make([]weeklyRankResult, 0, len(groups))
	var succeeded, skipped, failed int
	for _, group := range groups {
		result := weeklyRankResult{WxId: group.Wxid, Nickname: group.Nickname}
		var runErr error
		if p.DryRun {
			_, runErr = watergroup.BuildWeekRank(group.Wxid)
		} else {
			_, runErr = watergroup.SendWeekToGroup(group.Wxid)
		}

		switch {
		case errors.Is(runErr, watergroup.ErrNoWeekMessages):
			result.Status = "skipped"
			result.Message = "最近7天没有可排行的消息"
			skipped++
		case runErr != nil:
			result.Status = "failed"
			result.Message = runErr.Error()
			failed++
		default:
			result.Status = "success"
			if p.DryRun {
				result.Message = "预览生成成功，未发送"
			} else {
				result.Message = "发送成功"
			}
			succeeded++
		}
		results = append(results, result)
	}

	action := "批量发送"
	if p.DryRun {
		action = "批量预览"
	}
	message := fmt.Sprintf("%s完成：成功 %d 个，跳过 %d 个，失败 %d 个", action, succeeded, skipped, failed)
	log.Printf("%s", message)
	ctx.JSON(http.StatusOK, gin.H{
		"message": message,
		"total":   len(groups),
		"success": succeeded,
		"skipped": skipped,
		"failed":  failed,
		"results": results,
	})
}
