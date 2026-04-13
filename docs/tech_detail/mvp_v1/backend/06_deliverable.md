# 06 交付模块实现明细

## 1. 实现范围

本次实现对应：

* `docs/tech_plan/mvp_v1/backend/06_deliverable.md`

实现目录：

* `backend/`

已完成内容：

* 实现交付物模型 `Deliverable`
* 实现交付物生成
* 实现交付物列表查询
* 实现交付物详情查询
* 实现交付物验收通过和驳回
* 实现关键附件复制绑定到交付物
* 附件服务支持 `target_type = deliverable`

## 2. 新增和变更文件

本次新增或扩展的后端文件：

```text
backend/internal/domain/deliverable.go
backend/internal/model/deliverable.go
backend/internal/dto/deliverable.go
backend/internal/repository/deliverable_repository.go
backend/internal/service/deliverable_service.go
backend/internal/handler/deliverable_handler.go
backend/internal/db/migrate.go
backend/internal/app/router.go
backend/internal/repository/attachment_repository.go
backend/internal/service/collaboration_service.go
```

## 3. 数据模型

模型文件：

* `backend/internal/model/deliverable.go`

表名：

* `deliverables`

字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uint64` | 主键 |
| `run_id` | `uint64` | 流程实例 ID |
| `title` | `string` | 交付标题 |
| `summary` | `string` | 交付摘要 |
| `result_json` | `datatypes.JSON` | 交付结构化结果 |
| `reviewer_person_id` | `uint64` | 验收人 ID |
| `review_status` | `string` | 验收状态 |
| `reviewed_at` | `*time.Time` | 验收时间 |
| `created_at` | `time.Time` | 创建时间 |
| `updated_at` | `time.Time` | 更新时间 |

验收状态：

| 状态 | 说明 |
|---|---|
| `pending` | 待验收 |
| `approved` | 验收通过 |
| `rejected` | 验收驳回 |

## 4. API 实现

### 4.1 生成交付物

接口：

```text
POST /api/deliverables
```

请求示例：

```json
{
  "run_id": 1,
  "title": "三年级 A 班甩班申请交付结果",
  "summary": "已完成甩班处理并归档",
  "result_json": {
    "conclusion": "甩班完成",
    "exception_note": ""
  },
  "reviewer_person_id": 2,
  "attachment_ids": [10, 11]
}
```

规则：

* `run_id` 不能为空
* `title` 不能为空
* `summary` 不能为空
* `result_json` 必须是合法 JSON，未传时按 `{}` 处理
* `reviewer_person_id` 不能为空
* 验收人必须存在且 `status = 1`
* 流程必须存在
* 流程状态必须是 `completed`
* 当前用户必须是流程发起人、管理员或最终节点责任人
* 初始 `review_status = pending`
* 不自动生成 PDF 或 Excel

事务步骤：

* 加锁查询流程
* 查询流程节点
* 校验生成权限和流程状态
* 汇总节点完成情况写入 `result_json.node_summary`
* 创建 `deliverables`
* 复制关键附件记录到交付物
* 返回交付详情

### 4.2 关键附件绑定

实现策略：

* 不新增 `deliverable_attachment` 表
* 不移动原附件记录
* 原附件仍保留在流程、节点或评论下
* 对请求里的 `attachment_ids`，读取原 `attachments`
* 为每个原附件创建一条新的 `attachments` 记录
* 新记录使用 `target_type = deliverable`
* 新记录使用 `target_id = deliverable.id`
* 新记录复用原 `file_url`
* 不复制文件二进制内容

校验：

* 如果存在无效 `attachment_id`，返回校验错误

### 4.3 查询交付列表

接口：

```text
GET /api/deliverables
```

查询参数：

| 参数 | 说明 |
|---|---|
| `review_status` | 可选，`pending`、`approved`、`rejected` |
| `reviewer_person_id` | 可选，验收人 ID |
| `page` | 页码，默认 `1` |
| `page_size` | 每页条数，默认 `20`，最大 `100` |

规则：

* 默认按 `created_at desc`
* `review_status=pending` 可作为待验收列表
* `review_status=approved` 可作为已归档列表
* 返回关联流程摘要和验收人摘要

### 4.4 查询交付详情

接口：

```text
GET /api/deliverables/:id
```

返回内容：

* 交付物基础信息
* 关联流程信息
* 验收人信息
* 验收状态
* 验收时间
* 结构化结果 `result_json`
* 节点完成情况
* 关键附件

节点完成情况来源：

* `flow_run_nodes`

关键附件来源：

* `attachments.target_type = deliverable`
* `attachments.target_id = deliverable.id`

### 4.5 验收交付物

接口：

```text
POST /api/deliverables/:id/review
```

请求示例：

```json
{
  "review_status": "approved",
  "review_comment": "验收通过"
}
```

规则：

* 交付物必须存在
* 当前用户必须通过 `X-Person-ID` 传入
* 当前用户必须是验收人或 `X-Role-Type = admin`
* `review_status` 只能是 `approved` 或 `rejected`
* 写入 `review_status`
* 写入 `reviewed_at`
* 验收意见写入 `result_json.review_comment`
* 不改变流程状态
* 不重开流程

## 5. 附件服务变更

本模块解除 `deliverable` 作为附件目标的禁用限制。

现在 `POST /api/attachments` 和 `GET /api/attachments` 支持：

```text
target_type=deliverable
```

目标校验：

* `deliverable` 目标必须存在

说明：

* 交付物生成时复制关键附件记录
* 交付物生成后也可以通过附件接口继续向交付物追加附件

## 6. 当前边界

已实现：

* 已完成流程生成交付物
* 非完成流程生成交付物返回 `INVALID_STATE`
* 交付列表筛选
* 交付详情
* 交付验收通过
* 交付验收驳回
* 关键附件复制绑定
* 验收操作不修改流程状态

暂未实现：

* PDF / Excel 导出
* 自动生成富文本交付页
* 每个流程只允许一个交付物的唯一约束
* 交付物变更日志
* 完整可见性权限矩阵
* 关键附件是否属于当前流程的强校验

## 7. 验证结果

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
