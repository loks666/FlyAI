package plugins

import (
	"github.com/loks666/FlyAI/service"
	"testing"
)

func TestRunOwnerCommandDoesNotRequireScheduledFeatureSwitches(t *testing.T) {
	originalWeekly := sendWeeklyByOwner
	originalDormant := sendDormantByOwner
	originalNews := sendNewsByOwner
	defer func() {
		sendWeeklyByOwner = originalWeekly
		sendDormantByOwner = originalDormant
		sendNewsByOwner = originalNews
	}()

	called := make(map[string]string)
	sendWeeklyByOwner = func(groupId string) (string, error) {
		called[service.OwnerCommandWeekly] = groupId
		return "", nil
	}
	sendDormantByOwner = func(groupId string) (string, error) {
		called[service.OwnerCommandDormant] = groupId
		return "", nil
	}
	sendNewsByOwner = func(groupId string) (string, error) {
		called[service.OwnerCommandNews] = groupId
		return "", nil
	}

	tests := []struct {
		command string
	}{
		{service.OwnerCommandWeekly},
		{service.OwnerCommandDormant},
		{service.OwnerCommandNews},
	}
	for _, test := range tests {
		if err := runOwnerCommand("test@chatroom", test.command); err != nil {
			t.Fatalf("command %s unexpectedly required a scheduled feature switch: %v", test.command, err)
		}
		if called[test.command] != "test@chatroom" {
			t.Fatalf("command %s did not invoke its sender", test.command)
		}
	}
}
