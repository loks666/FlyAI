package app

import (
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/model/vo"
	"github.com/loks666/FlyAI/service"
	"net/http"
)

type webContact struct {
	Wxid             string `json:"wxid"`
	CustomAccount    string `json:"customAccount"`
	Nickname         string `json:"nickname"`
	LastActive       string `json:"lastActive"`
	EnableMonitor    bool   `json:"enableMonitor"`
	EnableWeeklyRank bool   `json:"enableWeeklyRank"`
	EnableYearlyRank bool   `json:"enableYearlyRank"`
	EnableNews       bool   `json:"enableNews"`
	EnableDormant    bool   `json:"enableDormant"`
}

func toWebContact(item vo.FriendItem) webContact {
	lastActive := ""
	if !item.LastActive.IsNil() {
		lastActive = item.LastActive.String()
	}
	return webContact{
		Wxid:             item.Wxid,
		CustomAccount:    item.CustomAccount,
		Nickname:         item.Nickname,
		LastActive:       lastActive,
		EnableMonitor:    item.EnableMonitor,
		EnableWeeklyRank: item.EnableWeeklyRank,
		EnableYearlyRank: item.EnableYearlyRank,
		EnableNews:       item.EnableNews,
		EnableDormant:    item.EnableDormant,
	}
}

func activeWebContacts(items []vo.FriendItem) []webContact {
	result := make([]webContact, 0, len(items))
	for _, item := range items {
		if item.IsOk {
			result = append(result, toWebContact(item))
		}
	}
	return result
}

// GetWebDashboard returns the existing home-page counters as JSON.
func GetWebDashboard(ctx *gin.Context) {
	service.SyncFriendCatalog()
	friends, groups, err := service.GetAllFriend()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取运行概览失败"})
		return
	}
	activeFriends := activeWebContacts(friends)
	activeGroups := activeWebContacts(groups)
	monitored := 0
	ranked := 0
	for _, group := range activeGroups {
		if group.EnableMonitor {
			monitored++
		}
		if group.EnableMonitor && group.EnableWeeklyRank {
			ranked++
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"contactCount": len(activeFriends),
		"groupCount":   len(activeGroups),
		"monitorCount": monitored,
		"rankCount":    ranked,
	})
}

// GetWebGroups returns current groups and all per-group feature switches.
func GetWebGroups(ctx *gin.Context) {
	service.SyncFriendCatalog()
	_, groups, err := service.GetAllFriend()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取群组失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"groups": activeWebContacts(groups)})
}

// GetWebContacts returns current direct contacts for the management UI.
func GetWebContacts(ctx *gin.Context) {
	service.SyncFriendCatalog()
	friends, _, err := service.GetAllFriend()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取通讯录失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"contacts": activeWebContacts(friends)})
}
