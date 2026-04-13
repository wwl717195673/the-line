# 05 协同留痕模块实现明细

## 1. 实现范围

本次实现对应：

* `docs/tech_plan/mvp_v1/backend/05_collaboration_trace.md`

实现目录：

* `backend/`

已完成内容：

* 实现评论模型 `Comment`
* 实现评论查询、创建、标记已解决
* 实现附件查询和创建
* 实现 multipart 本地文件上传
* 实现 JSON 方式绑定已有附件 URL
* 节点详情接入真实评论列表
* 节点附件上传时写入节点日志
* 实现可选节点日志接口 `GET /api/run-nodes/:id/logs`
* 补充最小路由注册测试，防止 Gin 路由冲突

## 2. 新增和变更文件

本次新增或扩展的后端文件：

```text
backend/internal/domain/collaboration.go
backend/internal/model/comment.go
backend/internal/dto/collaboration.go
backend/internal/repository/comment_repository.go
backend/internal/service/collaboration_service.go
backend/internal/handler/comment_handler.go
backend/internal/handler/attachment_handler.go
backend/internal/app/router_test.go
backend/internal/app/router.go
backend/internal/db/migrate.go
backend/internal/repository/attachment_repository.go
backend/internal/service/run_node_service.go
backend/internal/handler/run_node_handler.go
```

## 3. 数据模型

### 3.1 评论表

模型文件：

* `backend/internal/model/comment.go`

表名：

* `comments`

字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uint64` | 主键 |
| `target_type` | `string` | 目标类型 |
| `target_id` | `uint64` | 目标 ID |
| `author_person_id` | `uint64` | 评论人 |
| `content` | `string` | 评论内容 |
| `is_resolved` | `bool` | 是否已解决 |
| `created_at` | `time.Time` | 创建时间 |
| `updated_at` | `time.Time` | 更新时间 |

支持评论目标：

* `flow_run`
* `flow_run_node`

### 3.2 附件表

附件表在 `04_node_processing` 已创建，本模块补齐创建能力。

模型文件：

* `backend/internal/model/attachment.go`

表名：

* `attachments`

支持附件目标：

* `flow_run`
* `flow_run_node`
* `comment`

`deliverable` 会在交付模块实现后启用。

### 3.3 节点日志表

节点日志表沿用：

* `backend/internal/model/flow_run_node_log.go`

本模块补充的日志类型：

* `attachment_uploaded`

## 4. 评论接口

### 4.1 查询评论

接口：

```text
GET /api/comments
```

查询参数：

| 参数 | 说明 |
|---|---|
| `target_type` | 必填，`flow_run` 或 `flow_run_node` |
| `target_id` | 必填，目标 ID |

规则：

* 校验 `target_type`
* 校验 `target_id`
* 校验目标流程或节点存在
* 只返回目标对象下的评论
* 按 `created_at asc, id asc` 排序
* 返回评论作者简要信息

### 4.2 创建评论

接口：

```text
POST /api/comments
```

请求示例：

```json
{
  "target_type": "flow_run_node",
  "target_id": 1,
  "content": "请补充触达截图"
}
```

规则：

* 当前用户必须通过 `X-Person-ID` 传入
* `target_type` 必须是 `flow_run` 或 `flow_run_node`
* `target_id` 必须存在
* `content` 不能为空
* MVP 不解析真实 `@人`
* MVP 不发送通知

### 4.3 标记评论已解决

接口：

```text
POST /api/comments/:id/resolve
```

规则：

* 评论必须存在
* 当前用户必须通过 `X-Person-ID` 传入
* 更新 `is_resolved = true`
* 不改变流程和节点状态

## 5. 附件接口

### 5.1 查询附件

接口：

```text
GET /api/attachments
```

查询参数：

| 参数 | 说明 |
|---|---|
| `target_type` | 必填，`flow_run`、`flow_run_node` 或 `comment` |
| `target_id` | 必填，目标 ID |

规则：

* 校验附件目标类型
* 校验目标对象存在
* 只返回目标对象下的附件
* 按 `created_at asc, id asc` 排序

### 5.2 JSON 方式创建附件记录

接口：

```text
POST /api/attachments
Content-Type: application/json
```

请求示例：

```json
{
  "target_type": "flow_run_node",
  "target_id": 1,
  "file_name": "contact.png",
  "file_url": "/uploads/contact.png",
  "file_size": 1024,
  "file_type": "image/png"
}
```

规则：

* 当前用户必须通过 `X-Person-ID` 传入
* `target_type` 必须合法
* `target_id` 必须存在
* `file_name` 不能为空
* `file_url` 不能为空
* 如果目标是 `flow_run_node`，写入一条 `attachment_uploaded` 节点日志

### 5.3 multipart 文件上传

接口：

```text
POST /api/attachments
Content-Type: multipart/form-data
```

表单字段：

| 字段 | 说明 |
|---|---|
| `target_type` | 必填 |
| `target_id` | 必填 |
| `file` | 必填，上传文件 |
| `file_type` | 可选，文件类型兜底字段 |

处理规则：

* 文件保存到本地目录，默认 `uploads/`
* 可用环境变量 `UPLOAD_DIR` 覆盖保存目录
* 文件访问路径返回为 `/uploads/{stored_name}`
* Gin 路由通过 `router.Static("/uploads", "uploads")` 暴露本地文件
* 文件名会做基础清理，去掉路径并替换空格

说明：

* 本地上传是 MVP 实现，后续可替换为对象存储
* 如果数据库写入失败，已保存的本地文件当前不会自动回滚删除

## 6. 节点日志接口

接口：

```text
GET /api/run-nodes/:id/logs
```

规则：

* 节点必须存在
* 返回该节点下的日志
* 按 `created_at asc, id asc` 排序

节点详情接口也会继续返回节点日志：

```text
GET /api/run-nodes/:id
```

## 7. 节点详情增强

本模块把节点详情中的 `comments` 从空数组接成真实评论查询：

* 查询 `target_type = flow_run_node`
* 查询 `target_id = 当前节点 ID`
* 补充评论作者简要信息

节点详情继续返回：

* `attachments`
* `comments`
* `logs`
* `available_actions`

## 8. 当前边界

已实现：

* 流程评论创建和查询
* 节点评论创建和查询
* 评论标记已解决
* 附件查询
* 附件 JSON 元信息绑定
* 附件 multipart 本地上传
* 节点附件上传日志
* 节点日志查询

暂未实现：

* 真实 `@人` 解析
* 通知中心
* 完整审计日志后台
* 文件大小、文件类型白名单和病毒扫描
* 对象存储上传
* `deliverable` 附件目标校验，等交付模块实现后启用
* 完整目标对象权限矩阵，当前以目标存在性和当前用户 Header 为基础校验

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
