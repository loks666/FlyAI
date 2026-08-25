package app

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/utils"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"sync"
	"unicode/utf8"
)

type sendDirectMessageParam struct {
	WxId    string `json:"wxId" binding:"required"`
	Message string `json:"message" binding:"required"`
	DryRun  bool   `json:"dryRun"`
}

var directMessageLocks sync.Map

func directMessageLock(wxId string) *sync.Mutex {
	lock, _ := directMessageLocks.LoadOrStore(wxId, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// SendDirectMessage sends a plain-text private message to an active contact.
func SendDirectMessage(ctx *gin.Context) {
	var p sendDirectMessageParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "联系人和消息内容不能为空"})
		return
	}

	p.WxId = strings.TrimSpace(p.WxId)
	p.Message = strings.TrimSpace(p.Message)
	messageLength := utf8.RuneCountInString(p.Message)
	if p.WxId == "" || strings.HasSuffix(p.WxId, "@chatroom") {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "请选择个人联系人"})
		return
	}
	if messageLength == 0 || messageLength > 2000 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "消息内容需为 1～2000 个字符"})
		return
	}

	contact, err := service.GetFriendInfoById(p.WxId)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "联系人不存在，请刷新通讯录后重试"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取联系人失败"})
		return
	}
	if !contact.IsOk || strings.HasSuffix(contact.Wxid, "@chatroom") {
		ctx.JSON(http.StatusConflict, gin.H{"message": "该联系人当前不可私聊发送"})
		return
	}

	if p.DryRun {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "预览校验成功，未发送",
			"target":  contact.Nickname,
			"preview": p.Message,
		})
		return
	}

	lock := directMessageLock(p.WxId)
	lock.Lock()
	defer lock.Unlock()
	if err = utils.SendMessage(p.WxId, "", p.Message, 0); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "发送失败：" + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "消息已私聊发送", "target": contact.Nickname})
}
