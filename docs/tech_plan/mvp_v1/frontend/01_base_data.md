# 01 基础数据前端技术方案

---

## 1. 模块目标

实现人员和龙虾的基础管理页面，为流程发起、节点责任人分配和节点绑定龙虾提供基础数据。

---

## 2. 页面与路由

| 路由 | 页面 | 说明 |
|---|---|---|
| `/resources/persons` | 人员管理页 | 人员列表、新建、编辑、停用 |
| `/resources/agents` | 龙虾管理页 | 龙虾列表、新建、编辑、停用 |

---

## 3. 组件拆分

建议组件：

* `PersonListPage`
* `PersonTable`
* `PersonFormModal`
* `AgentListPage`
* `AgentTable`
* `AgentFormModal`
* `StatusTag`
* `PersonSelect`
* `AgentSelect`

`PersonSelect` 和 `AgentSelect` 需要作为公共组件，被节点详情、模板详情、交付验收人选择复用。

---

## 4. 前端状态

人员管理页状态：

* 当前筛选条件
* 当前分页
* 创建/编辑弹窗状态
* 当前编辑人员
* 停用确认弹窗状态

龙虾管理页状态：

* 当前筛选条件
* 当前分页
* 创建/编辑弹窗状态
* 当前编辑龙虾
* 停用确认弹窗状态

服务端缓存：

* `persons.list`
* `persons.detail`
* `agents.list`
* `agents.detail`

---

## 5. API 对接

人员 API：

* `GET /api/persons`
* `POST /api/persons`
* `PUT /api/persons/{id}`
* `POST /api/persons/{id}/disable`

龙虾 API：

* `GET /api/agents`
* `POST /api/agents`
* `PUT /api/agents/{id}`
* `POST /api/agents/{id}/disable`

前端封装建议：

```ts
export function usePersons(params: PersonQuery) {}
export function useCreatePerson() {}
export function useUpdatePerson() {}
export function useDisablePerson() {}

export function useAgents(params: AgentQuery) {}
export function useCreateAgent() {}
export function useUpdateAgent() {}
export function useDisableAgent() {}
```

---

## 6. 表单校验

人员表单：

* 姓名必填
* 邮箱必填
* 邮箱格式合法
* 默认角色类型必填
* 状态必填

龙虾表单：

* 龙虾名称必填
* 龙虾编码必填
* 来源默认 `openclaw`
* 版本必填
* 维护人必填
* 配置快照如果填写，必须是合法 JSON
* 状态必填

---

## 7. 交互细节

人员停用：

* 点击停用后弹出确认框
* 停用成功后刷新人员列表
* 停用人员不再出现在新节点责任人下拉中
* 历史流程中的人员展示不受影响

龙虾停用：

* 点击停用后弹出确认框
* 停用成功后刷新龙虾列表
* 停用龙虾不再出现在新节点绑定下拉中
* 历史节点绑定龙虾展示不受影响

列表空状态：

* 没有人员时提示“暂无人员”
* 没有龙虾时提示“暂无龙虾”

---

## 8. 验收标准

* 人员列表可查看、筛选、刷新
* 人员可新建、编辑、停用
* 邮箱格式错误时不能提交
* 停用人员不会出现在新责任人选择中
* 龙虾列表可查看、筛选、刷新
* 龙虾可新建、编辑、停用
* 配置快照 JSON 非法时不能提交
* 停用龙虾不会出现在新节点绑定选择中
