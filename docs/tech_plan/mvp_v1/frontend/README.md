# 虾线 MVP V1 前端技术方案

---

## 1. 文档目的

本目录基于 `docs/prd/mvp_v1/` 的模块化 PRD 编写，目标是把产品需求转成前端可执行的技术方案。

每个模块文档都会说明：

* 页面范围
* 路由设计
* 组件拆分
* 状态管理
* API 对接
* 表单和校验
* 权限和状态控制
* 异常处理
* 验收标准

---

## 2. 技术栈假设

当前仓库 `frontend/` 目录为空，尚未发现既有前端工程配置。因此 MVP V1 前端技术方案先按以下栈设计：

* 框架：React + TypeScript
* 构建：Vite
* 路由：React Router
* 服务端状态：TanStack Query 或同类 query cache
* 表单：React Hook Form 或同类表单库
* 组件库：Ant Design、Arco Design 或同类中后台组件库
* 请求层：基于 `fetch` 或 `axios` 封装统一 API client
* 样式：组件库主题 + CSS Modules 或普通 CSS

如果后续确定使用 Next.js，模块拆分、API hook、组件边界和状态机规则仍可复用，只需要调整路由和数据加载方式。

---

## 3. 前端模块文件

| 文件 | 模块 | 对应 PRD |
|---|---|---|
| `00_frontend_architecture.md` | 前端总体架构 | 全局 |
| `01_base_data.md` | 基础数据前端方案 | `01_base_data.md` |
| `02_template.md` | 模板前端方案 | `02_template.md` |
| `03_flow_run.md` | 流程实例前端方案 | `03_flow_run.md` |
| `04_node_processing.md` | 节点处理前端方案 | `04_node_processing.md` |
| `05_collaboration_trace.md` | 协同留痕前端方案 | `05_collaboration_trace.md` |
| `06_deliverable.md` | 交付前端方案 | `06_deliverable.md` |
| `07_pages.md` | 页面与布局前端方案 | `07_pages.md` |
| `08_fixed_flow_nodes.md` | 固定流程节点前端方案 | `08_fixed_flow_nodes.md` |
| `09_scope_acceptance.md` | 前端范围与验收 | `09_scope_acceptance.md` |

---

## 4. MVP 前端主链路

1. 管理员进入资源中心维护人员和龙虾
2. 用户进入模板中心查看“班主任甩班申请”
3. 用户点击使用模板进入流程发起页
4. 提交发起表单后进入流程详情页
5. 流程详情页展示 9 个节点的时间线
6. 用户点击当前节点，在节点详情区处理输入、附件、评论和操作
7. 节点完成后前端刷新流程详情，展示下一个待处理节点
8. 最后一个节点完成后展示生成交付物入口
9. 用户生成交付物并进入交付页验收

---

## 5. 全局实现原则

* 页面先保证业务闭环，不做复杂动效和高级可视化
* 流程详情用时间线或步骤条，不做 BPMN 画布
* 表单使用配置驱动和少量节点定制结合，不做完整动态表单系统
* 权限控制以前端隐藏按钮为体验优化，后端仍必须二次校验
* API 错误需要直接展示给用户，避免静默失败
* 所有节点操作成功后需要刷新流程详情和当前节点详情
* 已取消流程、已完成节点必须在前端呈现只读状态
