# 10 MVP V1 前后端联调验证记录

## 验证目标

验证 MVP V1 在真实运行环境下是否具备以下能力：

- 前端可正常安装、编译、启动
- 后端可正常编译、启动并连通数据库
- 固定模板、流程发起、节点推进、龙虾执行、交付生成、交付验收整链路可跑通
- 前端节点工作台依赖的 `available_actions` 可正确返回

## 验证环境

- 时间：2026-04-08
- 前端目录：`frontend/`
- 后端目录：`backend/`
- 前端运行地址：`http://127.0.0.1:4173`
- 后端运行地址：`http://127.0.0.1:8080`
- 数据库：Docker MySQL 8
- MySQL 容器名：`the-line-mysql`
- MySQL 端口：`3306`
- 数据库名：`the_line`

## 前端验证结果

执行结果：

- `npm install`：通过
- `npm run build`：通过
- `npm run dev -- --host 127.0.0.1 --port 4173`：通过

验证过程中修复：

- 修复 3 处 `??` 与 `||` 混用导致的 TypeScript 编译错误
- 修复发起流程后本地 actor 为空导致节点详情 `available_actions` 为空的问题

关联文件：

- `frontend/src/pages/DashboardPage.tsx`
- `frontend/src/pages/RunDetailPage.tsx`
- `frontend/src/pages/RunListPage.tsx`
- `frontend/src/pages/RunStartPage.tsx`

## 后端验证结果

执行结果：

- `go test ./...`：通过
- `go build ./cmd/api`：通过
- `go run ./cmd/api`：通过
- `GET /api/healthz`：通过，返回 `status=ok`、`database=ok`

运行环境处理：

- 本机原本无 MySQL 监听 `3306`
- 通过 Docker 拉起 `mysql:8.0`
- 启动参数：
  - `MYSQL_ROOT_PASSWORD=root`
  - `MYSQL_DATABASE=the_line`

## 联调前置数据

为保证整链路可推进，额外创建了测试数据：

- 人员
  - `id=1`，`role_type=leader`
  - `id=2`，`role_type=middle_office`
  - `id=3`，`role_type=operation`
- 龙虾
  - `id=1`
  - `code=shift_class_agent`

## 关键修复

### 1. 前端 actor 自动补齐

问题：

- 流程发起成功后，如果前端本地未设置 actor，请求节点详情时不会携带 `X-Person-ID`
- 后端因此返回 `available_actions=[]`

修复：

- 在流程发起成功后，若当前本地无 actor，则自动将 `initiator_person_id` 写入本地 actor

文件：

- `frontend/src/pages/RunStartPage.tsx`

### 2. 模板节点默认龙虾绑定同步

问题：

- 服务首次启动时若 `shift_class_agent` 尚不存在，`execute_transfer` 节点不会绑定默认龙虾
- 后续即使创建了龙虾，旧模板节点也不会自动补绑定

修复：

- 将固定模板种子从“存在即跳过”改为“幂等同步”
- 服务重启时会同步节点配置和 `default_agent_id`

文件：

- `backend/internal/db/fixed_template_seed.go`

## 整链路验证过程

验证流程：

- 新建流程：`run_id=2`
- 流程标题：`MVP 整链路流程`
- 交付物：`deliverable_id=1`

顺序验证结果：

1. `submit_application`
- 发起人完成节点
- 流程推进到 `middle_office_review`

2. `middle_office_review`
- 中台审核通过
- 流程推进到 `notify_teacher`

3. `notify_teacher`
- 中台暂存并完成
- 流程推进到 `upload_contact_record`

4. `upload_contact_record`
- 中台暂存输入
- 上传附件成功，附件 `id=1`
- 节点完成后流程推进到 `leader_confirm_contact`

5. `leader_confirm_contact`
- 发起人审核通过
- 流程推进到 `provide_receiver_list`

6. `provide_receiver_list`
- 中台暂存并完成
- 流程推进到 `operation_confirm_plan`

7. `operation_confirm_plan`
- 运营审核通过
- 流程推进到 `execute_transfer`

8. `execute_transfer`
- 运营暂存输入
- `run_agent` 成功执行
- 节点完成后流程推进到 `archive_result`

9. `archive_result`
- 运营暂存并完成
- 流程状态变为 `completed`

10. 交付物生成
- 发起人创建交付物成功
- 初始状态：`pending`

11. 交付物验收
- 中台作为验收人执行 `approved`
- 最终状态：`approved`

## 最终结果

- 流程状态：`completed`
- 当前节点：空
- 交付物状态：`approved`
- 交付详情包含节点数：`9`
- 整条主链路可跑通

## 当前已确认可用的能力

- 固定模板查询
- 流程发起
- 流程详情与节点时间线展示
- 节点动作：
  - `save_input`
  - `complete`
  - `approve`
  - `reject`
  - `request_material`
  - `fail`
  - `run_agent`
- 节点附件上传
- 节点日志写入
- 交付物生成
- 交付物验收

## 当前边界

- 本次联调以“主链路跑通”为目标，未覆盖所有异常分支
- 评论、驳回、补材料、取消流程等能力已有实现，但未在本次记录中逐项回归
- 联调数据仍保留在本地数据库中，后续如需清理应单独处理
