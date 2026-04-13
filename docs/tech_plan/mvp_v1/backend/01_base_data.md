# 01 基础数据后端技术方案

---

## 1. 模块目标

实现人员和龙虾基础数据的 CRUD，为流程节点责任人、审核人、协作者和绑定龙虾提供数据来源。

---

## 2. GORM 模型

### 2.1 `Person`

```go
type Person struct {
    ID        uint64    `gorm:"primaryKey"`
    Name      string    `gorm:"size:64;not null"`
    Email     string    `gorm:"size:128;not null;index"`
    RoleType  string    `gorm:"size:64;not null;index"`
    Status    int8      `gorm:"not null;index"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### 2.2 `Agent`

```go
type Agent struct {
    ID            uint64         `gorm:"primaryKey"`
    Name          string         `gorm:"size:128;not null"`
    Code          string         `gorm:"size:64;not null;uniqueIndex"`
    Provider      string         `gorm:"size:64;not null"`
    Version       string         `gorm:"size:64;not null"`
    OwnerPersonID uint64         `gorm:"index"`
    ConfigJSON    datatypes.JSON
    Status        int8           `gorm:"not null;index"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

状态建议：

* `1`：启用
* `0`：停用

---

## 3. Gin 路由

人员：

* `GET /api/persons`
* `POST /api/persons`
* `PUT /api/persons/:id`
* `POST /api/persons/:id/disable`

龙虾：

* `GET /api/agents`
* `POST /api/agents`
* `PUT /api/agents/:id`
* `POST /api/agents/:id/disable`

---

## 4. Handler 设计

### 4.1 `PersonHandler`

方法：

* `List`
* `Create`
* `Update`
* `Disable`

职责：

* 绑定请求参数
* 调用 `PersonService`
* 返回统一响应
* 不直接写 GORM 查询

### 4.2 `AgentHandler`

方法：

* `List`
* `Create`
* `Update`
* `Disable`

职责同 `PersonHandler`。

---

## 5. Service 规则

### 5.1 人员创建

校验：

* `name` 不能为空
* `email` 不能为空
* `email` 格式合法
* `role_type` 不能为空

处理：

* 创建 `person`
* 默认 `status = 1`

### 5.2 人员更新

规则：

* 允许更新 `name`、`email`、`role_type`、`status`
* 不删除历史记录
* 停用人员不影响历史流程展示

### 5.3 人员停用

规则：

* 更新 `status = 0`
* 不删除记录
* 查询可选责任人时默认过滤停用人员

### 5.4 龙虾创建

校验：

* `name` 不能为空
* `code` 不能为空且唯一
* `provider` 不能为空
* `version` 不能为空
* `owner_person_id` 必须存在
* `config_json` 必须是合法 JSON

处理：

* 创建 `agent`
* 默认 `status = 1`

### 5.5 龙虾更新和停用

规则：

* 停用只更新 `status = 0`
* 不删除历史记录
* 节点绑定下拉默认只返回启用龙虾

---

## 6. Repository 查询

人员查询：

* 按 `status` 过滤
* 按 `keyword` 模糊匹配 `name` 和 `email`
* 默认按 `created_at desc`

龙虾查询：

* 按 `status` 过滤
* 按 `keyword` 模糊匹配 `name` 和 `code`
* 默认按 `created_at desc`

---

## 7. 验收标准

* 人员可以创建、编辑、停用、查询
* 停用人员不会出现在新节点责任人选择中
* 历史流程仍可展示停用人员姓名
* 龙虾可以创建、编辑、停用、查询
* 龙虾编码唯一校验生效
* 非法 `config_json` 不能保存
* 停用龙虾不会出现在新节点绑定选择中
