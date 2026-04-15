import { useEffect, useMemo, useState } from "react";
import AttachmentList from "./AttachmentList";
import AttachmentUploader from "./AttachmentUploader";
import CommentEditor from "./CommentEditor";
import CommentList from "./CommentList";
import NodeLogTimeline from "./NodeLogTimeline";
import { useAgentTaskSnapshot } from "../hooks/useAgentTasks";
import { useCollaborationActions, useComments, useAttachments } from "../hooks/useCollaboration";
import { useRunNodeActions, useRunNodeDetail, useRunNodeLogs } from "../hooks/useRunNodes";

type RunNodeWorkbenchProps = {
  nodeID?: number;
  runStatus?: string;
  onMutated: () => Promise<void>;
};

type DeliverableMaterialItem = {
  name: string;
  description: string;
  oss_url: string;
};

function readRecord(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }
  return value as Record<string, unknown>;
}

function toTextValue(value: unknown, fallback = ""): string {
  if (typeof value === "string") {
    return value;
  }
  if (value === undefined || value === null) {
    return fallback;
  }
  return String(value);
}

function toMaterialItems(value: unknown): DeliverableMaterialItem[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value
    .map((item) => {
      const record = readRecord(item);
      return {
        name: toTextValue(record.name).trim(),
        description: toTextValue(record.description).trim(),
        oss_url: toTextValue(record.oss_url || record.url).trim()
      };
    })
    .filter((item) => item.name || item.description || item.oss_url);
}

function lobsterStatusLabel(status: string): string {
  return status === "done" || status === "waiting_confirm" ? "已完成" : "未完成";
}

function reviewStatusLabel(status: string): string {
  return status === "done" ? "已审核" : "未审核";
}

function normalizeMaterials(output: unknown, attachments: Array<{ file_name: string; file_type: string; file_url: string }>) {
  const outputRecord = readRecord(output);
  const materialRecord = readRecord(outputRecord.deliverable_materials);
  const note = toTextValue(materialRecord.summary);
  const files = toMaterialItems(materialRecord.files);
  if (files.length > 0 || note.trim()) {
    return { note, files };
  }
  return {
    note: "",
    files: attachments.map((attachment) => ({
      name: attachment.file_name,
      description: attachment.file_type,
      oss_url: attachment.file_url
    }))
  };
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
  const [matter, setMatter] = useState("");
  const [deliverableNote, setDeliverableNote] = useState("");
  const [deliverableItems, setDeliverableItems] = useState<DeliverableMaterialItem[]>([]);
  const [actionError, setActionError] = useState("");

  useEffect(() => {
    if (!data) {
      setMatter("");
      setDeliverableNote("");
      setDeliverableItems([]);
      setActionError("");
      return;
    }
    const inputRecord = readRecord(data.input_json);
    const outputRecord = readRecord(data.output_json);
    const materials = normalizeMaterials(outputRecord, attachmentsQuery.data);
    setMatter(toTextValue(inputRecord.matter || outputRecord.summary, data.node_name));
    setDeliverableNote(materials.note);
    setDeliverableItems(materials.files);
    setActionError("");
  }, [attachmentsQuery.data, data]);

  const readonly = runStatus === "cancelled";
  const can = useMemo(() => {
    const set = new Set(data?.available_actions ?? []);
    return {
      saveInput: set.has("save_input"),
      approve: set.has("approve"),
      complete: set.has("complete"),
      fail: set.has("fail"),
      runAgent: set.has("run_agent"),
      confirmAgentResult: set.has("confirm_agent_result"),
      takeover: set.has("takeover")
    };
  }, [data?.available_actions]);

  const isCurrentNode = !!data?.is_current && data?.status !== "done";
  const shouldPollAgentState = !!data?.bound_agent_id && isCurrentNode && ["running", "waiting_confirm", "blocked", "failed"].includes(data.status);

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

  const buildNodeOutput = () => ({
    summary: matter.trim() || data?.node_name || "",
    deliverable_materials: {
      summary: deliverableNote.trim(),
      files: deliverableItems.filter((item) => item.name.trim() || item.description.trim() || item.oss_url.trim())
    }
  });

  const updateMaterialItem = (index: number, key: keyof DeliverableMaterialItem, value: string) => {
    setDeliverableItems((prev) => prev.map((item, currentIndex) => (currentIndex === index ? { ...item, [key]: value } : item)));
  };

  const handleSave = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      await actions.saveInput(data.id, { matter: matter.trim() });
    });

  const handleComplete = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      await actions.complete(data.id, {
        output_json: buildNodeOutput()
      });
    });

  const handleApprove = (): Promise<void> =>
    exec(async () => {
      if (!data) {
        return;
      }
      await actions.approve(data.id, {
        review_comment: "已审核",
        output_json: buildNodeOutput()
      });
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
      const comment = window.prompt(action === "confirm" ? "请输入审核说明（可选）" : "请输入驳回说明", "") ?? "";
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
      const comment = window.prompt(action === "retry" ? "请输入重试说明（可选）" : "请输入人工接管说明", "") ?? "";
      await actions.takeover(data.id, {
        action,
        comment: comment.trim() || undefined,
        manual_result: action === "complete" ? buildNodeOutput() : undefined
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
          <div className="node-summary-card">
            <p>
              <b>哪个龙虾：</b>
              {data.bound_agent?.name ?? "未指定"}
            </p>
            <p>
              <b>哪个人：</b>
              {data.owner_person?.name ?? data.result_owner_person?.name ?? "-"}
            </p>
            <p>
              <b>龙虾状态：</b>
              {lobsterStatusLabel(data.status)}
            </p>
            <p>
              <b>审核状态：</b>
              {reviewStatusLabel(data.status)}
            </p>
          </div>

          <div className="node-form-grid">
            <label className="full-width">
              事项
              <input
                value={matter}
                onChange={(event) => setMatter(event.target.value)}
                disabled={readonly || !isCurrentNode}
                placeholder="1 句话解释当前节点完成了什么事"
              />
            </label>

            <label className="full-width">
              交付物料说明
              <textarea
                rows={3}
                value={deliverableNote}
                onChange={(event) => setDeliverableNote(event.target.value)}
                disabled={readonly || !isCurrentNode}
                placeholder="说明这组文件分别是什么，给下游怎么使用"
              />
            </label>
          </div>

          <div className="node-summary-card">
            <div className="page-title">
              <div>
                <span className="section-kicker">deliverables</span>
                <h4>交付物料</h4>
              </div>
              {!readonly && isCurrentNode ? (
                <button
                  type="button"
                  className="btn"
                  onClick={() => setDeliverableItems((prev) => [...prev, { name: "", description: "", oss_url: "" }])}
                >
                  新增文件
                </button>
              ) : null}
            </div>

            {deliverableItems.length ? (
              <div className="node-form-grid">
                {deliverableItems.map((item, index) => (
                  <div className="full-width node-summary-card" key={`${index}-${item.oss_url}`}>
                    <label>
                      文件名
                      <input
                        value={item.name}
                        onChange={(event) => updateMaterialItem(index, "name", event.target.value)}
                        disabled={readonly || !isCurrentNode}
                        placeholder="例如：甩班名单.xlsx"
                      />
                    </label>
                    <label>
                      文件说明
                      <input
                        value={item.description}
                        onChange={(event) => updateMaterialItem(index, "description", event.target.value)}
                        disabled={readonly || !isCurrentNode}
                        placeholder="说明这个文件是什么"
                      />
                    </label>
                    <label className="full-width">
                      OSS URL
                      <input
                        value={item.oss_url}
                        onChange={(event) => updateMaterialItem(index, "oss_url", event.target.value)}
                        disabled={readonly || !isCurrentNode}
                        placeholder="https://oss.example.com/path/file.pdf"
                      />
                    </label>
                    {!readonly && isCurrentNode ? (
                      <div className="toolbar">
                        <button
                          type="button"
                          className="btn danger"
                          onClick={() => setDeliverableItems((prev) => prev.filter((_, currentIndex) => currentIndex !== index))}
                        >
                          删除文件
                        </button>
                      </div>
                    ) : null}
                  </div>
                ))}
              </div>
            ) : (
              <p className="muted">还没有交付物料，可以直接补 OSS URL 列表，或继续使用下方附件区上传/挂载文件。</p>
            )}
          </div>

          {data.run ? (
            <div className="node-summary-card">
              <p>
                <b>流程标题：</b>
                {data.run.title}
              </p>
              <p>
                <b>流程状态：</b>
                {data.run.current_status}
              </p>
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
                  <details>
                    <summary>结果 JSON</summary>
                    <pre>{formatJSON(agentSnapshot.task.result_json)}</pre>
                  </details>
                </>
              ) : null}
            </div>
          ) : null}

          <div className="action-grid">
            <button
              type="button"
              className="btn"
              disabled={!can.saveInput || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleSave()}
            >
              保存事项
            </button>
            <button
              type="button"
              className="btn"
              disabled={!can.runAgent || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleRunAgent()}
            >
              让龙虾执行
            </button>
            <button
              type="button"
              className="btn btn-primary"
              disabled={!can.complete || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleComplete()}
            >
              标记已完成
            </button>
            <button
              type="button"
              className="btn btn-primary"
              disabled={!can.approve || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleApprove()}
            >
              标记已审核
            </button>
            <button
              type="button"
              className="btn btn-primary"
              disabled={!can.confirmAgentResult || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleConfirmAgentResult("confirm")}
            >
              标记已审核
            </button>
            <button
              type="button"
              className="btn danger"
              disabled={!can.confirmAgentResult || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleConfirmAgentResult("reject")}
            >
              驳回结果
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
              className="btn"
              disabled={!can.takeover || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleTakeover("complete")}
            >
              人工补完成
            </button>
            <button
              type="button"
              className="btn danger"
              disabled={!can.fail || readonly || actions.loading || !isCurrentNode}
              onClick={() => void handleFail()}
            >
              标记异常
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
