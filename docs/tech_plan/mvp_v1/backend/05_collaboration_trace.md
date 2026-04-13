# 05 协同留痕后端技术方案

---

## 1. 模块目标

实现评论、附件和节点日志的后端能力，用于沉淀流程协同过程和关键操作记录。

MVP 不做完整通知中心和完整审计日志后台。

---

## 2. GORM 模型

### 2.1 `Comment`

```go
type Comment struct {
    ID             uint64    `gorm:"primaryKey"`
    TargetType     string    `gorm:"size:32;not null;index:idx_comment_target"`
    TargetID       uint64    `gorm:"not null;index:idx_comment_target"`
    AuthorPersonID uint64    `gorm:"not null;index"`
    Content        string    `gorm:"type:text;not null"`
    IsResolved     bool      `gorm:"not null;default:false"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### 2.2 `Attachment`

```go
type Attachment struct {
    ID         uint64    `gorm:"primaryKey"`
    TargetType string    `gorm:"size:32;not null;index:idx_attachment_target"`
    TargetID   uint64    `gorm:"not null;index:idx_attachment_target"`
    FileName   string    `gorm:"size:256;not null"`
    FileURL    string    `gorm:"size:512;not null"`
    FileSize   int64
    FileType   string    `gorm:"size:64"`
    UploadedBy uint64    `gorm:"not null;index"`
    CreatedAt  time.Time
}
```

### 2.3 `FlowRunNodeLog`

```go
type FlowRunNodeLog struct {
    ID           uint64         `gorm:"primaryKey"`
    RunID        uint64         `gorm:"not null;index"`
    RunNodeID    uint64         `gorm:"not null;index"`
    LogType      string         `gorm:"size:32;not null"`
    OperatorType string         `gorm:"size:32;not null"`
    OperatorID   uint64         `gorm:"index"`
    Content      string         `gorm:"type:text;not null"`
    ExtraJSON    datatypes.JSON
    CreatedAt    time.Time
}
```

---

## 3. Gin 路由

评论：

* `GET /api/comments`
* `POST /api/comments`
* `POST /api/comments/:id/resolve`

附件：

* `GET /api/attachments`
* `POST /api/attachments`

日志：

* MVP 可随节点详情返回
* 可选：`GET /api/run-nodes/:id/logs`

---

## 4. 评论服务

### 4.1 发表评论

规则：

* `target_type` 必须是 `flow_run` 或 `flow_run_node`
* `target_id` 必须存在
* 当前用户必须有目标对象查看权限
* `content` 不能为空
* MVP 阶段不解析真实 `@人`
* MVP 阶段不发送通知

### 4.2 查询评论列表

规则：

* 按 `target_type` 和 `target_id` 查询
* 校验目标对象权限
* 按创建时间排序
* 只返回目标对象下的评论

### 4.3 标记已解决

规则：

* 评论必须存在
* 当前用户必须有处理权限
* 更新 `is_resolved = true`
* 不改变流程和节点状态

---

## 5. 附件服务

### 5.1 上传附件

规则：

* `target_type` 必须是 `flow_run`、`flow_run_node`、`comment` 或 `deliverable`
* `target_id` 必须存在
* 当前用户必须有目标对象上传权限
* 保存文件并生成 `file_url`
* 创建 `attachment` 记录
* 如果目标是节点，写入节点日志

文件存储：

* MVP 可先使用本地文件目录或对象存储抽象接口
* 数据库只保存 `file_url` 和元信息
* 不在数据库保存文件二进制内容

### 5.2 查询附件列表

规则：

* 按 `target_type` 和 `target_id` 查询
* 校验目标对象查看权限
* 不返回其他目标对象附件

---

## 6. 节点日志服务

建议接口：

```go
func (s *NodeLogService) Append(tx *gorm.DB, req AppendNodeLogRequest) error
```

入参：

* `run_id`
* `run_node_id`
* `log_type`
* `operator_type`
* `operator_id`
* `content`
* `extra_json`

规则：

* 日志只能追加
* 日志不能编辑
* 日志不能删除
* 人工动作使用 `operator_type = person`
* 龙虾模拟执行使用 `operator_type = agent`
* 系统推进使用 `operator_type = system`

---

## 7. 验收标准

* 流程评论能创建和查询
* 节点评论能创建和查询
* 流程评论和节点评论不混淆
* 评论内容为空时返回校验错误
* 附件能上传并绑定正确目标
* 附件查询只返回当前目标附件
* 节点关键动作能写入日志
* 日志能区分人工、龙虾和系统动作
