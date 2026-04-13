# 07 页面与布局前端实现总结

## 实现范围

已按 `docs/tech_plan/mvp_v1/frontend/07_pages.md` 完成页面与布局模块实现，重点覆盖：

- 统一中后台布局（顶部栏 + 左侧菜单 + 主内容区）
- 工作台页面（待办、进行中、最近动态、快速发起）
- 菜单一级/二级入口对齐
- 流程中心、资源中心、交付中心的页面入口统一

## 布局实现

新增组件：

- `frontend/src/components/Sidebar.tsx`

布局调整：

- `frontend/src/App.tsx`
  - 顶部栏保留全局快捷导航与 `ActorBar`
  - 主体改为 `app-body` 双栏布局
  - 左侧 `Sidebar` 提供一级菜单与二级菜单入口

样式支持：

- `frontend/src/styles.css`
  - 新增 `app-body`、`app-sidebar`、`menu-group`、`sub-nav`
  - 新增移动端单列降级样式

## 菜单结构对齐

左侧菜单已覆盖：

1. 工作台

- 首页工作台：`/`

2. 流程中心

- 全部流程：`/runs`
- 我发起的：`/runs/mine`
- 待我处理：`/runs/todo`

3. 模板中心

- 模板列表：`/templates`

4. 交付中心

- 全部交付：`/deliverables`
- 待验收：`/deliverables?review_status=pending`
- 已归档：`/deliverables?review_status=approved`

5. 资源中心

- 人员管理：`/resources/persons`
- 龙虾管理：`/resources/agents`

## 工作台页面实现

新增页面：

- `frontend/src/pages/DashboardPage.tsx`

展示模块：

- 我的待办（`GET /api/runs?scope=todo`）
- 进行中流程（`GET /api/runs?status=running`）
- 最近动态（`GET /api/activities/recent`）
- 快速发起入口（跳转模板中心）

对应数据层：

- `frontend/src/api/activities.ts`
- `frontend/src/hooks/useActivities.ts`

## 页面级补充

1. 流程中心二级导航

- 在 `RunListPage` 顶部新增 `/runs`、`/runs/mine`、`/runs/todo` 切换入口

2. 资源中心二级导航

- 在 `PersonListPage` / `AgentListPage` 顶部新增互跳入口

## 当前边界

- 交付中心与流程中心筛选参数当前仍以内存状态管理为主，尚未全部落 URL 持久化
- 节点详情仍采用嵌入式右侧工作台，不单独拆独立路由页（符合 MVP 承载方式）
