# 07 页面接口聚合后端方案

---

## 1. 模块目标

本模块定义为前端页面提供数据的接口聚合策略，避免前端为了一个页面发起过多零散请求。

MVP 可以先按基础 REST 接口实现；如果页面请求过多，再增加聚合接口。

---

## 2. 工作台接口

接口：

* `GET /api/runs?scope=todo`
* `GET /api/runs?status=running`
* `GET /api/activities/recent`

返回数据：

* 我的待办节点
* 进行中流程
* 最近动态

后端规则：

* 我的待办基于当前用户 ID 查询当前节点责任人为该用户的流程
* 进行中流程查询 `running`、`waiting`、`blocked`
* 最近动态来自 `flow_run_node_log`
* `GET /api/activities/recent` 返回最近节点日志，按 `created_at desc` 排序

`GET /api/activities/recent` 返回字段建议：

* `id`
* `run_id`
* `run_title`
* `run_node_id`
* `node_name`
* `log_type`
* `operator_type`
* `operator_name`
* `content`
* `created_at`

---

## 3. 流程详情聚合接口

接口：

* `GET /api/runs/:id`

返回建议：

* 流程基础信息
* 模板摘要
* 发起人摘要
* 当前节点摘要
* 节点列表
* 节点责任人摘要
* 节点审核人摘要
* 绑定龙虾摘要
* 是否已生成交付物

说明：

* 节点评论、附件、日志可以随节点详情加载
* 如果前端需要一次性展示，也可以在 `GET /api/runs/:id` 返回当前节点详情摘要

---

## 4. 节点详情聚合接口

接口：

* `GET /api/run-nodes/:id`

返回建议：

* 节点基础信息
* 节点输入
* 节点输出
* 责任人
* 审核人
* 绑定龙虾
* 附件列表
* 评论列表
* 日志列表
* 当前用户可执行动作

后端规则：

* 后端计算 `available_actions`
* 已取消流程下 `available_actions` 为空
* 已完成节点下处理动作为空

---

## 5. 模板详情聚合接口

接口：

* `GET /api/templates/:id`

返回建议：

* 模板基础信息
* 节点列表
* 默认责任人规则
* 默认绑定龙虾摘要
* 输入输出 schema 摘要

---

## 6. 交付详情聚合接口

接口：

* `GET /api/deliverables/:id`

返回建议：

* 交付基础信息
* 关联流程摘要
* 节点完成情况
* 关键附件
* 验收人摘要
* 验收状态
* 验收意见

---

## 7. DTO 设计建议

摘要对象：

```go
type PersonBrief struct {
    ID   uint64 `json:"id"`
    Name string `json:"name"`
}

type AgentBrief struct {
    ID      uint64 `json:"id"`
    Name    string `json:"name"`
    Code    string `json:"code"`
    Version string `json:"version"`
}
```

节点详情 DTO 必须包含：

```go
type RunNodeDetailDTO struct {
    ID               uint64          `json:"id"`
    RunID            uint64          `json:"run_id"`
    NodeCode         string          `json:"node_code"`
    NodeName         string          `json:"node_name"`
    NodeType         string          `json:"node_type"`
    Status           string          `json:"status"`
    Owner            *PersonBrief    `json:"owner"`
    Reviewer         *PersonBrief    `json:"reviewer"`
    BoundAgent       *AgentBrief     `json:"bound_agent"`
    Input            json.RawMessage `json:"input"`
    Output           json.RawMessage `json:"output"`
    AvailableActions []string        `json:"available_actions"`
}
```

---

## 8. 验收标准

* 流程详情接口能满足流程详情页首屏展示
* 节点详情接口能满足节点处理区展示
* 后端返回 `available_actions`
* DTO 不直接暴露无关数据库字段
* 关联人员和龙虾返回摘要对象，避免前端二次查询过多
