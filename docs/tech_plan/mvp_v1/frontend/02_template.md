# 02 模板前端技术方案

---

## 1. 模块目标

实现固定模板的只读展示，并提供从模板发起流程的入口。

MVP 只支持“班主任甩班申请”模板，不做拖拽式模板设计器。

---

## 2. 页面与路由

| 路由 | 页面 | 说明 |
|---|---|---|
| `/templates` | 模板列表页 | 展示已发布固定模板 |
| `/templates/:templateId` | 模板详情页 | 展示模板节点清单 |
| `/templates/:templateId/start` | 流程发起页 | 基于模板发起流程 |

---

## 3. 组件拆分

建议组件：

* `TemplateListPage`
* `TemplateTable`
* `TemplateDetailPage`
* `TemplateNodeTimeline`
* `TemplateNodeCard`
* `StartRunButton`
* `RunStartForm`

---

## 4. API 对接

接口：

* `GET /api/templates`
* `GET /api/templates/{id}`
* `POST /api/runs`

前端 hook：

```ts
export function useTemplates(params: TemplateQuery) {}
export function useTemplateDetail(templateId: string) {}
export function useStartRun() {}
```

---

## 5. 模板列表页实现

展示字段：

* 模板名称
* 模板编码
* 模板版本
* 模板分类
* 模板状态
* 模板说明
* 更新时间
* 操作

操作：

* 查看详情
* 使用模板

实现规则：

* 默认只查询 `published` 模板
* 不展示新建、编辑、发布、下线按钮
* 使用模板跳转 `/templates/:templateId/start`

---

## 6. 模板详情页实现

展示内容：

* 模板基础信息
* 节点时间线
* 节点名称
* 节点类型
* 默认责任人规则
* 默认绑定龙虾
* 输入输出结构摘要

实现规则：

* 页面只读
* 节点按 `sort_order` 展示
* 不提供拖拽交互
* 不提供节点编辑入口
* 不展示连线配置

---

## 7. 流程发起页实现

表单字段：

* `title`：实例标题
* `reason`：申请原因
* `class_info`：涉及班级
* `current_teacher`：当前班主任
* `expected_time`：期望处理时间
* `extra_note`：补充说明
* 附件

提交逻辑：

* 校验必填字段
* 组装 `input_payload_json`
* 调用 `POST /api/runs`
* 成功后跳转 `/runs/:runId`
* 失败后展示后端错误

---

## 8. 异常处理

* 模板不存在：展示“模板不存在或已下线”
* 模板未发布：隐藏使用模板按钮
* 发起失败：保留表单内容并展示错误
* 网络失败：展示重试按钮

---

## 9. 验收标准

* 模板列表能看到“班主任甩班申请”
* 模板详情能展示 9 个节点
* 模板详情页没有编辑和拖拽入口
* 用户能从模板进入发起页
* 发起页必填项校验有效
* 发起成功后跳转流程详情页
