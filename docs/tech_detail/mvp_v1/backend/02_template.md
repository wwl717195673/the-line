# 02 模板模块实现明细

## 1. 实现范围

本次实现对应：

* `docs/tech_plan/mvp_v1/backend/02_template.md`

实现目录：

* `backend/`

已完成内容：

* 实现模板表 `flow_templates`
* 实现模板节点表 `flow_template_nodes`
* 固化“班主任甩班申请”模板配置
* 固化 9 个固定流程节点配置
* 在自动迁移时执行模板初始化
* 实现模板列表接口
* 实现模板详情接口
* 模板接口只读，不暴露创建、编辑、发布、下线接口

## 2. 新增文件

本次新增或扩展的后端文件：

```text
backend/internal/model/flow_template.go
backend/internal/model/flow_template_node.go
backend/internal/domain/template.go
backend/internal/domain/fixed_template.go
backend/internal/db/fixed_template_seed.go
backend/internal/dto/template.go
backend/internal/repository/template_repository.go
backend/internal/service/template_service.go
backend/internal/handler/template_handler.go
backend/internal/app/router.go
backend/internal/db/migrate.go
backend/internal/repository/agent_repository.go
```

## 3. 数据模型

### 3.1 模板表

模型文件：

* `backend/internal/model/flow_template.go`

表名：

* `flow_templates`

字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uint64` | 主键 |
| `name` | `string` | 模板名称 |
| `code` | `string` | 模板编码，唯一索引 |
| `version` | `int` | 模板版本 |
| `category` | `string` | 模板分类 |
| `description` | `string` | 模板说明 |
| `status` | `string` | 模板状态，MVP 固定为 `published` |
| `created_by` | `uint64` | 创建人，MVP 初始化数据默认为 `0` |
| `created_at` | `time.Time` | 创建时间 |
| `updated_at` | `time.Time` | 更新时间 |

### 3.2 模板节点表

模型文件：

* `backend/internal/model/flow_template_node.go`

表名：

* `flow_template_nodes`

字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uint64` | 主键 |
| `template_id` | `uint64` | 模板 ID |
| `node_code` | `string` | 节点编码 |
| `node_name` | `string` | 节点名称 |
| `node_type` | `string` | 节点类型 |
| `sort_order` | `int` | 节点顺序 |
| `default_owner_rule` | `string` | 默认责任人规则 |
| `default_agent_id` | `*uint64` | 默认龙虾 ID |
| `input_schema_json` | `datatypes.JSON` | 输入 schema |
| `output_schema_json` | `datatypes.JSON` | 输出 schema |
| `config_json` | `datatypes.JSON` | 节点配置 |
| `created_at` | `time.Time` | 创建时间 |
| `updated_at` | `time.Time` | 更新时间 |

已配置唯一索引：

* `template_id + node_code`
* `template_id + sort_order`

## 4. 固定模板初始化

固定模板配置文件：

* `backend/internal/domain/fixed_template.go`

模板配置：

| 字段 | 值 |
|---|---|
| `code` | `teacher_class_transfer` |
| `name` | `班主任甩班申请` |
| `version` | `1` |
| `category` | `class_operation` |
| `status` | `published` |

初始化函数：

* `backend/internal/db/fixed_template_seed.go`
* `SeedTeacherClassTransferTemplate(database *gorm.DB) error`

初始化触发点：

* `backend/internal/db/migrate.go`
* `AutoMigrate` 完成表结构迁移后调用 `SeedTeacherClassTransferTemplate`

幂等规则：

* 如果模板 `code = teacher_class_transfer` 已存在，不重复创建模板
* 如果某个节点的 `template_id + node_code` 已存在，不重复创建节点
* 不覆盖已存在模板和节点，避免影响后续历史流程实例

## 5. 固定节点

已初始化 9 个节点：

| 顺序 | 编码 | 名称 | 类型 | 默认责任人规则 |
|---|---|---|---|---|
| 1 | `submit_application` | 小组长发起甩班申请 | `manual` | `initiator` |
| 2 | `middle_office_review` | 中台初审 | `review` | `middle_office` |
| 3 | `notify_teacher` | 通知班主任触达家长 | `notify` | `middle_office` |
| 4 | `upload_contact_record` | 上传触达记录 | `archive` | `current_owner` |
| 5 | `leader_confirm_contact` | 小组长确认触达完成 | `review` | `initiator` |
| 6 | `provide_receiver_list` | 提供接班名单 | `manual` | `middle_office` |
| 7 | `operation_confirm_plan` | 运营确认甩班方案 | `review` | `operation` |
| 8 | `execute_transfer` | 执行甩班 | `execute` | `operation` |
| 9 | `archive_result` | 输出结论并归档 | `archive` | `operation` |

节点配置写入 `config_json`：

```json
{
  "need_review": true,
  "required_fields": ["review_comment"],
  "require_attachment": false,
  "default_agent_code": ""
}
```

`execute_transfer` 节点配置了 `default_agent_code = shift_class_agent`。如果初始化模板时数据库中已经存在该编码的龙虾，则会自动写入 `default_agent_id`；如果不存在，则保持为空，避免强依赖基础数据。

## 6. API 实现

### 6.1 模板列表

接口：

```text
GET /api/templates
```

查询参数：

| 参数 | 说明 |
|---|---|
| `page` | 页码，默认 `1` |
| `page_size` | 每页条数，默认 `20`，最大 `100` |
| `keyword` | 可选，模糊匹配 `name` 和 `code` |

服务规则：

* 固定只返回 `status = published` 的模板
* 不返回草稿和下线模板
* 默认按 `created_at desc` 排序

响应结构：

```json
{
  "items": [
    {
      "id": 1,
      "name": "班主任甩班申请",
      "code": "teacher_class_transfer",
      "version": 1,
      "category": "class_operation",
      "description": "MVP 固定模板，用于跑通班主任甩班申请的人机协作闭环。",
      "status": "published",
      "created_by": 0,
      "created_at": "2026-04-07T00:00:00+08:00",
      "updated_at": "2026-04-07T00:00:00+08:00"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

### 6.2 模板详情

接口：

```text
GET /api/templates/:id
```

服务规则：

* 校验模板 ID 合法
* 模板不存在返回 `NOT_FOUND`
* 非 `published` 模板按不存在处理
* 节点按 `sort_order asc` 返回
* 如果节点绑定了 `default_agent_id`，返回 `default_agent` 简要信息

响应结构：

```json
{
  "id": 1,
  "name": "班主任甩班申请",
  "code": "teacher_class_transfer",
  "version": 1,
  "category": "class_operation",
  "description": "MVP 固定模板，用于跑通班主任甩班申请的人机协作闭环。",
  "status": "published",
  "created_by": 0,
  "created_at": "2026-04-07T00:00:00+08:00",
  "updated_at": "2026-04-07T00:00:00+08:00",
  "nodes": [
    {
      "id": 1,
      "template_id": 1,
      "node_code": "submit_application",
      "node_name": "小组长发起甩班申请",
      "node_type": "manual",
      "sort_order": 1,
      "default_owner_rule": "initiator",
      "default_agent_id": null,
      "input_schema_json": {
        "type": "object",
        "required_fields": ["reason", "class_info", "current_teacher", "expected_time"]
      },
      "output_schema_json": {
        "type": "object",
        "properties": {
          "summary": "string",
          "structured_data": "object",
          "decision": "string"
        }
      },
      "config_json": {
        "need_review": false,
        "required_fields": ["reason", "class_info", "current_teacher", "expected_time"],
        "require_attachment": false,
        "default_agent_code": ""
      },
      "created_at": "2026-04-07T00:00:00+08:00",
      "updated_at": "2026-04-07T00:00:00+08:00"
    }
  ]
}
```

## 7. 分层说明

Handler 层：

* `TemplateHandler.List` 绑定列表查询参数
* `TemplateHandler.Detail` 解析模板 ID
* Handler 不直接查询数据库

Service 层：

* `TemplateService.List` 固定过滤 `status = published`
* `TemplateService.Detail` 校验模板存在和发布状态
* `TemplateService.Detail` 组装节点列表和默认龙虾简要信息

Repository 层：

* `TemplateRepository.List` 封装模板列表查询
* `TemplateRepository.GetByID` 封装模板详情查询
* `TemplateRepository.ListNodesByTemplateID` 封装节点顺序查询
* `AgentRepository.GetByIDs` 支持模板详情批量补充默认龙虾信息

DB 层：

* `AutoMigrate` 负责迁移 `flow_templates` 和 `flow_template_nodes`
* `SeedTeacherClassTransferTemplate` 负责初始化固定模板和节点

## 8. 验证结果

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

说明：

* 当前验证覆盖编译检查
* 尚未连接真实 MySQL 执行模板初始化和接口联调
* 模板初始化会在服务启动且 `AUTO_MIGRATE=true` 时执行
