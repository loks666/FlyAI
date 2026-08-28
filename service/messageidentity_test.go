package service

import (
	"github.com/loks666/FlyAI/model/dto"
	"testing"
	"time"
)

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "wechat nickname", "group nickname", "wxid"); got != "wechat nickname" {
		t.Fatalf("firstNonEmpty() = %q", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Fatalf("firstNonEmpty() = %q, want empty", got)
	}
}

func TestFallbackMessageIdentity(t *testing.T) {
	groupMessage := &dto.Message{FromUser: "group@chatroom", GroupUser: "wxid_sender"}
	groupIdentity := fallbackMessageIdentity(groupMessage)
	if groupIdentity.SenderNickname != "wxid_sender" || groupIdentity.GroupNickname != "group@chatroom" {
		t.Fatalf("unexpected group fallback: %+v", groupIdentity)
	}

	privateIdentity := fallbackMessageIdentity(&dto.Message{FromUser: "wxid_friend"})
	if privateIdentity.SenderNickname != "wxid_friend" {
		t.Fatalf("unexpected private fallback: %+v", privateIdentity)
	}
}

func TestMessageIdentityCacheExpires(t *testing.T) {
	key := "test-group\x00test-user"
	putCachedMessageIdentity(key, messageIdentity{SenderNickname: "nickname"})
	messageIdentityCache.Lock()
	item := messageIdentityCache.items[key]
	item.expiresAt = time.Now().Add(-time.Second)
	messageIdentityCache.items[key] = item
	messageIdentityCache.Unlock()

	if _, exists := getCachedMessageIdentity(key); exists {
		t.Fatal("expired cache entry should not be returned")
	}
}
