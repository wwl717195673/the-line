# 04 节点处理模块

---

## 1. 模块目标

节点处理模块是 MVP V1 的核心。它负责让责任人、审核人和协作者围绕一个节点完成输入、附件、评论、审核、驳回、补材料、异常标记和龙虾模拟执行。

节点处理的核心原则是：节点是流程推进的最小单元，每个节点必须有清晰状态、责任人、输入、输出和日志。

---

## 2. 用户与入口

使用角色：

* 节点责任人
* 审核人
* 协作者
* 发起人
* 管理员
* 观察者

页面入口：

* 工作台 - 我的待办
* 流程详情 - 点击节点
* 流程列表 - 查看流程 - 节点详情

---

## 3. 前端页面需求

### 3.1 节点详情区域

展示内容：

* 节点名称
* 节点说明
* 节点类型
* 节点状态
* 当前责任人
* 审核人
* 绑定龙虾
* 输入表单
* 输出结果
* 附件列表
* 评论区
* 执行日志
* 操作按钮

交互规则：

* 当前用户没有查看权限时不能进入
* 当前用户没有处理权限时只能查看
* 已完成节点默认只读
* 已取消流程下所有节点只读
* 操作按钮根据节点状态和用户角色动态展示

### 3.2 输入表单

页面目标：让责任人录入当前节点处理所需信息。

交互规则：

* 表单结构来自模板节点配置或节点类型默认结构
* 必填字段需要有明确标识
* 暂存时不要求所有必填项都完成
* 提交确认或标记完成时必须校验必填项
* 刷新页面后已暂存内容不能丢失

### 3.3 操作按钮

按钮清单：

* 暂存
* 提交确认
* 审核通过
* 驳回
* 要求补材料
* 标记完成
* 标记异常
* 运行龙虾

展示规则：

* 责任人可看到暂存、提交确认、标记完成、标记异常、运行龙虾
* 审核人可看到审核通过、驳回、要求补材料、标记异常
* 管理员可看到所有处理按钮
* 协作者只可评论和上传附件
* 观察者只读

---

## 4. 后端功能需求

### 4.1 查询节点详情

接口建议：`GET /api/run-nodes/{id}`

返回内容：

* 节点基础信息
* 节点输入
* 节点输出
* 责任人信息
* 审核人信息
* 绑定龙虾信息
* 附件列表
* 评论列表
* 日志列表
* 当前用户可执行动作

处理规则：

* 校验用户查看权限
* 根据状态和角色计算可执行动作
* 已取消流程下返回只读动作集合

### 4.2 节点输入暂存

接口建议：`PUT /api/run-nodes/{id}/input`

输入：

* `input_json`

处理规则：

* 校验当前用户是责任人、审核人或管理员
* 校验节点不是 `done`
* 校验流程不是 `cancelled`
* 更新 `flow_run_node.input_json`
* 不改变节点状态
* 写入暂存日志

### 4.3 提交确认

接口建议：`POST /api/run-nodes/{id}/submit`

输入：

* `input_json`
* `output_json`
* `comment`

处理规则：

* 校验当前用户是责任人或管理员
* 校验节点状态为 `ready`、`running` 或 `waiting_material`
* 校验必填输入
* 校验必需附件
* 更新输入和输出
* 节点状态变为 `waiting_confirm`
* 写入提交日志

### 4.4 审核通过

接口建议：`POST /api/run-nodes/{id}/approve`

输入：

* `review_comment`

处理规则：

* 校验当前用户是审核人或管理员
* 校验节点状态为 `waiting_confirm`
* 节点状态更新为 `done`
* 写入审核意见
* 写入完成时间
* 调用流程串行推进逻辑
* 写入审核通过日志

### 4.5 驳回节点

接口建议：`POST /api/run-nodes/{id}/reject`

输入：

* `reason`

处理规则：

* 校验当前用户是审核人或管理员
* 校验节点状态为 `waiting_confirm`
* 校验驳回原因不能为空
* 节点状态更新为 `rejected`
* 流程当前节点保持不变
* 写入驳回日志
* 不退回上游节点

### 4.6 要求补材料

接口建议：`POST /api/run-nodes/{id}/request-material`

输入：

* `requirement`

处理规则：

* 校验当前用户是审核人或管理员
* 校验节点状态为 `waiting_confirm`
* 校验补充要求不能为空
* 节点状态更新为 `waiting_material`
* 写入补材料日志
* 责任人补充材料后可再次提交确认

### 4.7 标记完成

接口建议：`POST /api/run-nodes/{id}/complete`

输入：

* `input_json`
* `output_json`

处理规则：

* 校验当前用户是责任人或管理员
* 校验节点状态为 `ready`、`running` 或 `waiting_material`
* 校验必填输入
* 校验必需附件
* 校验节点配置为无需审核
* 节点状态更新为 `done`
* 写入完成时间
* 调用流程串行推进逻辑
* 写入完成日志

### 4.8 标记异常

接口建议：`POST /api/run-nodes/{id}/fail`

输入：

* `reason`

处理规则：

* 校验当前用户是责任人、审核人或管理员
* 校验节点不是 `done`
* 校验流程不是 `cancelled`
* 校验异常原因不能为空
* 节点状态更新为 `failed`
* 流程状态更新为 `blocked`
* 写入异常日志

### 4.9 运行龙虾

接口建议：`POST /api/run-nodes/{id}/run-agent`

处理规则：

* 校验节点绑定了启用龙虾
* 校验当前用户是责任人或管理员
* 校验节点状态为 `ready`、`running` 或 `waiting_material`
* 写入开始执行日志
* 生成模拟输出
* 更新 `flow_run_node.output_json`
* 写入执行完成日志
* MVP 阶段不调用真实 OpenClaw
* MVP 阶段不做自动重试

模拟输出建议：

```json
{
  "summary": "龙虾模拟执行完成",
  "structured_data": {},
  "decision": "mock_success",
  "logs": ["mock agent executed"],
  "next_actions": []
}
```

---

## 5. 数据落点

| 表 | 字段 | 说明 |
|---|---|---|
| `flow_run_node` | `status` | 节点状态 |
| `flow_run_node` | `input_json` | 节点输入 |
| `flow_run_node` | `output_json` | 节点输出 |
| `flow_run_node` | `bound_agent_id` | 绑定龙虾 |
| `flow_run_node` | `completed_at` | 完成时间 |
| `flow_run_node_log` | `content` | 操作日志 |
| `flow_run` | `current_status` | 流程状态 |
| `flow_run` | `current_node_code` | 当前节点 |

---

## 6. 状态规则

节点允许的主要流转：

* `not_started -> ready`
* `ready -> running`
* `ready -> waiting_confirm`
* `running -> waiting_confirm`
* `waiting_material -> waiting_confirm`
* `waiting_confirm -> done`
* `waiting_confirm -> rejected`
* `waiting_confirm -> waiting_material`
* `rejected -> ready`
* `running -> failed`
* `failed -> ready`

流程阻塞规则：

* 节点标记异常后，流程状态变为 `blocked`
* 阻塞流程不自动推进
* 管理员或责任人处理后可将节点重新置为 `ready`

---

## 7. 验收标准

* 节点详情能展示输入、输出、责任人、审核人、龙虾、附件、评论和日志
* 责任人能暂存输入且刷新不丢失
* 必填项缺失时不能提交或完成
* 提交确认后节点状态变为 `waiting_confirm`
* 审核通过后节点变为 `done`，流程自动推进
* 驳回必须填写原因，且流程停留当前节点
* 要求补材料必须填写要求，且责任人可再次提交
* 标记异常后流程状态变为 `blocked`
* 运行龙虾能生成模拟输出和执行日志
* 已取消流程下节点不能继续处理
