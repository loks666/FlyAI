package app

import (
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/service"
	"net/http"
	"sync"
)

var manualWechatSyncState = struct {
	sync.Mutex
	running bool
}{}

// SyncWechatCatalog manually refreshes groups, contacts, and current group members.
func SyncWechatCatalog(ctx *gin.Context) {
	manualWechatSyncState.Lock()
	if manualWechatSyncState.running {
		manualWechatSyncState.Unlock()
		ctx.JSON(http.StatusConflict, gin.H{"message": "微信数据正在同步，请稍候"})
		return
	}
	manualWechatSyncState.running = true
	manualWechatSyncState.Unlock()
	defer func() {
		manualWechatSyncState.Lock()
		manualWechatSyncState.running = false
		manualWechatSyncState.Unlock()
	}()

	// 桥接器没起来时整段同步都会失败，提前给出明确原因而不是逐个群报连接错误。
	if !service.BridgeReady() {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"message": service.ErrBridgeNotRunning.Error()})
		return
	}

	service.SyncFriendCatalog()
	friends, groups, err := service.GetAllFriend()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取同步结果失败"})
		return
	}

	var activeGroups, memberCount int
	for _, group := range groups {
		if !group.IsOk {
			continue
		}
		service.SyncGroupMembers(group.Wxid)
		activeGroups++
		if count, countErr := service.CountCurrentGroupMembers(group.Wxid); countErr == nil {
			memberCount += int(count)
		}
	}
	activeContacts := 0
	for _, friend := range friends {
		if friend.IsOk {
			activeContacts++
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "微信数据同步完成",
		"groups":   activeGroups,
		"contacts": activeContacts,
		"members":  memberCount,
	})
}
