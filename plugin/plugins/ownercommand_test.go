package plugins

import (
	"github.com/loks666/FlyAI/service"
	"strings"
	"testing"
)

func TestRunOwnerCommandDoesNotRequireScheduledFeatureSwitches(t *testing.T) {
	originalYesterday := sendYesterdayByOwner
	originalWeekly := sendWeeklyByOwner
	originalMonthly := sendMonthlyByOwner
	originalYearly := sendYearlyByOwner
	originalDormant := sendDormantByOwner
	originalNews := sendNewsByOwner
	originalAttack := sendAttackByOwner
	defer func() {
		sendYesterdayByOwner = originalYesterday
		sendWeeklyByOwner = originalWeekly
		sendMonthlyByOwner = originalMonthly
		sendYearlyByOwner = originalYearly
		sendDormantByOwner = originalDormant
		sendNewsByOwner = originalNews
		sendAttackByOwner = originalAttack
	}()

	called := make(map[string]string)
	sendYesterdayByOwner = func(groupId string) (string, error) {
		called[service.OwnerCommandYesterday] = groupId
		return "", nil
	}
	sendWeeklyByOwner = func(groupId string) (string, error) {
		called[service.OwnerCommandWeekly] = groupId
		return "", nil
	}
	sendMonthlyByOwner = func(groupId string) (string, error) {
		called[service.OwnerCommandMonthly] = groupId
		return "", nil
	}
	sendYearlyByOwner = func(groupId string) (string, error) {
		called[service.OwnerCommandYearly] = groupId
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
	sendAttackByOwner = func(groupId string, targetNames []string, mode service.OwnerAttackMode) (string, error) {
		called[service.OwnerCommandAttack] = groupId + ":" + strings.Join(targetNames, ",") + ":" + string(mode)
		return "", nil
	}

	tests := []struct {
		command string
	}{
		{service.OwnerCommandYesterday},
		{service.OwnerCommandWeekly},
		{service.OwnerCommandMonthly},
		{service.OwnerCommandYearly},
		{service.OwnerCommandDormant},
		{service.OwnerCommandNews},
		{service.OwnerCommandAttack},
	}
	for _, test := range tests {
		if err := runOwnerCommand("test@chatroom", []string{"昵称1", "昵称2"}, service.OwnerAttackMorse, test.command); err != nil {
			t.Fatalf("command %s unexpectedly required a scheduled feature switch: %v", test.command, err)
		}
		expected := "test@chatroom"
		if test.command == service.OwnerCommandAttack {
			expected += ":昵称1,昵称2:morse"
		}
		if called[test.command] != expected {
			t.Fatalf("command %s did not invoke its sender", test.command)
		}
	}
}
