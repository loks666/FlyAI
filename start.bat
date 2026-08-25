@echo off
setlocal
chcp 65001 >nul

set "FLYAI_LAUNCHER=%~f0"
set "FLYAI_ROOT=%~dp0"
set "FLYAI_RUNTIME=%~dp0.runtime"
set "FLYAI_PS=%~dp0.runtime\launcher.ps1"

if not exist "%FLYAI_RUNTIME%" mkdir "%FLYAI_RUNTIME%"

powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -Command ^
  "$lines = Get-Content -LiteralPath $env:FLYAI_LAUNCHER -Encoding UTF8; $marker = [Array]::IndexOf($lines, '# __POWERSHELL_PAYLOAD__'); if ($marker -lt 0) { exit 2 }; $lines[($marker + 1)..($lines.Length - 1)] | Set-Content -LiteralPath $env:FLYAI_PS -Encoding UTF8"
if errorlevel 1 (
  echo [ERROR] 无法提取启动器，请确认 PowerShell 可用。
  exit /b 1
)

powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%FLYAI_PS%" %*
set "FLYAI_EXIT=%ERRORLEVEL%"
exit /b %FLYAI_EXIT%

# __POWERSHELL_PAYLOAD__
param(
    [ValidateSet("start", "stop", "restart", "status", "logs", "update")]
    [string]$Action = "start"
)

$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$Root = [IO.Path]::GetFullPath($env:FLYAI_ROOT).TrimEnd('\')
$RuntimeDir = Join-Path $Root ".runtime"
$DownloadDir = Join-Path $RuntimeDir "downloads"
$BinDir = Join-Path $RuntimeDir "bin"
$PidDir = Join-Path $RuntimeDir "pids"
$LogDir = Join-Path $Root "logs"
$EnvFile = Join-Path $Root ".env"
$EnvExample = Join-Path $Root ".env.example"
$ComposeFile = Join-Path $Root "compose.yaml"
$ConfigFile = Join-Path $Root "config.yaml"
$BridgePidFile = Join-Path $PidDir "bridge.pid"
$HubPidFile = Join-Path $PidDir "hub.pid"

foreach ($dir in @($RuntimeDir, $DownloadDir, $BinDir, $PidDir, $LogDir)) {
    New-Item -ItemType Directory -Force -Path $dir | Out-Null
}

function Write-Step([string]$Message) {
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Write-Ok([string]$Message) {
    Write-Host "[OK] $Message" -ForegroundColor Green
}

function Write-Warn([string]$Message) {
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Read-DotEnv([string]$Path) {
    $values = New-Object 'System.Collections.Generic.Dictionary[string,string]' ([StringComparer]::OrdinalIgnoreCase)
    foreach ($line in Get-Content -LiteralPath $Path -Encoding UTF8) {
        $trimmed = $line.Trim()
        if (!$trimmed -or $trimmed.StartsWith("#")) { continue }
        $separator = $trimmed.IndexOf("=")
        if ($separator -lt 1) { continue }
        $name = $trimmed.Substring(0, $separator).Trim()
        $value = $trimmed.Substring($separator + 1).Trim()
        if ($value.Length -ge 2 -and (($value.StartsWith('"') -and $value.EndsWith('"')) -or ($value.StartsWith("'") -and $value.EndsWith("'")))) {
            $value = $value.Substring(1, $value.Length - 2)
        }
        $values[$name] = $value
    }
    return $values
}

function Get-Setting([string]$Name, [string]$Default = "") {
    if ($script:Settings.ContainsKey($Name)) { return $script:Settings[$Name] }
    return $Default
}

function Get-BoolSetting([string]$Name, [bool]$Default = $false) {
    $fallback = if ($Default) { "true" } else { "false" }
    $value = (Get-Setting $Name $fallback).Trim().ToLowerInvariant()
    return $value -in @("1", "true", "yes", "on")
}

function ConvertTo-YamlScalar([string]$Value) {
    return "'" + $Value.Replace("'", "''") + "'"
}

function ConvertTo-YamlList([string]$Value, [int]$Indent) {
    $padding = " " * $Indent
    $items = @($Value.Split(",", [StringSplitOptions]::RemoveEmptyEntries) | ForEach-Object { $_.Trim() } | Where-Object { $_ })
    if ($items.Count -eq 0) { return $padding + "[]" }
    return ($items | ForEach-Object { $padding + "- " + (ConvertTo-YamlScalar $_) }) -join "`r`n"
}

function Initialize-Environment {
    if (!(Test-Path -LiteralPath $EnvExample)) {
        throw "缺少 .env.example，仓库文件不完整。"
    }
    if (!(Test-Path -LiteralPath $EnvFile)) {
        Copy-Item -LiteralPath $EnvExample -Destination $EnvFile
        Write-Warn "已创建 .env。首次运行可先关闭窗口并编辑该文件；本次将使用默认本机配置。"
    }
    $script:Settings = Read-DotEnv $EnvFile
    foreach ($entry in $script:Settings.GetEnumerator()) {
        [Environment]::SetEnvironmentVariable($entry.Key, $entry.Value, "Process")
    }
}

function Write-RuntimeConfig {
    $notifyUsers = ConvertTo-YamlList (Get-Setting "NEW_FRIEND_NOTIFY_USERS") 6
    $forwardTargets = ConvertTo-YamlList (Get-Setting "WECHAT_FORWARD") 4
    $welcomeHash = Get-Setting "WELCOME_EMOTION_HASH"
    $wordcloudPath = Get-Setting "WORDCLOUD_PATH"

    $config = @"
system:
  alApiToken: $(ConvertTo-YamlScalar (Get-Setting "ALAPI_TOKEN" "xxx"))
  domain: $(ConvertTo-YamlScalar (Get-Setting "SYSTEM_DOMAIN" "http://127.0.0.1:8080"))
  newFriendNotify:
    enable: $((Get-BoolSetting "NEW_FRIEND_NOTIFY_ENABLE").ToString().ToLowerInvariant())
    toUser:
$notifyUsers
  defaultRule:
    ai: $((Get-BoolSetting "DEFAULT_AI").ToString().ToLowerInvariant())
    chatRank: $((Get-BoolSetting "DEFAULT_CHAT_RANK" $true).ToString().ToLowerInvariant())
    summary: $((Get-BoolSetting "DEFAULT_SUMMARY").ToString().ToLowerInvariant())
    welcome: $((Get-BoolSetting "DEFAULT_WELCOME" $true).ToString().ToLowerInvariant())
    news: $((Get-BoolSetting "DEFAULT_NEWS" $true).ToString().ToLowerInvariant())
    hotTop: $((Get-BoolSetting "DEFAULT_HOT_TOP").ToString().ToLowerInvariant())
    aiFreeLimit: $(Get-Setting "DEFAULT_AI_FREE_LIMIT" "50")

wechat:
  host: $(ConvertTo-YamlScalar (Get-Setting "WECHAT_HOST" "127.0.0.1:19089"))
  forward:
$forwardTargets

mysql:
  host: $(ConvertTo-YamlScalar (Get-Setting "MYSQL_HOST" "127.0.0.1"))
  port: $(Get-Setting "MYSQL_PORT" "3307")
  user: $(ConvertTo-YamlScalar (Get-Setting "MYSQL_USER" "wechat"))
  password: $(ConvertTo-YamlScalar (Get-Setting "MYSQL_PASSWORD" "wechat"))
  db: $(ConvertTo-YamlScalar (Get-Setting "MYSQL_DATABASE" "wechat"))

task:
  enable: $((Get-BoolSetting "TASK_ENABLE" $true).ToString().ToLowerInvariant())
  news:
    enable: $((Get-BoolSetting "TASK_NEWS_ENABLE").ToString().ToLowerInvariant())
    cron: $(ConvertTo-YamlScalar (Get-Setting "TASK_NEWS_CRON" "0 8 * * *"))
  hotTop:
    enable: $((Get-BoolSetting "TASK_HOT_TOP_ENABLE").ToString().ToLowerInvariant())
    cron: $(ConvertTo-YamlScalar (Get-Setting "TASK_HOT_TOP_CRON" "0 */1 * * *"))
  syncFriends:
    enable: $((Get-BoolSetting "TASK_SYNC_FRIENDS_ENABLE" $true).ToString().ToLowerInvariant())
    cron: $(ConvertTo-YamlScalar (Get-Setting "TASK_SYNC_FRIENDS_CRON" "*/1 * * * *"))
  groupSummary:
    enable: $((Get-BoolSetting "TASK_GROUP_SUMMARY_ENABLE").ToString().ToLowerInvariant())
    cron: $(ConvertTo-YamlScalar (Get-Setting "TASK_GROUP_SUMMARY_CRON" "30 0 * * *"))
  waterGroup:
    enable: $((Get-BoolSetting "TASK_WATER_GROUP_ENABLE").ToString().ToLowerInvariant())
    cron:
      yesterday: $(ConvertTo-YamlScalar (Get-Setting "TASK_WATER_GROUP_YESTERDAY_CRON" "30 9 * * *"))
      week: $(ConvertTo-YamlScalar (Get-Setting "TASK_WATER_GROUP_WEEK_CRON" "30 9 * * 1"))
      month: $(ConvertTo-YamlScalar (Get-Setting "TASK_WATER_GROUP_MONTH_CRON" "30 9 1 * *"))
      year: $(ConvertTo-YamlScalar (Get-Setting "TASK_WATER_GROUP_YEAR_CRON" "0 9 1 1 *"))

mq:
  enable: $((Get-BoolSetting "MQ_ENABLE").ToString().ToLowerInvariant())
  rabbitmq:
    host: $(ConvertTo-YamlScalar (Get-Setting "MQ_HOST" "127.0.0.1"))
    port: $(Get-Setting "MQ_PORT" "5672")
    user: $(ConvertTo-YamlScalar (Get-Setting "MQ_USER" "wechat"))
    password: $(ConvertTo-YamlScalar (Get-Setting "MQ_PASSWORD" "wechat"))
    vhost: $(ConvertTo-YamlScalar (Get-Setting "MQ_VHOST" "wechat"))

ai:
  enable: $((Get-BoolSetting "AI_ENABLE").ToString().ToLowerInvariant())
  model: $(ConvertTo-YamlScalar (Get-Setting "AI_MODEL" "gpt-4o-mini"))
  summaryModel: $(ConvertTo-YamlScalar (Get-Setting "AI_SUMMARY_MODEL" "gpt-4o-mini"))
  apiKey: $(ConvertTo-YamlScalar (Get-Setting "AI_API_KEY" "sk-xxxx"))
  baseUrl: $(ConvertTo-YamlScalar (Get-Setting "AI_BASE_URL" "https://api.openai.com"))
  personality: $(ConvertTo-YamlScalar (Get-Setting "AI_PERSONALITY" "你是一个友好、简洁且可靠的群聊助手。"))
  models:
    - name: '默认'
      model: $(ConvertTo-YamlScalar (Get-Setting "AI_MODEL" "gpt-4o-mini"))
      canManager: true

resource:
  introduce:
    type: 'text'
    path: $(ConvertTo-YamlScalar (Get-Setting "INTRODUCE_TEXT" "大家好，我是一个群聊自动化助手，可以直接@我询问问题。"))
  welcome-new:
    type: 'emotion'
    path: $(ConvertTo-YamlScalar $welcomeHash)
  wordcloud:
    type: 'image'
    path: $(ConvertTo-YamlScalar $wordcloudPath)
  invitation-join-group:
    type: 'text'
    path: $(ConvertTo-YamlScalar (Get-Setting "INVITATION_REPLY" "您的邀请消息已收到，管理员稍后会处理。"))
"@
    Set-Content -LiteralPath $ConfigFile -Value $config -Encoding UTF8
}

function Get-DockerExecutable {
    $command = Get-Command docker -ErrorAction SilentlyContinue
    if ($command) { return $command.Source }
    $candidates = @(
        (Join-Path $env:ProgramFiles "Docker\Docker\resources\bin\docker.exe"),
        (Join-Path ${env:ProgramFiles(x86)} "Docker\Docker\resources\bin\docker.exe")
    ) | Where-Object { $_ -and (Test-Path -LiteralPath $_) }
    if ($candidates.Count -gt 0) { return $candidates[0] }
    throw "未找到 Docker Desktop。请先安装：https://www.docker.com/products/docker-desktop/"
}

function Wait-Docker([string]$DockerExe) {
    try {
        & $DockerExe info *> $null
        if ($LASTEXITCODE -eq 0) { return }
    } catch {}

    $desktop = Join-Path $env:ProgramFiles "Docker\Docker\Docker Desktop.exe"
    if (Test-Path -LiteralPath $desktop) {
        Write-Step "正在启动 Docker Desktop"
        Start-Process -FilePath $desktop -WindowStyle Hidden | Out-Null
    }
    $deadline = [DateTime]::UtcNow.AddMinutes(3)
    do {
        Start-Sleep -Seconds 3
        try {
            & $DockerExe info *> $null
            if ($LASTEXITCODE -eq 0) { return }
        } catch {}
    } while ([DateTime]::UtcNow -lt $deadline)
    throw "Docker Desktop 未在 3 分钟内就绪。请打开 Docker Desktop 查看状态。"
}

function Invoke-Compose([string[]]$Arguments) {
    & $script:DockerExe compose --project-name flyai --env-file $EnvFile -f $ComposeFile @Arguments
    if ($LASTEXITCODE -ne 0) { throw "docker compose 执行失败。" }
}

function Wait-MySQL {
    $deadline = [DateTime]::UtcNow.AddMinutes(3)
    do {
        $containerId = ([string](& $script:DockerExe compose --project-name flyai --env-file $EnvFile -f $ComposeFile ps -q mysql)).Trim()
        if ($containerId) {
            $health = ([string](& $script:DockerExe inspect --format "{{.State.Health.Status}}" $containerId 2>$null)).Trim()
            if ($health -eq "healthy") { return }
            if ($health -eq "unhealthy") { throw "MySQL 容器健康检查失败，请运行 start.bat logs 查看。" }
        }
        Start-Sleep -Seconds 3
    } while ([DateTime]::UtcNow -lt $deadline)
    throw "MySQL 未在 3 分钟内就绪。"
}

function Get-ArchitectureName {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { return "arm64" }
    return "amd64"
}

function Get-GoExecutable {
    $localGo = Join-Path $RuntimeDir "go\bin\go.exe"
    if (Test-Path -LiteralPath $localGo) { return $localGo }
    $systemGo = Get-Command go -ErrorAction SilentlyContinue
    if ($systemGo) { return $systemGo.Source }

    $version = Get-Setting "GO_VERSION" "1.23.12"
    $arch = Get-ArchitectureName
    $archive = Join-Path $DownloadDir "go$version.windows-$arch.zip"
    $url = "https://go.dev/dl/go$version.windows-$arch.zip"
    Write-Step "首次运行：下载便携式 Go $version"
    Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $archive
    Expand-Archive -LiteralPath $archive -DestinationPath $RuntimeDir -Force
    if (!(Test-Path -LiteralPath $localGo)) { throw "Go 解压后未找到 go.exe。" }
    return $localGo
}

function Get-UvExecutable {
    $uvDir = Join-Path $RuntimeDir "uv"
    $uvExe = Join-Path $uvDir "uv.exe"
    if (Test-Path -LiteralPath $uvExe) { return $uvExe }
    New-Item -ItemType Directory -Force -Path $uvDir | Out-Null
    $uvArch = if ((Get-ArchitectureName) -eq "arm64") { "aarch64" } else { "x86_64" }
    $archive = Join-Path $DownloadDir "uv-windows.zip"
    $url = "https://github.com/astral-sh/uv/releases/latest/download/uv-$uvArch-pc-windows-msvc.zip"
    Write-Step "首次运行：下载便携式 Python 管理器"
    Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $archive
    Expand-Archive -LiteralPath $archive -DestinationPath $uvDir -Force
    if (!(Test-Path -LiteralPath $uvExe)) { throw "uv 解压后未找到 uv.exe。" }
    return $uvExe
}

function Initialize-Python {
    $venv = Join-Path $RuntimeDir "venv"
    $python = Join-Path $venv "Scripts\python.exe"
    $uv = Get-UvExecutable
    $version = Get-Setting "PYTHON_VERSION" "3.12"
    if (!(Test-Path -LiteralPath $python)) {
        Write-Step "首次运行：准备 Python $version"
        & $uv python install $version
        if ($LASTEXITCODE -ne 0) { throw "Python 下载失败。" }
        & $uv venv --python $version $venv
        if ($LASTEXITCODE -ne 0) { throw "Python 虚拟环境创建失败。" }
    }
    $requirements = Join-Path $Root "bridge\requirements.txt"
    $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $requirements).Hash
    $stamp = Join-Path $RuntimeDir "requirements.sha256"
    $installedHash = if (Test-Path -LiteralPath $stamp) { (Get-Content -LiteralPath $stamp -Raw).Trim() } else { "" }
    if ($hash -ne $installedHash) {
        Write-Step "安装微信桥接器依赖"
        & $uv pip install --python $python -r $requirements
        if ($LASTEXITCODE -ne 0) { throw "Python 依赖安装失败。" }
        Set-Content -LiteralPath $stamp -Value $hash -Encoding ASCII
    }
    return $python
}

function Build-Hub([string]$GoExe, [bool]$Force) {
    $hubExe = Join-Path $BinDir "FlyAI.exe"
    if (!$Force -and (Test-Path -LiteralPath $hubExe)) {
        $latestSource = Get-ChildItem -LiteralPath $Root -Recurse -File -Include "*.go", "go.mod", "go.sum" |
            Where-Object { $_.FullName -notlike "$RuntimeDir*" -and $_.FullName -notlike "$(Join-Path $Root '.git')*" } |
            Sort-Object LastWriteTimeUtc -Descending |
            Select-Object -First 1
        if (!$latestSource -or $latestSource.LastWriteTimeUtc -le (Get-Item -LiteralPath $hubExe).LastWriteTimeUtc) {
            Write-Ok "Go 服务无需重新构建"
            return $hubExe
        }
    }
    Write-Step "构建 Go 服务"
    Push-Location $Root
    try {
        $env:CGO_ENABLED = "0"
        & $GoExe build -trimpath -o $hubExe .
        if ($LASTEXITCODE -ne 0) { throw "Go 服务构建失败。" }
    } finally {
        Pop-Location
    }
    return $hubExe
}

function Test-Http([string]$Url) {
    try {
        $response = Invoke-WebRequest -UseBasicParsing -Uri $Url -TimeoutSec 2
        return $response.StatusCode -ge 200 -and $response.StatusCode -lt 500
    } catch { return $false }
}

function Test-Tcp([string]$HostName, [int]$Port) {
    $client = New-Object Net.Sockets.TcpClient
    try {
        $async = $client.BeginConnect($HostName, $Port, $null, $null)
        if (!$async.AsyncWaitHandle.WaitOne(1000, $false)) { return $false }
        $client.EndConnect($async)
        return $true
    } catch { return $false } finally { $client.Dispose() }
}

function Wait-Condition([scriptblock]$Condition, [int]$TimeoutSeconds, [string]$FailureMessage) {
    $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
    do {
        if (& $Condition) { return }
        Start-Sleep -Seconds 1
    } while ([DateTime]::UtcNow -lt $deadline)
    throw $FailureMessage
}

function Quote-ProcessArgument([string]$Value) {
    return '"' + $Value.Replace('"', '\"') + '"'
}

function Save-ManagedProcess([string]$PidFile, [Diagnostics.Process]$Process, [string]$Signature) {
    [pscustomobject]@{
        pid       = $Process.Id
        signature = $Signature
        startedAt = $Process.StartTime.ToUniversalTime().ToString("o")
    } | ConvertTo-Json | Set-Content -LiteralPath $PidFile -Encoding UTF8
}

function Start-Bridge([string]$PythonExe) {
    $bridgePort = [int](Get-Setting "BRIDGE_HTTP_PORT" "19089")
    if (Test-Http "http://127.0.0.1:$bridgePort/health") {
        Write-Ok "微信桥接器已在运行"
        return
    }
    $bridgeScript = Join-Path $Root "bridge\wechatauto_bridge.py"
    $workdir = Join-Path $RuntimeDir "bridge-state"
    $arguments = @(
        (Quote-ProcessArgument $bridgeScript),
        "--listen", "127.0.0.1",
        "--port", $bridgePort,
        "--ingest-host", "127.0.0.1",
        "--ingest-port", (Get-Setting "HUB_INGEST_PORT" "19099"),
        "--monitor-url", (Quote-ProcessArgument ("http://127.0.0.1:" + (Get-Setting "HUB_HTTP_PORT" "8080") + "/api/monitor/groups")),
        "--workdir", (Quote-ProcessArgument $workdir),
        "--poll", (Get-Setting "BRIDGE_POLL_SECONDS" "2")
    )
    $dbDir = Get-Setting "WECHAT_DB_DIR"
    if ($dbDir) { $arguments += @("--db-dir", (Quote-ProcessArgument $dbDir)) }
    if (!(Get-BoolSetting "BRIDGE_BACKFILL" $true)) { $arguments += "--no-backfill" }

    Write-Step "启动微信本地桥接器"
    $process = Start-Process -FilePath $PythonExe -ArgumentList ($arguments -join " ") -WorkingDirectory $Root -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $LogDir "bridge.log") -RedirectStandardError (Join-Path $LogDir "bridge-error.log")
    Save-ManagedProcess $BridgePidFile $process $bridgeScript
    Wait-Condition { Test-Http "http://127.0.0.1:$bridgePort/health" } 45 "微信桥接器启动失败，请查看 logs\bridge-error.log。"
    Write-Ok "微信桥接器已启动"
}

function Start-Hub([string]$HubExe) {
    $httpPort = [int](Get-Setting "HUB_HTTP_PORT" "8080")
    if (Test-Tcp "127.0.0.1" $httpPort) {
        Write-Ok "Go 服务已在运行"
        return
    }
    Write-Step "启动 Go 服务"
    $process = Start-Process -FilePath $HubExe -WorkingDirectory $Root -WindowStyle Hidden -PassThru -RedirectStandardOutput (Join-Path $LogDir "hub.log") -RedirectStandardError (Join-Path $LogDir "hub-error.log")
    Save-ManagedProcess $HubPidFile $process $HubExe
    Wait-Condition { Test-Tcp "127.0.0.1" $httpPort } 45 "Go 服务启动失败，请查看 logs\hub-error.log。"
    Write-Ok "Go 服务已启动"
}

function Stop-ManagedProcess([string]$PidFile, [string]$Label) {
    if (!(Test-Path -LiteralPath $PidFile)) { return }
    try {
        $record = Get-Content -LiteralPath $PidFile -Raw -Encoding UTF8 | ConvertFrom-Json
        $pidValue = [int]$record.pid
        $process = Get-Process -Id $pidValue -ErrorAction SilentlyContinue
        if (!$process) { return }
        $processInfo = Get-CimInstance Win32_Process -Filter "ProcessId = $pidValue" -ErrorAction SilentlyContinue
        $commandLine = [string]$processInfo.CommandLine
        $signature = [string]$record.signature
        if (!$signature -or $commandLine.IndexOf($signature, [StringComparison]::OrdinalIgnoreCase) -lt 0) {
            Write-Warn "$Label 的 PID 已被其他进程占用，为避免误杀已跳过停止。"
            return
        }
        Stop-Process -Id $pidValue -Force
        Write-Ok "$Label 已停止"
    } catch {
        Write-Warn "读取 $Label 的进程记录失败，为避免误杀已跳过停止。"
    } finally {
        Remove-Item -LiteralPath $PidFile -Force -ErrorAction SilentlyContinue
    }
}

function Stop-All {
    Write-Step "停止本地服务"
    Stop-ManagedProcess $HubPidFile "Go 服务"
    Stop-ManagedProcess $BridgePidFile "微信桥接器"
    if ((Test-Path -LiteralPath $EnvFile) -and (Test-Path -LiteralPath $ComposeFile)) {
        try {
            $script:DockerExe = Get-DockerExecutable
            & $script:DockerExe compose --project-name flyai --env-file $EnvFile -f $ComposeFile stop *> $null
            Write-Ok "MySQL 容器已停止"
        } catch {
            Write-Warn "未能停止 MySQL：$($_.Exception.Message)"
        }
    }
}

function Show-Status {
    Initialize-Environment
    $bridgePort = [int](Get-Setting "BRIDGE_HTTP_PORT" "19089")
    $httpPort = [int](Get-Setting "HUB_HTTP_PORT" "8080")
    $mysqlPort = [int](Get-Setting "MYSQL_PORT" "3307")
    Write-Host "微信桥接器 : $(if (Test-Http "http://127.0.0.1:$bridgePort/health") { '运行中' } else { '未运行' })"
    Write-Host "Go 管理服务 : $(if (Test-Tcp '127.0.0.1' $httpPort) { '运行中' } else { '未运行' })"
    Write-Host "MySQL       : $(if (Test-Tcp '127.0.0.1' $mysqlPort) { '端口可访问' } else { '未运行' })"
    Write-Host "管理页面    : http://127.0.0.1:$httpPort"
}

function Assert-WeChat {
    $wechat = Get-Process -Name "Weixin", "WeChat" -ErrorAction SilentlyContinue
    if (!$wechat) {
        throw "未检测到正在运行的微信。请先启动并登录 Windows 微信 4.x，然后重新运行 start.bat。"
    }
}

function Start-All([bool]$ForceUpdate) {
    Initialize-Environment
    Write-RuntimeConfig
    Assert-WeChat

    $script:DockerExe = Get-DockerExecutable
    Wait-Docker $script:DockerExe
    Write-Step "启动 MySQL"
    Invoke-Compose @("up", "-d", "mysql")
    Wait-MySQL
    Write-Ok "MySQL 已就绪"

    $bridgePort = [int](Get-Setting "BRIDGE_HTTP_PORT" "19089")
    if (Test-Http "http://127.0.0.1:$bridgePort/health") {
        Write-Ok "微信桥接器已在运行"
    } else {
        $pythonExe = Initialize-Python
        Start-Bridge $pythonExe
    }

    $httpPort = [int](Get-Setting "HUB_HTTP_PORT" "8080")
    $hubExe = Join-Path $BinDir "FlyAI.exe"
    if (Test-Tcp "127.0.0.1" $httpPort) {
        Write-Ok "Go 服务已在运行"
    } else {
        $goExe = Get-GoExecutable
        $hubExe = Build-Hub $goExe $ForceUpdate
        Start-Hub $hubExe
    }

    $url = "http://127.0.0.1:$httpPort"
    Write-Host "`n启动完成：$url" -ForegroundColor Green
    Write-Host "配置文件：$EnvFile"
    Write-Host "日志目录：$LogDir"
    Start-Process $url | Out-Null
}

$mutex = New-Object Threading.Mutex($false, "Local\FlyAILauncher")
$locked = $false
try {
    $locked = $mutex.WaitOne(0)
    if (!$locked) { throw "另一个启动脚本正在运行，请稍候。" }
    switch ($Action) {
        "start"   { Start-All $false }
        "update"  {
            Initialize-Environment
            Stop-All
            Start-All $true
        }
        "stop"    { Initialize-Environment; Stop-All }
        "restart" {
            Initialize-Environment
            Stop-All
            Start-All $false
        }
        "status"  { Show-Status }
        "logs"    { Start-Process explorer.exe -ArgumentList (Quote-ProcessArgument $LogDir) | Out-Null }
    }
} catch {
    Write-Host "`n[ERROR] $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "日志目录：$LogDir"
    exit 1
} finally {
    if ($locked) { $mutex.ReleaseMutex() }
    $mutex.Dispose()
}
