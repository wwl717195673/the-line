# OpenClaw 自动注册人员与龙虾归属绑定实现计划

## 1. 背景与目标

当前 OpenClaw 接入链路已经具备：

- 注册码创建与校验
- Bridge 实例注册
- `OpenClawIntegration` 落库
- `bound_agent_id` 绑定能力

但接入时仍存在一个明显断点：

- 龙虾成功接入平台，不代表龙虾拥有者已经作为 `Person` 注册到平台
- `Agent.owner_person_id` 可能为空，导致归属关系、权限治理、通知审计都不完整

本方案目标是在 OpenClaw Bridge 注册成功时，自动完成：

1. 识别 / 注册龙虾拥有者为平台 `Person`
2. 识别 / 创建对应 `Agent`
3. 将 `Agent.owner_person_id` 绑定到该 `Person`
4. 将 `OpenClawIntegration.bound_agent_id` 绑定到该 `Agent`

从用户视角上实现“一次接入，自动完成人员与龙虾归属注册”。

---

## 2. 方案结论

推荐采用：

**在现有 `POST /api/integrations/openclaw/register` 注册链路中，扩展 owner 信息与 agent 信息，并在后端事务中完成 Person / Agent / Integration 的统一落库与绑定。**

不建议新增一条独立“先注册人员，再接入龙虾”的前置链路。

原因：

- 用户操作路径更短
- 归属信息在首次接入时即可补全
- 避免出现“龙虾已接入但 owner 为空”的脏数据
- 更容易做幂等与审计

---

## 3. 现状梳理

### 3.1 当前相关数据对象

#### Person

文件：`backend/internal/model/person.go`

当前字段：

- `id`
- `name`
- `email`
- `role_type`
- `status`

问题：

- 没有外部身份映射字段
- 无法稳定识别同一个 OpenClaw 用户是否已注册

#### Agent

文件：`backend/internal/model/agent.go`

当前字段：

- `id`
- `name`
- `code`
- `provider`
- `version`
- `owner_person_id`
- `config_json`
- `status`

现状：

- 已具备 owner 归属字段
- 已具备唯一业务标识 `code`

#### OpenClawIntegration

文件：`backend/internal/model/openclaw_integration.go`

当前字段：

- `instance_fingerprint`
- `bound_agent_id`
- `callback_url`
- `callback_secret`
- `status`

现状：

- 已能绑定 `Agent`
- 但无法直接建立“拥有者人员”关系

### 3.2 当前链路的主要问题

1. 接入和人员注册分离，用户心智不连续
2. `Agent.owner_person_id` 可能为空
3. 无法判断同一个外部拥有者是否已存在
4. 后续权限、通知、审计都只能落在龙虾层，不能落在人层

---

## 4. 核心设计

### 4.1 注册请求扩展

在当前 `BridgeRegisterRequest` 中增加拥有者与龙虾基础信息：

建议新增字段：

- `owner_name`
- `owner_email`
- `owner_role_type`
- `owner_external_id`
- `agent_name`
- `agent_code`

建议语义：

- `owner_name`：龙虾拥有者在平台侧显示名
- `owner_email`：用于匹配 / 通知 / fallback 去重
- `owner_role_type`：平台角色，可选
- `owner_external_id`：外部稳定用户标识，强烈建议传
- `agent_name`：接入龙虾显示名
- `agent_code`：龙虾业务唯一编码，建议由外部传入，避免后端猜测

### 4.2 人员匹配策略

推荐优先级：

1. `owner_external_id`
2. `owner_email`
3. 都没有则注册失败

不建议只按 `owner_name` 匹配。

### 4.3 龙虾匹配策略

推荐优先级：

1. `agent_code`
2. 没有 `agent_code` 时可回退到 `instance_fingerprint`

但工程上建议要求外部显式传 `agent_code`。

### 4.4 统一事务链路

`register` 接口内部按以下顺序执行：

1. 校验注册码
2. 校验注册请求参数
3. 查找 / 创建 `Person`
4. 查找 / 创建 `Agent`
5. 查找 / 创建 `OpenClawIntegration`
6. 补齐绑定关系：
   - `Agent.owner_person_id = Person.ID`
   - `OpenClawIntegration.bound_agent_id = Agent.ID`
7. 提交事务

---

## 5. 数据模型改造方案

### 5.1 推荐最小改造：直接扩展 `persons`

在 `persons` 表增加：

- `external_source`
- `external_user_id`

建议含义：

- `external_source`：例如 `openclaw`
- `external_user_id`：外部系统里的稳定用户 ID

建议索引：

- 普通索引：`external_source`
- 普通索引：`external_user_id`
- 联合唯一索引：`(external_source, external_user_id)`

优点：

- 实现成本最低
- 与当前仓库结构最兼容
- 足够支撑本阶段 OpenClaw 场景

缺点：

- 如果未来一个人要映射多个外部系统，扩展性一般

### 5.2 可选增强方案：独立身份映射表

新增 `person_identities`：

- `id`
- `person_id`
- `source`
- `external_user_id`
- `email_snapshot`
- `created_at`
- `updated_at`

适合未来：

- 对接多个外部用户体系
- 一个 `Person` 映射多个外部身份

本期不建议优先做，除非 OpenClaw 只是多个接入源之一。

---

## 6. 详细实现流程

### 6.1 Person Upsert

后端逻辑：

1. 若 `owner_external_id` 非空：
   - 按 `(external_source = openclaw, external_user_id = owner_external_id)` 查询
2. 若未命中且 `owner_email` 非空：
   - 按 `email` 查询
3. 若仍未命中：
   - 创建新 `Person`
4. 若命中：
   - 视情况补齐 `name` / `email` / `external_*`

补齐规则建议：

- `name` 为空时可补
- `email` 为空时可补
- 已有不同 `email` 时不要无脑覆盖，记录日志并保留原值
- `role_type` 若未传，使用默认值

### 6.2 Agent Upsert

后端逻辑：

1. 按 `agent_code` 查询
2. 不存在则创建：
   - `name = agent_name`
   - `code = agent_code`
   - `provider = openclaw`
   - `version = bridge_version / openclaw_version`
   - `owner_person_id = person.ID`
   - `status = enabled`
3. 已存在则更新：
   - 补齐 `owner_person_id`
   - 更新 `name` / `version`
   - 将接入元数据写入 `config_json`

建议写入 `config_json` 的内容：

- `instance_fingerprint`
- `callback_url`
- `bridge_version`
- `openclaw_version`
- `registration_source`

### 6.3 Integration Upsert

后端逻辑：

1. 按 `instance_fingerprint` 查询
2. 已存在则走幂等更新
3. 不存在则创建
4. 最终写入：
   - `bound_agent_id = agent.ID`
   - `display_name`
   - `callback_url`
   - `capabilities_json`
   - `status`

---

## 7. 幂等与冲突策略

### 7.1 Integration 幂等

规则：

- 同一个 `instance_fingerprint` 重复注册，视为同一个 Bridge 实例重放注册
- 不重复创建 integration
- 更新版本号、回调地址、能力集等可变字段

### 7.2 Person 幂等

规则：

- `(external_source, external_user_id)` 命中则复用同一 `Person`
- 若 external id 不存在，则尝试按 `email` fallback

### 7.3 Agent 幂等

规则：

- `agent_code` 命中则复用同一 `Agent`

### 7.4 冲突处理

需要显式报错的情况：

1. `agent_code` 已存在，但 owner 指向另一人且不允许接管
2. `owner_external_id` 指向已有人员，但注册邮箱与历史邮箱明显冲突
3. `instance_fingerprint` 已绑定另一只不一致的 Agent

建议响应：

- 返回 409 Conflict
- 给出明确冲突说明，便于接入方排查

---

## 8. 默认值与约束建议

### 8.1 owner_role_type 默认值

如果外部未传 `owner_role_type`，建议默认：

- `operation`

如果后续平台希望区分“龙虾拥有者”，可以再新增：

- `agent_owner`

本期不强制新增角色类型。

### 8.2 必填建议

建议作为注册接口必填：

- `registration_code`
- `instance_fingerprint`
- `callback_url`
- `bridge_version`
- `owner_name`
- `owner_email`
- `agent_name`
- `agent_code`

建议作为强烈推荐：

- `owner_external_id`

---

## 9. 代码改造范围

### 9.1 DTO

文件：

- `backend/internal/dto/openclaw_integration.go`

改动：

- 扩展 `BridgeRegisterRequest`

### 9.2 Model

文件：

- `backend/internal/model/person.go`

改动：

- 新增 `external_source`
- 新增 `external_user_id`

### 9.3 Repository

文件：

- `backend/internal/repository/person_repository.go`
- `backend/internal/repository/agent_repository.go`
- `backend/internal/repository/openclaw_integration_repository.go`

新增能力：

- 按 `(external_source, external_user_id)` 查人
- 按 `email` 查人
- 按 `agent_code` 查龙虾
- 按 `instance_fingerprint` 查 integration
- 事务内 upsert/update

### 9.4 Service

文件：

- `backend/internal/service/openclaw_integration_service.go`
- `backend/internal/service/person_service.go`
- `backend/internal/service/agent_service.go`

改动：

- 在 integration register 主链路中增加 upsert Person / Agent 的事务流程
- 沉淀必要的校验与冲突策略

### 9.5 DB Migration

文件：

- `backend/internal/db/migrate.go`

改动：

- 让新字段随迁移创建
- 补充索引与唯一约束

---

## 10. 建议实施步骤

### P1. 数据层改造

目标：

- `Person` 支持外部身份映射

工作项：

- `persons` 增加 `external_source` / `external_user_id`
- 增加联合唯一索引
- 调整 model 与 migration

### P2. 接口扩展

目标：

- 注册接口支持 owner / agent 信息输入

工作项：

- 扩展 `BridgeRegisterRequest`
- 更新参数校验
- 更新接口协议文档

### P3. 注册链路事务化

目标：

- 接入时自动完成人员 / 龙虾 / integration 绑定

工作项：

- `upsert person`
- `upsert agent`
- `upsert integration`
- 串成单事务

### P4. 冲突与幂等治理

目标：

- 避免重复人、重复龙虾、重复 integration

工作项：

- 实现冲突判断
- 明确 409 响应
- 记录关键日志

### P5. 前端接入页联调

目标：

- 接入表单收集 owner / agent 信息

工作项：

- 更新接入向导 / 注册表单
- 联调注册接口
- 验证自动注册效果

---

## 11. 验收标准

满足以下条件视为完成：

1. 用户首次接入 OpenClaw Bridge 时，平台自动创建或复用 `Person`
2. 平台自动创建或复用 `Agent`
3. `Agent.owner_person_id` 正确指向拥有者 `Person`
4. `OpenClawIntegration.bound_agent_id` 正确指向对应 `Agent`
5. 相同 `instance_fingerprint` 重复注册不会生成重复 integration
6. 相同 owner 重复注册不会生成重复人员
7. `agent_code` 冲突时，系统能返回明确错误
8. 整条注册链路在事务中执行，异常时不出现半成功数据

---

## 12. 不做范围

本期不做：

- 通用外部身份中心
- 多外部来源统一身份映射平台
- 自动账号登录 / SSO
- 龙虾 owner 自助认领历史 Agent
- 复杂的人员合并与冲突修复后台

这些能力可在后续平台治理阶段继续演进。

---

## 13. 一句话总结

本方案的核心是把“OpenClaw 接入”从单纯的桥接注册，升级成一条“Bridge 注册 + 人员自动注册 + 龙虾归属绑定”的统一事务链路，从而让龙虾一接入平台，就天然具备可治理、可审计、可归属的人和龙虾关系。
