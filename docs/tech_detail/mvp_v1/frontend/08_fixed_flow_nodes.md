# 08 固定流程节点前端实现总结

## 实现范围

已按 `docs/tech_plan/mvp_v1/frontend/08_fixed_flow_nodes.md` 完成固定节点前端首版实现，覆盖：

- 9 个固定节点按 `node_code` 渲染专属表单
- 未命中固定节点时按 `node_type` 渲染默认兜底表单
- 节点必填校验、审核通过附加校验、附件必传校验
- 节点动作执行后统一刷新流程详情、节点详情、评论、附件、日志
- 执行节点（`execute_transfer`）展示龙虾执行输出

## 固定节点映射

新增文件：

- `frontend/src/components/fixedNodeForms.ts`

核心内容：

- 固定节点配置 `FIXED_NODE_FORM_CONFIGS`
- 配置字段类型 `NodeFormField`、节点配置类型 `NodeFormConfig`
- 入口方法 `resolveNodeFormConfig(nodeCode, nodeType)`

已映射节点：

1. `submit_application`
2. `middle_office_review`
3. `notify_teacher`
4. `upload_contact_record`
5. `leader_confirm_contact`
6. `provide_receiver_list`
7. `operation_confirm_plan`
8. `execute_transfer`
9. `archive_result`

## 节点类型兜底

未命中固定 `node_code` 时：

- `node_type` 包含 `review/approve` -> 默认审核意见表单
- `node_type` 包含 `execute/agent` -> 默认执行结果表单
- 其他 -> 默认处理说明表单

满足“先匹配 `node_code`，再按 `node_type` 兜底”的渲染规则。

## 工作台改造

改造文件：

- `frontend/src/components/RunNodeWorkbench.tsx`

主要调整：

- 从 JSON 文本编辑改为结构化字段表单编辑
- 根据 `resolveNodeFormConfig` 渲染固定/兜底表单
- 节点动作与表单校验绑定：
  - `complete` / `submit` 前执行必填校验
  - `approve` 前执行 `requiredOnApprove` 校验（如 `final_plan`）
  - `upload_contact_record` 完成前要求至少 1 个附件
  - 驳回/补材料/异常继续使用弹窗并校验非空
- 所有动作统一走 `exec()`，成功后刷新：
  - 节点详情
  - 节点评论
  - 节点附件
  - 节点日志
  - 流程详情（通过 `onMutated`）

## 展示能力补充

在节点工作台增加：

- 表单来源标识（固定节点 / 节点类型兜底）
- 节点表单描述文本
- 上游申请摘要（流程 `input_payload_json`）
- 执行节点 `execute_transfer` 的“龙虾模拟执行结果”展示块
- 上传触达记录节点的附件校验提示文案

## 样式改动

更新文件：

- `frontend/src/styles.css`

新增样式：

- `.node-form-grid`
- `.node-summary-card`
- `.node-form-hint`

## 当前边界

- 驳回/补材料/异常原因仍使用 `window.prompt`，未升级为自定义弹窗组件
- 上游摘要当前展示整段 JSON，未按业务字段做卡片化拆解
- 节点字段值统一按字符串处理，未引入日期/人员选择等专用控件
