package orchestration

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const minimumBridgeSize = 1024 * 1024

type Runtime struct {
	root         string
	bridgeURL    string
	bridgePath   string
	cancel       context.CancelFunc
	mu           sync.Mutex
	bridge       *managedProcess
	desiredState string
	lastError    string
}

type BridgeStatus struct {
	Status           string `json:"status"`
	DesiredState     string `json:"desiredState"`
	WechatAccount    string `json:"wechatAccount"`
	ResolvedWxid     string `json:"resolvedWxid"`
	AccountDirectory string `json:"accountDirectory"`
	ProcessRunning   bool   `json:"processRunning"`
	ServiceHealthy   bool   `json:"serviceHealthy"`
	DatabaseVerified bool   `json:"databaseVerified"`
	WindowVerified   bool   `json:"windowVerified"`
	WindowPID        int    `json:"windowPid"`
	WindowHandle     int64  `json:"windowHandle"`
	SendReady        bool   `json:"sendReady"`
	LastError        string `json:"lastError"`
	CheckedAt        string `json:"checkedAt"`
}

type bridgeHealthData struct {
	ConfiguredAccount string `json:"configuredAccount"`
	ResolvedWxid      string `json:"resolvedWxid"`
	AccountDirectory  string `json:"accountDirectory"`
	DatabaseVerified  bool   `json:"databaseVerified"`
	WindowVerified    bool   `json:"windowVerified"`
	WindowPID         int    `json:"windowPid"`
	WindowHandle      int64  `json:"windowHandle"`
	SendReady         bool   `json:"sendReady"`
	Ingest            string `json:"ingest"`
}

func Start(parent context.Context) (*Runtime, error) {
	root, err := findProjectRoot()
	if err != nil {
		return nil, err
	}
	if err := loadEnvironment(root); err != nil {
		return nil, err
	}
	if err := ensureRuntimeConfig(root); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(parent)
	runtimeState := &Runtime{
		root:         root,
		bridgeURL:    bridgeHealthURL(),
		cancel:       cancel,
		desiredState: "stopped",
	}
	if !envBool("ORCHESTRATION_ENABLE", true) {
		log.Printf("依赖编排已通过 ORCHESTRATION_ENABLE=false 关闭")
		return runtimeState, nil
	}

	if runtime.GOOS != "windows" {
		cancel()
		return nil, fmt.Errorf("FlyAI 自动编排目前仅支持 Windows")
	}
	if err := ensureHubPortsFree(); err != nil {
		cancel()
		return nil, err
	}
	if err := ensureDocker(ctx, root); err != nil {
		cancel()
		return nil, err
	}
	log.Printf("基础服务已就绪；微信桥接器请在管理页面配置账号后启动")
	return runtimeState, nil
}

func ensureHubPortsFree() error {
	ports := []struct {
		name string
		port int
	}{
		{name: "Web 管理服务", port: envInt("HUB_HTTP_PORT", 9528)},
		{name: "消息接收服务", port: envInt("HUB_INGEST_PORT", 19099)},
	}
	for _, item := range ports {
		address := net.JoinHostPort("127.0.0.1", strconv.Itoa(item.port))
		if portOpen(address, time.Second) {
			return fmt.Errorf("%s端口 %d 已被占用，FlyAI 可能已经运行", item.name, item.port)
		}
	}
	return nil
}

func (r *Runtime) Close() {
	if r == nil {
		return
	}
	r.cancel()
	r.mu.Lock()
	r.desiredState = "stopped"
	bridge := r.bridge
	r.bridge = nil
	r.mu.Unlock()
	if bridge != nil {
		bridge.stop()
	}
}

func (r *Runtime) StartBridge(ctx context.Context, wechatAccount string) (BridgeStatus, error) {
	wechatAccount = strings.TrimSpace(wechatAccount)
	if wechatAccount == "" {
		return r.Status(ctx, wechatAccount), fmt.Errorf("请先配置机器人微信号")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.desiredState = "running"
	r.lastError = ""
	if health, ok := readBridgeHealth(ctx, r.bridgeURL); ok {
		if !strings.EqualFold(health.ConfiguredAccount, wechatAccount) {
			return BridgeStatus{}, fmt.Errorf("桥接器端口已由另一个微信账号占用: %s", health.ConfiguredAccount)
		}
		return r.statusFromHealthLocked(wechatAccount, health, r.bridge != nil && r.bridge.isRunning()), nil
	}

	bridgeHost, bridgePort, err := bridgeAddress()
	if err != nil {
		return BridgeStatus{}, err
	}
	if portOpen(net.JoinHostPort(bridgeHost, strconv.Itoa(bridgePort)), time.Second) {
		return BridgeStatus{}, fmt.Errorf("桥接端口 %d 已被非 FlyAI 健康服务占用", bridgePort)
	}
	if r.bridge != nil {
		r.bridge.stop()
		r.bridge = nil
	}

	bridgeExe, err := prepareBridgeExecutable(ctx, r.root)
	if err != nil {
		r.lastError = err.Error()
		return BridgeStatus{}, err
	}
	r.bridgePath = bridgeExe
	logsDir := filepath.Join(r.root, "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return BridgeStatus{}, fmt.Errorf("创建日志目录失败: %w", err)
	}
	stdout, err := os.OpenFile(filepath.Join(logsDir, "bridge.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return BridgeStatus{}, err
	}
	defer stdout.Close()
	stderr, err := os.OpenFile(filepath.Join(logsDir, "bridge-error.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return BridgeStatus{}, err
	}
	defer stderr.Close()

	ingestPort := envInt("HUB_INGEST_PORT", 19099)
	pollSeconds := envString("BRIDGE_POLL_SECONDS", "2")
	args := []string{
		"--account", wechatAccount,
		"--listen", bridgeHost,
		"--port", strconv.Itoa(bridgePort),
		"--ingest-host", "127.0.0.1",
		"--ingest-port", strconv.Itoa(ingestPort),
		"--monitor-url", monitorURL(),
		"--workdir", filepath.Join(r.root, ".runtime", "bridge-state"),
		"--poll", pollSeconds,
		"--backfill-batch-size", strconv.Itoa(envInt("BRIDGE_BACKFILL_BATCH_SIZE", 1000)),
	}
	if dbDir := strings.TrimSpace(os.Getenv("WECHAT_DB_DIR")); dbDir != "" {
		args = append(args, "--db-dir", dbDir)
	}
	if !envBool("BRIDGE_BACKFILL", true) {
		args = append(args, "--no-backfill")
	}

	process, err := startManagedProcess(bridgeExe, args, r.root, stdout, stderr)
	if err != nil {
		r.lastError = err.Error()
		return BridgeStatus{}, fmt.Errorf("启动微信桥接器失败: %w", err)
	}
	r.bridge = process
	deadline := time.Now().Add(envDurationSeconds("BRIDGE_START_TIMEOUT_SECONDS", 90*time.Second))
	for time.Now().Before(deadline) {
		if health, ok := readBridgeHealth(ctx, r.bridgeURL); ok {
			if !strings.EqualFold(health.ConfiguredAccount, wechatAccount) {
				process.stop()
				r.bridge = nil
				r.lastError = "桥接器启动账号与配置不一致"
				return BridgeStatus{}, errors.New(r.lastError)
			}
			log.Printf("微信桥接器已就绪: %s", r.bridgeURL)
			return r.statusFromHealthLocked(wechatAccount, health, true), nil
		}
		select {
		case <-process.done:
			r.bridge = nil
			r.lastError = fmt.Sprintf("微信桥接器启动后提前退出: %v，请查看 logs/bridge-error.log", process.err())
			return BridgeStatus{}, errors.New(r.lastError)
		case <-ctx.Done():
			process.stop()
			r.bridge = nil
			r.lastError = ctx.Err().Error()
			return BridgeStatus{}, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	process.stop()
	r.bridge = nil
	r.lastError = "微信桥接器未在启动期限内就绪，请查看 logs/bridge-error.log"
	return BridgeStatus{}, errors.New(r.lastError)
}

func (r *Runtime) StopBridge(ctx context.Context, wechatAccount string) BridgeStatus {
	r.mu.Lock()
	r.desiredState = "stopped"
	r.lastError = ""
	bridge := r.bridge
	r.bridge = nil
	r.mu.Unlock()
	if bridge != nil {
		bridge.stop()
	}
	return r.Status(ctx, wechatAccount)
}

func (r *Runtime) Status(ctx context.Context, wechatAccount string) BridgeStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	checkedAt := time.Now().Format("2006-01-02 15:04:05")
	if strings.TrimSpace(wechatAccount) == "" {
		return BridgeStatus{Status: "unconfigured", DesiredState: r.desiredState, CheckedAt: checkedAt}
	}
	processRunning := r.bridge != nil && r.bridge.isRunning()
	if health, ok := readBridgeHealth(ctx, r.bridgeURL); ok {
		return r.statusFromHealthLocked(wechatAccount, health, processRunning)
	}
	status := "stopped"
	lastError := r.lastError
	if r.desiredState == "running" {
		status = "error"
		if lastError == "" {
			if r.bridge != nil && !r.bridge.isRunning() {
				lastError = fmt.Sprintf("桥接器进程已退出: %v", r.bridge.err())
			} else {
				lastError = "桥接器健康检查失败"
			}
		}
	}
	return BridgeStatus{Status: status, DesiredState: r.desiredState, WechatAccount: wechatAccount,
		ProcessRunning: processRunning, LastError: lastError, CheckedAt: checkedAt}
}

func (r *Runtime) statusFromHealthLocked(wechatAccount string, health bridgeHealthData, processRunning bool) BridgeStatus {
	status := "degraded"
	lastError := r.lastError
	if health.SendReady && strings.EqualFold(health.ConfiguredAccount, wechatAccount) {
		status = "running"
		lastError = ""
	} else if !strings.EqualFold(health.ConfiguredAccount, wechatAccount) {
		lastError = fmt.Sprintf("运行中的桥接器绑定了其他账号: %s", health.ConfiguredAccount)
	} else if lastError == "" {
		lastError = "桥接器已启动，但数据库或微信窗口尚未验证完成"
	}
	return BridgeStatus{
		Status: status, DesiredState: r.desiredState, WechatAccount: health.ConfiguredAccount,
		ResolvedWxid: health.ResolvedWxid, AccountDirectory: health.AccountDirectory,
		ProcessRunning: processRunning, ServiceHealthy: true,
		DatabaseVerified: health.DatabaseVerified, WindowVerified: health.WindowVerified,
		WindowPID: health.WindowPID, WindowHandle: health.WindowHandle,
		SendReady: health.SendReady, LastError: lastError,
		CheckedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
}

func ensureDocker(ctx context.Context, root string) error {
	docker, err := exec.LookPath("docker")
	if err != nil {
		return fmt.Errorf("未找到 Docker CLI，请先安装 Docker Desktop: %w", err)
	}
	if err := dockerInfo(ctx, docker); err != nil {
		if !envBool("DOCKER_DESKTOP_AUTO_START", true) {
			return fmt.Errorf("Docker Engine 未运行: %w", err)
		}
		if err := startDockerDesktop(); err != nil {
			return fmt.Errorf("Docker Engine 未运行且无法启动 Docker Desktop: %w", err)
		}
		deadline := time.Now().Add(envDurationSeconds("DOCKER_START_TIMEOUT_SECONDS", 180*time.Second))
		for time.Now().Before(deadline) {
			if dockerInfo(ctx, docker) == nil {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}
		if err := dockerInfo(ctx, docker); err != nil {
			return fmt.Errorf("Docker Desktop 启动超时: %w", err)
		}
	}

	services := []string{"mysql"}
	if envBool("REDIS_ENABLE", true) {
		services = append(services, "redis")
	}
	args := composeArgs(root, append([]string{"up", "-d"}, services...)...)
	if output, err := runCommand(ctx, root, docker, args...); err != nil {
		return fmt.Errorf("启动 Docker 服务失败: %w\n%s", err, output)
	}
	for _, service := range services {
		if err := waitContainerHealthy(ctx, root, docker, service); err != nil {
			return err
		}
	}
	log.Printf("Docker 服务已就绪: %s", strings.Join(services, ", "))
	return nil
}

func dockerInfo(ctx context.Context, docker string) error {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(checkCtx, docker, "info", "--format", "{{.ServerVersion}}")
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

func waitContainerHealthy(ctx context.Context, root, docker, service string) error {
	deadline := time.Now().Add(envDurationSeconds("DOCKER_HEALTH_TIMEOUT_SECONDS", 300*time.Second))
	for time.Now().Before(deadline) {
		output, err := runCommand(ctx, root, docker, composeArgs(root, "ps", "-q", service)...)
		containerID := strings.TrimSpace(output)
		if err == nil && containerID != "" {
			status, inspectErr := runCommand(ctx, root, docker, "inspect", "--format", "{{.State.Health.Status}}", containerID)
			switch strings.TrimSpace(status) {
			case "healthy":
				return nil
			case "unhealthy":
				return fmt.Errorf("Docker 服务 %s 健康检查失败", service)
			}
			if inspectErr != nil && !strings.Contains(status, "no such object") {
				log.Printf("读取 Docker 服务 %s 状态失败: %v", service, inspectErr)
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("Docker 服务 %s 未在启动期限内就绪", service)
}

func composeArgs(root string, args ...string) []string {
	prefix := []string{"compose", "--project-name", "flyai", "--env-file", filepath.Join(root, ".env"), "-f", filepath.Join(root, "compose.yaml")}
	return append(prefix, args...)
}

func runCommand(ctx context.Context, dir, executable string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func bridgeAddress() (string, int, error) {
	host := strings.TrimSpace(os.Getenv("WECHAT_HOST"))
	if host == "" {
		host = "127.0.0.1:19089"
	}
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	parsed, err := url.Parse(host)
	if err != nil {
		return "", 0, fmt.Errorf("WECHAT_HOST 无效: %w", err)
	}
	name := parsed.Hostname()
	if name == "" {
		name = "127.0.0.1"
	}
	port := envInt("BRIDGE_HTTP_PORT", 19089)
	if parsed.Port() != "" {
		parsedPort, parseErr := strconv.Atoi(parsed.Port())
		if parseErr != nil {
			return "", 0, fmt.Errorf("WECHAT_HOST 端口无效: %w", parseErr)
		}
		port = parsedPort
	}
	return name, port, nil
}

func bridgeHealthURL() string {
	host := strings.TrimSpace(os.Getenv("WECHAT_HOST"))
	if host == "" {
		host = "127.0.0.1:19089"
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	return strings.TrimRight(host, "/") + "/health"
}

func monitorURL() string {
	domain := envString("SYSTEM_DOMAIN", "http://127.0.0.1:9528")
	return strings.TrimRight(domain, "/") + "/api/monitor/groups"
}

func bridgeHealthy(endpoint string) bool {
	_, ok := readBridgeHealth(context.Background(), endpoint)
	return ok
}

func readBridgeHealth(ctx context.Context, endpoint string) (bridgeHealthData, bool) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return bridgeHealthData{}, false
	}
	client := &http.Client{Timeout: 3 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return bridgeHealthData{}, false
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return bridgeHealthData{}, false
	}
	var payload struct {
		Code int              `json:"code"`
		Data bridgeHealthData `json:"data"`
	}
	if json.NewDecoder(io.LimitReader(response.Body, 1024*1024)).Decode(&payload) != nil || payload.Code != 1 || payload.Data.Ingest == "" {
		return bridgeHealthData{}, false
	}
	return payload.Data, true
}

func portOpen(address string, timeout time.Duration) bool {
	connection, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}

func prepareBridgeExecutable(ctx context.Context, root string) (string, error) {
	runtimePath := filepath.Join(root, ".runtime", "bin", "FlyAI-Bridge.exe")
	candidates := []string{
		strings.TrimSpace(os.Getenv("BRIDGE_EXECUTABLE_PATH")),
		filepath.Join(root, "FlyAI-Bridge.exe"),
		filepath.Join(filepath.Dir(root), "FlyAI-Bridge", "dist", "FlyAI-Bridge.exe"),
		runtimePath,
	}
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if !filepath.IsAbs(candidate) {
			candidate = filepath.Join(root, candidate)
		}
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() || info.Size() < minimumBridgeSize {
			continue
		}
		if samePath(candidate, runtimePath) {
			return runtimePath, nil
		}
		if err := copyIfChanged(candidate, runtimePath); err != nil {
			return "", fmt.Errorf("准备本地桥接器失败: %w", err)
		}
		return runtimePath, nil
	}

	downloadURL := strings.TrimSpace(os.Getenv("BRIDGE_DOWNLOAD_URL"))
	if downloadURL == "" {
		return "", fmt.Errorf("未找到 FlyAI-Bridge.exe；请设置 BRIDGE_EXECUTABLE_PATH，或把 EXE 放到项目根目录")
	}
	if err := downloadFile(ctx, downloadURL, runtimePath); err != nil {
		return "", fmt.Errorf("下载桥接器失败: %w", err)
	}
	return runtimePath, nil
}

func copyIfChanged(source, destination string) error {
	sourceHash, err := fileSHA256(source)
	if err != nil {
		return err
	}
	if destinationHash, hashErr := fileSHA256(destination); hashErr == nil && sourceHash == destinationHash {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	temporary := destination + ".new"
	if err := copyFile(source, temporary); err != nil {
		return err
	}
	_ = os.Remove(destination)
	if err := os.Rename(temporary, destination); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}

func downloadFile(ctx context.Context, sourceURL, destination string) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("HTTP %s", response.Status)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	temporary := destination + ".download"
	file, err := os.OpenFile(temporary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(file, io.LimitReader(response.Body, 512*1024*1024))
	closeErr := file.Close()
	if copyErr != nil {
		_ = os.Remove(temporary)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(temporary)
		return closeErr
	}
	if written < minimumBridgeSize {
		_ = os.Remove(temporary)
		return fmt.Errorf("下载结果过小: %d bytes", written)
	}
	if expected := strings.ToLower(strings.TrimSpace(os.Getenv("BRIDGE_SHA256"))); expected != "" {
		actual, hashErr := fileSHA256(temporary)
		if hashErr != nil || actual != expected {
			_ = os.Remove(temporary)
			return fmt.Errorf("SHA-256 校验失败")
		}
	}
	_ = os.Remove(destination)
	return os.Rename(temporary, destination)
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	return errors.Join(copyErr, closeErr)
}

func samePath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	return leftErr == nil && rightErr == nil && strings.EqualFold(filepath.Clean(leftAbs), filepath.Clean(rightAbs))
}

func findProjectRoot() (string, error) {
	var candidates []string
	if configured := strings.TrimSpace(os.Getenv("FLYAI_ROOT")); configured != "" {
		candidates = append(candidates, configured)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}
	if executable, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Dir(executable))
	}
	seen := map[string]bool{}
	for _, candidate := range candidates {
		current, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		for depth := 0; depth < 5; depth++ {
			if seen[current] {
				break
			}
			seen[current] = true
			if fileExists(filepath.Join(current, "compose.yaml")) && fileExists(filepath.Join(current, ".env.example")) {
				return current, nil
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}
	return "", fmt.Errorf("未找到 FlyAI 项目根目录（需要 compose.yaml 和 .env.example）")
}

func loadEnvironment(root string) error {
	examplePath := filepath.Join(root, ".env.example")
	envPath := filepath.Join(root, ".env")
	if !fileExists(envPath) {
		if err := copyFile(examplePath, envPath); err != nil {
			return fmt.Errorf("创建 .env 失败: %w", err)
		}
		log.Printf("已根据 .env.example 创建 .env")
	}
	file, err := os.Open(envPath)
	if err != nil {
		return fmt.Errorf("读取 .env 失败: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		separator := strings.IndexByte(line, '=')
		if separator < 1 {
			continue
		}
		key := strings.TrimSpace(line[:separator])
		value := strings.TrimSpace(line[separator+1:])
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func ensureRuntimeConfig(root string) error {
	configPath := filepath.Join(root, "config.yaml")
	if fileExists(configPath) {
		return nil
	}
	content := renderInitialConfig()
	temporary := configPath + ".new"
	if err := os.WriteFile(temporary, []byte(content), 0o600); err != nil {
		return fmt.Errorf("生成 config.yaml 失败: %w", err)
	}
	if err := os.Rename(temporary, configPath); err != nil {
		_ = os.Remove(temporary)
		return fmt.Errorf("生成 config.yaml 失败: %w", err)
	}
	log.Printf("已根据 .env 生成 config.yaml")
	return nil
}

func renderInitialConfig() string {
	return fmt.Sprintf(`system:
  alApiToken: %s
  domain: %s
  newFriendNotify:
    enable: %t
    toUser:
%s
  defaultRule:
    ai: %t
    chatRank: %t
    summary: %t
    welcome: %t
    news: %t
    hotTop: %t
    aiFreeLimit: %d

wechat:
  host: %s
  forward:
%s

mysql:
  host: %s
  port: %d
  user: %s
  password: %s
  db: %s

redis:
  enable: %t
  host: %s
  port: %d
  password: %s
  db: %d

task:
  enable: %t
  news:
    enable: %t
    cron: %s
  hotTop:
    enable: %t
    cron: %s
  syncFriends:
    enable: %t
    cron: %s
  groupSummary:
    enable: %t
    cron: %s
mq:
  enable: %t
  rabbitmq:
    host: %s
    port: %d
    user: %s
    password: %s
    vhost: %s

ai:
  enable: %t
  model: %s
  summaryModel: %s
  apiKey: %s
  baseUrl: %s
  personality: %s
  models:
    - name: "默认"
      model: %s
      canManager: true

resource:
  introduce:
    type: "text"
    path: %s
  welcome-new:
    type: "emotion"
    path: %s
  wordcloud:
    type: "image"
    path: %s
  invitation-join-group:
    type: "text"
    path: %s
`,
		yamlString(envString("ALAPI_TOKEN", "xxx")),
		yamlString(envString("SYSTEM_DOMAIN", "http://127.0.0.1:9528")),
		envBool("NEW_FRIEND_NOTIFY_ENABLE", false), yamlList(os.Getenv("NEW_FRIEND_NOTIFY_USERS"), 6),
		envBool("DEFAULT_AI", false), envBool("DEFAULT_CHAT_RANK", true), envBool("DEFAULT_SUMMARY", false),
		envBool("DEFAULT_WELCOME", true), envBool("DEFAULT_NEWS", true), envBool("DEFAULT_HOT_TOP", false), envInt("DEFAULT_AI_FREE_LIMIT", 50),
		yamlString(envString("WECHAT_HOST", "127.0.0.1:19089")), yamlList(os.Getenv("WECHAT_FORWARD"), 4),
		yamlString(envString("MYSQL_HOST", "127.0.0.1")), envInt("MYSQL_PORT", 3307), yamlString(envString("MYSQL_USER", "wechat")), yamlString(envString("MYSQL_PASSWORD", "wechat")), yamlString(envString("MYSQL_DATABASE", "wechat")),
		envBool("REDIS_ENABLE", true), yamlString(envString("REDIS_HOST", "127.0.0.1")), envInt("REDIS_PORT", 6379), yamlString(envString("REDIS_PASSWORD", "flyai-redis")), envInt("REDIS_DB", 0),
		envBool("TASK_ENABLE", true), envBool("TASK_NEWS_ENABLE", false), yamlString(envString("TASK_NEWS_CRON", "0 8 * * *")),
		envBool("TASK_HOT_TOP_ENABLE", false), yamlString(envString("TASK_HOT_TOP_CRON", "0 */1 * * *")), envBool("TASK_SYNC_FRIENDS_ENABLE", true), yamlString(envString("TASK_SYNC_FRIENDS_CRON", "*/1 * * * *")),
		envBool("TASK_GROUP_SUMMARY_ENABLE", false), yamlString(envString("TASK_GROUP_SUMMARY_CRON", "30 0 * * *")),
		envBool("MQ_ENABLE", false), yamlString(envString("MQ_HOST", "127.0.0.1")), envInt("MQ_PORT", 5672), yamlString(envString("MQ_USER", "wechat")), yamlString(envString("MQ_PASSWORD", "wechat")), yamlString(envString("MQ_VHOST", "wechat")),
		envBool("AI_ENABLE", false), yamlString(envString("AI_MODEL", "gpt-4o-mini")), yamlString(envString("AI_SUMMARY_MODEL", "gpt-4o-mini")), yamlString(envString("AI_API_KEY", "sk-xxxx")), yamlString(envString("AI_BASE_URL", "https://api.openai.com")), yamlString(envString("AI_PERSONALITY", "你是一个友好、简洁且可靠的群聊助手。")), yamlString(envString("AI_MODEL", "gpt-4o-mini")),
		yamlString(envString("INTRODUCE_TEXT", "大家好，我是一个群聊自动化助手，可以直接@我询问问题。")), yamlString(os.Getenv("WELCOME_EMOTION_HASH")), yamlString(os.Getenv("WORDCLOUD_PATH")), yamlString(envString("INVITATION_REPLY", "您的邀请消息已收到，管理员稍后会处理。")),
	)
}

func yamlString(value string) string { return strconv.QuoteToGraphic(value) }

func yamlList(value string, indent int) string {
	padding := strings.Repeat(" ", indent)
	var lines []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			lines = append(lines, padding+"- "+yamlString(item))
		}
	}
	if len(lines) == 0 {
		return padding + "[]"
	}
	return strings.Join(lines, "\n")
}

func envString(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv(name)))
	if err != nil {
		return fallback
	}
	return value
}

func envBool(name string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(name)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func envDurationSeconds(name string, fallback time.Duration) time.Duration {
	seconds := envInt(name, int(fallback/time.Second))
	if seconds < 1 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
