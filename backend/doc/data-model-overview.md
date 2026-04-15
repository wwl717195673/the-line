# Backend Data Model Overview

这份文档梳理 `backend` 当前的主要数据对象，基于真实后端代码中的 `model`、`dto`、`domain` 三层定义整理而成。

## 1. 文档范围

- `internal/model`：数据库实体定义
- `internal/dto`：接口入参 / 出参对象
- `internal/domain`：状态枚举、节点类型、业务常量

本文以 `internal/model` 中的落库对象为主。

## 2. 核心业务主线

后端当前的主要业务链路如下：

1. 维护人员与龙虾资源：`Person`、`Agent`
2. 设计流程：`FlowDraft`
3. 确认草案生成模板：`FlowTemplate`、`FlowTemplateNode`
4. 基于模板发起流程实例：`FlowRun`
5. 流程推进到具体节点：`FlowRunNode`
6. 节点协作沉淀评论、附件、日志：`Comment`、`Attachment`、`FlowRunNodeLog`
7. 节点需要龙虾执行时派发任务：`AgentTask`、`AgentTaskReceipt`
8. 流程完成后生成交付物并审核：`Deliverable`
9. 对接外部 OpenClaw 桥接能力：`OpenClawIntegration`、`RegistrationCode`

## 3. 数据对象分组

### 3.1 人员与龙虾资源

#### Person

文件：`internal/model/person.go`

表示系统中的人工参与者。

主要字段：

- `id`
- `name`
- `email`
- `role_type`
- `status`
- `created_at`
- `updated_at`

常见角色由业务侧约定，例如：

- `leader`
- `middle_office`
- `operation`
- `admin`

#### Agent

文件：`internal/model/agent.go`

表示系统中的龙虾执行体或自动化代理。

主要字段：

- `id`
- `name`
- `code`
- `provider`
- `version`
- `owner_person_id`
- `config_json`
- `status`
- `created_at`
- `updated_at`

说明：

- `code` 是业务唯一标识
- `owner_person_id` 表示该龙虾的负责人
- `config_json` 用于保存运行配置

### 3.2 流程草案与模板

#### FlowDraft

文件：`internal/model/flow_draft.go`

表示流程设计草案，通常由人或龙虾先生成一版，再确认成正式模板。

主要字段：

- `id`
- `title`
- `description`
- `source_prompt`
- `creator_person_id`
- `planner_agent_id`
- `status`
- `structured_plan_json`
- `confirmed_template_id`
- `created_at`
- `updated_at`
- `confirmed_at`

说明：

- `structured_plan_json` 是草案主体，保存完整编排结构
- `planner_agent_id` 表示生成该草案的规划龙虾
- `confirmed_template_id` 表示草案确认后生成的模板

#### FlowTemplate

文件：`internal/model/flow_template.go`

表示正式流程模板。

主要字段：

- `id`
- `name`
- `code`
- `version`
- `category`
- `description`
- `status`
- `created_by`
- `created_at`
- `updated_at`

说明：

- `code` 是模板唯一标识
- `version` 支持模板版本化

#### FlowTemplateNode

文件：`internal/model/flow_template_node.go`

表示模板下的节点定义。

主要字段：

- `id`
- `template_id`
- `node_code`
- `node_name`
- `node_type`
- `sort_order`
- `default_owner_rule`
- `default_owner_person_id`
- `default_agent_id`
- `result_owner_rule`
- `result_owner_person_id`
- `input_schema_json`
- `output_schema_json`
- `config_json`
- `created_at`
- `updated_at`

说明：

- 一个模板包含多个模板节点
- `sort_order` 表示模板中的顺序
- `default_owner_rule` 表示默认责任归属规则
- `default_agent_id` 表示默认绑定的龙虾
- `input_schema_json` / `output_schema_json` 描述节点输入输出结构

### 3.3 流程运行态

#### FlowRun

文件：`internal/model/flow_run.go`

表示一次实际发起的流程实例。

主要字段：

- `id`
- `template_id`
- `template_version`
- `title`
- `biz_key`
- `initiator_person_id`
- `current_status`
- `current_node_code`
- `input_payload_json`
- `output_payload_json`
- `started_at`
- `completed_at`
- `created_at`
- `updated_at`

说明：

- `template_id` 指向模板
- `template_version` 固化发起时使用的模板版本
- `current_status` 表示流程整体状态
- `current_node_code` 表示当前推进到哪个节点

#### FlowRunNode

文件：`internal/model/flow_run_node.go`

表示流程实例里的单个运行节点。

主要字段：

- `id`
- `run_id`
- `template_node_id`
- `node_code`
- `node_name`
- `node_type`
- `sort_order`
- `owner_person_id`
- `reviewer_person_id`
- `result_owner_person_id`
- `bound_agent_id`
- `status`
- `input_json`
- `output_json`
- `started_at`
- `completed_at`
- `created_at`
- `updated_at`

说明：

- 一个 `FlowRun` 对应多个 `FlowRunNode`
- `template_node_id` 指向模板节点定义
- `bound_agent_id` 表示运行时实际绑定的龙虾
- `status` 表示节点当前状态

#### FlowRunNodeLog

文件：`internal/model/flow_run_node_log.go`

表示流程节点上的日志与操作轨迹。

主要字段：

- `id`
- `run_id`
- `run_node_id`
- `log_type`
- `operator_type`
- `operator_id`
- `content`
- `extra_json`
- `created_at`

说明：

- 用于记录提交、审批、驳回、系统推进、龙虾执行等事件
- `operator_type` 支持人、龙虾、系统

### 3.4 协作对象

#### Comment

文件：`internal/model/comment.go`

表示评论。

主要字段：

- `id`
- `target_type`
- `target_id`
- `author_person_id`
- `content`
- `is_resolved`
- `created_at`
- `updated_at`

说明：

- 当前评论目标主要是 `flow_run` 和 `flow_run_node`

#### Attachment

文件：`internal/model/attachment.go`

表示附件。

主要字段：

- `id`
- `target_type`
- `target_id`
- `file_name`
- `file_url`
- `file_size`
- `file_type`
- `uploaded_by`
- `created_at`

说明：

- 附件可挂到流程、节点、评论、交付物等对象上

### 3.5 交付物

#### Deliverable

文件：`internal/model/deliverable.go`

表示流程最终产出的交付物。

主要字段：

- `id`
- `run_id`
- `title`
- `summary`
- `result_json`
- `reviewer_person_id`
- `review_status`
- `reviewed_at`
- `created_at`
- `updated_at`

说明：

- 一个流程通常对应一个交付物
- `review_status` 表示审核状态

### 3.6 龙虾执行任务

#### AgentTask

文件：`internal/model/agent_task.go`

表示派发给龙虾的一次执行任务。

主要字段：

- `id`
- `run_id`
- `run_node_id`
- `agent_id`
- `task_type`
- `input_json`
- `status`
- `started_at`
- `finished_at`
- `error_message`
- `result_json`
- `artifacts_json`
- `external_runtime`
- `external_session_key`
- `external_run_id`
- `created_at`
- `updated_at`

说明：

- 任务挂在某个流程节点上
- `task_type` 目前包括查询、批量操作、导出等
- `external_*` 字段用于关联外部运行时

#### AgentTaskReceipt

文件：`internal/model/agent_task_receipt.go`

表示龙虾任务回执。

主要字段：

- `id`
- `agent_task_id`
- `run_id`
- `run_node_id`
- `agent_id`
- `receipt_status`
- `payload_json`
- `received_at`

说明：

- 用于异步接收任务完成、失败、阻塞、待确认等结果

### 3.7 OpenClaw 集成

#### OpenClawIntegration

文件：`internal/model/openclaw_integration.go`

表示一个外部 OpenClaw 桥接实例。

主要字段：

- `id`
- `display_name`
- `status`
- `bridge_version`
- `openclaw_version`
- `instance_fingerprint`
- `bound_agent_id`
- `capabilities_json`
- `callback_url`
- `callback_secret`
- `heartbeat_interval`
- `last_heartbeat_at`
- `last_error_message`
- `created_at`
- `updated_at`

说明：

- 用于桥接外部运行时与本系统龙虾任务体系

#### RegistrationCode

文件：`internal/model/registration_code.go`

表示 OpenClaw 桥接注册邀请码。

主要字段：

- `id`
- `code`
- `status`
- `integration_id`
- `expires_at`
- `created_at`
- `updated_at`

说明：

- 用于桥接实例首次注册接入

## 4. 对象关系梳理

### 4.1 资源关系

- `Person` 1:N `Agent`
  - 一个人员可以拥有多个龙虾

### 4.2 草案与模板关系

- `FlowDraft` N:1 `Person`
  - 草案由创建人发起
- `FlowDraft` N:1 `Agent`
  - 草案可由规划龙虾生成
- `FlowDraft` 1:0..1 `FlowTemplate`
  - 草案确认后生成模板
- `FlowTemplate` 1:N `FlowTemplateNode`
  - 一个模板有多个模板节点

### 4.3 运行态关系

- `FlowRun` N:1 `FlowTemplate`
  - 流程实例基于模板发起
- `FlowRun` N:1 `Person`
  - `initiator_person_id` 表示发起人
- `FlowRun` 1:N `FlowRunNode`
  - 一个流程实例包含多个运行节点
- `FlowRunNode` N:1 `FlowTemplateNode`
  - 运行节点来源于模板节点
- `FlowRunNode` N:1 `Person`
  - 节点可关联执行人、审核人、结果责任人
- `FlowRunNode` N:1 `Agent`
  - 节点运行时可绑定龙虾
- `FlowRunNode` 1:N `FlowRunNodeLog`
  - 节点行为日志

### 4.4 协作关系

- `Comment` 通过 `target_type + target_id` 关联目标对象
- `Attachment` 通过 `target_type + target_id` 关联目标对象

### 4.5 交付关系

- `Deliverable` N:1 `FlowRun`
  - 交付物来源于流程实例
- `Deliverable` N:1 `Person`
  - `reviewer_person_id` 表示审核人

### 4.6 龙虾任务关系

- `AgentTask` N:1 `FlowRun`
- `AgentTask` N:1 `FlowRunNode`
- `AgentTask` N:1 `Agent`
- `AgentTaskReceipt` N:1 `AgentTask`

### 4.7 OpenClaw 集成关系

- `OpenClawIntegration` N:1 `Agent`
  - 一个桥接实例可绑定一个龙虾
- `RegistrationCode` 0..1:1 `OpenClawIntegration`
  - 注册码使用后可指向某个桥接实例

## 5. 当前主要状态枚举

### 5.1 通用启停状态

定义位置：`internal/domain/status.go`

- `0`：`disabled`
- `1`：`enabled`

### 5.2 草案状态

定义位置：`internal/domain/draft.go`

- `draft`
- `confirmed`
- `discarded`

### 5.3 模板状态与节点类型

定义位置：`internal/domain/template.go`

模板状态：

- `published`

模板节点类型：

- `manual`
- `review`
- `notify`
- `execute`
- `archive`

草案编排节点类型：

- `human_input`
- `human_review`
- `agent_execute`
- `agent_export`
- `human_acceptance`

### 5.4 流程与节点状态

定义位置：`internal/domain/run.go`

流程状态：

- `running`
- `waiting`
- `blocked`
- `completed`
- `cancelled`

节点状态：

- `not_started`
- `ready`
- `running`
- `waiting_confirm`
- `waiting_material`
- `rejected`
- `done`
- `failed`
- `blocked`
- `cancelled`

### 5.5 龙虾任务状态

定义位置：`internal/domain/agent_task.go`

任务状态：

- `queued`
- `running`
- `completed`
- `needs_review`
- `failed`
- `blocked`
- `cancelled`

任务类型：

- `query`
- `batch_operation`
- `export`

回执状态：

- `completed`
- `needs_review`
- `failed`
- `blocked`

### 5.6 交付物审核状态

定义位置：`internal/domain/deliverable.go`

- `pending`
- `approved`
- `rejected`

### 5.7 OpenClaw 集成状态

定义位置：`internal/domain/openclaw_integration.go`

- `pending`
- `active`
- `degraded`
- `disabled`
- `revoked`

### 5.8 注册码状态

定义位置：`internal/domain/registration_code.go`

- `active`
- `used`
- `expired`
- `revoked`

## 6. DTO 视角的接口对象

接口层对象主要位于 `internal/dto`，作用是：

- 规范请求参数
- 规范响应结构
- 对实体进行裁剪、补充关联对象、展开详情

常见 DTO 分组：

- 人员：`person.go`
- 龙虾：`agent.go`
- 流程草案：`draft.go`
- 模板：`template.go`
- 流程实例：`run.go`
- 节点动作：`run_node.go`
- 评论与附件：`collaboration.go`
- 交付物：`deliverable.go`
- 龙虾任务：`agent_task.go`
- OpenClaw 集成：`openclaw_integration.go`
- 最近动态：`activity.go`

## 7. 当前数据模型的一句话总结

当前后端本质上是一个“流程模板 + 流程运行 + 人机协作 + 交付审核 + 外部龙虾桥接”的数据模型体系：

- 模板定义流程结构
- 运行实例承载流程执行态
- 节点承载责任人与龙虾执行
- 评论、附件、日志承载协作过程
- 交付物承载最终结果
- AgentTask 与 OpenClawIntegration 承载外部自动化执行能力
