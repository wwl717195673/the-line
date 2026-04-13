# 07 页面与布局前端技术方案

---

## 1. 模块目标

统一 MVP V1 的页面布局、导航结构和页面级交互，确保各业务模块在前端有一致体验。

---

## 2. 应用布局

建议采用中后台经典布局：

* 顶部栏：Logo、当前用户、全局操作
* 左侧菜单：工作台、流程中心、模板中心、交付中心、资源中心
* 主内容区：当前页面
* 可选右侧区域：节点详情侧栏或辅助信息

---

## 3. 菜单结构

一级菜单：

* 工作台：`/`
* 流程中心：`/runs`
* 模板中心：`/templates`
* 交付中心：`/deliverables`
* 资源中心：`/resources/persons`

二级菜单：

* 流程中心：全部流程、我发起的、待我处理
* 资源中心：人员管理、龙虾管理
* 交付中心：全部交付、待验收、已归档

---

## 4. 工作台页面

页面组件：

* `DashboardPage`
* `TodoNodeCardList`
* `RunningRunList`
* `RecentActivityList`
* `StartFlowEntry`

后端依赖：

* `GET /api/runs?scope=todo`
* `GET /api/runs?status=running`
* `GET /api/activities/recent`

展示规则：

* 我的待办展示当前用户负责的待处理节点
* 进行中流程展示运行中的流程
* 最近动态展示节点日志
* 发起流程按钮跳转模板列表或发起页

---

## 5. 流程中心页面

页面组件：

* `RunListPage`
* `RunFilterBar`
* `RunTable`

后端依赖：

* `GET /api/runs`

展示规则：

* 全部流程使用 `scope=all`
* 我发起的使用 `scope=initiated_by_me`
* 待我处理使用 `scope=todo`
* 筛选条件保存在 URL query 中，便于刷新和分享

---

## 6. 流程详情页面

页面组件：

* `RunDetailPage`
* `RunSummaryPanel`
* `RunNodeTimeline`
* `RunNodeDetail`
* `RunCommentPanel`

后端依赖：

* `GET /api/runs/{id}`
* `GET /api/run-nodes/{id}`

布局规则：

* 顶部展示流程全局信息
* 中间展示节点时间线
* 右侧展示节点详情
* 默认选中当前节点
* 点击节点更新选中状态

---

## 7. 节点详情页面区域

页面组件：

* `NodeHeader`
* `NodeInputForm`
* `NodeOutputPanel`
* `NodeActionBar`
* `NodeAttachmentPanel`
* `NodeCommentPanel`
* `NodeLogTimeline`

展示规则：

* 操作按钮统一放在节点详情顶部或底部固定区域
* 输入、输出、附件、评论、日志分区展示
* 已完成或只读状态下表单禁用

---

## 8. 模板中心页面

页面组件：

* `TemplateListPage`
* `TemplateDetailPage`
* `TemplateNodeTimeline`

展示规则：

* 模板详情只读
* 不展示拖拽编辑入口
* 使用模板按钮跳转发起页

---

## 9. 资源中心页面

页面组件：

* `PersonListPage`
* `AgentListPage`

展示规则：

* 人员和龙虾使用表格 + 弹窗表单
* 停用操作必须二次确认
* 停用对象用状态标签标识

---

## 10. 交付中心页面

页面组件：

* `DeliverableListPage`
* `DeliverableDetailPage`

展示规则：

* 待验收和已归档可以使用 tab 或筛选
* 交付详情页展示验收状态和验收操作
* 不展示导出按钮

---

## 11. 验收标准

* 左侧菜单能进入所有 MVP 页面
* 工作台能看到待办、进行中流程和最近动态
* 流程详情默认选中当前节点
* 节点详情操作区清晰可见
* 模板详情只读
* 资源中心能维护人员和龙虾
* 交付中心能查看和验收交付物
