package service

import "testing"

func TestContactCacheKeysAreScoped(t *testing.T) {
	if got := contactCacheKey("wxid_user"); got != "flyai:contact:v1:wxid_user" {
		t.Fatalf("contactCacheKey() = %q", got)
	}
	if got := groupMemberCacheKey("group@chatroom", "wxid_user"); got != "flyai:group-member:v1:group@chatroom:wxid_user" {
		t.Fatalf("groupMemberCacheKey() = %q", got)
	}
	if got := groupMemberIndexKey("group@chatroom"); got != "flyai:group-member:index:v1:group@chatroom" {
		t.Fatalf("groupMemberIndexKey() = %q", got)
	}
}

func TestInvalidateGroupMessageIdentityCache(t *testing.T) {
	invalidateAllMessageIdentityCache()
	putCachedMessageIdentity("group-one\x00user", messageIdentity{SenderNickname: "one"})
	putCachedMessageIdentity("group-two\x00user", messageIdentity{SenderNickname: "two"})

	invalidateGroupMessageIdentityCache("group-one")
	if _, exists := getCachedMessageIdentity("group-one\x00user"); exists {
		t.Fatal("target group cache should be invalidated")
	}
	if _, exists := getCachedMessageIdentity("group-two\x00user"); !exists {
		t.Fatal("other group cache should remain")
	}
}
