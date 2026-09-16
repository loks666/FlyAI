package service

import (
	"testing"

	"github.com/loks666/FlyAI/model/entity"
)

func TestOwnerWechatAccountMatchesUniqueAccount(t *testing.T) {
	owner := entity.Owner{
		WechatAccount: "owner_account",
		Wxid:          "wxid_owner",
		Nickname:      "shared nickname",
	}

	if !ownerWechatAccountMatches(owner, "owner_account") {
		t.Fatal("matching owner WeChat account should be accepted")
	}
	if ownerWechatAccountMatches(owner, "attacker_account") {
		t.Fatal("a different WeChat account must be rejected")
	}
	if ownerWechatAccountMatches(owner, "") {
		t.Fatal("missing WeChat account must fail closed")
	}
}
