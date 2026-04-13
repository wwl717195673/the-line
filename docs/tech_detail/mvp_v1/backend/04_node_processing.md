# 04 节点处理模块实现明细

## 1. 实现范围

本次实现对应：

* `docs/tech_plan/mvp_v1/backend/04_node_processing.md`

实现目录：

* `backend/`

已完成内容：

* 实现节点详情接口
* 实现节点输入暂存接口
* 实现节点提交确认接口
* 实现审核通过接口
* 实现驳回接口
* 实现要求补材料接口
* 实现标记完成接口
* 实现标记异常接口
* 实现龙虾模拟执行接口
* 实现节点动作可用按钮 `available_actions`
* 实现节点动作权限校验、状态校验、输入校验和日志写入
* 补充最小附件模型和查询能力，用于节点附件校验和详情回显

## 2. 新增和变更文件

本次新增或扩展的后端文件：

```text
backend/internal/dto/run_node.go
backend/internal/handler/run_node_handler.go
backend/internal/service/run_node_service.go
backend/internal/model/attachment.go
backend/internal/repository/attachment_repository.go
backend/internal/domain/run.go
backend/internal/response/errors.go
backend/internal/db/migrate.go
backend/internal/repository/run_node_repository.go
backend/internal/repository/node_log_repository.go
backend/internal/app/router.go
```

## 3. 新增路由

节点处理接口：

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET` | `/api/run-nodes/:id` | 节点详情 |
| `PUT` | `/api/run-nodes/:id/input` | 暂存节点输入 |
| `POST` | `/api/run-nodes/:id/submit` | 提交确认 |
| `POST` | `/api/run-nodes/:id/approve` | 审核通过 |
| `POST` | `/api/run-nodes/:id/reject` | 驳回 |
| `POST` | `/api/run-nodes/:id/request-material` | 要求补材料 |
| `POST` | `/api/run-nodes/:id/complete` | 标记完成 |
| `POST` | `/api/run-nodes/:id/fail` | 标记异常 |
| `POST` | `/api/run-nodes/:id/run-agent` | 运行龙虾模拟执行 |

Actor Header 仍沿用流程实例模块约定：

| Header | 说明 |
|---|---|
| `X-Person-ID` | 当前操作人员 ID |
| `X-Role-Type` | 当前操作人员角色，管理员使用 `admin` |

## 4. 节点详情

接口：

```text
GET /api/run-nodes/:id
```

返回内容：

* 节点基础信息
* 所属流程基础信息
* 节点输入 `input_json`
* 节点输出 `output_json`
* 责任人
* 审核人
* 绑定龙虾
* 节点附件列表
* 节点日志列表
* 当前用户可执行动作 `available_actions`
* 评论列表，当前返回空数组，完整评论能力在 `05_collaboration_trace` 模块实现

`available_actions` 当前可能返回：

| action | 说明 |
|---|---|
| `save_input` | 暂存输入 |
| `submit` | 提交确认 |
| `approve` | 审核通过 |
| `reject` | 驳回 |
| `request_material` | 要求补材料 |
| `complete` | 标记完成 |
| `fail` | 标记异常 |
| `run_agent` | 运行龙虾 |

## 5. 输入暂存

接口：

```text
PUT /api/run-nodes/:id/input
```

请求示例：

```json
{
  "input_json": {
    "form_data": {
      "notify_result": "已通知班主任触达家长"
    }
  }
}
```

规则：

* 当前用户必须是节点责任人、审核人或管理员
* 节点状态不能是 `done`
* 所属流程不能是 `cancelled` 或 `completed`
* 只更新 `flow_run_nodes.input_json`
* 不改变节点状态
* 写入 `save_input` 日志

## 6. 提交确认

接口：

```text
POST /api/run-nodes/:id/submit
```

请求示例：

```json
{
  "comment": "请审核"
}
```

规则：

* 当前用户必须是节点责任人或管理员
* 节点状态必须是 `ready`、`running` 或 `waiting_material`
* 校验节点必填输入
* 如果节点要求附件，校验至少存在 1 个附件
* 更新节点状态为 `waiting_confirm`
* 更新流程状态为 `waiting`
* 写入 `submit` 日志

说明：

* 当前固定流程的审核节点前端直接使用审核按钮，因此 `submit` 主要作为通用节点能力保留
* `POST /api/run-nodes/:id/submit` 支持空 body

## 7. 审核通过

接口：

```text
POST /api/run-nodes/:id/approve
```

请求示例：

```json
{
  "review_comment": "审核通过",
  "final_plan": "按运营确认方案甩班"
}
```

规则：

* 当前用户必须是节点审核人、节点责任人且未单独设置审核人，或管理员
* 审核节点允许从 `ready`、`running`、`waiting_confirm`、`waiting_material` 直接审核
* `middle_office_review` 审核通过时 `review_comment` 必填
* `operation_confirm_plan` 审核通过时 `final_plan` 必填
* 更新节点状态为 `done`
* 写入节点 `completed_at`
* 写入审核输出到 `output_json`
* 调用 `RunService.AdvanceAfterNodeDone` 推进到下一个节点
* 写入 `approve` 日志

说明：

* 允许审核节点从 `ready` 直接审核，是为了匹配固定节点前端方案：审核节点页面只展示“审核通过/驳回/补材料”，不展示“提交确认”

## 8. 驳回

接口：

```text
POST /api/run-nodes/:id/reject
```

请求示例：

```json
{
  "reason": "材料信息不完整"
}
```

规则：

* 当前用户必须是节点审核人、节点责任人且未单独设置审核人，或管理员
* 驳回原因 `reason` 不能为空
* 审核节点允许从 `ready`、`running`、`waiting_confirm`、`waiting_material` 驳回
* 更新节点状态为 `rejected`
* 更新流程状态为 `waiting`
* 流程当前节点保持不变
* 不退回上游节点
* 写入 `reject` 日志

## 9. 要求补材料

接口：

```text
POST /api/run-nodes/:id/request-material
```

请求示例：

```json
{
  "requirement": "请补充触达截图"
}
```

规则：

* 当前用户必须是节点审核人、节点责任人且未单独设置审核人，或管理员
* 补材料要求 `requirement` 不能为空
* 审核节点允许从 `ready`、`running`、`waiting_confirm`、`waiting_material` 要求补材料
* 更新节点状态为 `waiting_material`
* 更新流程状态为 `waiting`
* 写入 `request_material` 日志

## 10. 标记完成

接口：

```text
POST /api/run-nodes/:id/complete
```

请求示例：

```json
{
  "comment": "已处理完成",
  "output_json": {
    "summary": "处理完成"
  }
}
```

规则：

* 当前用户必须是节点责任人或管理员
* 节点状态必须是 `ready`、`running` 或 `waiting_material`
* 审核节点不能直接标记完成，必须走审核通过
* 校验节点必填输入
* 如果节点要求附件，校验至少存在 1 个附件
* 更新节点状态为 `done`
* 写入节点 `completed_at`
* 调用 `RunService.AdvanceAfterNodeDone` 推进到下一个节点
* 写入 `complete` 日志

说明：

* `POST /api/run-nodes/:id/complete` 支持空 body
* 空 `output_json` 默认保存为 `{}`

## 11. 标记异常

接口：

```text
POST /api/run-nodes/:id/fail
```

请求示例：

```json
{
  "reason": "甩班工具执行失败"
}
```

规则：

* 当前用户必须是节点责任人、审核人或管理员
* 异常原因 `reason` 不能为空
* 节点不能是 `done`
* 所属流程不能是 `cancelled` 或 `completed`
* 更新节点状态为 `failed`
* 更新流程状态为 `blocked`
* 写入 `fail` 日志

## 12. 运行龙虾

接口：

```text
POST /api/run-nodes/:id/run-agent
```

规则：

* 当前用户必须是节点责任人或管理员
* 节点状态必须是 `ready`、`running` 或 `waiting_material`
* 节点必须绑定 `bound_agent_id`
* 绑定龙虾必须存在且 `status = 1`
* 不调用真实 OpenClaw
* 生成模拟输出并写入 `flow_run_nodes.output_json`
* 节点状态更新为 `running`
* 写入人工触发日志和龙虾执行完成日志

模拟输出：

```json
{
  "summary": "龙虾模拟执行完成",
  "structured_data": {},
  "decision": "mock_success",
  "logs": ["mock agent executed"],
  "next_actions": []
}
```

## 13. 固定节点校验

必填字段来自 `domain.TeacherClassTransferTemplate()` 中的固定节点配置。

当前校验覆盖：

| 节点 | 校验 |
|---|---|
| `submit_application` | `reason`、`class_info`、`current_teacher`、`expected_time` |
| `notify_teacher` | `notify_result` |
| `upload_contact_record` | `contact_description`，且至少 1 个节点附件 |
| `provide_receiver_list` | `receiver_teacher`、`receiver_class`、`handover_description` |
| `operation_confirm_plan` | 审核通过时 `final_plan` 必填 |
| `execute_transfer` | `execute_result` |
| `archive_result` | `deliverable_summary`、`archive_result` |

说明：

* `middle_office_review` 的 `review_comment` 在审核通过动作校验
* `leader_confirm_contact` 的 `review_comment` 按前端方案为建议填写，当前不强制

## 14. 当前边界

已实现：

* 节点详情
* 节点输入暂存
* 节点提交、审核、驳回、补材料、完成、异常、运行龙虾
* 动作权限和状态校验
* 固定节点必填字段校验
* `upload_contact_record` 附件数量校验
* 节点日志追加
* `available_actions` 后端计算

暂未实现：

* 评论 CRUD，当前节点详情 `comments` 返回空数组
* 附件上传接口，本轮只补了最小 `attachments` 表和查询能力
* 完整角色到责任人/审核人的自动分配
* 完整可见性权限矩阵
* 真实 OpenClaw 调用

## 15. 验证结果

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
