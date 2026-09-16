package initialization

import "testing"

func TestNonOwnerCommandMatchesSupportedOwnerCommands(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "ranking command", content: "@Fly AI /周排", want: true},
		{name: "ranking command without slash", content: "@Fly AI 周排", want: true},
		{name: "dormant command", content: "@Fly AI /潜水", want: true},
		{name: "dormant command without slash", content: "@Fly AI 潜水", want: true},
		{name: "attack with robot mention", content: "@Fly AI 给我骂他@目标", want: true},
		{name: "attack without robot mention", content: "给我骂他@目标", want: true},
		{name: "morse attack", content: "给我骂他.@目标", want: true},
		{name: "unknown slash command", content: "@Fly AI /不存在", want: false},
		{name: "ordinary message", content: "今天吃什么", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := nonOwnerCommandMatches(test.content, "Fly AI"); got != test.want {
				t.Fatalf("nonOwnerCommandMatches(%q) = %v, want %v", test.content, got, test.want)
			}
		})
	}
}
