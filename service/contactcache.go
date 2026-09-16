package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/common/applog"
	"github.com/loks666/FlyAI/model/entity"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"time"
)

const contactCacheTTL = time.Hour

const contactIndexKey = "flyai:contact:index:v1"

func contactCacheKey(wxid string) string {
	return "flyai:contact:v1:" + wxid
}

func groupMemberCacheKey(groupId, wxid string) string {
	return "flyai:group-member:v1:" + groupId + ":" + wxid
}

func groupMemberIndexKey(groupId string) string {
	return "flyai:group-member:index:v1:" + groupId
}

func getContact(wxid string) (entity.Friend, error) {
	var contact entity.Friend
	if wxid == "" {
		return contact, fmt.Errorf("联系人 wxid 为空")
	}
	if getContactCacheValue(contactCacheKey(wxid), &contact) {
		return contact, nil
	}
	if client.MySQL == nil {
		return contact, fmt.Errorf("MySQL 未初始化")
	}
	query := client.MySQL.Session(&gorm.Session{Logger: gormLogger.Discard}).Where("wxid = ?", wxid).Take(&contact)
	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		applog.ContactCachePrintf("record_not_found table=t_friend wxid=%q", wxid)
		return contact, query.Error
	}
	if query.Error != nil {
		log.Printf("查询联系人失败[wxid=%s]: %v", wxid, query.Error)
		return contact, query.Error
	}
	setContactCacheValue(contactCacheKey(wxid), contact)
	return contact, nil
}

func getGroupMember(groupId, wxid string) (entity.GroupUser, error) {
	var member entity.GroupUser
	if groupId == "" || wxid == "" {
		return member, fmt.Errorf("群或成员 wxid 为空")
	}
	if getContactCacheValue(groupMemberCacheKey(groupId, wxid), &member) {
		return member, nil
	}
	if client.MySQL == nil {
		return member, fmt.Errorf("MySQL 未初始化")
	}
	query := client.MySQL.Session(&gorm.Session{Logger: gormLogger.Discard}).Where("group_id = ? AND wxid = ?", groupId, wxid).Take(&member)
	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		applog.ContactCachePrintf("record_not_found table=t_group_user group_id=%q wxid=%q", groupId, wxid)
		return member, query.Error
	}
	if query.Error != nil {
		log.Printf("查询群成员失败[group=%s wxid=%s]: %v", groupId, wxid, query.Error)
		return member, query.Error
	}
	setContactCacheValue(groupMemberCacheKey(groupId, wxid), member)
	return member, nil
}

func getContactCacheValue(key string, destination any) bool {
	redisClient := client.GetRedisClient()
	if redisClient == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	data, err := redisClient.Get(ctx, key).Bytes()
	cancel()
	if err != nil {
		return false
	}
	if err = json.Unmarshal(data, destination); err != nil {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		_ = redisClient.Del(ctx, key).Err()
		cancel()
		return false
	}
	return true
}

func setContactCacheValue(key string, value any) {
	redisClient := client.GetRedisClient()
	if redisClient == nil {
		return
	}
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_ = redisClient.Set(ctx, key, data, contactCacheTTL).Err()
	cancel()
}

// WarmContactCache rebuilds Redis contact records and lookup indexes from the
// authoritative MySQL tables. Failures are non-fatal because all reads fall
// back to MySQL.
func WarmContactCache() {
	if client.GetRedisClient() == nil || client.MySQL == nil {
		return
	}
	var contacts []entity.Friend
	if err := client.MySQL.Find(&contacts).Error; err != nil {
		log.Printf("预热 Redis 通讯录失败: %v", err)
		return
	}
	var members []entity.GroupUser
	if err := client.MySQL.Find(&members).Error; err != nil {
		log.Printf("预热 Redis 群成员失败: %v", err)
		return
	}
	if err := writeContactCache(contacts, members); err != nil {
		log.Printf("写入 Redis 通讯录缓存失败，继续使用 MySQL: %v", err)
		return
	}
	log.Printf("Redis 通讯录缓存已预热: contacts=%d groupMembers=%d ttl=%s", len(contacts), len(members), contactCacheTTL)
}

func refreshFriendContactCache() {
	if client.GetRedisClient() == nil {
		return
	}
	var contacts []entity.Friend
	if err := client.MySQL.Find(&contacts).Error; err != nil {
		log.Printf("刷新 Redis 联系人缓存失败: %v", err)
		return
	}
	if err := writeFriendCache(contacts); err != nil {
		log.Printf("刷新 Redis 联系人缓存失败: %v", err)
	}
}

func refreshGroupMemberContactCache(groupId string) {
	if client.GetRedisClient() == nil {
		return
	}
	var members []entity.GroupUser
	if err := client.MySQL.Where("group_id = ?", groupId).Find(&members).Error; err != nil {
		log.Printf("刷新 Redis 群成员缓存失败[%s]: %v", groupId, err)
		return
	}
	if err := writeGroupMemberCache(groupId, members); err != nil {
		log.Printf("刷新 Redis 群成员缓存失败[%s]: %v", groupId, err)
	}
}

func writeContactCache(contacts []entity.Friend, members []entity.GroupUser) error {
	if err := writeFriendCache(contacts); err != nil {
		return err
	}
	grouped := make(map[string][]entity.GroupUser)
	for _, member := range members {
		grouped[member.GroupId] = append(grouped[member.GroupId], member)
	}
	for groupId, groupMembers := range grouped {
		if err := writeGroupMemberCache(groupId, groupMembers); err != nil {
			return err
		}
	}
	return nil
}

func writeFriendCache(contacts []entity.Friend) error {
	redisClient := client.GetRedisClient()
	if redisClient == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pipe := redisClient.Pipeline()
	pipe.Del(ctx, contactIndexKey)
	for _, contact := range contacts {
		data, err := json.Marshal(contact)
		if err != nil {
			return err
		}
		pipe.Set(ctx, contactCacheKey(contact.Wxid), data, contactCacheTTL)
		if contact.IsOk {
			pipe.SAdd(ctx, contactIndexKey, contact.Wxid)
		}
	}
	pipe.Expire(ctx, contactIndexKey, contactCacheTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func writeGroupMemberCache(groupId string, members []entity.GroupUser) error {
	redisClient := client.GetRedisClient()
	if redisClient == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pipe := redisClient.Pipeline()
	indexKey := groupMemberIndexKey(groupId)
	pipe.Del(ctx, indexKey)
	for _, member := range members {
		data, err := json.Marshal(member)
		if err != nil {
			return err
		}
		pipe.Set(ctx, groupMemberCacheKey(groupId, member.Wxid), data, contactCacheTTL)
		if member.IsMember {
			pipe.SAdd(ctx, indexKey, member.Wxid)
		}
	}
	pipe.Expire(ctx, indexKey, contactCacheTTL)
	_, err := pipe.Exec(ctx)
	return err
}
