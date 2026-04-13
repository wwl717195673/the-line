# 03 流程实例模块实现明细

## 1. 实现范围

本次实现对应：

* `docs/tech_plan/mvp_v1/backend/03_flow_run.md`

实现目录：

* `backend/`

已完成内容：

* 实现流程实例表 `flow_runs`
* 实现流程实例节点表 `flow_run_nodes`
* 实现最小节点日志表 `flow_run_node_logs`
* 实现流程发起接口
* 实现流程列表接口
* 实现流程详情接口
* 实现流程取消接口
* 实现串行推进基础函数 `AdvanceAfterNodeDone`
* 接入 `X-Person-ID` 和 `X-Role-Type` 简化 Actor 上下文

## 2. 新增和变更文件

本次新增或扩展的后端文件：

```text
backend/internal/domain/actor.go
backend/internal/domain/run.go
backend/internal/model/flow_run.go
backend/internal/model/flow_run_node.go
backend/internal/model/flow_run_node_log.go
backend/internal/dto/run.go
backend/internal/repository/run_repository.go
backend/internal/repository/run_node_repository.go
backend/internal/repository/node_log_repository.go
backend/internal/service/run_service.go
backend/internal/handler/run_handler.go
backend/internal/db/migrate.go
backend/internal/app/router.go
backend/internal/repository/person_repository.go
backend/internal/repository/template_repository.go
```

## 3. Actor 上下文

当前没有登录模块，因此流程接口采用轻量 Header 约定：

| Header | 说明 |
|---|---|
| `X-Person-ID` | 当前操作人员 ID |
| `X-Role-Type` | 当前操作人员角色，管理员使用 `admin` |

用途：

* `POST /api/runs` 未传 `initiator_person_id` 时，使用 `X-Person-ID` 作为发起人
* `GET /api/runs?scope=initiated_by_me` 使用 `X-Person-ID` 过滤发起人
* `GET /api/runs?scope=todo` 使用 `X-Person-ID` 过滤当前节点责任人
* `POST /api/runs/:id/cancel` 使用 `X-Person-ID` 和 `X-Role-Type` 做取消权限校验

## 4. 数据模型

### 4.1 流程实例表

模型文件：

* `backend/internal/model/flow_run.go`

表名：

* `flow_runs`

字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uint64` | 主键 |
| `template_id` | `uint64` | 模板 ID |
| `template_version` | `int` | 模板版本 |
| `title` | `string` | 流程标题 |
| `biz_key` | `string` | 业务标识 |
| `initiator_person_id` | `uint64` | 发起人 ID |
| `current_status` | `string` | 当前流程状态 |
| `current_node_code` | `string` | 当前节点编码 |
| `input_payload_json` | `datatypes.JSON` | 发起输入 |
| `output_payload_json` | `datatypes.JSON` | 流程输出 |
| `started_at` | `*time.Time` | 开始时间 |
| `completed_at` | `*time.Time` | 完成或取消时间 |
| `created_at` | `time.Time` | 创建时间 |
| `updated_at` | `time.Time` | 更新时间 |

### 4.2 流程实例节点表

模型文件：

* `backend/internal/model/flow_run_node.go`

表名：

* `flow_run_nodes`

字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uint64` | 主键 |
| `run_id` | `uint64` | 流程实例 ID |
| `template_node_id` | `uint64` | 模板节点 ID |
| `node_code` | `string` | 节点编码 |
| `node_name` | `string` | 节点名称 |
| `node_type` | `string` | 节点类型 |
| `sort_order` | `int` | 节点顺序 |
| `owner_person_id` | `*uint64` | 责任人 ID |
| `reviewer_person_id` | `*uint64` | 审核人 ID |
| `bound_agent_id` | `*uint64` | 绑定龙虾 ID |
| `status` | `string` | 节点状态 |
| `input_json` | `datatypes.JSON` | 节点输入 |
| `output_json` | `datatypes.JSON` | 节点输出 |
| `started_at` | `*time.Time` | 开始时间 |
| `completed_at` | `*time.Time` | 完成时间 |
| `created_at` | `time.Time` | 创建时间 |
| `updated_at` | `time.Time` | 更新时间 |

### 4.3 节点日志表

模型文件：

* `backend/internal/model/flow_run_node_log.go`

表名：

* `flow_run_node_logs`

说明：

* 本模块先实现最小日志表，用于流程创建和取消留痕
* 后续 `05_collaboration_trace` 模块会继续扩展评论、附件和日志查询能力

## 5. API 实现

### 5.1 发起流程

接口：

```text
POST /api/runs
```

请求示例：

```json
{
  "template_id": 1,
  "title": "三年级 A 班甩班申请",
  "biz_key": "class-transfer-20260407-001",
  "initiator_person_id": 1,
  "input_payload_json": {
    "form_data": {
      "reason": "班主任调整",
      "class_info": "三年级 A 班",
      "current_teacher": "张老师",
      "expected_time": "2026-04-10"
    }
  }
}
```

处理规则：

* 校验 `template_id` 非空
* 校验 `title` 非空
* `initiator_person_id` 为空时使用 Header `X-Person-ID`
* 校验发起人存在
* 校验模板存在且 `status = published`
* 查询模板节点并按 `sort_order asc`
* 校验第一个节点的必填字段
* 在事务中创建 1 条 `flow_run`
* 在事务中复制模板节点并创建 9 条 `flow_run_node`
* 第一个实例节点状态为 `ready`
* 其他实例节点状态为 `not_started`
* 流程状态为 `running`
* 流程当前节点为第一个节点编码
* 写入 `run_created` 节点日志

字段校验说明：

* 发起输入支持直接传字段，也支持传 `form_data` 嵌套字段
* 当前固定模板第一个节点要求：`reason`、`class_info`、`current_teacher`、`expected_time`

### 5.2 流程列表

接口：

```text
GET /api/runs
```

查询参数：

| 参数 | 说明 |
|---|---|
| `page` | 页码，默认 `1` |
| `page_size` | 每页条数，默认 `20`，最大 `100` |
| `status` | 可选，按 `current_status` 过滤 |
| `owner_person_id` | 可选，按当前节点责任人过滤 |
| `initiator_person_id` | 可选，按发起人过滤 |
| `scope` | 可选，`all`、`initiated_by_me`、`todo` |

`scope` 规则：

| scope | 当前实现 |
|---|---|
| `all` | 返回全部流程，MVP 暂未接入完整可见性矩阵 |
| `initiated_by_me` | 返回 `initiator_person_id = X-Person-ID` 的流程 |
| `todo` | 返回当前节点责任人为 `X-Person-ID` 的流程 |

列表响应会补充：

* 发起人简要信息
* 当前节点信息
* 当前节点责任人、审核人、绑定龙虾简要信息

### 5.3 流程详情

接口：

```text
GET /api/runs/:id
```

返回内容：

* 流程基础信息
* 模板基础信息
* 发起人简要信息
* 当前节点信息
* 全量节点列表
* 节点责任人信息
* 节点审核人信息
* 绑定龙虾信息
* 流程级节点日志列表

节点排序：

* 按 `sort_order asc` 返回

当前节点标识：

* 节点响应中返回 `is_current`
* 流程响应中返回 `current_node`

### 5.4 取消流程

接口：

```text
POST /api/runs/:id/cancel
```

请求示例：

```json
{
  "reason": "业务方撤回申请"
}
```

处理规则：

* `reason` 不能为空
* 当前用户必须通过 `X-Person-ID` 传入
* 只有发起人或 `X-Role-Type = admin` 的用户可以取消
* `completed` 流程不能取消
* `cancelled` 流程不能重复取消
* 在事务中更新 `flow_run.current_status = cancelled`
* 写入 `completed_at`
* 写入 `run_cancel` 节点日志

## 6. 串行推进函数

实现函数：

```go
func (s *RunService) AdvanceAfterNodeDone(tx *gorm.DB, runID uint64, nodeID uint64, actor domain.Actor) error
```

当前能力：

* 查询已完成节点
* 查找同一流程中 `sort_order` 更大的下一个节点
* 如果存在下一个节点，将下一个节点状态更新为 `ready`
* 将流程状态更新为 `running`
* 将流程当前节点更新为下一个节点编码
* 如果不存在下一个节点，将流程状态更新为 `completed`
* 最后节点完成时写入 `completed_at`

说明：

* 该函数当前由后续 `04_node_processing` 模块调用
* 本模块先实现可复用基础能力

## 7. 分层说明

Handler 层：

* `RunHandler.Create`
* `RunHandler.List`
* `RunHandler.Detail`
* `RunHandler.Cancel`
* `actorFromContext` 从 Header 解析当前用户

Service 层：

* `RunService.CreateRun` 负责发起事务和实例节点复制
* `RunService.List` 负责列表聚合
* `RunService.Detail` 负责详情聚合
* `RunService.CancelRun` 负责取消权限、状态校验和日志写入
* `RunService.AdvanceAfterNodeDone` 负责节点完成后的串行推进

Repository 层：

* `RunRepository`
* `RunNodeRepository`
* `NodeLogRepository`
* `TemplateRepository` 增加事务内查询方法
* `PersonRepository` 增加批量查询方法

## 8. 当前边界

已实现：

* 流程从固定模板发起
* 发起后创建 9 个实例节点
* 第一个节点 `ready`，其他节点 `not_started`
* 流程列表和详情返回当前节点
* 取消流程做基础权限和状态校验
* 发起和取消写入节点日志

暂未实现：

* 完整可见性权限矩阵
* 节点责任人规则的完整解析
* 审核人自动分配
* 节点处理接口
* 评论和附件聚合
* 真实 MySQL 接口级联调

节点责任人当前规则：

* `default_owner_rule = initiator` 时，实例节点责任人为发起人
* 其他规则暂时置空，后续在人员角色和分配规则明确后补齐

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
* 构建产物输出到 `/tmp/the-line-api`，没有在 `backend/` 下残留二进制
