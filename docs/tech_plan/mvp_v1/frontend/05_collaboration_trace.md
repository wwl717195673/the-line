# 05 协同留痕前端技术方案

---

## 1. 模块目标

实现评论、附件和节点日志的前端能力，让协同信息沉淀在流程和节点上下文中。

---

## 2. 页面承载

评论：

* 流程详情页 - 流程评论区
* 节点详情页 - 节点评论区

附件：

* 节点详情页 - 附件区域
* 评论区 - 评论附件
* 交付页 - 关键附件

日志：

* 节点详情页 - 执行日志区域

---

## 3. 组件拆分

建议组件：

* `CommentList`
* `CommentEditor`
* `CommentItem`
* `AttachmentUploader`
* `AttachmentList`
* `AttachmentItem`
* `NodeLogTimeline`
* `LogItem`

这些组件需要支持 `target_type` 和 `target_id` 参数，以便复用在流程、节点、评论和交付物上。

---

## 4. API 对接

评论接口：

* `POST /api/comments`
* `GET /api/comments`
* `POST /api/comments/{id}/resolve`

附件接口：

* `POST /api/attachments`
* `GET /api/attachments`

日志接口：

* 日志通常随节点详情返回
* 如果独立查询，建议 `GET /api/run-nodes/{id}/logs`

hook：

```ts
export function useComments(targetType: string, targetId: string) {}
export function useCreateComment() {}
export function useResolveComment() {}
export function useAttachments(targetType: string, targetId: string) {}
export function useUploadAttachment() {}
```

---

## 5. 评论实现

评论编辑器：

* 多行文本输入
* 支持输入 `@张三`，MVP 阶段按普通文本保存
* 提交按钮
* 提交中 loading

评论列表：

* 展示作者
* 展示内容
* 展示创建时间
* 展示已解决状态
* 展示标记已解决按钮

规则：

* 评论内容不能为空
* 评论发布成功后清空输入框
* 评论发布成功后刷新评论列表
* 流程评论和节点评论不能混用
* MVP 不做评论编辑和删除

---

## 6. 附件实现

上传组件：

* 支持选择文件
* 展示上传中状态
* 上传成功后刷新附件列表
* 上传失败后保留页面上下文

附件列表：

* 文件名
* 文件类型
* 文件大小
* 上传人
* 上传时间
* 下载或查看链接

规则：

* 上传时必须传入 `target_type` 和 `target_id`
* 节点附件只显示在对应节点
* 评论附件只显示在对应评论
* 交付附件只显示在交付页
* MVP 不做在线预览和在线编辑

---

## 7. 日志实现

日志展示：

* 时间线样式
* 日志类型标签
* 操作者类型标签
* 操作者名称
* 日志内容
* 创建时间

展示规则：

* `operator_type=person` 显示为人工操作
* `operator_type=agent` 显示为龙虾执行
* `operator_type=system` 显示为系统动作
* `log_type=error` 需要高亮
* 日志只读

---

## 8. 状态刷新规则

评论发布成功：

* 失效当前目标评论查询
* 如评论数量在节点详情中展示，也刷新节点详情

附件上传成功：

* 失效当前目标附件查询
* 如果附件影响节点必填校验，刷新节点详情

节点操作成功：

* 刷新节点日志
* 刷新节点详情
* 刷新流程详情

---

## 9. 验收标准

* 流程评论能发布并只展示在流程评论区
* 节点评论能发布并只展示在对应节点
* 评论内容为空时不能提交
* 评论能标记已解决
* 附件能上传并绑定到正确目标
* 上传失败不会清空节点表单
* 节点日志能展示关键操作
* 龙虾日志和人工日志能区分显示
