# 08 固定流程节点后端方案

---

## 1. 模块目标

定义“班主任甩班申请”9 个固定节点在后端的初始化配置、输入校验和完成条件。

---

## 2. 固定节点配置

建议在 `domain/fixed_template.go` 中定义：

```go
type FixedNodeConfig struct {
    NodeCode         string
    NodeName         string
    NodeType         string
    SortOrder        int
    DefaultOwnerRule string
    DefaultAgentCode string
    NeedReview       bool
    RequiredFields   []string
    RequireAttachment bool
}
```

固定节点：

| 顺序 | 编码 | 名称 | 类型 | 是否审核 |
|---|---|---|---|---|
| 1 | `submit_application` | 小组长发起甩班申请 | manual | 否 |
| 2 | `middle_office_review` | 中台初审 | review | 是 |
| 3 | `notify_teacher` | 通知班主任触达家长 | notify | 否 |
| 4 | `upload_contact_record` | 上传触达记录 | archive | 否 |
| 5 | `leader_confirm_contact` | 小组长确认触达完成 | review | 是 |
| 6 | `provide_receiver_list` | 提供接班名单 | manual | 否 |
| 7 | `operation_confirm_plan` | 运营确认甩班方案 | review | 是 |
| 8 | `execute_transfer` | 执行甩班 | execute | 否 |
| 9 | `archive_result` | 输出结论并归档 | archive | 否 |

---

## 3. 节点输入校验

### 3.1 `submit_application`

必填字段：

* `reason`
* `class_info`
* `current_teacher`
* `expected_time`

完成规则：

* 必填字段全部存在
* 可直接完成并推进到 `middle_office_review`

### 3.2 `middle_office_review`

必填字段：

* 审核通过时需要 `review_comment`
* 驳回时需要 `reason`
* 要求补材料时需要 `requirement`

完成规则：

* 审核通过后推进到 `notify_teacher`

### 3.3 `notify_teacher`

必填字段：

* `notify_result`

完成规则：

* 可直接完成并推进到 `upload_contact_record`

### 3.4 `upload_contact_record`

必填字段：

* `contact_description`

附件规则：

* 至少 1 个附件绑定到当前节点

完成规则：

* 必填字段和附件校验通过后可直接完成
* 完成后推进到 `leader_confirm_contact`
* 触达凭证是否有效由下一个节点 `leader_confirm_contact` 统一审核，避免 NODE-004 和 NODE-005 重复审核

### 3.5 `leader_confirm_contact`

必填字段：

* 审核通过时建议填写 `review_comment`
* 驳回时需要 `reason`
* 要求补材料时需要 `requirement`

完成规则：

* 审核通过后推进到 `provide_receiver_list`

### 3.6 `provide_receiver_list`

必填字段：

* `receiver_teacher`
* `receiver_class`
* `handover_description`

完成规则：

* 可直接完成并推进到 `operation_confirm_plan`

### 3.7 `operation_confirm_plan`

必填字段：

* 审核通过时需要 `final_plan`
* 驳回时需要 `reason`
* 要求补材料时需要 `requirement`
* 标记异常时需要 `reason`

完成规则：

* 审核通过后推进到 `execute_transfer`

### 3.8 `execute_transfer`

必填字段：

* `execute_result`

完成规则：

* 可运行龙虾生成模拟结果
* 可直接完成并推进到 `archive_result`

### 3.9 `archive_result`

必填字段：

* `deliverable_summary`
* `archive_result`

完成规则：

* 完成后流程状态变为 `completed`

---

## 4. 输入结构建议

节点输入统一保存在 `flow_run_node.input_json`。

示例：

```json
{
  "form_data": {
    "reason": "",
    "class_info": "",
    "current_teacher": "",
    "expected_time": ""
  },
  "attachments": [],
  "comments_context": []
}
```

节点输出统一保存在 `flow_run_node.output_json`。

示例：

```json
{
  "summary": "",
  "structured_data": {},
  "decision": "",
  "logs": [],
  "next_actions": []
}
```

---

## 5. 校验实现建议

建议实现：

```go
type NodeValidator interface {
    ValidateSubmit(node FlowRunNode, input datatypes.JSON) error
    ValidateComplete(node FlowRunNode, input datatypes.JSON) error
    ValidateApprove(node FlowRunNode, req ApproveNodeRequest) error
}
```

MVP 可以先使用 `switch node.NodeCode` 实现固定校验，后续再抽象注册器。

---

## 6. 验收标准

* 9 个固定节点可初始化到 `flow_template_node`
* 发起流程后 9 个实例节点顺序正确
* 每个节点必填字段校验生效
* `upload_contact_record` 至少需要 1 个附件
* 审核节点的驳回和补材料原因校验生效
* `archive_result` 完成后流程状态变为 `completed`
