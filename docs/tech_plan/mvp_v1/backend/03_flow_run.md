# 03 流程实例后端技术方案

---

## 1. 模块目标

流程实例模块负责从固定模板创建流程实例，并按节点顺序串行推进。

---

## 2. GORM 模型

### 2.1 `FlowRun`

```go
type FlowRun struct {
    ID                uint64         `gorm:"primaryKey"`
    TemplateID        uint64         `gorm:"not null;index"`
    TemplateVersion   int            `gorm:"not null"`
    Title             string         `gorm:"size:256;not null"`
    BizKey            string         `gorm:"size:128;index"`
    InitiatorPersonID uint64         `gorm:"not null;index"`
    CurrentStatus     string         `gorm:"size:32;not null;index"`
    CurrentNodeCode   string         `gorm:"size:64;index"`
    InputPayloadJSON  datatypes.JSON
    OutputPayloadJSON datatypes.JSON
    StartedAt         *time.Time
    CompletedAt       *time.Time
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

### 2.2 `FlowRunNode`

```go
type FlowRunNode struct {
    ID               uint64         `gorm:"primaryKey"`
    RunID            uint64         `gorm:"not null;index"`
    TemplateNodeID   uint64         `gorm:"not null;index"`
    NodeCode         string         `gorm:"size:64;not null;index"`
    NodeName         string         `gorm:"size:128;not null"`
    NodeType         string         `gorm:"size:32;not null;index"`
    SortOrder        int            `gorm:"not null;index"`
    OwnerPersonID    *uint64        `gorm:"index"`
    ReviewerPersonID *uint64        `gorm:"index"`
    BoundAgentID     *uint64        `gorm:"index"`
    Status           string         `gorm:"size:32;not null;index"`
    InputJSON        datatypes.JSON
    OutputJSON       datatypes.JSON
    StartedAt        *time.Time
    CompletedAt      *time.Time
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

说明：PRD 表中未显式包含 `sort_order`，但后端串行推进需要稳定顺序，建议实例节点表增加该字段。

---

## 3. Gin 路由

* `POST /api/runs`
* `GET /api/runs`
* `GET /api/runs/:id`
* `POST /api/runs/:id/cancel`

---

## 4. 发起流程

Service 方法：

```go
func (s *RunService) CreateRun(ctx context.Context, req CreateRunRequest, actor Actor) (*RunDetailDTO, error)
```

事务步骤：

1. 校验模板存在且 `status = published`
2. 校验发起人有效
3. 校验发起表单必填字段
4. 查询模板节点并按 `sort_order asc`
5. 创建 `flow_run`
6. 批量创建 `flow_run_node`
7. 第一个节点状态设置为 `ready`
8. 其他节点状态设置为 `not_started`
9. 写入流程创建日志
10. 返回流程详情

状态写入：

* `flow_run.current_status = running`
* `flow_run.current_node_code = first_node.node_code`
* `flow_run.started_at = now`

---

## 5. 查询流程列表

接口参数：

* `status`
* `owner_person_id`
* `initiator_person_id`
* `scope`
* `page`
* `page_size`

`scope` 规则：

* `all`：返回当前用户可见流程
* `initiated_by_me`：返回当前用户发起流程
* `todo`：返回当前节点责任人为当前用户的流程

查询实现：

* 主表查询 `flow_run`
* 关联当前节点 `flow_run_node`
* 关联发起人 `person`
* 关联当前责任人 `person`
* 默认 `updated_at desc`

---

## 6. 查询流程详情

返回内容：

* 流程基础信息
* 模板信息
* 发起人信息
* 当前节点信息
* 节点列表
* 节点责任人
* 节点审核人
* 绑定龙虾
* 流程评论摘要

处理规则：

* 校验查看权限
* 节点按 `sort_order asc`
* 当前节点需要明确标识
* 已取消流程返回只读状态

---

## 7. 取消流程

Service 方法：

```go
func (s *RunService) CancelRun(ctx context.Context, runID uint64, reason string, actor Actor) error
```

规则：

* 只有发起人或管理员可取消
* `completed` 不能取消
* `cancelled` 不能重复取消
* 取消原因不能为空

事务步骤：

1. 查询流程并加锁
2. 校验状态和权限
3. 更新 `flow_run.current_status = cancelled`
4. 写入取消日志

---

## 8. 串行推进服务

建议放在 `RunService` 或独立 `FlowAdvanceService`：

```go
func (s *RunService) AdvanceAfterNodeDone(tx *gorm.DB, runID uint64, nodeID uint64, actor Actor) error
```

处理规则：

* 查询当前完成节点
* 查询同一 `run_id` 下 `sort_order` 更大的下一个节点
* 如果存在下一个节点，将其状态更新为 `ready`
* 更新 `flow_run.current_node_code`
* 更新 `flow_run.current_status = running`
* 如果不存在下一个节点，更新 `flow_run.current_status = completed`
* 写入完成时间

流程状态同步规则：

* 节点进入 `waiting_confirm`、`waiting_material` 或 `rejected` 时，流程状态为 `waiting`
* 节点标记异常时，流程状态为 `blocked`
* 节点完成并激活下一个节点时，流程状态为 `running`
* 最后一个节点完成时，流程状态为 `completed`

---

## 9. 验收标准

* 发起流程能创建 1 条 `flow_run` 和 9 条 `flow_run_node`
* 第一个节点状态为 `ready`
* 其他节点状态为 `not_started`
* 流程列表支持 `all`、`initiated_by_me`、`todo`
* 流程详情按顺序返回 9 个节点
* 取消流程权限和状态校验正确
* 节点完成后能自动激活下一个节点
* 最后节点完成后流程状态变为 `completed`
