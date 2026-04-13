# 07 页面接口聚合模块实现明细

## 1. 实现范围

本次实现对应：

* `docs/tech_plan/mvp_v1/backend/07_api_for_pages.md`

实现目录：

* `backend/`

已完成内容：

* 新增最近动态接口 `GET /api/activities/recent`
* 最近动态基于 `flow_run_node_logs` 聚合流程、节点和操作人信息
* 流程详情 `GET /api/runs/:id` 增加是否已生成交付物标识
* 复用现有流程、节点、模板、交付详情接口作为页面聚合接口
* 保留前端按页面分区懒加载节点详情、评论、附件和日志的能力

## 2. 新增和变更文件

本次新增或扩展的后端文件：

```text
backend/internal/dto/activity.go
backend/internal/service/activity_service.go
backend/internal/handler/activity_handler.go
backend/internal/repository/node_log_repository.go
backend/internal/repository/run_repository.go
backend/internal/repository/run_node_repository.go
backend/internal/repository/deliverable_repository.go
backend/internal/service/run_service.go
backend/internal/dto/run.go
backend/internal/app/router.go
```

## 3. 工作台接口

### 3.1 我的待办

复用接口：

```text
GET /api/runs?scope=todo
```

规则：

* 基于 Header `X-Person-ID`
* 查询当前节点责任人为当前用户的流程
* 返回流程摘要和当前节点摘要

### 3.2 进行中流程

复用接口：

```text
GET /api/runs?status=running
```

规则：

* 基于 `flow_runs.current_status`
* 当前实现支持单状态查询
* 如果前端需要一次性查询 `running`、`waiting`、`blocked`，后续可以扩展为 `status_in=running,waiting,blocked`

### 3.3 最近动态

新增接口：

```text
GET /api/activities/recent
```

查询参数：

| 参数 | 说明 |
|---|---|
| `limit` | 可选，默认 `20`，最大 `100` |

数据来源：

* `flow_run_node_logs`
* `flow_runs`
* `flow_run_nodes`
* `persons`
* `agents`

排序：

* `created_at desc, id desc`

返回字段：

| 字段 | 说明 |
|---|---|
| `id` | 日志 ID |
| `run_id` | 流程 ID |
| `run_title` | 流程标题 |
| `run_node_id` | 节点 ID |
| `node_name` | 节点名称 |
| `log_type` | 日志类型 |
| `operator_type` | 操作人类型 |
| `operator_id` | 操作人 ID |
| `operator_name` | 操作人名称 |
| `content` | 日志内容 |
| `created_at` | 创建时间 |

操作人名称规则：

* `operator_type = person` 时，从 `persons` 读取姓名
* `operator_type = agent` 时，从 `agents` 读取名称
* `operator_type = system` 时，返回 `系统`

## 4. 流程详情聚合

复用接口：

```text
GET /api/runs/:id
```

已有返回内容：

* 流程基础信息
* 模板摘要
* 发起人摘要
* 当前节点摘要
* 节点列表
* 节点责任人摘要
* 节点审核人摘要
* 绑定龙虾摘要
* 流程日志

本轮增强字段：

| 字段 | 说明 |
|---|---|
| `has_deliverable` | 当前流程是否已生成交付物 |
| `deliverable_id` | 最新交付物 ID，有交付物时返回 |

说明：

* `has_deliverable` 基于 `deliverables.run_id` 查询
* 如果同一流程后续存在多个交付物，当前返回最新一条
* 当前未强制每个流程只能生成一个交付物

## 5. 节点详情聚合

复用接口：

```text
GET /api/run-nodes/:id
```

已有返回内容：

* 节点基础信息
* 节点输入
* 节点输出
* 责任人
* 审核人
* 绑定龙虾
* 附件列表
* 评论列表
* 日志列表
* 当前用户可执行动作 `available_actions`

规则：

* `available_actions` 由后端计算
* 已取消流程下操作为空
* 已完成节点下操作为空

## 6. 模板详情聚合

复用接口：

```text
GET /api/templates/:id
```

已有返回内容：

* 模板基础信息
* 节点列表
* 默认责任人规则
* 默认绑定龙虾摘要
* 输入 schema
* 输出 schema
* 节点配置

## 7. 交付详情聚合

复用接口：

```text
GET /api/deliverables/:id
```

已有返回内容：

* 交付基础信息
* 关联流程摘要
* 节点完成情况
* 关键附件
* 验收人摘要
* 验收状态
* 验收意见，位于 `result_json.review_comment`

## 8. 当前边界

已实现：

* 工作台最近动态接口
* 流程详情首屏所需的主信息
* 节点详情处理区所需信息
* 模板详情只读信息
* 交付详情页所需主信息
* 流程详情是否已生成交付物标识

暂未实现：

* 单独的 Dashboard 一体化接口
* `status_in` 多状态批量筛选
* 完整权限可见性矩阵
* 面向页面裁剪的更薄 DTO

## 9. 验证结果

已执行：

```bash
cd backend
gofmt -w ./cmd ./internal
go test ./...
go build -o /tmp/the-line-api ./cmd/api
```

验证结果：

* `gofmt` 通过
* `go test ./...` 通过
* `go build -o /tmp/the-line-api ./cmd/api` 通过
* `backend/internal/app/router_test.go` 已覆盖路由注册，不存在 Gin 路由冲突
* 构建产物输出到 `/tmp/the-line-api`，没有在 `backend/` 下残留二进制
