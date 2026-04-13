# 02 模板后端技术方案

---

## 1. 模块目标

实现固定模板“班主任甩班申请”的查询和初始化，为流程发起提供模板节点来源。

MVP 不实现前台模板编辑、模板发布、模板版本管理和流程图编排。

---

## 2. GORM 模型

### 2.1 `FlowTemplate`

```go
type FlowTemplate struct {
    ID          uint64    `gorm:"primaryKey"`
    Name        string    `gorm:"size:128;not null"`
    Code        string    `gorm:"size:64;not null;uniqueIndex"`
    Version     int       `gorm:"not null"`
    Category    string    `gorm:"size:64;index"`
    Description string    `gorm:"type:text"`
    Status      string    `gorm:"size:32;not null;index"`
    CreatedBy   uint64
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### 2.2 `FlowTemplateNode`

```go
type FlowTemplateNode struct {
    ID               uint64         `gorm:"primaryKey"`
    TemplateID       uint64         `gorm:"not null;index"`
    NodeCode         string         `gorm:"size:64;not null;index"`
    NodeName         string         `gorm:"size:128;not null"`
    NodeType         string         `gorm:"size:32;not null;index"`
    SortOrder        int            `gorm:"not null;index"`
    DefaultOwnerRule string         `gorm:"size:128"`
    DefaultAgentID   *uint64        `gorm:"index"`
    InputSchemaJSON  datatypes.JSON
    OutputSchemaJSON datatypes.JSON
    ConfigJSON       datatypes.JSON
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

建议唯一索引：

* `template_id + node_code`
* `template_id + sort_order`

---

## 3. Gin 路由

* `GET /api/templates`
* `GET /api/templates/:id`

MVP 不提供：

* `POST /api/templates`
* `PUT /api/templates/:id`
* `POST /api/templates/:id/publish`
* `POST /api/templates/:id/offline`

---

## 4. Handler 设计

`TemplateHandler` 方法：

* `List`
* `Detail`

职责：

* 绑定筛选参数
* 调用 `TemplateService`
* 返回模板基础信息和节点列表

---

## 5. Service 规则

### 5.1 查询模板列表

规则：

* 默认只返回 `status = published`
* 支持 `keyword` 查询
* 返回模板基础信息
* 不返回草稿和下线模板

### 5.2 查询模板详情

规则：

* 校验模板存在
* 查询模板节点
* 节点按 `sort_order asc` 返回
* 返回默认责任人规则和默认龙虾信息

### 5.3 固定模板初始化

建议在 `domain/fixed_template.go` 中定义固定模板：

* 模板编码：`teacher_class_transfer`
* 模板名称：`班主任甩班申请`
* 模板版本：`1`
* 模板状态：`published`

初始化服务：

```go
func SeedTeacherClassTransferTemplate(db *gorm.DB) error
```

处理规则：

* 如果模板不存在则创建
* 如果模板节点不存在则创建
* 如果已存在则不重复插入
* 不覆盖历史流程实例

---

## 6. 节点类型

MVP 支持：

* `manual`
* `review`
* `notify`
* `execute`
* `archive`

不支持：

* `data`
* `rule`
* `ai`
* 复杂自定义 DSL

---

## 7. 验收标准

* `GET /api/templates` 能返回“班主任甩班申请”
* `GET /api/templates/:id` 能返回 9 个节点
* 节点按 `sort_order` 升序返回
* 前台不能通过 API 创建或编辑模板
* 初始化脚本可重复执行且不重复插入节点
* 发起流程时能读取模板节点生成实例节点
