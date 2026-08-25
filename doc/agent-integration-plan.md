# FlyAI：Agent 与群聊分析改造规划

> 状态：Draft  
> 创建日期：2026-08-25  
> 目标版本：首个 Agent MVP  
> 建议项目名：FlyAI（GitHub：`FlyAI`）

## 1. 背景

当前项目已经具备以下基础能力：

- 从 Windows 微信 4.x 本地数据库同步受监控群聊的消息；
- 通过本地桥接器和 UI Automation 发送文本消息；
- 将消息、联系人和群成员保存到 MySQL；
- 生成最近 7 天发言排行、潜水成员报告、每日热点早报和群聊总结；
- 通过网页启用群监控、排行榜、潜水统计、早报等功能；
- 调用 OpenAI 兼容的 Chat Completions 接口生成普通聊天回复。

现有 AI 逻辑属于“上下文聊天机器人”：模型只能根据 Prompt 和最近消息生成文本，不能主动查询数据库或调用项目功能。本次改造的目标是增加 Agent Runtime，把群聊统计、报告预览、定时推送等现有能力注册成结构化工具，让模型能根据自然语言选择并调用工具。

例如：

- “最近 7 天谁发言最多？”
- “这个群现在有哪些人潜水？”
- “总结一下昨天大家主要讨论了什么。”
- “预览本群的周报。”
- “把本群周报改到每周五 18:30 推送。”

其中统计结果必须由数据库查询和确定性代码计算，AI 只负责理解意图、选择工具和解释结果，不允许由模型猜测人数、排名或发言数量。

## 2. 目标与非目标

### 2.1 第一阶段目标

1. 网页端可以配置 AI Provider、Base URL、API Key、模型及超时等参数。
2. 支持 OpenAI Responses API 的函数工具调用。
3. 保留旧的 OpenAI 兼容 Chat Completions 模式，避免破坏不支持 Responses API 的第三方模型服务。
4. 支持在群内通过 `@机器人` 使用自然语言查询群聊数据。
5. 支持私聊管理员查询其有权限管理的群。
6. 将发言排行、潜水成员、热点早报和已有推送能力封装成 Agent 工具。
7. 所有工具都有权限等级、参数校验、超时、调用日志和结果大小限制。
8. 写操作默认需要明确授权；批量发送、删除成员等高风险操作不进入首个 MVP。

### 2.2 暂不纳入首个 MVP

- 让模型执行任意 SQL、Shell 或本地文件操作；
- 自动踢人、退群、添加好友或批量群发；
- 未经确认修改多个群的配置；
- 依赖模型计算精确统计数据；
- 在桥接器尚未支持时承诺图片、文件、语音等发送能力；
- 第一版直接接入任意远程 MCP Server。

## 3. 关键产品定义

### 3.1 “潜水成员”必须区分口径

当前代码中的潜水统计定义为：当前仍在群内，并且在本机可读取的群历史中从未发言的成员。

Agent 工具需要明确支持以下口径，回答中必须回显统计口径和时间范围：

| 口径 | 定义 | 推荐用途 |
| --- | --- | --- |
| `never_spoken` | 可读取历史中从未发言 | 现有潜水报告兼容模式 |
| `no_message_in_range` | 指定时间范围内发言数为 0 | “最近 7/30 天谁在潜水” |
| `inactive_since` | 最后发言时间早于指定日期 | 社群召回和活跃分析 |
| `low_activity` | 指定范围内发言数低于阈值 | 分层运营 |

不能把“本机没有读取到消息”等同于“此人从未发言”。工具结果必须附带 `data_scope`，说明监控起始时间、查询时间范围、当前成员数以及数据是否完整。

### 3.2 Agent 触发规则

- 群聊：默认只响应 `@机器人` 的文本消息，避免对普通群消息逐条调用模型。
- 私聊：仅对已启用 Agent 的联系人或管理员响应。
- 原有 `/` 指令继续优先执行，Agent 不重复处理已识别的命令。
- 回填消息只用于修复数据库，绝不触发 Agent、工具或主动发送。
- 同一会话串行执行，防止连续消息产生交叉工具调用和乱序回复。

## 4. 总体架构

```text
Windows WeChat 4.x
        |
        v
Python Local Bridge -- TCP event --> Go Message Dispatcher
        ^                              |
        |                              v
        +---- HTTP text send ----- Agent Router
                                       |
                         +-------------+-------------+
                         |                           |
                         v                           v
                  Legacy Chat Provider        Agent Runtime
                  (Chat Completions)          (Responses API)
                                                     |
                                             filtered tool set
                                                     |
                          +--------------------------+--------------------------+
                          |                          |                          |
                          v                          v                          v
                    Analytics Tools            Feature Tools              Action Tools
                    read-only SQL            preview/schedule          guarded writes
```

建议新增目录：

```text
agent/
├── runtime.go          # Responses 工具循环、轮次和超时控制
├── router.go           # 触发规则、会话范围、Provider 选择
├── provider.go         # AgentProvider / ChatProvider 接口
├── permission.go       # 工具权限和执行策略
├── session.go          # 会话状态与并发控制
├── registry.go         # 工具注册表
├── schema.go           # 工具 JSON Schema 和通用输入输出
├── audit.go            # 调用审计
├── providers/
│   ├── responses.go    # OpenAI Responses API
│   └── chat.go         # 旧 Chat Completions 兼容实现
└── tools/
    ├── activity.go
    ├── dormant.go
    ├── messages.go
    ├── reports.go
    ├── schedules.go
    └── send.go
```

现有 `plugin/plugins/ai.go` 后续只负责把消息交给 `agent.Router`，不要继续同时包含消息查询、Provider 调用、Token 统计和微信发送等所有逻辑。

## 5. Provider 与 Agent Runtime 设计

### 5.1 双 Provider 模式

定义两个明确的能力接口：

```go
type ChatProvider interface {
    Reply(ctx context.Context, req ChatRequest) (ChatResponse, error)
}

type AgentProvider interface {
    CreateResponse(ctx context.Context, req AgentRequest) (AgentResponse, error)
    ContinueWithToolOutputs(ctx context.Context, req ToolOutputRequest) (AgentResponse, error)
}
```

- `chat_compatible`：继续兼容当前使用 Chat Completions 的服务，只能普通聊天。
- `openai_responses`：支持函数工具、结构化输出和多轮工具循环。

网页测试 Provider 时必须分别验证：

1. Base URL 可访问；
2. API Key 是否有效；
3. 指定模型是否能完成最小文本请求；
4. Agent 模式是否真正支持函数调用，不能仅凭服务自称兼容；
5. 错误只返回经过脱敏的摘要。

### 5.2 工具调用循环

一次 Agent 请求的建议流程：

1. 根据消息来源解析 `actor_wxid`、`group_id` 和管理员身份。
2. 读取群/联系人的 Agent 配置和 Provider。
3. 依据调用者权限过滤可见工具，未授权工具不能出现在模型的工具列表中。
4. 调用 Responses API。
5. 遍历模型返回的全部函数调用，而不是只取第一个输出项。
6. 严格校验 JSON Schema；`group_id` 等安全范围参数由服务器注入或复核，不能完全信任模型参数。
7. 执行工具，将结构化结果作为 tool output 返回模型。
8. 重复以上过程，直到得到最终文本或达到最大工具轮次。
9. 保存审计记录并通过微信回复。

建议默认限制：

- 最大工具轮次：5；
- 单次请求总超时：60 秒；
- 单个 SQL 查询超时：5 秒；
- 最大返回成员数：100；
- 最大历史消息条数：200；
- 同一会话并发数：1；
- 同一用户增加分钟级请求频率限制。

Responses API 支持自定义函数工具，并可以使用 `conversation` 或 `previous_response_id` 维护多轮状态。本项目第一版建议由 MySQL 保存自己的会话映射和摘要，同时保留关闭远端存储的能力。

## 6. Agent 工具规划

### 6.1 只读分析工具（MVP 自动允许）

#### `get_group_activity_summary`

用途：回答谁发言最多、活跃人数、消息数量和消息类型构成。

输入：

```json
{
  "group_id": "string",
  "start_time": "RFC3339",
  "end_time": "RFC3339",
  "top_n": 10
}
```

输出至少包含：

- 时间范围及 `Asia/Shanghai` 时区；
- 群成员总数、发言人数、未发言人数、活跃率；
- 总消息数、日均消息数；
- 文本、图片、视频、表情、语音、分享及其他消息数量；
- 按发言数排序的成员列表；
- 数据范围和完整性说明。

现有 `tasks/watergroup` 的 SQL 与格式化逻辑可复用，但应先拆成“结构化统计 Service”和“文本 Formatter”。Agent 工具返回 JSON，定时推送继续调用 Formatter。

#### `get_group_dormant_members`

用途：查询目前有哪些人潜水。

输入：

```json
{
  "group_id": "string",
  "mode": "never_spoken | no_message_in_range | inactive_since | low_activity",
  "start_time": "RFC3339 or null",
  "end_time": "RFC3339 or null",
  "max_messages": 0,
  "limit": 100
}
```

输出包含当前群成员数、符合条件人数、成员昵称、wxid、最后活跃时间、范围内发言数和数据口径。现有 `tasks/dormant.BuildReport` 只覆盖 `never_spoken`，需要扩展其他模式。

#### `get_group_message_trend`

用途：分析每日/每小时活跃趋势、峰值时段和活跃变化。

#### `search_group_messages`

用途：按关键词、发言人和时间范围检索消息。默认不向模型返回 `raw` 字段，限制最大条数，并对敏感字段脱敏。

#### `get_group_topic_material`

用途：取得经过筛选的文本消息，供模型做话题归纳、情绪/诉求聚类和群聊总结。工具只负责可靠取数，不在 SQL 工具中混入模型结论。

### 6.2 现有报告工具（读取/预览自动允许）

| 工具名 | 复用能力 | 默认权限 |
| --- | --- | --- |
| `preview_weekly_rank` | `watergroup.BuildWeekRank` | Read |
| `preview_dormant_report` | `dormant.BuildReport` | Read |
| `preview_morning_news` | `news.BuildMorningPost` | Read |
| `preview_group_summary` | 群聊总结构建逻辑 | Read |

### 6.3 配置工具（管理员允许）

| 工具名 | 功能 | 执行策略 |
| --- | --- | --- |
| `get_group_feature_settings` | 查询群功能开关 | 自动执行 |
| `get_push_schedules` | 查询统一推送计划 | 自动执行 |
| `set_group_feature_enabled` | 开关排行、潜水、早报、总结 | 要求群管理员身份及明确目标 |
| `set_push_schedule` | 修改推送周期和时间 | 要求管理员身份，回复变更前后值 |

设置类工具必须使用枚举字段，禁止模型传递任意数据库列名。

### 6.4 发送工具（首版受控）

| 工具名 | 功能 | 默认策略 |
| --- | --- | --- |
| `send_current_chat_message` | 回复当前会话 | Agent 最终回复可用 |
| `send_weekly_rank` | 向当前群发送周报 | 明确要求后执行 |
| `send_dormant_report` | 向当前群发送潜水报告 | 明确要求后执行 |
| `send_morning_news` | 向当前群发送早报 | 明确要求后执行 |
| `send_message_to_contact` | 给其他联系人发消息 | MVP 默认关闭，后续增加审批 |

工具执行层必须复用已有的生成/发送函数，不能通过调用本项目自己的 HTTP 路由绕一圈。预览和发送要保持分离，保证模型可以先展示内容再执行。

不把 `DeleteGroupMember`、`QuitChatroom`、任意批量发送注册成首版 Agent 工具。

## 7. 数据模型

### 7.1 `t_ai_provider`

建议字段：

| 字段 | 说明 |
| --- | --- |
| `id` | UUID 主键 |
| `name` | 网页显示名称 |
| `provider_type` | `openai_responses` / `chat_compatible` |
| `base_url` | Provider 地址 |
| `api_key_ciphertext` | 加密后的 API Key |
| `model` | 默认模型 |
| `summary_model` | 可选总结模型 |
| `enabled` | 是否启用 |
| `timeout_seconds` | 请求超时 |
| `max_tool_rounds` | 最大工具轮次 |
| `created_at` / `updated_at` | 审计时间 |

### 7.2 `t_agent_scope_setting`

建议按会话范围配置，避免继续向 `t_friend` 塞入大量 Agent 字段：

| 字段 | 说明 |
| --- | --- |
| `scope_id` | 好友 wxid 或群 wxid |
| `enabled` | 是否启用 Agent |
| `provider_id` | 使用的 Provider |
| `mention_only` | 群内是否必须 @ |
| `system_prompt` | 此范围的角色和业务说明 |
| `tool_policy` | 允许的工具策略 JSON |
| `history_window_minutes` | 普通聊天上下文窗口 |
| `max_daily_requests` | 每日限制 |

### 7.3 `t_agent_session`

- `scope_id`、`actor_wxid`；
- Provider conversation/response 标识；
- 本地摘要；
- 最后活动时间和过期时间；
- 状态版本，用于并发更新。

### 7.4 `t_agent_tool_audit`

- 请求 ID、消息 ID、调用者、群 ID；
- 工具名和脱敏后的参数；
- 权限决策；
- 执行状态、耗时和错误码；
- 结果摘要，不记录完整聊天内容和 API Key。

## 8. 网页端 AI 配置

在现有“功能配置”页面增加独立的“AI 与 Agent”区域，建议包含：

### 8.1 Provider 配置

- 启用开关；
- Provider 名称；
- 模式：Responses Agent / Chat Completions Compatible；
- Base URL；
- API Key 密码框；
- 默认模型；
- 总结模型；
- 请求超时；
- 最大工具轮次；
- “测试连接”按钮；
- 最近一次测试时间和脱敏后的状态。

### 8.2 API Key 安全要求

1. API 获取接口永远不能返回明文 Key，只返回 `configured: true` 和类似 `sk-...abcd` 的掩码。
2. 编辑时 Key 留空表示保持原值；清除 Key 使用单独且明确的操作。
3. 浏览器提交后禁止将 Key 写入前端日志、URL、HTML模板或错误提示。
4. 数据库使用 AES-256-GCM 加密，主密钥来自环境变量（建议 `FLYAI_SECRET_KEY`），不能与密文放在同一数据库中。
5. 服务日志禁止打印完整配置。当前初始化代码打印整个配置对象，实施前必须先移除或脱敏，否则网页保存的 Key 可能进入日志。
6. Provider 测试错误需要移除认证头、Key、内部响应体中的敏感信息。

### 8.3 建议 API

```text
GET    /api/ai/providers
POST   /api/ai/providers
PUT    /api/ai/providers/:id
DELETE /api/ai/providers/:id
POST   /api/ai/providers/:id/test

GET    /api/agent/tools
GET    /api/agent/scopes/:scopeId
PUT    /api/agent/scopes/:scopeId
POST   /api/agent/scopes/:scopeId/test

GET    /api/agent/audit
```

当前管理页面只监听 `127.0.0.1`，仍应增加管理端认证与 CSRF 防护。不能因为是本机服务就默认所有网页请求可信。

## 9. 权限模型

定义四个工具等级：

| 等级 | 能力 | 默认执行方式 |
| --- | --- | --- |
| `read` | 查询消息、成员和统计 | 自动 |
| `preview` | 生成报告但不发送 | 自动 |
| `write` | 发消息、修改单群配置 | 管理员且用户意图明确 |
| `admin` | 批量发送、成员管理、外部系统写入 | MVP 禁用，后续人工审批 |

额外规则：

- 群内普通成员只能查询当前群的允许数据，不能通过参数切换到其他群。
- 群管理员可以修改当前群的功能开关。
- 系统管理员白名单可以通过私聊管理多个群。
- 工具注册表为每次请求动态生成，不把未授权工具暴露给模型。
- 模型输出不是授权依据，最终权限判断必须由 Go 服务完成。
- 群消息和历史消息都视为不可信数据，防止聊天内容中的 Prompt Injection 诱导工具越权。

## 10. 群聊分析的数据与性能要求

### 10.1 统计查询

- 所有时间范围使用半开区间 `[start_time, end_time)`；
- 业务默认时区为 `Asia/Shanghai`；
- 过滤机器人自身消息、系统消息和历史自动推送时必须采用明确规则；
- 排行中的跳过成员继续兼容 `skip_chat_rank`；
- 查询返回 wxid 和展示昵称，昵称优先使用群昵称；
- 不允许模型拼接 SQL，所有查询使用固定 Repository 方法和绑定参数。

### 10.2 索引建议

根据实际执行计划评估并补充组合索引：

```text
t_message(from_user, create_at, group_user, type)
t_message(group_user, create_at)
t_group_user(group_id, is_member, wxid)
t_friend(wxid)
```

迁移前需要检查现有数据量、重复记录和数据库类型。当前部分查询使用 MySQL 专有函数，因此文档中声明的 PostgreSQL 支持在实现 Agent 分析前需要重新验证。

### 10.3 数据可信度

- 统计页面和 Agent 回答都显示时间范围；
- 显示监控启用时间或最早可用消息时间；
- 监控未开启、历史不足或群成员同步失败时返回明确错误；
- “没有查到”不等于“实际没有”；
- 不让 AI 补齐缺失数据。

## 11. 与现有推送功能的重构关系

现有功能普遍已经具备 `Build...` 和 `Send...` 两层，但返回值多数是最终文本。建议统一为三层：

```text
Query/Service -> Structured Result -> Formatter -> Sender
                                      |
                                      +-> Web preview
                                      +-> Scheduled push
                                      +-> Agent tool output
```

例如周排行：

```go
type GroupActivityReport struct {
    GroupID       string
    StartTime     time.Time
    EndTime       time.Time
    MemberCount   int64
    ActiveCount   int
    SilentCount   int64
    MessageCount  int64
    ActivityRate  float64
    MessageMix    MessageBreakdown
    Ranking       []MemberActivity
    DataScope     DataScope
}

func QueryGroupActivity(...) (GroupActivityReport, error)
func FormatWeeklyRank(GroupActivityReport) string
func SendWeeklyRank(groupID string) error
```

这样网页、定时器和 Agent 使用同一份统计结果，避免不同入口出现不同口径。

## 12. 错误处理与用户体验

- Agent 开始耗时任务时可先回复“正在统计最近 7 天消息，请稍候”；
- 超时后给出可理解的错误，不向微信发送堆栈或 Provider 原始响应；
- 微信文本过长时按段落安全分片，保持顺序并控制发送间隔；
- 工具结果为空时说明查询范围和可能的数据限制；
- 工具调用失败不能自动重试写操作；只读调用可进行有限重试；
- 使用消息 ID 作为幂等键，防止重复事件触发重复发送；
- Agent 自己发送并回填到数据库的消息不能再次触发 Agent。

## 13. 测试计划

### 13.1 单元测试

- 各种潜水口径的边界情况；
- 时间范围、时区和跨天/跨周计算；
- 发言排行、并列排名、跳过成员；
- 工具 JSON Schema 参数校验；
- 权限矩阵和跨群参数注入；
- API Key 加密、掩码、保留和清除；
- 工具最大轮次和超时；
- 微信长消息分片。

### 13.2 集成测试

- 使用假 Provider 验证“模型请求工具 -> 执行 -> 返回结果 -> 最终回答”完整循环；
- Responses API 不支持或 Base URL 不兼容时返回明确诊断；
- 回填消息不会触发 Agent；
- 连续消息在同一会话中保持顺序；
- Preview 不发送、Send 只发送一次；
- 网页 API 不泄露明文 Key。

### 13.3 验收示例

输入：

```text
@机器人 最近七天谁说话最多？目前有多少人没发过言？
```

期望行为：

1. Agent 调用 `get_group_activity_summary`；
2. Agent 调用 `get_group_dormant_members(mode=never_spoken)`；
3. 两个工具均被限制在当前群；
4. 最终回复列出第一名、发言数、当前成员数、潜水人数、统计时间范围和数据口径；
5. 所有数字与直接 SQL/Service 结果一致；
6. 审计表记录两个只读工具调用，但不记录完整消息正文。

## 14. 分阶段实施

### Phase 0：安全与配置基础

- 增加 Provider、Scope、Session、Audit 数据表；
- 增加 `FLYAI_SECRET_KEY` 和 Key 加密存储；
- 移除完整配置日志；
- 完成网页 Provider CRUD、Key 掩码和测试连接；
- 为管理页面补充身份认证方案。

验收：网页能够安全保存并测试 Provider，API 和日志均不出现明文 Key。

### Phase 1：只读群聊分析 Agent

- 引入 Agent Runtime 和 Responses Provider；
- 实现工具注册表、权限过滤、会话锁和审计；
- 重构最近 7 天排行，输出结构化结果；
- 实现四种潜水成员口径；
- 接入群内 @ 和管理员私聊。

验收：自然语言可以可靠回答“谁发言最多”和“谁在潜水”，数字可由数据库复核。

### Phase 2：现有推送工具化

- 封装周报、潜水报告、早报、总结的 Preview 工具；
- 封装当前群单次发送工具；
- 封装单群功能开关和推送计划工具；
- 实现写操作的明确意图检查和幂等控制。

验收：Agent 能预览现有推送，并在管理员明确要求时对当前群执行一次发送或配置变更。

### Phase 3：高级分析与长期记忆

- 话题、诉求和异常趋势分析；
- 长会话摘要和本地记忆；
- 自定义知识库；
- 报表导出与可视化；
- 成本、Token 和响应质量监控。

### Phase 4：外部工具与 MCP（可选）

- GitHub Issue、Todo、工单和文档系统集成；
- 外部 MCP Server 白名单；
- 写操作审批队列；
- 多 Agent 分工仅在实际业务复杂度需要时引入。

## 15. 首个版本的完成标准

首个 Agent 版本只有同时满足以下条件才算完成：

- 网页可以安全配置并测试 Key、Base URL 和模型；
- 兼容模式与 Agent 模式界限清晰；
- 群内必须 @ 才触发 Agent；
- 发言排行和潜水统计来自结构化查询；
- Agent 不能访问当前调用者无权访问的群；
- 回填、机器人自身消息和重复消息不会触发重复执行；
- 工具调用有审计、超时、限流和最大轮次；
- Preview 与 Send 行为严格分离；
- 不注册踢人、退群、任意 SQL、Shell、文件操作等高风险工具；
- 自动化测试覆盖主要统计口径、权限和 Key 安全。

## 16. 参考资料

- OpenAI Responses API：<https://developers.openai.com/api/reference/cli/resources/responses/methods/create>
- OpenAI 官方 Go SDK Responses 示例：<https://github.com/openai/openai-go/tree/main/examples/responses>

实现时以官方文档和目标 Provider 的实际兼容性测试结果为准，不假设所有 OpenAI 兼容 Base URL 都支持 Responses API 或工具调用。
