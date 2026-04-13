# 01 基础数据模块实现明细

## 1. 实现范围

本次实现对应：

* `docs/tech_plan/mvp_v1/backend/README.md`
* `docs/tech_plan/mvp_v1/backend/01_base_data.md`

实现目录：

* `backend/`

已完成内容：

* 初始化 Go 后端工程
* 接入 Gin HTTP 服务
* 接入 GORM 和 MySQL Driver
* 实现人员 `Person` 基础数据 CRUD
* 实现龙虾 `Agent` 基础数据 CRUD
* 实现人员和龙虾的停用能力
* 实现统一错误响应
* 实现列表分页、状态过滤和关键字搜索
* 实现 GORM 自动迁移入口

## 2. 工程结构

当前后端工程结构：

```text
backend/
  go.mod
  go.sum
  cmd/
    api/
      main.go
  internal/
    app/
      router.go
      server.go
    config/
      config.go
    db/
      mysql.go
      migrate.go
    domain/
      status.go
    dto/
      common.go
      person.go
      agent.go
    handler/
      person_handler.go
      agent_handler.go
    model/
      person.go
      agent.go
    repository/
      person_repository.go
      agent_repository.go
    response/
      errors.go
      response.go
    service/
      person_service.go
      agent_service.go
```

## 3. 启动方式

默认配置来自环境变量：

```text
APP_PORT=8080
GIN_MODE=debug
MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/the_line?charset=utf8mb4&parseTime=True&loc=Local
AUTO_MIGRATE=true
```

启动命令：

```bash
cd backend
go run ./cmd/api
```

如果本地 MySQL 配置不同，需要覆盖 `MYSQL_DSN`。

## 4. 数据模型

### 4.1 人员表

模型文件：

* `backend/internal/model/person.go`

表名：

* `persons`

字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uint64` | 主键 |
| `name` | `string` | 姓名 |
| `email` | `string` | 邮箱 |
| `role_type` | `string` | 角色类型 |
| `status` | `int8` | 状态，`1` 启用，`0` 停用 |
| `created_at` | `time.Time` | 创建时间 |
| `updated_at` | `time.Time` | 更新时间 |

### 4.2 龙虾表

模型文件：

* `backend/internal/model/agent.go`

表名：

* `agents`

字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | `uint64` | 主键 |
| `name` | `string` | 名称 |
| `code` | `string` | 编码，唯一索引 |
| `provider` | `string` | 提供方 |
| `version` | `string` | 版本 |
| `owner_person_id` | `uint64` | 负责人 ID |
| `config_json` | `datatypes.JSON` | 配置 JSON |
| `status` | `int8` | 状态，`1` 启用，`0` 停用 |
| `created_at` | `time.Time` | 创建时间 |
| `updated_at` | `time.Time` | 更新时间 |

## 5. API 实现

### 5.1 人员接口

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET` | `/api/persons` | 人员列表 |
| `POST` | `/api/persons` | 创建人员 |
| `PUT` | `/api/persons/:id` | 更新人员 |
| `POST` | `/api/persons/:id/disable` | 停用人员 |

列表查询参数：

| 参数 | 说明 |
|---|---|
| `page` | 页码，默认 `1` |
| `page_size` | 每页条数，默认 `20`，最大 `100` |
| `status` | 可选，`1` 启用，`0` 停用 |
| `keyword` | 可选，模糊匹配 `name` 和 `email` |

创建请求：

```json
{
  "name": "张三",
  "email": "zhangsan@example.com",
  "role_type": "operator"
}
```

更新请求：

```json
{
  "name": "张三",
  "email": "zhangsan@example.com",
  "role_type": "operator",
  "status": 1
}
```

更新接口所有字段都是可选字段。

### 5.2 龙虾接口

| 方法 | 路径 | 说明 |
|---|---|---|
| `GET` | `/api/agents` | 龙虾列表 |
| `POST` | `/api/agents` | 创建龙虾 |
| `PUT` | `/api/agents/:id` | 更新龙虾 |
| `POST` | `/api/agents/:id/disable` | 停用龙虾 |

列表查询参数：

| 参数 | 说明 |
|---|---|
| `page` | 页码，默认 `1` |
| `page_size` | 每页条数，默认 `20`，最大 `100` |
| `status` | 可选，`1` 启用，`0` 停用 |
| `keyword` | 可选，模糊匹配 `name` 和 `code` |

创建请求：

```json
{
  "name": "甩班执行龙虾",
  "code": "shift_class_agent",
  "provider": "openclaw_mock",
  "version": "v1",
  "owner_person_id": 1,
  "config_json": {
    "mode": "mock"
  }
}
```

更新请求：

```json
{
  "name": "甩班执行龙虾",
  "code": "shift_class_agent",
  "provider": "openclaw_mock",
  "version": "v1",
  "owner_person_id": 1,
  "config_json": {
    "mode": "mock"
  },
  "status": 1
}
```

更新接口所有字段都是可选字段。

## 6. 校验规则

人员创建和更新：

* `name` 不能为空
* `email` 不能为空
* `email` 必须是合法邮箱格式
* `role_type` 不能为空
* `status` 只能是 `0` 或 `1`

人员停用：

* 将 `status` 更新为 `0`
* 不删除人员记录

龙虾创建和更新：

* `name` 不能为空
* `code` 不能为空
* `code` 必须唯一
* `provider` 不能为空
* `version` 不能为空
* `owner_person_id` 不能为空
* `owner_person_id` 必须指向已存在人员
* `config_json` 为空时默认写入 `{}`
* `config_json` 非空时必须是合法 JSON
* `status` 只能是 `0` 或 `1`

龙虾停用：

* 将 `status` 更新为 `0`
* 不删除龙虾记录

## 7. 响应格式

普通成功响应直接返回资源对象。

列表成功响应：

```json
{
  "items": [],
  "total": 0,
  "page": 1,
  "page_size": 20
}
```

错误响应：

```json
{
  "code": "VALIDATION_ERROR",
  "message": "人员姓名不能为空",
  "details": {}
}
```

已实现错误码：

* `VALIDATION_ERROR`
* `NOT_FOUND`
* `CONFLICT`
* `INTERNAL_ERROR`

## 8. 分层说明

Handler 层：

* 绑定 query 和 body 参数
* 解析 path id
* 调用 Service
* 返回统一响应

Service 层：

* 实现字段校验
* 实现邮箱和 JSON 校验
* 实现龙虾编码唯一校验
* 实现龙虾负责人存在性校验
* 组装 Repository 查询条件

Repository 层：

* 封装 GORM 查询
* 支持列表过滤和分页
* 支持创建、按 ID 查询、更新、唯一性检查

Model 层：

* 定义 `Person` 和 `Agent` 两个 GORM 模型
* 表结构通过 `AutoMigrate` 自动迁移

## 9. 验证结果

已执行：

```bash
cd backend
go mod tidy
gofmt -w ./cmd ./internal
go test ./...
go build ./cmd/api
```

验证结果：

* `go mod tidy` 通过，并生成 `go.sum`
* `gofmt` 通过
* `go test ./...` 通过
* `go build ./cmd/api` 通过

说明：

* `go build ./cmd/api` 会在 `backend/` 下生成临时二进制 `api`
* 本次验证后已删除该临时二进制
* 当前验证覆盖编译检查，尚未连接真实 MySQL 执行接口级联调
