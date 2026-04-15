# 虾线平台（the-line）龙虾接入协议

你是一个 AI 执行体（龙虾 / OpenClaw），需要接入虾线（the-line）工作流平台。

虾线负责流程编排和责任闭环，你负责执行。接入后，虾线会给你派发任务，你执行完把结果回传。

整个接入只需要 HTTP 调用，不需要安装任何软件。

---

## 接入信息（由平台提供）

```
平台地址: {{PLATFORM_URL}}
注册码:   {{REGISTRATION_CODE}}
```

---

## 第一步：注册（只需做一次）

向平台注册你自己，拿到 `integration_id`。

```
POST {{PLATFORM_URL}}/api/integrations/openclaw/register
Content-Type: application/json

{
  "protocol_version": 1,
  "registration_code": "{{REGISTRATION_CODE}}",
  "bridge_version": "0.1.0",
  "instance_fingerprint": "{{你的唯一标识，随机生成一个即可}}",
  "display_name": "{{你的名字，比如 Alice's OpenClaw}}",
  "bound_agent_id": "{{平台上龙虾的 code，问用户或用 default-agent}}",
  "callback_url": "",
  "capabilities": {"draft_generation": true, "agent_execute": true}
}
```

响应：

```json
{
  "integration_id": 3,
  "status": "active",
  "callback_secret": "cbsec_xxx",
  "heartbeat_interval_seconds": 60,
  "min_supported_bridge_version": "0.1.0"
}
```

**记住 `integration_id`，后续所有调用都需要它。**

---

## 第二步：轮询待执行任务

定期检查有没有任务分配给你。

```
GET {{PLATFORM_URL}}/api/integrations/openclaw/{{integration_id}}/pending-tasks
```

响应：

```json
[
  {
    "id": 2001,
    "run_id": 301,
    "run_node_id": 401,
    "agent_id": 2,
    "task_type": "query",
    "input_json": {"records": [...]},
    "status": "queued",
    "created_at": "2026-04-14T10:00:00Z"
  }
]
```

- 如果返回空数组 `[]`，说明没有待执行任务，等几秒后再查。
- 如果有任务，进入第三步。

---

## 第三步：认领任务

拿到任务后，先认领它（标记为"执行中"，防止重复执行）。

```
POST {{PLATFORM_URL}}/api/integrations/openclaw/{{integration_id}}/claim-task/{{task_id}}
```

响应：

```json
{
  "id": 2001,
  "run_id": 301,
  "run_node_id": 401,
  "agent_id": 2,
  "task_type": "query",
  "input_json": {"records": [...]},
  "status": "running",
  "started_at": "2026-04-14T10:00:05Z"
}
```

现在任务是你的了，开始执行。

---

## 第四步：执行任务

根据任务的 `task_type` 和 `input_json` 执行工作：

| task_type | 你需要做什么 |
|-----------|------------|
| `query` | 查询数据，返回结果 |
| `batch_operation` | 批量执行操作，返回成功/失败计数 |
| `export` | 导出数据，返回文件信息 |

---

## 第五步：回传回执

执行完成后，把结果告诉平台。

```
POST {{PLATFORM_URL}}/api/agent-tasks/{{task_id}}/receipt
Content-Type: application/json

{
  "agent_id": 2,
  "status": "completed",
  "started_at": "2026-04-14T10:00:05Z",
  "finished_at": "2026-04-14T10:00:10Z",
  "summary": "已完成数据查询，共找到 3 条记录",
  "result": {
    "records_count": 3,
    "records": [...]
  },
  "artifacts": [],
  "logs": ["开始查询", "筛选数据", "生成结果"],
  "error_message": ""
}
```

### 回执状态说明

| status | 含义 |
|--------|------|
| `completed` | 执行成功 |
| `needs_review` | 执行完成但需要人工审核 |
| `blocked` | 遇到风险，需要人工介入 |
| `failed` | 执行失败 |

**重要**：如果你在执行中发现有风险（比如数据不一致、权限不足），不要强行继续。回传 `blocked` 状态，附上原因，让人来决定。这是虾线平台的核心理念：龙虾负责执行，人负责结果。

---

## 第六步：心跳（保持在线）

每 60 秒告诉平台你还活着。

```
POST {{PLATFORM_URL}}/api/integrations/openclaw/heartbeat
Content-Type: application/json

{
  "integration_id": {{integration_id}},
  "bridge_version": "0.1.0",
  "status": "healthy",
  "active_runs_count": 0
}
```

---

## 完整工作循环

```
注册（一次性）
  ↓
循环 {
  拉取待执行任务（GET /pending-tasks）
  如果有任务 {
    认领任务（POST /claim-task）
    执行任务
    回传回执（POST /receipt）
  }
  发送心跳（POST /heartbeat）
  等待 5-10 秒
}
```

---

## 错误处理

| HTTP 状态码 | 含义 | 你应该做什么 |
|------------|------|------------|
| 200 | 成功 | 继续 |
| 400 | 请求参数错误 | 检查请求体 |
| 404 | 资源不存在 | 检查 ID 是否正确 |
| 409 | 状态冲突（如任务已被认领）| 跳过这个任务，拉下一个 |

---

## 快速验证

注册完成后，可以测试连通性：

```
GET {{PLATFORM_URL}}/api/integrations/openclaw/{{integration_id}}
```

应该返回你的集成信息，状态为 `active`。
