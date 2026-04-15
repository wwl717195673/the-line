# 接入虾线平台（the-line）

你需要帮我把当前的 OpenClaw 接入虾线（the-line）工作流平台。

## 接入信息

- 平台地址: `{{PLATFORM_URL}}`
- 注册码: `{{REGISTRATION_CODE}}`（30 分钟内有效）

## 你需要做什么

### 1. 获取 bridge 代码

```bash
git clone https://github.com/xxx/the-line-bridge.git
cd the-line-bridge
```

> 如果已经有这个目录，直接 `cd the-line-bridge && git pull` 即可。

### 2. 编译

```bash
go build -o the-line-bridge ./cmd/bridge/
```

需要 Go 1.22+。如果没有 Go 环境，请先安装。

### 3. 运行接入向导

```bash
MOCK_MODE=false ./the-line-bridge setup \
  --platform-url={{PLATFORM_URL}} \
  --registration-code={{REGISTRATION_CODE}}
```

> 如果暂时没有真实的 OpenClaw Runtime API 可对接，加上 `MOCK_MODE=true` 可以用模拟模式先跑通。

向导会自动完成：
- 检查平台是否可达
- 检查 OpenClaw 运行时状态
- 选择默认执行 Agent
- 向平台注册当前实例
- 保存接入配置到 `data/bridge-config.json`

成功后你会看到：

```
=== 接入完成 ===
  平台地址:     http://...
  Integration:  <数字ID>
  绑定 Agent:   <agent名称>
  Bridge 版本:  0.1.0
```

### 4. 启动 bridge 服务

```bash
./the-line-bridge serve
```

bridge 会在 `9090` 端口启动 HTTP 服务，并自动向平台发送心跳。

### 5. 验证接入状态

```bash
curl http://localhost:9090/bridge/health
```

应返回：

```json
{"ok": true, "data": {"status": "healthy", "bridge_version": "0.1.0"}}
```

## bridge 做了什么

接入成功后，bridge 作为中间层运行：

```
虾线平台 ──执行请求──> bridge ──调用──> OpenClaw Runtime
虾线平台 <──回执──── bridge <──结果── OpenClaw Runtime
```

- 虾线发来草案生成请求 → bridge 调用你的 planner 能力 → 返回结构化流程草案
- 虾线发来自动节点执行请求 → bridge 调用你的执行能力 → 执行完成后回传回执
- bridge 每 60 秒向平台发送心跳，报告健康状态

## 环境变量说明

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `BRIDGE_PORT` | `9090` | bridge HTTP 服务端口 |
| `PLATFORM_URL` | `http://localhost:8080` | 虾线平台地址 |
| `OPENCLAW_API_URL` | `http://localhost:8081` | OpenClaw Runtime API 地址 |
| `MOCK_MODE` | `true` | 是否使用模拟运行时 |
| `DATA_DIR` | `data` | 配置文件存储目录 |

## 如果接入失败

| 错误信息 | 原因 | 解决方式 |
|---------|------|---------|
| 无法访问平台 | 平台地址错误或网络不通 | 检查 PLATFORM_URL 和网络 |
| 注册码无效 | 注册码已使用或拼写错误 | 去平台重新生成注册码 |
| 注册码已过期 | 超过有效期 | 去平台重新生成注册码 |
| 绑定的龙虾编码不存在 | 平台上没有对应 agent | 先在平台创建 agent |

## 如何对接真实的 OpenClaw Runtime

bridge 通过 `OpenClawRuntime` 接口与 OpenClaw 交互。接口定义在 `internal/runtime/runtime.go`：

```go
type OpenClawRuntime interface {
    PlanDraft(ctx, req)      // 草案生成
    ExecuteTask(ctx, req)    // 启动任务执行
    WaitForResult(ctx, key)  // 等待执行结果
    CancelTask(ctx, key)     // 取消任务
    Health(ctx)              // 健康检查
    ListAgents(ctx)          // 列出可用 agent
}
```

当前 `MOCK_MODE=true` 使用模拟实现。要对接真实 Runtime：

1. 在 `internal/runtime/` 下新建一个实现（例如 `openclaw_runtime.go`）
2. 实现上述 6 个方法，内部调用 OpenClaw 的 HTTP API（`chat.send`、`agent.wait` 等）
3. 在 `cmd/bridge/main.go` 的 `runServe` 中，当 `MOCK_MODE=false` 时使用你的实现

## 目录结构

```
the-line-bridge/
├── cmd/bridge/main.go          # 入口：setup 和 serve 两个命令
├── data/bridge-config.json     # setup 后生成的配置
├── internal/
│   ├── app/                    # HTTP 路由和服务器
│   ├── client/                 # 调用虾线平台的 HTTP 客户端
│   ├── config/                 # 环境变量配置
│   ├── handler/                # HTTP 请求处理
│   │   ├── draft_handler.go    # POST /bridge/drafts/generate
│   │   ├── execution_handler.go# POST /bridge/executions
│   │   ├── health_handler.go   # GET  /bridge/health
│   │   └── test_ping_handler.go# POST /bridge/test-ping
│   ├── receipt/                # 执行结果 → 回执格式转换
│   ├── runtime/                # OpenClaw Runtime 接口和实现
│   │   ├── runtime.go          # 接口定义
│   │   └── mock_runtime.go     # 模拟实现
│   ├── service/                # 心跳、Setup 等业务逻辑
│   └── store/                  # 本地配置持久化
└── go.mod
```
