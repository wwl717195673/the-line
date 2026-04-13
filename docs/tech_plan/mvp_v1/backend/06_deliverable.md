# 06 交付后端技术方案

---

## 1. 模块目标

实现流程完成后的交付物生成、查询和验收。

交付模块只影响交付物状态，不反向修改流程节点历史记录。

---

## 2. GORM 模型

```go
type Deliverable struct {
    ID               uint64         `gorm:"primaryKey"`
    RunID            uint64         `gorm:"not null;index"`
    Title            string         `gorm:"size:256;not null"`
    Summary          string         `gorm:"type:text"`
    ResultJSON       datatypes.JSON
    ReviewerPersonID uint64         `gorm:"index"`
    ReviewStatus     string         `gorm:"size:32;not null;index"`
    ReviewedAt       *time.Time
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

状态：

* `pending`
* `approved`
* `rejected`

---

## 3. Gin 路由

* `GET /api/deliverables`
* `POST /api/deliverables`
* `GET /api/deliverables/:id`
* `POST /api/deliverables/:id/review`

---

## 4. 生成交付物

Service 方法：

```go
func (s *DeliverableService) Create(ctx context.Context, req CreateDeliverableRequest, actor Actor) (*DeliverableDTO, error)
```

输入：

* `run_id`
* `title`
* `summary`
* `result_json`
* `reviewer_person_id`
* `attachment_ids`

规则：

* 流程必须存在
* 流程状态必须是 `completed`
* 当前用户必须是发起人、管理员或最终节点责任人
* 标题不能为空
* 摘要不能为空
* 验收人必须存在且启用
* 对 `attachment_ids` 对应的附件创建交付物附件绑定
* 初始 `review_status = pending`
* MVP 不自动生成 PDF 或 Excel

事务步骤：

1. 查询流程和节点结果
2. 校验权限和流程状态
3. 汇总节点完成情况
4. 汇总关键附件
5. 创建 `deliverable`
6. 复制关键附件绑定记录到交付物
7. 返回交付详情

附件绑定规则：

* 不新增 `deliverable_attachment` 关联表
* 不移动原附件记录，原附件仍保留在流程、节点或评论下
* 对用户选中的 `attachment_ids`，复制附件元信息并创建新的 `attachment` 记录
* 新记录使用 `target_type = deliverable`、`target_id = deliverable.id`
* 新记录复用原 `file_url`，不复制文件二进制内容

---

## 5. 查询交付列表

查询参数：

* `review_status`
* `reviewer_person_id`
* `page`
* `page_size`

规则：

* 默认按 `created_at desc`
* `pending` 用于待验收列表
* `approved` 用于已归档列表
* 返回关联流程标题、发起人和验收人摘要信息

---

## 6. 查询交付详情

返回内容：

* 交付基础信息
* 关联流程信息
* 节点完成情况
* 关键附件
* 验收人
* 验收状态
* 验收意见

数据来源：

* `deliverable`
* `flow_run`
* `flow_run_node`
* `attachment`
* `person`

---

## 7. 验收交付物

接口：`POST /api/deliverables/:id/review`

输入：

* `review_status`
* `review_comment`

规则：

* 交付物必须存在
* 当前用户必须是验收人或管理员
* `review_status` 只能是 `approved` 或 `rejected`
* 写入 `reviewer_person_id`
* 写入 `reviewed_at`
* 验收意见写入 `result_json` 或扩展字段
* 不改变流程状态
* 不重开流程

---

## 8. 验收标准

* 已完成流程能生成交付物
* 未完成流程生成交付物返回状态错误
* 交付列表支持待验收和已归档筛选
* 交付详情能返回节点完成情况和关键附件
* 验收人能通过交付物
* 验收人能驳回交付物
* 验收操作不改变流程状态
