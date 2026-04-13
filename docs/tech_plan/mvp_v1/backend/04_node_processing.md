# 04 节点处理后端技术方案

---

## 1. 模块目标

节点处理模块负责节点详情查询、输入保存、提交、审核、驳回、要求补材料、完成、异常和龙虾模拟执行。

所有节点动作必须做权限校验、状态校验、数据校验和日志写入。

---

## 2. Gin 路由

* `GET /api/run-nodes/:id`
* `PUT /api/run-nodes/:id/input`
* `POST /api/run-nodes/:id/submit`
* `POST /api/run-nodes/:id/approve`
* `POST /api/run-nodes/:id/reject`
* `POST /api/run-nodes/:id/request-material`
* `POST /api/run-nodes/:id/complete`
* `POST /api/run-nodes/:id/fail`
* `POST /api/run-nodes/:id/run-agent`

---

## 3. Handler 设计

`RunNodeHandler` 方法：

* `Detail`
* `SaveInput`
* `Submit`
* `Approve`
* `Reject`
* `RequestMaterial`
* `Complete`
* `Fail`
* `RunAgent`

Handler 职责：

* 读取路径参数 `id`
* 绑定请求 JSON
* 读取当前用户上下文
* 调用 `RunNodeService`
* 返回统一响应

---

## 4. Service 状态规则

节点状态：

* `not_started`
* `ready`
* `running`
* `waiting_confirm`
* `waiting_material`
* `rejected`
* `done`
* `failed`

允许流转：

* `not_started -> ready`
* `ready -> running`
* `ready -> waiting_confirm`
* `running -> waiting_confirm`
* `waiting_material -> waiting_confirm`
* `waiting_confirm -> done`
* `waiting_confirm -> rejected`
* `waiting_confirm -> waiting_material`
* `rejected -> ready`
* `running -> failed`
* `failed -> ready`

MVP 可以用显式方法校验，不需要引入状态机库。

---

## 5. 节点详情

Service 方法：

```go
func (s *RunNodeService) GetDetail(ctx context.Context, nodeID uint64, actor Actor) (*RunNodeDetailDTO, error)
```

返回内容：

* 节点基础信息
* 节点输入
* 节点输出
* 责任人
* 审核人
* 绑定龙虾
* 附件列表
* 评论列表
* 日志列表
* 当前用户可执行动作 `available_actions`

`available_actions` 由后端计算，前端直接使用。

---

## 6. 输入暂存

规则：

* 当前用户必须是责任人、审核人或管理员
* 节点不能是 `done`
* 流程不能是 `cancelled`
* 只更新 `input_json`
* 不改变节点状态
* 写入暂存日志

---

## 7. 提交确认

规则：

* 当前用户必须是责任人或管理员
* 节点状态必须是 `ready`、`running` 或 `waiting_material`
* 必填输入必须通过校验
* 必需附件必须存在
* 节点状态更新为 `waiting_confirm`
* 流程状态更新为 `waiting`
* 写入提交日志

---

## 8. 审核通过

事务步骤：

1. 查询节点和流程并加锁
2. 校验当前用户是审核人或管理员
3. 校验节点状态为 `waiting_confirm`
4. 更新节点状态为 `done`
5. 写入 `completed_at`
6. 写入审核意见到 `output_json` 或日志
7. 调用串行推进服务
8. 写入审核通过日志

---

## 9. 驳回

规则：

* 当前用户必须是审核人或管理员
* 节点状态必须是 `waiting_confirm`
* 驳回原因不能为空
* 节点状态更新为 `rejected`
* 流程状态更新为 `waiting`
* 流程当前节点保持不变
* 不退回上游节点
* 写入驳回日志

---

## 10. 要求补材料

规则：

* 当前用户必须是审核人或管理员
* 节点状态必须是 `waiting_confirm`
* 补充要求不能为空
* 节点状态更新为 `waiting_material`
* 流程状态更新为 `waiting`
* 写入补材料日志
* 责任人补充后可再次提交确认

---

## 11. 标记完成

事务步骤：

1. 查询节点和流程并加锁
2. 校验当前用户是责任人或管理员
3. 校验节点状态为 `ready`、`running` 或 `waiting_material`
4. 校验必填输入和必需附件
5. 校验该节点无需审核或允许直接完成
6. 更新节点状态为 `done`
7. 写入 `completed_at`
8. 调用串行推进服务
9. 写入完成日志

---

## 12. 标记异常

规则：

* 当前用户必须是责任人、审核人或管理员
* 节点不能是 `done`
* 流程不能是 `cancelled`
* 异常原因不能为空
* 节点状态更新为 `failed`
* 流程状态更新为 `blocked`
* 写入异常日志

---

## 13. 运行龙虾

规则：

* 节点必须绑定启用龙虾
* 当前用户必须是责任人或管理员
* 节点状态为 `ready`、`running` 或 `waiting_material`
* 不调用真实 OpenClaw
* 生成模拟输出
* 写入开始执行和执行完成日志

模拟输出：

```json
{
  "summary": "龙虾模拟执行完成",
  "structured_data": {},
  "decision": "mock_success",
  "logs": ["mock agent executed"],
  "next_actions": []
}
```

---

## 14. 验收标准

* 节点详情返回 `available_actions`
* 暂存不改变节点状态
* 提交确认后状态变为 `waiting_confirm`
* 审核通过后节点完成并推进流程
* 驳回必须填写原因
* 补材料必须填写要求
* 标记异常后流程变为 `blocked`
* 运行龙虾能写入模拟输出和日志
* 所有节点动作都有权限和状态校验
* 所有节点动作都有日志记录
