package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMonitorAccessIsWrittenToFileLoggerOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var console bytes.Buffer
	var fileLog strings.Builder
	router := gin.New()
	router.Use(monitorAccessFileLogger(func(format string, values ...any) {
		fileLog.WriteString(fmt.Sprintf(format, values...))
	}))
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    &console,
		SkipPaths: []string{"/api/monitor/groups"},
	}))
	router.GET("/api/monitor/groups", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	request := httptest.NewRequest(http.MethodGet, "/api/monitor/groups", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if console.Len() != 0 {
		t.Fatalf("monitor access unexpectedly written to console: %s", console.String())
	}
	if !strings.Contains(fileLog.String(), "path=\"/api/monitor/groups\"") {
		t.Fatalf("monitor access missing from file logger: %s", fileLog.String())
	}
}

func TestRegularAccessRemainsInConsole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var console bytes.Buffer
	var fileLog strings.Builder
	router := gin.New()
	router.Use(monitorAccessFileLogger(func(format string, values ...any) {
		fileLog.WriteString(fmt.Sprintf(format, values...))
	}))
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output:    &console,
		SkipPaths: []string{"/api/monitor/groups"},
	}))
	router.GET("/health", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if !strings.Contains(console.String(), "/health") {
		t.Fatalf("regular access missing from console: %s", console.String())
	}
	if fileLog.Len() != 0 {
		t.Fatalf("regular access unexpectedly written to monitor file: %s", fileLog.String())
	}
}

func TestManagementURLDefaultsToLocalHTTPPort(t *testing.T) {
	t.Setenv("SYSTEM_DOMAIN", "")
	if got := managementURL("9528"); got != "http://127.0.0.1:9528" {
		t.Fatalf("managementURL() = %q", got)
	}
}

func TestAutoOpenBrowserCanBeDisabled(t *testing.T) {
	t.Setenv("OPEN_BROWSER", "false")
	if autoOpenBrowserEnabled() {
		t.Fatal("auto open browser should be disabled")
	}
}

func TestBridgeAutoStartDefaultsToEnabled(t *testing.T) {
	t.Setenv("BRIDGE_AUTOSTART", "")
	if !bridgeAutoStartEnabled() {
		t.Fatal("bridge auto start should be enabled by default")
	}
}

func TestBridgeAutoStartCanBeDisabled(t *testing.T) {
	t.Setenv("BRIDGE_AUTOSTART", "false")
	if bridgeAutoStartEnabled() {
		t.Fatal("bridge auto start should be disabled")
	}
}
