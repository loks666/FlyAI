package utils

import (
	"reflect"
	"testing"
)

func TestMorningHotboardSourceLimits(t *testing.T) {
	want := []hotboardSource{
		{Type: "zhihu", Name: "知乎热榜", Limit: 10},
	}
	if !reflect.DeepEqual(morningHotboardSources, want) {
		t.Fatalf("sources=%+v, want %+v", morningHotboardSources, want)
	}
}

func TestPlainHTTPURL(t *testing.T) {
	want := "https://www.zhihu.com/search?q=Zhihu+CLI%E4%B8%8A%E7%BA%BF&search_source=Trending"
	if got := plainHTTPURL("https://www.zhihu.com/search?q=Zhihu+CLI%E4%B8%8A%E7%BA%BF&amp;search_source=Trending"); got != want {
		t.Fatalf("url=%q, want %q", got, want)
	}
	if got := plainHTTPURL("javascript:alert(1)"); got != "" {
		t.Fatalf("unsafe url=%q, want empty", got)
	}
}
