package orchestration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestRenderInitialConfigProducesValidYAML(t *testing.T) {
	t.Setenv("MYSQL_PASSWORD", "p@ss'word")
	t.Setenv("AI_PERSONALITY", "中文助手")
	t.Setenv("NEW_FRIEND_NOTIFY_USERS", "wxid_one, wxid_two")

	var decoded map[string]any
	if err := yaml.Unmarshal([]byte(renderInitialConfig()), &decoded); err != nil {
		t.Fatalf("rendered config is invalid YAML: %v", err)
	}
	mysqlConfig := decoded["mysql"].(map[string]any)
	if mysqlConfig["password"] != "p@ss'word" {
		t.Fatalf("unexpected mysql password: %#v", mysqlConfig["password"])
	}
	systemConfig := decoded["system"].(map[string]any)
	notify := systemConfig["newFriendNotify"].(map[string]any)
	users := notify["toUser"].([]any)
	if len(users) != 2 || users[0] != "wxid_one" || users[1] != "wxid_two" {
		t.Fatalf("unexpected notify users: %#v", users)
	}
}

func TestLoadEnvironmentCreatesLocalEnvAndLoadsValues(t *testing.T) {
	root := t.TempDir()
	example := "TEST_FLYAI_VALUE='hello world'\nEMPTY_VALUE=\n"
	if err := os.WriteFile(filepath.Join(root, ".env.example"), []byte(example), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadEnvironment(root); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("TEST_FLYAI_VALUE"); got != "hello world" {
		t.Fatalf("unexpected loaded value: %q", got)
	}
	if _, err := os.Stat(filepath.Join(root, ".env")); err != nil {
		t.Fatalf(".env was not created: %v", err)
	}
}

func TestBridgeHealthyValidatesFlyAIPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"code":1,"data":{"ingest":"127.0.0.1:19099"}}`))
	}))
	defer server.Close()
	if !bridgeHealthy(server.URL) {
		t.Fatal("expected bridge health response to be accepted")
	}
}

func TestBridgeHealthyRejectsUnrelatedHTTPService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()
	if bridgeHealthy(server.URL) {
		t.Fatal("unrelated HTTP service must not be accepted as FlyAI bridge")
	}
}

func TestReadBridgeHealthReturnsAccountAndWindowBinding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"code":1,"data":{"configuredAccount":"niubaidely","resolvedWxid":"wxid_robot","databaseVerified":true,"windowVerified":true,"windowPid":27296,"windowHandle":70700,"sendReady":true,"ingest":"127.0.0.1:19099"}}`))
	}))
	defer server.Close()
	health, ok := readBridgeHealth(context.Background(), server.URL)
	if !ok {
		t.Fatal("expected bridge health response to be accepted")
	}
	if health.ConfiguredAccount != "niubaidely" || health.ResolvedWxid != "wxid_robot" || !health.SendReady || health.WindowPID != 27296 {
		t.Fatalf("unexpected health payload: %#v", health)
	}
}
