# 05 协同留痕前端实现总结

## 实现范围

已按 `docs/tech_plan/mvp_v1/frontend/05_collaboration_trace.md` 补齐协同留痕前端能力，覆盖：

- 流程评论区（流程详情页）
- 节点评论区（节点工作台）
- 节点附件区（URL 绑定 + 本地文件上传）
- 节点日志时间线展示
- 评论“标记已解决”动作

## API 对接

评论：

- `GET /api/comments`
- `POST /api/comments`
- `POST /api/comments/:id/resolve`

附件：

- `GET /api/attachments`
- `POST /api/attachments`（JSON URL 绑定）
- `POST /api/attachments`（multipart 文件上传）

日志：

- `GET /api/run-nodes/:id/logs`

对应代码：

- `frontend/src/api/collaboration.ts`
- `frontend/src/api/runNodes.ts`
- `frontend/src/hooks/useCollaboration.ts`
- `frontend/src/hooks/useRunNodes.ts`

## 组件化落地

新增可复用组件：

- `CommentEditor`
- `CommentList`
- `AttachmentUploader`
- `AttachmentList`
- `NodeLogTimeline`

文件：

- `frontend/src/components/CommentEditor.tsx`
- `frontend/src/components/CommentList.tsx`
- `frontend/src/components/AttachmentUploader.tsx`
- `frontend/src/components/AttachmentList.tsx`
- `frontend/src/components/NodeLogTimeline.tsx`

## 页面接入

1. 流程详情页

- 在 `RunDetailPage` 增加流程评论区（target: `flow_run`）
- 支持评论发布、评论列表、标记已解决
- 流程取消后评论区只读

2. 节点工作台

- `RunNodeWorkbench` 使用通用组件接入：
  - 节点评论（target: `flow_run_node`）
  - 节点附件（target: `flow_run_node`）
  - 节点日志（独立日志接口）
- 任一节点动作成功后统一刷新：
  - 节点详情
  - 节点评论
  - 节点附件
  - 节点日志
  - 流程详情

## 请求层增强

为支持 multipart 上传，更新请求层：

- `requestJSON` 在 body 为 `FormData` 时不再强制设置 `Content-Type: application/json`
- 暴露 `getAPIBaseURL()`，供文件上传函数复用

文件：

- `frontend/src/lib/http.ts`

## 样式补充

新增样式：

- `flow-comment-section`
- `comment-row`
- `attachment-uploader`
- `log-error`

文件：

- `frontend/src/styles.css`

## 当前边界

- 评论附件（绑定到 comment target）前端入口暂未单独放出，当前以节点/流程上下文为主
- 附件预览仍为链接跳转，不做在线预览
- `@人` 仍按普通文本保存，不做解析和提醒
