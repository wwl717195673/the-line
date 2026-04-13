# Backend

MVP V1 后端服务，基于 Golang + Gin + Gorm。

## 技术栈

- Go 1.22
- Gin
- Gorm
- MySQL 8

## 功能范围

- 人员管理
- 龙虾管理
- 固定模板管理
- 流程实例创建 / 查询 / 取消
- 节点详情 / 节点动作
- 评论 / 附件 / 节点日志
- 交付物生成 / 查询 / 验收
- 最近动态

## 环境变量

服务启动时读取以下配置：

- `APP_PORT`
  - 默认：`8080`
- `GIN_MODE`
  - 默认：`debug`
- `MYSQL_DSN`
  - 默认：`root:root@tcp(127.0.0.1:3306)/the_line?charset=utf8mb4&parseTime=True&loc=Local`
- `AUTO_MIGRATE`
  - 默认：`true`

## 本地数据库准备

需要可连接的 MySQL 8。

示例 Docker 启动方式：

```bash
docker run -d \
  --name the-line-mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=the_line \
  -p 3306:3306 \
  mysql:8.0
```

## 启动服务

在 `backend/` 目录执行：

```bash
go run ./cmd/api
```

默认监听：

- `http://127.0.0.1:8080`

## 编译

```bash
go build ./cmd/api
```

## 测试

```bash
go test ./...
```

## 健康检查

启动后可验证：

```bash
curl http://127.0.0.1:8080/api/healthz
```

预期返回：

```json
{
  "status": "ok",
  "database": "ok"
}
```

## 固定模板与种子数据

服务在 `AUTO_MIGRATE=true` 时会自动：

- 建表
- 初始化固定模板 `teacher_class_transfer`
- 同步固定模板节点配置

当前模板种子是幂等同步的：

- 模板节点已存在时不会跳过配置同步
- 如果后续补充了默认龙虾，重启服务后会同步 `default_agent_id`

## 主要接口

- `GET /api/healthz`
- `GET /api/persons`
- `POST /api/persons`
- `GET /api/agents`
- `POST /api/agents`
- `GET /api/templates`
- `GET /api/templates/:id`
- `POST /api/runs`
- `GET /api/runs`
- `GET /api/runs/:id`
- `POST /api/runs/:id/cancel`
- `GET /api/run-nodes/:id`
- `PUT /api/run-nodes/:id/input`
- `POST /api/run-nodes/:id/approve`
- `POST /api/run-nodes/:id/complete`
- `POST /api/run-nodes/:id/run-agent`
- `GET /api/deliverables`
- `POST /api/deliverables`
- `POST /api/deliverables/:id/review`

## 联调说明

部分接口会基于以下请求头判断当前操作者：

- `X-Person-ID`
- `X-Role-Type`

推荐联调测试身份：

- `1` / `leader`
- `2` / `middle_office`
- `3` / `operation`

## 当前状态

已完成一轮运行验证：

- `go test ./...` 通过
- `go build ./cmd/api` 通过
- 服务可成功连接 MySQL 并启动
- 主链路已完成真实联调：
  - 发起流程
  - 9 节点推进
  - 龙虾执行
  - 交付物生成
  - 交付物验收
