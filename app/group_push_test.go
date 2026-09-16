package app

import (
	"testing"

	"github.com/loks666/FlyAI/service"
)

func TestFindGroupFeaturePusher(t *testing.T) {
	tests := map[string]string{
		service.FeatureWeekly:  "排行榜",
		service.FeatureYearly:  "年排行榜",
		service.FeatureNews:    "热点早报",
		service.FeatureDormant: "潜水统计",
	}
	for _, removed := range []string{"yesterday", "monthly"} {
		if _, ok := findGroupFeaturePusher(removed); ok {
			t.Fatalf("removed feature %q should not have a pusher", removed)
		}
	}
	for feature, wantLabel := range tests {
		pusher, ok := findGroupFeaturePusher(feature)
		if !ok || pusher.run == nil || pusher.label != wantLabel {
			t.Fatalf("feature %q mapping = %#v, %v; want label %q and runnable pusher", feature, pusher, ok, wantLabel)
		}
	}
	if _, ok := findGroupFeaturePusher("unknown"); ok {
		t.Fatal("unknown feature should not have a pusher")
	}
}
