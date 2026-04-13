# 01 基础数据前端实现总结

## 实现范围

已按 `docs/tech_plan/mvp_v1/frontend/README.md` 与 `docs/tech_plan/mvp_v1/frontend/01_base_data.md` 完成前端基础数据模块首版实现，落地在 `frontend/` 目录。

实现模块包含：

- 人员管理页：列表、筛选、分页、新建、编辑、停用
- 龙虾管理页：列表、筛选、分页、新建、编辑、停用
- 公共选择器：`PersonSelect`、`AgentSelect`
- 统一请求层与错误处理

## 工程初始化

新建前端工程骨架（React + TypeScript + Vite）：

- `frontend/package.json`
- `frontend/vite.config.ts`
- `frontend/index.html`
- `frontend/tsconfig.json`
- `frontend/tsconfig.app.json`
- `frontend/tsconfig.node.json`
- `frontend/src/vite-env.d.ts`

## 路由与页面

已实现路由：

- `/resources/persons`：人员管理
- `/resources/agents`：龙虾管理

对应文件：

- `frontend/src/App.tsx`
- `frontend/src/pages/PersonListPage.tsx`
- `frontend/src/pages/AgentListPage.tsx`

## API 对接

人员 API：

- `GET /api/persons`
- `POST /api/persons`
- `PUT /api/persons/:id`
- `POST /api/persons/:id/disable`

龙虾 API：

- `GET /api/agents`
- `POST /api/agents`
- `PUT /api/agents/:id`
- `POST /api/agents/:id/disable`

前端实现文件：

- `frontend/src/api/persons.ts`
- `frontend/src/api/agents.ts`
- `frontend/src/lib/http.ts`

说明：

- 支持 `VITE_API_BASE_URL` 环境变量；默认 `http://localhost:8080`
- 分页响应按后端 `PageData` 结构解析：`items/total/page/page_size`
- 统一解析后端错误响应 `code/message`

## 组件与状态

已实现组件：

- `frontend/src/components/StatusTag.tsx`
- `frontend/src/components/Modal.tsx`
- `frontend/src/components/PersonFormModal.tsx`
- `frontend/src/components/AgentFormModal.tsx`
- `frontend/src/components/PersonSelect.tsx`
- `frontend/src/components/AgentSelect.tsx`

已实现 hooks：

- `frontend/src/hooks/usePersons.ts`
- `frontend/src/hooks/useAgents.ts`

状态覆盖：

- 列表筛选与分页
- 查询/刷新
- 弹窗开关
- 新建/编辑态切换
- 停用确认
- 请求 loading/error 状态

## 校验与交互规则

人员表单：

- 姓名必填
- 邮箱必填且格式合法
- 角色必填
- 新建时默认启用，编辑时可改状态

龙虾表单：

- 名称、编码、来源、版本必填
- 维护人必填（使用 `PersonSelect`）
- `config_json` 必须是合法 JSON
- 新建时默认启用，编辑时可改状态

停用动作：

- 二次确认后调用 disable 接口
- 成功后刷新列表

## 样式与响应式

已实现基础样式与移动端适配：

- `frontend/src/styles.css`
- 桌面端表格 + 移动端横向滚动降级
- 弹窗在小屏幕自动单列布局

## 当前边界

- 当前未引入 TanStack Query / React Hook Form，先以轻量 hooks + 受控表单完成 MVP
- 当前未接入鉴权 Header（如 `X-Person-ID`），后续联调节点类接口时再统一补充
- 当前未执行 `npm install`/`npm run build` 编译验证（代码已落地，待安装依赖后验证）

## 联调建议

1. 后端启动：`backend/cmd/api`
2. 前端目录安装依赖：`frontend` 下执行 `npm install`
3. 启动前端：`npm run dev`
4. 如后端地址非默认值，设置 `.env`：
   - `VITE_API_BASE_URL=http://<host>:<port>`
