package service

import "testing"

func TestFilterDormantMembersUsesStrictThresholdAndIncludesZero(t *testing.T) {
	members := []DormantMember{
		{Wxid: "zero", Nickname: "零条"},
		{Wxid: "nineteen", Nickname: "十九条"},
		{Wxid: "twenty", Nickname: "二十条"},
		{Wxid: "more", Nickname: "更多"},
	}
	counts := map[string]int64{"nineteen": 19, "twenty": 20, "more": 21}

	got := filterDormantMembers(members, counts, 20)
	if len(got) != 2 {
		t.Fatalf("got %d members, want 2: %+v", len(got), got)
	}
	if got[0].Wxid != "zero" || got[0].Count != 0 {
		t.Fatalf("zero-message member missing or miscounted: %+v", got)
	}
	if got[1].Wxid != "nineteen" || got[1].Count != 19 {
		t.Fatalf("19-message member missing or miscounted: %+v", got)
	}
}
