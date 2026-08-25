package service

import "testing"

func TestParseOwnerCommand(t *testing.T) {
	tests := []struct {
		content string
		want    string
		ok      bool
	}{
		{"@Fly AI /周报", OwnerCommandWeekly, true},
		{"@Fly AI\u2005/潜水", OwnerCommandDormant, true},
		{"  @Fly AI /热点  ", OwnerCommandNews, true},
		{"/周报", "", false},
		{"@Fly AI /未知", "", false},
		{"@Fly AI2 /周报", "", false},
		{"@Fly AI /周报 现在", "", false},
	}
	for _, test := range tests {
		got, ok := ParseOwnerCommand(test.content, "Fly AI")
		if got != test.want || ok != test.ok {
			t.Fatalf("ParseOwnerCommand(%q)=(%q,%v), want (%q,%v)", test.content, got, ok, test.want, test.ok)
		}
	}
}
