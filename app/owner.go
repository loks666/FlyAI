package app

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/service"
	"net/http"
	"strconv"
)

type ownerParam struct {
	WechatAccount string `json:"wechatAccount" binding:"required"`
}

func ownerErrorStatus(err error) int {
	switch {
	case errors.Is(err, service.ErrInvalidWechatAccount),
		errors.Is(err, service.ErrDirectWxidNotAllowed),
		errors.Is(err, service.ErrOwnerContactNotFound),
		errors.Is(err, service.ErrOwnerAlreadyExists),
		errors.Is(err, service.ErrOwnerNotFound):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func GetOwners(ctx *gin.Context) {
	owners, err := service.GetOwners()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "读取主人配置失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"owners": owners})
}

func CreateOwner(ctx *gin.Context) {
	var p ownerParam
	if err := ctx.ShouldBindJSON(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "请输入微信号"})
		return
	}
	service.SyncFriendCatalog()
	owner, err := service.CreateOwner(p.WechatAccount)
	if err != nil {
		ctx.JSON(ownerErrorStatus(err), gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "主人微信号已绑定", "owner": owner})
}

func UpdateOwner(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "主人参数错误"})
		return
	}
	var p ownerParam
	if err = ctx.ShouldBindJSON(&p); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "请输入微信号"})
		return
	}
	service.SyncFriendCatalog()
	owner, err := service.UpdateOwner(uint(id), p.WechatAccount)
	if err != nil {
		ctx.JSON(ownerErrorStatus(err), gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "主人微信号已更新", "owner": owner})
}

func DeleteOwner(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || id == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "主人参数错误"})
		return
	}
	if err = service.DeleteOwner(uint(id)); err != nil {
		ctx.JSON(ownerErrorStatus(err), gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "主人微信号已删除"})
}
