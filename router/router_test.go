package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/loks666/FlyAI/webui"
)

func TestWebFallbackServesStaticExportNavigationFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	for name, content := range map[string]string{
		"index.html": "index",
		"group.html": "group",
		"group.txt":  "group-data",
		"404.html":   "not-found",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "_next"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("WEB_DIST_PATH", root)
	assets, err := webui.Load()
	if err != nil {
		t.Fatal(err)
	}
	handler := WebFallback(assets)

	for requestPath, expected := range map[string]string{
		"/group":     "group",
		"/group.txt": "group-data",
	} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, requestPath, nil)
		handler(ctx)
		if recorder.Code != http.StatusOK {
			t.Fatalf("%s status = %d", requestPath, recorder.Code)
		}
		if recorder.Body.String() != expected {
			t.Fatalf("%s body = %q", requestPath, recorder.Body.String())
		}
	}
}
