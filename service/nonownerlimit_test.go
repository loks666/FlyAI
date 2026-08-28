package service

import (
	"testing"
	"time"
)

func TestNonOwnerMemoryStateFirstSecondMuteAndExpire(t *testing.T) {
	state := newNonOwnerMemoryState()
	now := time.Date(2026, 8, 26, 10, 0, 0, 0, time.Local)
	state.now = func() time.Time { return now }
	userId := "wxid_user"

	first := state.begin(userId)
	if first != NonOwnerReplyRandom {
		t.Fatalf("first action = %d", first)
	}
	state.complete(userId, first, true)

	second := state.begin(userId)
	if second != NonOwnerReplyBalanceAndMute {
		t.Fatalf("second action = %d", second)
	}
	state.complete(userId, second, true)

	if !state.isSilenced(userId) {
		t.Fatal("muted sender should be silenced")
	}
	if got := state.begin(userId); got != NonOwnerReplyIgnore {
		t.Fatalf("muted action = %d", got)
	}
	now = now.Add(time.Hour)
	if state.isSilenced(userId) {
		t.Fatal("sender should not remain silenced after one hour")
	}
	if got := state.begin(userId); got != NonOwnerReplyRandom {
		t.Fatalf("action after expiry = %d", got)
	}
}

func TestNonOwnerStateUsesOnlySenderWxidAcrossGroups(t *testing.T) {
	if got := nonOwnerReplyKey("wxid_user"); got != "flyai:non-owner-reply:v2:wxid_user" {
		t.Fatalf("key = %q", got)
	}
}

func TestNonOwnerFailedSendRollsBackStage(t *testing.T) {
	state := newNonOwnerMemoryState()
	first := state.begin("wxid_user")
	state.complete("wxid_user", first, false)
	if got := state.begin("wxid_user"); got != NonOwnerReplyRandom {
		t.Fatalf("failed first send should retry random, got %d", got)
	}
	state.complete("wxid_user", NonOwnerReplyRandom, true)
	second := state.begin("wxid_user")
	state.complete("wxid_user", second, false)
	if got := state.begin("wxid_user"); got != NonOwnerReplyBalanceAndMute {
		t.Fatalf("failed balance send should retry balance, got %d", got)
	}
}
