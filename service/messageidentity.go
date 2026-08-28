package service

import (
	"github.com/loks666/FlyAI/model/dto"
	"strings"
	"sync"
	"time"
)

const messageIdentityCacheTTL = 5 * time.Minute

type messageIdentity struct {
	SenderNickname      string
	SenderGroupNickname string
	SenderWechatAccount string
	GroupNickname       string
}

type cachedMessageIdentity struct {
	identity  messageIdentity
	expiresAt time.Time
}

var messageIdentityCache = struct {
	sync.Mutex
	items map[string]cachedMessageIdentity
}{items: make(map[string]cachedMessageIdentity)}

// PopulateMessageIdentity enriches a parsed message from the locally synced
// contact tables. Missing contact data never blocks message persistence.
func PopulateMessageIdentity(message *dto.Message) {
	if message == nil || message.FromUser == "" {
		return
	}

	key := message.FromUser + "\x00" + message.GroupUser
	identity, exists := getCachedMessageIdentity(key)
	if !exists {
		identity = loadMessageIdentity(message)
		putCachedMessageIdentity(key, identity)
	}
	message.SenderNickname = identity.SenderNickname
	message.SenderGroupNickname = identity.SenderGroupNickname
	message.SenderWechatAccount = identity.SenderWechatAccount
	message.GroupNickname = identity.GroupNickname
}

func loadMessageIdentity(message *dto.Message) messageIdentity {
	if !message.IsGroup() {
		sender, _ := getContact(message.FromUser)
		identity := messageIdentity{
			SenderNickname:      sender.Nickname,
			SenderWechatAccount: sender.CustomAccount,
		}
		if identity.SenderNickname == "" {
			identity.SenderNickname = message.FromUser
		}
		return identity
	}

	group, _ := getContact(message.FromUser)
	identity := messageIdentity{GroupNickname: group.Nickname}
	if identity.GroupNickname == "" {
		identity.GroupNickname = message.FromUser
	}
	if message.GroupUser == "" {
		return identity
	}

	member, _ := getGroupMember(message.FromUser, message.GroupUser)
	sender, _ := getContact(message.GroupUser)

	identity.SenderGroupNickname = member.Nickname
	identity.SenderNickname = firstNonEmpty(member.WechatNickname, sender.Nickname, member.Nickname, message.GroupUser)
	identity.SenderWechatAccount = firstNonEmpty(member.Account, sender.CustomAccount)
	return identity
}

func fallbackMessageIdentity(message *dto.Message) messageIdentity {
	if message.IsGroup() {
		identity := messageIdentity{GroupNickname: message.FromUser}
		if message.GroupUser != "" {
			identity.SenderNickname = message.GroupUser
		}
		return identity
	}
	return messageIdentity{SenderNickname: message.FromUser}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func getCachedMessageIdentity(key string) (messageIdentity, bool) {
	now := time.Now()
	messageIdentityCache.Lock()
	defer messageIdentityCache.Unlock()
	item, exists := messageIdentityCache.items[key]
	if !exists || !now.Before(item.expiresAt) {
		delete(messageIdentityCache.items, key)
		return messageIdentity{}, false
	}
	return item.identity, true
}

func putCachedMessageIdentity(key string, identity messageIdentity) {
	now := time.Now()
	messageIdentityCache.Lock()
	defer messageIdentityCache.Unlock()
	messageIdentityCache.items[key] = cachedMessageIdentity{
		identity:  identity,
		expiresAt: now.Add(messageIdentityCacheTTL),
	}
	if len(messageIdentityCache.items) <= 10000 {
		return
	}
	for storedKey, item := range messageIdentityCache.items {
		if !now.Before(item.expiresAt) {
			delete(messageIdentityCache.items, storedKey)
		}
	}
}

func invalidateAllMessageIdentityCache() {
	messageIdentityCache.Lock()
	messageIdentityCache.items = make(map[string]cachedMessageIdentity)
	messageIdentityCache.Unlock()
}

func invalidateGroupMessageIdentityCache(groupId string) {
	prefix := groupId + "\x00"
	messageIdentityCache.Lock()
	for key := range messageIdentityCache.items {
		if strings.HasPrefix(key, prefix) {
			delete(messageIdentityCache.items, key)
		}
	}
	messageIdentityCache.Unlock()
}
