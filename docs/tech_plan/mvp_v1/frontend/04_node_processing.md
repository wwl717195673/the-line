# 04 节点处理前端技术方案

---

## 1. 模块目标

实现节点详情和节点操作。节点详情是 MVP V1 的核心交互区，承载输入、输出、附件、评论、日志和动作按钮。

---

## 2. 页面承载方式

节点详情不单独设计一级路由，默认嵌入流程详情页：

* `/runs/:runId`
* 通过 query 或本地状态保存当前选中节点，如 `?nodeId=123`

如果后续需要独立页面，可扩展：

* `/runs/:runId/nodes/:nodeId`

---

## 3. 组件拆分

建议组件：

* `RunNodeDetail`
* `NodeHeader`
* `NodeInputForm`
* `NodeOutputPanel`
* `NodeAgentPanel`
* `NodeActionBar`
* `NodeAttachmentPanel`
* `NodeCommentPanel`
* `NodeLogTimeline`
* `RejectNodeModal`
* `RequestMaterialModal`
* `FailNodeModal`
* `RunAgentResultPanel`

---

## 4. API 对接

接口：

* `GET /api/run-nodes/{id}`
* `PUT /api/run-nodes/{id}/input`
* `POST /api/run-nodes/{id}/submit`
* `POST /api/run-nodes/{id}/approve`
* `POST /api/run-nodes/{id}/reject`
* `POST /api/run-nodes/{id}/request-material`
* `POST /api/run-nodes/{id}/complete`
* `POST /api/run-nodes/{id}/fail`
* `POST /api/run-nodes/{id}/run-agent`

hook：

```ts
export function useRunNodeDetail(nodeId: string) {}
export function useSaveNodeInput() {}
export function useSubmitNode() {}
export function useApproveNode() {}
export function useRejectNode() {}
export function useRequestMaterial() {}
export function useCompleteNode() {}
export function useFailNode() {}
export function useRunAgent() {}
```

---

## 5. 动作按钮计算

优先使用后端返回的 `available_actions`。

如果后端未返回，前端按以下规则兜底：

| 动作 | 显示条件 |
|---|---|
| 暂存 | 责任人或管理员，节点未完成，流程未取消 |
| 提交确认 | 责任人或管理员，节点状态为 `ready`、`running`、`waiting_material` |
| 标记完成 | 责任人或管理员，节点无需审核 |
| 审核通过 | 审核人或管理员，节点状态为 `waiting_confirm` |
| 驳回 | 审核人或管理员，节点状态为 `waiting_confirm` |
| 要求补材料 | 审核人或管理员，节点状态为 `waiting_confirm` |
| 标记异常 | 责任人、审核人或管理员，节点未完成，流程未取消 |
| 运行龙虾 | 责任人或管理员，节点绑定启用龙虾 |

---

## 6. 表单实现

MVP 采用“节点类型默认表单 + 固定节点覆盖”的方式。

默认表单：

* `manual`：文本输入、附件、备注
* `review`：审核意见
* `notify`：通知结果、附件
* `execute`：执行结果、异常说明
* `archive`：归档摘要、附件

固定节点可覆盖字段，详见 `08_fixed_flow_nodes.md`。

校验规则：

* 暂存不校验必填
* 提交确认校验必填
* 标记完成校验必填
* 驳回必须填写原因
* 要求补材料必须填写补充要求
* 标记异常必须填写异常原因

---

## 7. 节点操作交互

暂存：

* 调用保存输入接口
* 成功后提示“已暂存”
* 不改变节点状态

提交确认：

* 校验表单
* 调用提交接口
* 成功后刷新节点详情和流程详情

审核通过：

* 可填写审核意见
* 成功后刷新流程详情，自动展示新当前节点

驳回：

* 弹窗输入驳回原因
* 原因不能为空
* 成功后节点显示 `rejected`

要求补材料：

* 弹窗输入补充要求
* 要求不能为空
* 成功后节点显示 `waiting_material`

标记异常：

* 弹窗输入异常原因
* 成功后节点显示 `failed`，流程显示 `blocked`

运行龙虾：

* 按钮进入 loading
* 调用模拟执行接口
* 成功后展示模拟输出并刷新日志

---

## 8. 只读规则

以下情况节点详情只读：

* 流程状态为 `cancelled`
* 当前用户是观察者
* 节点状态为 `done`
* 后端返回无可执行动作

只读状态下仍可展示：

* 输入
* 输出
* 附件
* 评论
* 日志

是否允许评论由后端权限决定。

---

## 9. 验收标准

* 节点详情展示信息完整
* 暂存不推进流程
* 提交确认后节点进入待确认
* 审核通过后流程进入下一节点
* 驳回必须填写原因
* 补材料必须填写要求
* 标记异常后流程显示阻塞
* 运行龙虾能展示 loading、结果和日志
* 已取消流程和已完成节点不能继续处理
