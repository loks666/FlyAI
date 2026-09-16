<div align="center">
  <img src="webui/frontend/public/flyai-logo.jpg" alt="FlyAI" width="140">
</div>

<h1 align="center">🪽 FlyAI【群聊无忧】</h1>

<div align="center">

[![Stars](https://img.shields.io/github/stars/loks666/FlyAI?style=flat&label=%F0%9F%8C%9Fstars&labelColor=ff4f4f&color=ff8383)](https://github.com/loks666/FlyAI)
[![Forks](https://img.shields.io/github/forks/loks666/FlyAI?style=flat&label=%F0%9F%8F%85%20forks&labelColor=800080&color=912CEE)](https://github.com/loks666/FlyAI/forks)
![closed](https://img.shields.io/github/issues-search?query=repo%3Aloks666/FlyAI+is%3Aclosed&label=%F0%9F%A4%8F%F0%9F%8F%BB%20closed&labelColor=008B8B&color=00CCCC)
[![updated](https://img.shields.io/github/last-commit/loks666/FlyAI/main?style=flat&label=%F0%9F%95%92%20updated&labelColor=4B0082&color=7B68EE)](https://github.com/loks666/FlyAI/commits/main)

[![Go](https://img.shields.io/badge/%F0%9F%90%B9Go-1.21-00ADD8?labelColor=007D9C&style=flat)](https://go.dev/)
[![Gin](https://img.shields.io/badge/%F0%9F%8D%B8Gin-1.10-00B5A0?labelColor=00796B&style=flat)](https://gin-gonic.com/)
[![MySQL](https://img.shields.io/badge/%F0%9F%90%ACMySQL-8.4-4479A1?labelColor=00618A&style=flat)](https://www.mysql.com/)
[![Redis](https://img.shields.io/badge/%F0%9F%A7%B1Redis-7.4-FF4438?labelColor=A41E11&style=flat)](https://redis.io/)
[![Next.js](https://img.shields.io/badge/%E2%9A%A1Next.js-16-B39DDB?labelColor=5E35B1&style=flat)](https://nextjs.org/)
[![React](https://img.shields.io/badge/%E2%9A%9B%EF%B8%8FReact-19-61DAFB?labelColor=1E88E5&style=flat)](https://react.dev/)
[![Nuitka](https://img.shields.io/badge/%F0%9F%90%8DNuitka-Bridge-FFD43B?labelColor=B8860B&style=flat)](https://github.com/loks666/FlyAI-Bridge)

</div>

<div align="center">
<h3 align="center">让群聊里的数据自己说话</h3>
<p><strong>采集你选择监控的群聊，在网页端看清活跃度、发言排行、潜水成员与热点，</strong><br>
<strong>并用兼容 OpenAI 接口的模型完成群聊问答。</strong></p>
</div>

---

- 🔔 *重要说明*
    - 目前**只支持 Windows 微信 4.x**（建议 4.1.12 或更高版本），不支持服务器部署。
    - 项目只读取本机已登录微信的数据，通过 Windows UI Automation 发送文本消息，**不注入 DLL**。
    - 微信并未为这些能力提供官方接口，使用前请充分了解风险并遵守相关平台规则。
    - 所有服务默认只监听 `127.0.0.1`，不要在不清楚风险时暴露到公网。

### 🌴 源码地址

- 主项目：https://github.com/loks666/FlyAI
- 微信桥接器：https://github.com/loks666/FlyAI-Bridge

## 🌟 特色功能

- **📊 群聊数据分析**：消息数量、发言排行、消息类型、活跃率等统计。
- **🤿 潜水成员识别**：按统计周期和阈值生成低活跃成员报告，只发送文字；默认展示前 20 人的群昵称或微信昵称，文本超过 2000 字时缩减为前 10 人。
- **🏆 群聊排行榜**：汇总群聊活跃情况和成员表现。
- **🔥 热点推送**：定时获取热点内容，并按群独立配置推送策略。
- **🤖 AI 群聊助手**：支持 OpenAI 兼容接口，可配置 API Key、Base URL、模型和群聊人设。
- **👑 群主命令**：绑定管理员后，可在群聊中触发排行榜、潜水报告和热点等操作。
- **📇 联系人管理**：同步好友和群聊，在网页端查看联系人并发送文本消息。
- **🛟 安全回填**：桥接器重启后可回填本周遗漏消息；回填只修复数据，不触发 AI 或自动回复。
- **🖥️ 网页控制台**：集中配置监控群、定时任务、AI、端口和功能开关。
- **⚙️ Go 原生编排**：运行 FlyAI 后自动启动 Docker 服务、管理页和微信桥接器。
- **🔒 微信多开保护**：按机器人微信号解析本地账号数据库，并用数据库密钥绑定唯一微信进程和窗口；无法唯一确认时拒绝发送。

### 🔞️ 注意事项

- 🖥️ 只支持 Windows 桌面环境，服务器 / Linux / macOS 无法运行，UIA 依赖真实的微信窗口。
- 🪟 目标微信可以最小化，发送时会临时恢复并在完成后重新最小化；**不要退出微信或将其关闭到托盘**。
- 🐳 首次启动需要联网拉取 MySQL 和 Redis 镜像，Docker Desktop 未运行时 FlyAI 会尝试自动拉起。
- 🧪 建议先只开启群聊监控，确认数据正常后再逐项启用自动推送和 AI，避免首次接入就产生意外回复。

## 🚀 如何使用？

### 1️⃣ 环境要求

- Windows 10/11 x64
- 已安装 [Docker Desktop](https://www.docker.com/products/docker-desktop/)（FlyAI 会在需要时自动启动）
- 已登录的 Windows 微信 4.x（建议 4.1.12 或更高版本）
- 从源码运行时需要 Go 1.21 或更高版本
- 项目根目录存在 `FlyAI-Bridge.exe`，或在 `.env` 中设置 `BRIDGE_EXECUTABLE_PATH`
- 首次启动时可连接互联网，用于下载 Docker 镜像

### 2️⃣ 拉取代码

```powershell
git clone https://github.com/loks666/FlyAI.git
cd FlyAI
```

### 3️⃣ 一键启动

```powershell
go run .
```

编译后也可以直接运行生成的 `FlyAI.exe`。Go 服务会按以下顺序自动完成启动：

1. 从 `.env.example` 生成仅保存在本地的 `.env`，并在缺失时生成 `config.yaml`；
2. 检查 Docker Engine，必要时隐藏启动 Docker Desktop；
3. 使用 Docker Compose 启动并等待 MySQL、Redis 健康；
4. 初始化数据库、插件、任务及 Web/TCP 服务，并打开管理页面；
5. 若数据库中已保存机器人微信号，自动拉起微信桥接器并完成账号绑定。

首次创建 `data/mysql` 时，MySQL 会自动执行 [`database/schema.sql`](database/schema.sql)，一次性创建项目使用的全部数据表和默认功能时间配置；已有数据库不会重复初始化或覆盖数据。

### 4️⃣ 绑定微信桥接器

首次使用时数据库里还没有机器人微信号，需要手动绑定一次：

1. 打开 http://127.0.0.1:9528 ，进入「功能配置 → 微信桥接器」；
2. 填写**机器人自己的唯一微信号**并保存；
3. 点击「启动桥接器」，等待状态显示「运行正常」。

绑定成功后微信号会持久化，**之后每次启动 FlyAI 都会自动拉起桥接器**，无需再点击。想恢复成手动启动，在 `.env` 中设置 `BRIDGE_AUTOSTART=false`。

> 桥接器由 Windows Job Object 绑定到 Go 进程：停止 Go 或结束 GoLand 调试时，Go 启动的桥接器父子进程会一起退出。MySQL 和 Redis 使用 Docker 的 `restart: unless-stopped` 策略，不随 Go 频繁重启，避免数据库反复启停。

### 5️⃣ 开启监控与功能

1. 在联系人页面同步好友和群聊；
2. 进入群聊管理，**只为需要分析的群开启消息监控**；
3. 按需开启排行榜、潜水报告、热点推送或 AI 助手。

## 🏗️ 运行结构

```text
已登录的 Windows 微信 4.x
          │
          ▼
Nuitka 编译桥接器 ──读取本机数据库 / UIA 发送文本
          │ 127.0.0.1:19099
          ▼
FlyAI Go 服务 ─────消息处理、统计、任务与 AI
          │
          ├── MySQL（Docker）
          ├── Redis（Docker）
          └── Web 控制台 http://127.0.0.1:9528
```

## 🤖 AI 配置

FlyAI 使用 OpenAI Chat Completions 兼容接口。复制环境变量示例后即可配置：

```powershell
Copy-Item .env.example .env
notepad .env
```

```dotenv
AI_ENABLE=false
AI_API_KEY=
AI_BASE_URL=https://api.openai.com
AI_MODEL=gpt-4o-mini
```

- `AI_ENABLE`：是否启用全局 AI 能力。
- `AI_API_KEY`：模型服务商提供的密钥。
- `AI_BASE_URL`：OpenAI 兼容 API 地址，请按服务商文档填写。
- `AI_MODEL`：服务商支持的模型名称。

网页端还可以为不同群聊设置是否启用 AI、模型和角色设定。密钥保存在本地配置中；`.env` 和自动生成的 `config.yaml` 已被 Git 忽略，不会被提交。

## ⚙️ 常用配置

完整示例见 [`.env.example`](.env.example)。常用项目包括：

| 配置 | 默认值 | 说明 |
| --- | --- | --- |
| `HUB_HTTP_PORT` | `9528` | 管理页面端口 |
| `HUB_INGEST_PORT` | `19099` | Go 服务接收消息的端口 |
| `BRIDGE_HTTP_PORT` | `19089` | 微信桥接器监听端口 |
| `SYSTEM_DOMAIN` | `http://127.0.0.1:9528` | 管理页地址，**必须与 `HUB_HTTP_PORT` 一致**；桥接器靠它回调拉取监控群列表 |
| `BRIDGE_AUTOSTART` | `true` | 启动 FlyAI 后按已保存的机器人微信号自动拉起桥接器；设为 `false` 时改为在管理页面手动启动 |
| `OPEN_BROWSER` | `true` | HTTP 服务成功监听后自动打开默认浏览器 |
| `WEB_DIST_PATH` | 空 | 可选的前端 `dist` 目录；默认自动查找 `webui/frontend/dist` |
| `BRIDGE_EXECUTABLE_PATH` | 空 | 本地 Nuitka 编译产物路径；也可将 EXE 放在项目根目录 |
| `BRIDGE_DOWNLOAD_URL` | 空 | 找不到本地桥接器时使用的完整下载地址 |
| `BRIDGE_SHA256` | 空 | 可选的下载文件 SHA-256 校验值 |
| `WECHAT_HOST` | `127.0.0.1:19089` | Go 服务访问桥接器的地址 |
| `MYSQL_PORT` | `3307` | 本地 MySQL 映射端口 |
| `MYSQL_USER` | `wechat` | FlyAI 数据库用户 |
| `MYSQL_PASSWORD` | `wechat` | FlyAI 数据库密码 |
| `MYSQL_DATABASE` | `wechat` | 数据库名 |
| `REDIS_PORT` | `6379` | Redis 本机端口 |
| `REDIS_PASSWORD` | `flyai-redis` | Redis 本地密码 |

消息处理默认使用有界队列，避免消息集中到达时无限创建 goroutine：

| 配置 | 默认值 | 说明 |
| --- | --- | --- |
| `RECEIVE_WORKERS` | `4` | TCP 消息接收工作器数量 |
| `RECEIVE_QUEUE_SIZE` | `1000` | 待接收连接队列容量 |
| `PLUGIN_WORKERS` | `4` | 插件工作器数量，同一群固定进入同一工作器 |
| `PLUGIN_QUEUE_SIZE` | `500` | 插件待处理消息总容量 |
| `SEND_QUEUE_SIZE` | `100` | 微信发送队列容量，发送工作器固定为 1 |
| `SYNC_QUEUE_SIZE` | `100` | 联系人及群成员同步队列容量，同步工作器固定为 1 |

> 如果本机端口已占用，修改 `.env` 后重新运行 FlyAI 即可。改 `HUB_HTTP_PORT` 时记得同步改 `SYSTEM_DOMAIN`。

### 📋 日志分流

- 接收到的消息以单行 JSON 输出到日志，内容字段位于最前面，并附带发送者微信昵称、群内昵称、微信号和群聊名称。以上身份信息也会随消息保存到 MySQL；尚未同步到通讯录时使用 wxid 兜底，通讯录信息短期缓存 5 分钟以减少高频数据库查询。
- 「获取好友列表结果」和「同步好友列表完成」不会输出到消息控制台，而是单独追加到 `logs/contact-sync.log`；通讯录获取或入库失败等异常仍保留在控制台。
- 群成员不在好友通讯录时产生的预期 `record_not_found` 只写入 `logs/contact-cache.log`；桥接器定期访问 `/api/monitor/groups` 的高频访问记录只写入 `logs/http-monitor.log`。这两类正常诊断信息不会刷 Debug 控制台，其他数据库错误和 HTTP 请求仍正常输出。

## 🖥️ 网页控制台

- **功能配置**只管理统一执行时间：排行榜统计滚动最近 7 天，可选择每周自动执行的星期和时间。年排行榜固定在每年 1 月 1 日 00:00 自动发送，所有已开启消息监控的群默认参与，因此不再占用群组页面的开关卡片。
- **群组页面**可分别控制排行榜、知乎热榜和潜水统计；绿色圆点表示已开启，关闭后状态标签灰显。每项功能旁边的独立绿色「发送」按钮可立即向当前群发送对应内容，手动发送不受该项定时开关限制。昨日排行和月排行已移除。
- **微信桥接器卡片**保存机器人微信号，并提供启动、停止和手动刷新。页面只在首次打开、启动/停止后以及用户点击刷新图标时检查状态，不会定时轮询。多开微信时，桥接器先从每个本地账号数据库解析微信号和 wxid，再用数据库密钥确认对应微信进程并固定窗口句柄；所有文本、文件和收藏表情发送都只能使用这个窗口。窗口关闭、账号不匹配或出现多个候选窗口时会拒绝发送，避免串号。

## 👑 群主命令

在网页端绑定管理员后，可通过 `@机器人` 触发；群功能开关只控制定时投送，不限制管理员手动命令：

```text
@机器人 排行榜（滚动最近 7 天）
@机器人 周排（排行榜的兼容别名）
@机器人 年排
@机器人 潜水
@机器人 热点
@机器人 给我骂他@昵称1 @昵称2
给我骂他@昵称1 @昵称2（主人可不艾特机器人）
给我骂他.@昵称1 @昵称2（从摩斯密码列表随机回复）
```

排行榜、周排、年排、潜水和热点同时兼容 `/` 前缀，例如 `@机器人 /潜水`。命令只对已绑定的管理员生效，避免普通群成员触发管理操作。

主人权限只按消息发送者的唯一微信号判断，群昵称、微信昵称和 wxid 均不参与授权；微信号缺失时默认拒绝主人指令。

## 🧱 Redis 使用边界

- **回复状态**：非管理员 `@机器人` 或冒用主人命令时，只以发送者 wxid 为键，不包含群 ID，因此同一个人更换群聊也不会重新触发回复。第一次触发时机器人会艾特发送者并从中文消息池随机回复；第二次触发时艾特发送者、提示管理员昵称和微信号、发送第一张收藏表情，并在发送成功后设置 1 小时静默 TTL；静默期间后续群艾特、冒用主人命令、私聊和私聊指令均直接忽略，普通群消息仍正常入库但不会回复。超过一小时未发生第二次艾特则重新从第一次开始。
- **通讯录缓存**：采用「进程内存 → Redis → MySQL」三级读取。Redis 按联系人和群成员分别保存结构化记录，同时维护联系人索引及每群成员索引，默认 TTL 为 1 小时；项目启动时自动预热，同步通讯录或群成员成功后立即刷新对应缓存。群功能开关等业务配置仍直接读取 MySQL，避免缓存延迟影响管理操作。
- **降级**：Redis 不可用时自动降级为当前进程内存，服务仍能继续工作。
- **适用与不适用**：Redis 还适合后续保存短期 AI 会话上下文、接口限流、幂等键、分布式任务锁、临时任务状态和高频统计缓存。不应使用 Redis 代替 MySQL 保存群聊历史、联系人、功能配置等需要长期可靠保留的数据。

## 🔐 数据与隐私

- 微信数据库只在本机读取，不会由 FlyAI 自动上传。
- 通讯录中的群聊消息仅在本机入库，用于随时生成手动报告；只有明确开启功能的群才会自动投送。
- AI 未启用时，不会向模型服务商发送群聊内容。
- AI 启用后，参与回答所需的内容会发送到你配置的 API 地址，请自行确认服务商的隐私政策。
- 默认 Web、桥接器和消息入口均绑定 `127.0.0.1`，不要在不清楚风险时暴露到公网。
- `.env`、`config.yaml`、数据库和运行日志均已加入 Git 忽略规则。

## 🛠️ 开发运行

新版管理端使用 Next.js + shadcn/ui。构建结果位于 `webui/frontend/dist/`，Go 后端启动时直接读取该目录；修改网页后只需重新构建前端，无需重新编译 Go。Docker、MySQL、Redis 和桥接器仍由 Go 编排：

```powershell
npm --prefix webui/frontend ci
npm --prefix webui/frontend run build
go mod download
go test ./...
go run .
```

桥接器源码单独留存在 [FlyAI-Bridge](https://github.com/loks666/FlyAI-Bridge)；运行时只需要 Nuitka 编译后的 `FlyAI-Bridge.exe`。本地开发默认还会查找相邻目录 `FlyAI-Bridge/dist/FlyAI-Bridge.exe`。

### 📁 项目目录

```text
app/        Web 路由与页面接口
config/     配置定义与加载
orchestration/ Docker、桥接器等依赖的启动编排
plugin/     消息处理、AI 和群主命令插件
service/    群聊、联系人、功能设置等业务逻辑
tasks/      排行榜、潜水报告、热点等定时任务
webui/      Next.js + shadcn/ui 管理端及 Go 运行时读取的 dist 静态产物
views/      消息卡片继续使用的品牌图片资源
doc/        设计与后续规划文档
```

## 🤖 Agent 规划

当前版本已经提供 AI 群聊回复能力。下一阶段计划将推送、排行榜、潜水分析、联系人查询等能力封装为受控工具，让 Agent 可以根据自然语言组合调用。设计边界和分阶段方案见 [`doc/agent-integration-plan.md`](doc/agent-integration-plan.md)。

## 🚩 已知限制

- 目前只支持 Windows 微信 4.x。
- 当前发送能力以文本消息为主。
- UI Automation 会受微信界面或版本更新影响。
- 本项目不是微信官方产品，也不提供账号安全保证。

## 🤝 参与贡献

欢迎提交 Issue 和 Pull Request。报告问题时，请附上 Windows、微信版本、复现步骤和已脱敏的相关日志。

FlyAI 基于原项目历史代码继续维护，并保留 Git 提交历史以便追溯贡献。

<a href="https://github.com/loks666/FlyAI/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=loks666/FlyAI" alt="贡献者" />
</a>

---

## ⚠️ 免责声明

本项目仅用于个人学习、研究与本地自动化。使用者应遵守所在地法律法规、微信服务条款及群成员隐私约定，并自行承担使用风险。

---

### ☕️ Github Star 历史

<a href="https://star-history.com/#loks666/FlyAI&Date">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=loks666/FlyAI&type=Date&theme=dark" />
    <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=loks666/FlyAI&type=Date" />
    <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=loks666/FlyAI&type=Date" />
  </picture>
</a>
