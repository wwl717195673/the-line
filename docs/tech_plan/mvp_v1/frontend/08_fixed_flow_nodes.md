# 08 固定流程节点前端技术方案

---

## 1. 模块目标

定义“班主任甩班申请”9 个固定节点在前端的表单字段、展示内容和操作按钮。

MVP 采用固定节点表单映射，不做完整动态表单设计器。

---

## 2. 节点表单映射

建议维护一个前端配置：

```ts
export const teacherTransferNodeForms = {
  submit_application: SubmitApplicationForm,
  middle_office_review: ReviewForm,
  notify_teacher: NotifyTeacherForm,
  upload_contact_record: UploadContactRecordForm,
  leader_confirm_contact: ReviewForm,
  provide_receiver_list: ProvideReceiverListForm,
  operation_confirm_plan: OperationConfirmPlanForm,
  execute_transfer: ExecuteTransferForm,
  archive_result: ArchiveResultForm,
};
```

渲染规则：

* 优先按 `node_code` 匹配固定表单
* 未匹配时按 `node_type` 渲染默认表单
* 表单提交统一交给节点处理模块

---

## 3. NODE-001 小组长发起甩班申请

前端表单字段：

* `reason`：申请原因，必填，多行文本
* `class_info`：涉及班级，必填，文本或多选
* `current_teacher`：当前班主任，必填，人员选择或文本
* `expected_time`：期望处理时间，必填，日期时间
* `extra_note`：补充说明，选填，多行文本
* 附件，选填

按钮：

* 暂存
* 标记完成

校验：

* 申请原因不能为空
* 涉及班级不能为空
* 当前班主任不能为空
* 期望处理时间不能为空

完成后前端行为：

* 刷新流程详情
* 当前节点切换到“中台初审”

---

## 4. NODE-002 中台初审

前端展示：

* 上游申请单摘要
* 发起人补充材料
* `review_comment`：审核意见输入框

按钮：

* 审核通过
* 驳回
* 要求补材料

校验：

* `reason`：驳回原因必填
* `requirement`：补材料要求必填

完成后前端行为：

* 审核通过后刷新流程详情
* 当前节点切换到“通知班主任触达家长”
* 驳回或补材料后当前节点保持选中

---

## 5. NODE-003 通知班主任触达家长

前端表单字段：

* `notify_result`：通知结果，必填，多行文本
* `notify_note`：备注，选填
* 通知截图，选填附件

按钮：

* 暂存
* 标记完成

校验：

* 通知结果不能为空

完成后前端行为：

* 刷新流程详情
* 当前节点切换到“上传触达记录”

---

## 6. NODE-004 上传触达记录

前端表单字段：

* `contact_description`：触达说明，必填，多行文本
* `special_note`：特殊情况说明，选填
* 触达截图或附件，必填附件

按钮：

* 暂存
* 标记完成

校验：

* 触达说明不能为空
* 至少上传一份附件

完成后前端行为：

* 标记完成后流程详情刷新
* 当前节点切换到“小组长确认触达完成”
* 触达凭证是否有效由下一个节点统一审核

---

## 7. NODE-005 小组长确认触达完成

前端展示：

* 触达说明
* 触达附件
* 评论和补充材料
* `review_comment`：确认意见输入框

按钮：

* 审核通过
* 驳回
* 要求补材料

校验：

* `reason`：驳回原因必填
* `requirement`：补材料要求必填

完成后前端行为：

* 审核通过后切换到“提供接班名单”
* 补材料后当前节点显示 `waiting_material`

---

## 8. NODE-006 提供接班名单

前端表单字段：

* `receiver_teacher`：接班班主任，必填，人员选择或文本
* `receiver_class`：接班班级，必填，文本
* `handover_description`：承接说明，必填，多行文本
* `risk_note`：风险备注，选填
* 接班名单附件，选填

按钮：

* 暂存
* 标记完成

校验：

* 接班班主任不能为空
* 接班班级不能为空
* 承接说明不能为空

完成后前端行为：

* 刷新流程详情
* 当前节点切换到“运营确认甩班方案”

---

## 9. NODE-007 运营确认甩班方案

前端展示：

* 申请单摘要
* 初审意见
* 触达凭证
* 接班名单
* `review_comment`：运营确认意见输入框
* `final_plan`：最终甩班方案输入框

按钮：

* 审核通过
* 驳回
* 要求补材料
* 标记异常

校验：

* 审核通过时最终甩班方案不能为空
* `reason`：驳回原因必填
* `requirement`：补材料要求必填
* `reason`：标记异常时异常原因必填

完成后前端行为：

* 审核通过后切换到“执行甩班”
* 标记异常后流程状态显示为 `blocked`

---

## 10. NODE-008 执行甩班

前端表单字段：

* `execute_result`：执行结果，必填，多行文本
* `exception_note`：异常说明，选填
* 执行附件，选填

展示：

* 最终甩班方案
* 绑定龙虾信息
* 龙虾模拟执行结果

按钮：

* 暂存
* 运行龙虾
* 标记完成
* 标记异常

校验：

* 执行结果不能为空
* `reason`：标记异常时异常原因必填

完成后前端行为：

* 运行龙虾后刷新输出结果和日志
* 标记完成后切换到“输出结论并归档”

---

## 11. NODE-009 输出结论并归档

前端表单字段：

* `deliverable_summary`：交付摘要，必填，多行文本
* `archive_result`：归档结论，必填，多行文本
* 关键附件选择，选填

展示：

* 申请单摘要
* 触达凭证
* 接班名单
* 执行结果
* 异常说明

按钮：

* 暂存
* 标记完成

校验：

* 交付摘要不能为空
* 归档结论不能为空

完成后前端行为：

* 流程状态刷新为 `completed`
* 展示生成交付物入口

---

## 12. 验收标准

* 9 个固定节点都能渲染对应表单
* 未匹配节点能使用节点类型默认表单兜底
* 每个节点必填校验生效
* 每个节点成功操作后刷新流程详情
* 审核节点的驳回和补材料弹窗校验生效
* 执行节点能展示运行龙虾结果
* 最后节点完成后展示生成交付物入口
