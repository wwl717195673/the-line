# 00 后端总体架构

---

## 1. 架构目标

后端需要支撑 MVP V1 的固定流程闭环：基础数据、固定模板、流程实例、节点处理、协同留痕、交付验收。

技术栈：

* Golang
* Gin
* GORM
* MySQL

核心原则：

* Handler 只做参数绑定、鉴权上下文读取和响应转换
* Service 承载业务规则、状态机和事务
* Repository 封装 GORM 查询
* Model 对应数据库表
* DTO 区分请求、响应和内部领域对象

---

## 2. 推荐工程结构

建议在 `backend/` 下采用如下结构：

```text
backend/
  cmd/
    api/
      main.go
  internal/
    app/
      router.go
      middleware.go
      server.go
    config/
      config.go
    db/
      mysql.go
      migrate.go
    model/
      person.go
      agent.go
      flow_template.go
      flow_template_node.go
      flow_run.go
      flow_run_node.go
      flow_run_node_log.go
      comment.go
      attachment.go
      deliverable.go
    dto/
      common.go
      person.go
      agent.go
      template.go
      run.go
      run_node.go
      comment.go
      attachment.go
      deliverable.go
    repository/
      person_repository.go
      agent_repository.go
      template_repository.go
      run_repository.go
      run_node_repository.go
      comment_repository.go
      attachment_repository.go
      deliverable_repository.go
    service/
      person_service.go
      agent_service.go
      template_service.go
      run_service.go
      run_node_service.go
      comment_service.go
      attachment_service.go
      deliverable_service.go
      node_log_service.go
    handler/
      person_handler.go
      agent_handler.go
      template_handler.go
      run_handler.go
      run_node_handler.go
      comment_handler.go
      attachment_handler.go
      deliverable_handler.go
    domain/
      enums.go
      permissions.go
      node_actions.go
      fixed_template.go
    response/
      response.go
      errors.go
```

---

## 3. Gin 路由分组

建议路由：

```text
/api
  /persons
  /agents
  /templates
  /runs
  /run-nodes
  /comments
  /attachments
  /deliverables
```

MVP 接口：

* `GET /api/persons`
* `POST /api/persons`
* `PUT /api/persons/:id`
* `POST /api/persons/:id/disable`
* `GET /api/agents`
* `POST /api/agents`
* `PUT /api/agents/:id`
* `POST /api/agents/:id/disable`
* `GET /api/templates`
* `GET /api/templates/:id`
* `POST /api/runs`
* `GET /api/runs`
* `GET /api/runs/:id`
* `POST /api/runs/:id/cancel`
* `GET /api/activities/recent`
* `GET /api/run-nodes/:id`
* `PUT /api/run-nodes/:id/input`
* `POST /api/run-nodes/:id/submit`
* `POST /api/run-nodes/:id/approve`
* `POST /api/run-nodes/:id/reject`
* `POST /api/run-nodes/:id/request-material`
* `POST /api/run-nodes/:id/complete`
* `POST /api/run-nodes/:id/fail`
* `POST /api/run-nodes/:id/run-agent`
* `GET /api/comments`
* `POST /api/comments`
* `POST /api/comments/:id/resolve`
* `GET /api/attachments`
* `POST /api/attachments`
* `GET /api/deliverables`
* `POST /api/deliverables`
* `GET /api/deliverables/:id`
* `POST /api/deliverables/:id/review`

---

## 4. GORM 模型约定

模型约定：

* 主键统一使用 `uint64` 或 `int64`
* 时间字段统一使用 `time.Time`
* JSON 字段使用 `datatypes.JSON`
* 状态字段使用 `string`
* 软删除 MVP 阶段不强制使用，停用用 `status`

示例：

```go
type FlowRun struct {
    ID                uint64         `gorm:"primaryKey"`
    TemplateID        uint64         `gorm:"index"`
    TemplateVersion   int
    Title             string         `gorm:"size:256"`
    BizKey            string         `gorm:"size:128;index"`
    InitiatorPersonID uint64         `gorm:"index"`
    CurrentStatus     string         `gorm:"size:32;index"`
    CurrentNodeCode   string         `gorm:"size:64;index"`
    InputPayloadJSON  datatypes.JSON
    OutputPayloadJSON datatypes.JSON
    StartedAt         *time.Time
    CompletedAt       *time.Time
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

---

## 5. 错误响应约定

建议统一响应：

```json
{
  "code": "VALIDATION_ERROR",
  "message": "申请原因不能为空",
  "details": {}
}
```

常见错误码：

* `VALIDATION_ERROR`
* `NOT_FOUND`
* `FORBIDDEN`
* `CONFLICT`
* `INVALID_STATE`
* `INTERNAL_ERROR`

---

## 6. 事务边界

必须使用事务的场景：

* 发起流程：创建 `flow_run` 和 9 个 `flow_run_node`
* 取消流程：更新流程状态并写日志
* 审核通过：更新节点状态、推进流程、写日志
* 标记完成：更新节点状态、推进流程、写日志
* 标记异常：更新节点状态、更新流程状态、写日志
* 生成交付物：读取流程结果、创建交付物、绑定附件

---

## 7. 权限上下文

MVP 可以先用简化用户上下文：

* `CurrentPersonID`
* `RoleType`

来源可以是请求头、登录中间件或开发阶段 mock。

后端必须校验：

* 发起人能取消自己未完成流程
* 管理员拥有全部管理权限
* 节点责任人能处理责任节点
* 审核人能执行审核动作
* 协作者能评论和上传附件
* 观察者只读

---

## 8. 验收标准

* Gin 路由覆盖 MVP 所需接口
* GORM 模型覆盖 MVP 10 张表
* 写操作使用 Service 事务
* 状态机逻辑集中在 Service 层
* Handler 不直接写复杂业务逻辑
* 错误响应格式统一
* 节点动作全部写日志
