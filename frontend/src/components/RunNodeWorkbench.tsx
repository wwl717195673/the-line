import { useEffect, useMemo, useState } from "react";
import AttachmentList from "./AttachmentList";
import AttachmentUploader from "./AttachmentUploader";
import CommentEditor from "./CommentEditor";
import CommentList from "./CommentList";
import NodeLogTimeline from "./NodeLogTimeline";
import { resolveNodeFormConfig, type NodeFormField } from "./fixedNodeForms";
import { useAgentTaskSnapshot } from "../hooks/useAgentTasks";
import { useCollaborationActions, useComments, useAttachments } from "../hooks/useCollaboration";
import { useRunNodeActions, useRunNodeDetail, useRunNodeLogs } from "../hooks/useRunNodes";

type RunNodeWorkbenchProps = {
  nodeID?: number;
  runStatus?: string;
  onMutated: () => Promise<void>;
};

function readRecord(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }
  return value as Record<string, unknown>;
}

function toTextValue(value: unknown): string {
  if (value === undefined || value === null) {
    return "";
  }
  return String(value);
}

function buildPayload(values: Record<string, string>): Record<string, string> {
  const payload: Record<string, string> = {};
  Object.entries(values).forEach(([key, value]) => {
    if (!value.trim()) {
      return;
    }
    payload[key] = value.trim();
  });
  return payload;
}

function validateRequired(values: Record<string, string>, fields: NodeFormField[]): string | null {
  for (const field of fields) {
    if (field.required && !values[field.key]?.trim()) {
      return `${field.label}不能为空`;
    }
  }
  return null;
}

function validateRequiredOnApprove(values: Record<string, string>, fields: NodeFormField[]): string | null {
  for (const field of fields) {
    if (field.requiredOnApprove && !values[field.key]?.trim()) {
      return `${field.label}不能为空`;
    }
  }
  return null;
}

function formatJSON(value: unknown): string {
  try {
    return JSON.stringify(value ?? {}, null, 2);
  } catch {
    return "{}";
  }
}

function renderResultSummary(result: Record<string, unknown>): string[] {
  const lines: string[] = [];
  const summary = result.summary;
  if (typeof summary === "string" && summary.trim()) {
    lines.push(summary.trim());
  }
  const successCount = result.success_count;
  const failedCount = result.failed_count;
  if (typeof successCount === "number" || typeof failedCount === "number") {
    lines.push(`成功 ${successCount ?? 0} 条 / 失败 ${failedCount ?? 0} 条`);
  }
  const records = result.records;
  if (Array.isArray(records)) {
    lines.push(`输出记录 ${records.length} 条`);
  }
  return lines;
}

function RunNodeWorkbench({ nodeID, runStatus, onMutated }: RunNodeWorkbenchProps) {
  const { data, loading, error, refetch } = useRunNodeDetail(nodeID);
  const commentsQuery = useComments("flow_run_node", nodeID);
  const attachmentsQuery = useAttachments("flow_run_node", nodeID);
  const logsQuery = useRunNodeLogs(nodeID);
  const agentSnapshot = useAgentTaskSnapshot(nodeID, !!nodeID);
  const actions = useRunNodeActions();
  const collaborationActions = useCollaborationActions();
  const [formValues, setFormValues] = useState<Record<string, string>>({});
  const [actionError, setActionError] = useState("");

  const formConfig = useMemo(() => {
    if (!data) {
      return null;
    }
    return resolveNodeFormConfig(data.node_code, data.node_type);
  }, [data]);

  useEffect(() => {
    if (!data || !formConfig) {
      setFormValues({});
      setActionError("");
      return;
    }
    const inputRecord = readRecord(data.input_json);
    const outputRecord = readRecord(data.output_json);
    const nextValues: Record<string, string> = {};
    formConfig.fields.forEach((field) => {
      const inputValue = inputRecord[field.key];
      const outputValue = outputRecord[field.key];
      nextValues[field.key] = toTextValue(inputValue ?? outputValue);
    });
    setFormValues(nextValues);
    setActionError("");
  }, [data, formConfig]);

  const readonly = runStatus === "cancelled";
  const can = useMemo(() => {
    const set = new Set(data?.available_actions ?? []);
    return {
      saveInput: set.has("save_input"),
      submit: set.has("submit"),
      approve: set.has("approve"),
      reject: set.has("reject"),
      requestMaterial: set.has("request_material"),
      complete: set.has("complete"),
      fail: set.has("fail"),
      runAgent: set.has("run_agent"),
      confirmAgentResult: set.has("confirm_agent_result"),
      takeover: set.has("takeover")
    };
  }, [data?.available_actions]);

  const isCurrentNode = !!data?.is_current && data?.status !== "done";
  const shouldPollAgentState = !!data?.bound_agent_id && isCurrentNode && ["running", "wait_confirm", "blocked", "failed"].includes(data.status);

  useEffect(() => {
    if (!shouldPollAgentState) {
      return;
    }
    const timer = window.setInterval(() => {
      void refetch();
      void logsQuery.refetch();
      void agentSnapshot.refetch();
    }, 1500);
    return () => window.clearInterval(timer);
  }, [agentSnapshot, logsQuery, refetch, shouldPollAgentState]);

  const exec = async (fn: () => Promise<unknown>) => {
    setActionError("");
    try {
      await fn();
      await refetch();
      await commentsQuery.refetch();
      await attachmentsQuery.refetch();
      await logsQuery.refetch();
      await onMutated();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "节点操作失败");
    }
  };

  const validateCompleteOrSubmit = (): void => {
    if (!formConfig) {
      return;
    }
    const requiredError = validateRequired(formValues, formConfig.fields);
    if (requiredError) {
      throw new Error(requiredError);
    }
    if (formConfig.requiresAttachmentOnComplete && attachmentsQuery.data.length < 1) {
      throw new Error("当前节点至少需要上传 1 份附件");
    }
  };

  const validateApprove = (): void => {
    if (!formConfig) {
      return;
    }
    const requiredApproveError = validateRequiredOnApprove(formValues, formConfig.fields);
    if (requiredApproveError) {
      throw new Error(requiredApproveError);
    }
  };

  const handleSave = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      await actions.saveInput(data.id, buildPayload(formValues));
    });

  const handleSubmit = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      validateCompleteOrSubmit();
      await actions.submit(data.id, formValues.review_comment?.trim() ?? "");
    });

  const handleComplete = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      validateCompleteOrSubmit();
      await actions.complete(data.id, {
        output_json: buildPayload(formValues)
      });
    });

  const handleApprove = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      validateApprove();
      await actions.approve(data.id, {
        review_comment: formValues.review_comment?.trim() || undefined,
        final_plan: formValues.final_plan?.trim() || undefined,
        output_json: buildPayload(formValues)
      });
    });

  const handleReject = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      const reason = window.prompt("请输入驳回原因", "") ?? "";
      if (!reason.trim()) {
        throw new Error("驳回原因不能为空");
      }
      await actions.reject(data.id, reason.trim());
    });

  const handleRequestMaterial = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      const requirement = window.prompt("请输入补充材料要求", "") ?? "";
      if (!requirement.trim()) {
        throw new Error("补充材料要求不能为空");
      }
      await actions.requestMaterial(data.id, requirement.trim());
    });

  const handleFail = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      const reason = window.prompt("请输入异常原因", "") ?? "";
      if (!reason.trim()) {
        throw new Error("异常原因不能为空");
      }
      await actions.fail(data.id, reason.trim());
    });

  const handleRunAgent = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      await actions.runAgent(data.id);
    });

  const handleConfirmAgentResult = (action: "confirm" | "reject"): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      const comment = window.prompt(action === "confirm" ? "请输入确认说明（可选）" : "请输入驳回说明", "") ?? "";
      if (action === "reject" && !comment.trim()) {
        throw new Error("驳回龙虾结果时必须填写说明");
      }
      await actions.confirmAgentResult(data.id, {
        action,
        comment: comment.trim() || undefined
      });
    });

  const handleTakeover = (action: "complete" | "retry"): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      const comment = window.prompt(
        action === "retry" ? "请输入重试说明（可选）" : "请输入人工接管说明，并在下一步补充结果 JSON",
        ""
      ) ?? "";
      let manualResult: Record<string, unknown> | undefined;
      if (action === "complete") {
        const raw = window.prompt("请输入人工结果 JSON", "{}") ?? "{}";
        try {
          manualResult = readRecord(JSON.parse(raw));
        } catch {
          throw new Error("人工结果 JSON 不合法");
        }
      }
      await actions.takeover(data.id, {
        action,
        comment: comment.trim() || undefined,
        manual_result: manualResult
      });
    });

  if (!nodeID) {
    return <p className="muted">请选择节点查看详情</p>;
  }

  return (
    <section className="run-node-detail-panel">
      <h3>节点工作台</h3>
      {loading ? <p>加载中...</p> : null}
      {error ? <p className="error-text">{error}</p> : null}
      {actionError ? <p className="error-text">{actionError}</p> : null}
      {data ? (
        <>
          <p>
            <b>节点：</b>
            {data.node_name}（{data.node_code}）
          </p>
          <p>
            <b>状态：</b>
            {data.status}
          </p>
          <p>
            <b>执行主体：</b>
            {data.bound_agent?.name ?? "人工节点"}
          </p>
          <p>
            <b>结果责任人：</b>
            {data.result_owner_person?.name ?? "-"}
          </p>
          <p>
            <b>表单渲染：</b>
            {formConfig?.source === "fixed" ? "固定节点表单" : "节点类型默认表单"}
          </p>
          {formConfig ? (
            <>
              <p className="muted">{formConfig.title}</p>
              <p className="muted">{formConfig.description}</p>
            </>
          ) : null}
          <p>
            <b>可执行动作：</b>
            {(data.available_actions || []).join(", ") || "无"}
          </p>

          {formConfig ? (
            <div className="node-form-grid">
              {formConfig.fields.map((field) => (
                <label className="full-width" key={field.key}>
                  {field.label}
                  {field.required ? " *" : ""}
                  {field.requiredOnApprove ? "（审核通过时必填）" : ""}
                  {field.multiline ? (
                    <textarea
                      rows={4}
                      value={formValues[field.key] ?? ""}
                      onChange={(event) =>
                        setFormValues((prev) => ({
                          ...prev,
                          [field.key]: event.target.value
                        }))
                      }
                      placeholder={field.placeholder}
                      disabled={readonly || !isCurrentNode}
                    />
                  ) : (
                    <input
                      value={formValues[field.key] ?? ""}
                      onChange={(event) =>
                        setFormValues((prev) => ({
                          ...prev,
                          [field.key]: event.target.value
                        }))
                      }
                      placeholder={field.placeholder}
                      disabled={readonly || !isCurrentNode}
                    />
                  )}
                </label>
              ))}
            </div>
          ) : null}

          {data.run ? (
            <div className="node-summary-card">
              <p>
                <b>流程标题：</b>
                {data.run.title}
              </p>
              <p>
                <b>流程当前状态：</b>
                {data.run.current_status}
              </p>
              <p>
                <b>上游申请摘要：</b>
              </p>
              <pre>{JSON.stringify(data.run.input_payload_json ?? {}, null, 2)}</pre>
            </div>
          ) : null}
          {agentSnapshot.task ? (
            <div className="node-summary-card agent-task-card">
              <div className="page-title">
                <div>
                  <span className="section-kicker">agent task</span>
                  <h4>龙虾执行回执</h4>
                </div>
                <span className={`pill run-status-tag ${agentSnapshot.task.status}`}>{agentSnapshot.task.status}</span>
              </div>
              <div className="kv-grid">
                <p>
                  <b>任务类型：</b>
                  {agentSnapshot.task.task_type}
                </p>
                <p>
                  <b>执行龙虾：</b>
                  {data.bound_agent?.name ?? `#${agentSnapshot.task.agent_id}`}
                </p>
                <p>
                  <b>开始时间：</b>
                  {agentSnapshot.task.started_at ? new Date(agentSnapshot.task.started_at).toLocaleString() : "-"}
                </p>
                <p>
                  <b>结束时间：</b>
                  {agentSnapshot.task.finished_at ? new Date(agentSnapshot.task.finished_at).toLocaleString() : "-"}
                </p>
              </div>
              {agentSnapshot.task.error_message ? <p className="error-text">{agentSnapshot.task.error_message}</p> : null}
              {renderResultSummary(readRecord(agentSnapshot.task.result_json)).length ? (
                <ul className="plain-list agent-result-points">
                  {renderResultSummary(readRecord(agentSnapshot.task.result_json)).map((line) => (
                    <li key={line}>{line}</li>
                  ))}
                </ul>
              ) : null}
              {Array.isArray(agentSnapshot.task.artifacts_json) && agentSnapshot.task.artifacts_json.length ? (
                <>
                  <h4>产物</h4>
                  <ul className="plain-list agent-artifact-list">
                    {agentSnapshot.task.artifacts_json.map((artifact) => (
                      <li key={`${artifact.url}-${artifact.name}`}>
                        <a href={artifact.url} target="_blank" rel="noreferrer">
                          {artifact.name}
                        </a>
                        <span className="muted">{artifact.type}</span>
                      </li>
                    ))}
                  </ul>
                </>
              ) : null}
              {agentSnapshot.receipt ? (
                <>
                  <p>
                    <b>最新回执：</b>
                    {agentSnapshot.receipt.receipt_status} / {new Date(agentSnapshot.receipt.received_at).toLocaleString()}
                  </p>
                  {Array.isArray(agentSnapshot.receipt.payload_json.logs) && agentSnapshot.receipt.payload_json.logs.length ? (
                    <details>
                      <summary>执行日志</summary>
                      <pre>{formatJSON(agentSnapshot.receipt.payload_json.logs)}</pre>
                    </details>
                  ) : null}
                  <details>
                    <summary>结果 JSON</summary>
                    <pre>{formatJSON(agentSnapshot.task.result_json)}</pre>
                  </details>
                </>
              ) : null}
            </div>
          ) : null}
          {formConfig?.requiresAttachmentOnComplete ? <p className="node-form-hint">本节点标记完成前至少上传 1 份触达凭证附件。</p> : null}

          <div className="action-grid">
            <button
              type="button"
              className="btn"
              disabled={!can.saveInput || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleSave()}
            >
              暂存
            </button>
            <button
              type="button"
              className="btn"
              disabled={!can.submit || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleSubmit()}
            >
              提交确认
            </button>
            <button
              type="button"
              className="btn btn-primary"
              disabled={!can.complete || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleComplete()}
            >
              标记完成
            </button>
            <button
              type="button"
              className="btn btn-primary"
              disabled={!can.approve || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleApprove()}
            >
              审核通过
            </button>
            <button
              type="button"
              className="btn danger"
              disabled={!can.reject || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleReject()}
            >
              驳回
            </button>
            <button
              type="button"
              className="btn"
              disabled={!can.requestMaterial || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleRequestMaterial()}
            >
              要求补材料
            </button>
            <button
              type="button"
              className="btn danger"
              disabled={!can.fail || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleFail()}
            >
              标记异常
            </button>
            <button
              type="button"
              className="btn"
              disabled={!can.runAgent || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleRunAgent()}
            >
              运行龙虾
            </button>
            <button
              type="button"
              className="btn btn-primary"
              disabled={!can.confirmAgentResult || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleConfirmAgentResult("confirm")}
            >
              确认龙虾结果
            </button>
            <button
              type="button"
              className="btn danger"
              disabled={!can.confirmAgentResult || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleConfirmAgentResult("reject")}
            >
              驳回龙虾结果
            </button>
            <button
              type="button"
              className="btn"
              disabled={!can.takeover || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleTakeover("retry")}
            >
              要求重试
            </button>
            <button
              type="button"
              className="btn btn-primary"
              disabled={!can.takeover || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleTakeover("complete")}
            >
              人工接管完成
            </button>
          </div>

          <h4>附件</h4>
          <AttachmentUploader
            disabled={readonly || !isCurrentNode}
            loading={actions.loading || collaborationActions.loading}
            onUploadByURL={(payload) =>
              exec(async () => {
                await collaborationActions.createAttachment("flow_run_node", data.id, payload);
              })
            }
            onUploadFile={(file) =>
              exec(async () => {
                await collaborationActions.uploadAttachmentFile("flow_run_node", data.id, file);
              })
            }
          />
          {attachmentsQuery.error ? <p className="error-text">{attachmentsQuery.error}</p> : null}
          <AttachmentList attachments={attachmentsQuery.data} />

          <h4>评论</h4>
          {!readonly && isCurrentNode ? (
            <CommentEditor
              submitting={collaborationActions.loading}
              onSubmit={(content) =>
                exec(async () => {
                  await collaborationActions.createComment("flow_run_node", data.id, content);
                })
              }
            />
          ) : null}
          {commentsQuery.error ? <p className="error-text">{commentsQuery.error}</p> : null}
          <CommentList
            comments={commentsQuery.data}
            resolving={collaborationActions.loading}
            onResolve={
              !readonly && isCurrentNode
                ? (commentID) =>
                    exec(async () => {
                      await collaborationActions.resolveComment(commentID);
                    })
                : undefined
            }
          />

          <h4>日志</h4>
          {logsQuery.error ? <p className="error-text">{logsQuery.error}</p> : null}
          <NodeLogTimeline logs={logsQuery.data} />
        </>
      ) : null}
    </section>
  );
}

export default RunNodeWorkbench;
