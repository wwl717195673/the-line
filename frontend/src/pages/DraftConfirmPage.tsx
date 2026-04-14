import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import PersonSelect from "../components/PersonSelect";
import { useAgents } from "../hooks/useAgents";
import { useConfirmDraft, useDiscardDraft, useDraftDetail, useUpdateDraft } from "../hooks/useDrafts";
import { getActor } from "../lib/actor";
import type { DraftNode, DraftPlan } from "../types/api";

function parseID(value?: string): number | undefined {
  if (!value) {
    return undefined;
  }
  const next = Number(value);
  return Number.isInteger(next) && next > 0 ? next : undefined;
}

function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, "_")
    .replace(/^_+|_+$/g, "") || "node";
}

function createEmptyNode(sortOrder: number): DraftNode {
  return {
    node_code: `step_${sortOrder}`,
    node_name: `新节点 ${sortOrder}`,
    node_type: "human_review",
    sort_order: sortOrder,
    executor_type: "human",
    owner_rule: "initiator",
    result_owner_rule: "initiator",
    input_schema: {},
    output_schema: {},
    completion_condition: "",
    failure_condition: "",
    escalation_rule: ""
  };
}

function isAgentNodeType(nodeType: string): boolean {
  return nodeType === "agent_execute" || nodeType === "agent_export";
}

function normalizeNode(node: DraftNode, index: number): DraftNode {
  const nextNodeType = isAgentNodeType(node.node_type) ? node.node_type : node.executor_type === "agent" ? "agent_execute" : node.node_type;
  return {
    ...node,
    node_code: slugify(node.node_code || node.node_name || `step_${index + 1}`),
    sort_order: index + 1,
    node_type: nextNodeType,
    executor_type: isAgentNodeType(nextNodeType) ? "agent" : "human",
    task_type: isAgentNodeType(nextNodeType) ? node.task_type || (nextNodeType === "agent_export" ? "export" : "query") : "",
    executor_agent_code: isAgentNodeType(nextNodeType) ? node.executor_agent_code || "" : "",
    owner_person_id: node.owner_rule === "specified_person" ? node.owner_person_id : undefined,
    result_owner_person_id: node.result_owner_rule === "specified_person" ? node.result_owner_person_id : undefined
  };
}

function getNodeTypeLabel(nodeType: string): string {
  switch (nodeType) {
    case "human_input":
      return "人工录入";
    case "human_review":
      return "人工审核";
    case "agent_execute":
      return "龙虾执行";
    case "agent_export":
      return "龙虾导出";
    case "human_acceptance":
      return "最终签收";
    default:
      return nodeType;
  }
}

function getNodeAccent(nodeType: string): string {
  switch (nodeType) {
    case "human_input":
      return "cyan";
    case "human_review":
      return "amber";
    case "agent_execute":
      return "blue";
    case "agent_export":
      return "emerald";
    case "human_acceptance":
      return "rose";
    default:
      return "slate";
  }
}

function summarizeOwner(rule: string, personID?: number): string {
  if (rule === "specified_person" && personID) {
    return `指定人员 #${personID}`;
  }
  switch (rule) {
    case "initiator":
      return "发起人";
    case "middle_office":
      return "中台角色";
    case "operation":
      return "运营角色";
    case "current_owner":
      return "沿用当前责任人";
    case "node_owner":
      return "节点执行责任人";
    default:
      return rule || "-";
  }
}

function DraftConfirmPage() {
  const params = useParams();
  const navigate = useNavigate();
  const actor = getActor();
  const draftID = parseID(params.id);
  const agentQuery = useMemo(
    () => ({
      page: 1,
      page_size: 100,
      status: 1 as const
    }),
    []
  );
  const agents = useAgents(agentQuery);
  const { data, loading, error, refetch } = useDraftDetail(draftID);
  const updateDraftMutation = useUpdateDraft();
  const confirmDraftMutation = useConfirmDraft();
  const discardDraftMutation = useDiscardDraft();
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [finalDeliverable, setFinalDeliverable] = useState("");
  const [nodes, setNodes] = useState<DraftNode[]>([]);
  const [selectedNodeIndex, setSelectedNodeIndex] = useState(0);
  const [actionError, setActionError] = useState("");

  useEffect(() => {
    if (!data) {
      return;
    }
    const nextNodes = (data.structured_plan_json.nodes ?? []).map((node, index) => normalizeNode(node, index));
    setTitle(data.structured_plan_json.title || data.title);
    setDescription(data.structured_plan_json.description || data.description);
    setFinalDeliverable(data.structured_plan_json.final_deliverable || "");
    setNodes(nextNodes);
    setSelectedNodeIndex(nextNodes.length ? 0 : -1);
    setActionError("");
  }, [data]);

  useEffect(() => {
    if (!nodes.length) {
      setSelectedNodeIndex(-1);
      return;
    }
    if (selectedNodeIndex < 0 || selectedNodeIndex >= nodes.length) {
      setSelectedNodeIndex(0);
    }
  }, [nodes, selectedNodeIndex]);

  const confirmedBy = actor.personId || data?.creator_person_id;
  const canEdit = data?.status === "draft";
  const planPreview = useMemo<DraftPlan>(
    () => ({
      title: title.trim(),
      description: description.trim(),
      final_deliverable: finalDeliverable.trim(),
      nodes: nodes.map((node, index) => normalizeNode(node, index))
    }),
    [description, finalDeliverable, nodes, title]
  );
  const selectedNode = selectedNodeIndex >= 0 ? nodes[selectedNodeIndex] : undefined;
  const normalizedSelectedNode = selectedNode ? normalizeNode(selectedNode, selectedNodeIndex) : undefined;
  const agentNodeCount = nodes.filter((node) => normalizeNode(node, 0).executor_type === "agent").length;
  const humanNodeCount = Math.max(0, nodes.length - agentNodeCount);

  function updateNode(index: number, updater: (node: DraftNode) => DraftNode) {
    setNodes((prev) => prev.map((item, currentIndex) => (currentIndex === index ? updater(item) : item)));
  }

  async function handleSaveDraft() {
    if (!draftID) {
      return;
    }
    setActionError("");
    try {
      await updateDraftMutation.run(draftID, {
        title: title.trim() || undefined,
        description: description.trim() || undefined,
        structured_plan_json: planPreview
      });
      await refetch();
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "保存草案失败");
    }
  }

  async function handleConfirmDraft() {
    if (!draftID || !confirmedBy) {
      setActionError("缺少确认人，无法创建模板");
      return;
    }
    setActionError("");
    try {
      await updateDraftMutation.run(draftID, {
        title: title.trim() || undefined,
        description: description.trim() || undefined,
        structured_plan_json: planPreview
      });
      const result = await confirmDraftMutation.run(draftID, confirmedBy);
      navigate(`/templates/${result.template_id}/start`);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "确认草案失败");
    }
  }

  async function handleDiscardDraft() {
    if (!draftID || !confirmedBy) {
      setActionError("缺少废弃操作人");
      return;
    }
    const reason = window.prompt("请输入废弃原因", "") ?? "";
    setActionError("");
    try {
      await discardDraftMutation.run(draftID, confirmedBy, reason.trim());
      navigate("/drafts/create");
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "废弃草案失败");
    }
  }

  if (!draftID) {
    return (
      <section className="page-card">
        <p className="error-text">草案 ID 不合法</p>
      </section>
    );
  }

  return (
    <section className="draft-studio">
      <article className="draft-studio-shell">
        <div className="page-title draft-studio-header">
          <div>
            <span className="section-kicker">draft canvas</span>
            <h2>{canEdit ? "继续编辑流程草案" : "查看流程草案"}</h2>
            <p className="page-note">将线性列表收拢成节点画布。中心区域负责浏览流程，右侧只编辑当前选中的节点。</p>
          </div>
          <div className="toolbar">
            <Link className="btn" to="/drafts">
              草案列表
            </Link>
            <Link className="btn" to="/drafts/create">
              再生成一份
            </Link>
            <button type="button" className="btn" onClick={() => void refetch()}>
              刷新
            </button>
          </div>
        </div>

        {loading ? (
          <div className="draft-canvas-empty">
            <strong>加载草案中...</strong>
            <p>正在同步节点编排、责任人规则和最终交付定义。</p>
          </div>
        ) : null}
        {error ? <p className="error-text">{error}</p> : null}
        {actionError ? <p className="error-text">{actionError}</p> : null}

        {data ? (
          <div className="draft-studio-grid">
            <div className="draft-main-column">
              <aside className="draft-summary-banner">
                <div className="draft-summary-floating">
                  <div className="draft-summary-inline-bar">
                    <div className="draft-summary-head">
                      <div>
                        <span className="section-kicker">overview</span>
                        <h3>{title || data.title || `草案 #${data.id}`}</h3>
                      </div>
                      <span className={`pill draft-status-${data.status}`}>{data.status}</span>
                    </div>

                    <div className="draft-summary-inline-metrics">
                      <span>{nodes.length} 个节点</span>
                      <span>{agentNodeCount} 个龙虾节点</span>
                      <span>{humanNodeCount} 个人工节点</span>
                    </div>

                    <div className="draft-summary-inline-meta">
                      <span>发起人 #{data.creator_person_id}</span>
                      <span>确认人 {confirmedBy ? `#${confirmedBy}` : "未识别"}</span>
                    </div>

                    <div className="draft-summary-actions">
                      {canEdit ? (
                        <>
                          <button type="button" className="btn" disabled={updateDraftMutation.loading} onClick={() => void handleSaveDraft()}>
                            {updateDraftMutation.loading ? "保存中..." : "保存草案"}
                          </button>
                          <button type="button" className="btn btn-primary" disabled={confirmDraftMutation.loading} onClick={() => void handleConfirmDraft()}>
                            {confirmDraftMutation.loading ? "创建模板中..." : "确认并创建模板"}
                          </button>
                          <button type="button" className="btn danger" disabled={discardDraftMutation.loading} onClick={() => void handleDiscardDraft()}>
                            废弃草案
                          </button>
                        </>
                      ) : (
                        <Link className="btn btn-primary" to={data.confirmed_template_id ? `/templates/${data.confirmed_template_id}/start` : "/templates"}>
                          {data.status === "confirmed" ? "进入模板发起流程" : "返回模板列表"}
                        </Link>
                      )}
                    </div>
                  </div>

                  <details className="draft-summary-editor">
                    <summary>展开编辑</summary>
                    <p className="draft-summary-brief">{description || "草案说明未填写，当前画布按节点完成条件串联主流程。"}</p>

                    <div className="draft-summary-inline-editors">
                      <label className="draft-summary-inline-field">
                        <span>草案标题</span>
                        <input value={title} onChange={(event) => setTitle(event.target.value)} disabled={!canEdit} />
                      </label>

                      <label className="draft-summary-inline-field">
                        <span>最终交付定义</span>
                        <textarea rows={3} value={finalDeliverable} onChange={(event) => setFinalDeliverable(event.target.value)} disabled={!canEdit} />
                      </label>
                    </div>

                    <details className="draft-summary-advanced">
                      <summary>更多说明</summary>
                      <label className="full-width">
                        草案说明
                        <textarea rows={4} value={description} onChange={(event) => setDescription(event.target.value)} disabled={!canEdit} />
                      </label>
                      <p className="draft-summary-source">
                        <b>原始需求：</b>
                        {data.source_prompt}
                      </p>
                    </details>
                  </details>
                </div>
              </aside>

              <div className="page-card draft-canvas-panel">
              <div className="draft-canvas-toolbar">
                <div>
                  <span className="section-kicker">canvas</span>
                  <h3>节点流画布</h3>
                </div>
                <div className="toolbar">
                  {canEdit ? (
                    <button
                      type="button"
                      className="btn"
                      onClick={() => {
                        setNodes((prev) => [...prev, createEmptyNode(prev.length + 1)]);
                        setSelectedNodeIndex(nodes.length);
                      }}
                    >
                      新增节点
                    </button>
                  ) : null}
                </div>
              </div>

              {nodes.length ? (
                <div className="draft-canvas-scroller">
                  <div className="draft-canvas-board">
                    <div className="draft-flow-origin">
                      <span className="draft-flow-origin-label">Start</span>
                      <strong>{title || data.title || `草案 #${data.id}`}</strong>
                      <p>{description || "从需求描述进入第一段节点执行链路。"}</p>
                    </div>

                    <div className="draft-flow-lane">
                      {nodes.map((node, index) => {
                        const normalized = normalizeNode(node, index);
                        const accent = getNodeAccent(normalized.node_type);
                        return (
                          <div key={`${node.node_code}-${index}`} className={`draft-flow-step draft-flow-step-${index % 3}`}>
                            {index > 0 ? <div className="draft-flow-connector" aria-hidden="true" /> : null}
                            <button
                              type="button"
                              className={`draft-flow-node draft-flow-node-${accent} ${selectedNodeIndex === index ? "active" : ""}`}
                              onClick={() => setSelectedNodeIndex(index)}
                            >
                              <span className="draft-flow-node-index">{String(index + 1).padStart(2, "0")}</span>
                              <span className="draft-flow-node-type">{getNodeTypeLabel(normalized.node_type)}</span>
                              <strong>{normalized.node_name}</strong>
                              <p>{normalized.completion_condition || "未填写完成条件，建议补充节点通过标准。"}</p>
                              <dl>
                                <div>
                                  <dt>执行</dt>
                                  <dd>{normalized.executor_type === "agent" ? normalized.executor_agent_code || "待绑定龙虾" : summarizeOwner(normalized.owner_rule, normalized.owner_person_id)}</dd>
                                </div>
                                <div>
                                  <dt>结果</dt>
                                  <dd>{summarizeOwner(normalized.result_owner_rule, normalized.result_owner_person_id)}</dd>
                                </div>
                              </dl>
                            </button>
                          </div>
                        );
                      })}

                      <div className="draft-flow-endcap">
                        <span className="draft-flow-origin-label">Deliverable</span>
                        <strong>最终交付</strong>
                        <p>{finalDeliverable || "尚未定义最终交付内容。"}</p>
                      </div>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="draft-canvas-empty">
                  <strong>当前没有节点</strong>
                  <p>可以先新增第一个节点，开始搭建这条流程草案的主链路。</p>
                </div>
              )}

              <div className="draft-canvas-footer">
                <p className="muted">节点按顺序串接显示。当前方案先聚焦阅读和编辑，不处理拖拽、缩放和自由连线。</p>
              </div>
            </div>

              <div className="page-card draft-inspector-dock">
                <div className="draft-inspector-head">
                  <div>
                    <span className="section-kicker">inspector</span>
                    <h3>{normalizedSelectedNode ? normalizedSelectedNode.node_name : "节点属性"}</h3>
                  </div>
                  {normalizedSelectedNode ? (
                    <span className={`pill draft-node-pill-${getNodeAccent(normalizedSelectedNode.node_type)}`}>
                      {getNodeTypeLabel(normalizedSelectedNode.node_type)}
                    </span>
                  ) : null}
                </div>
              </div>
              <aside className="page-card draft-inspector-panel">
                {normalizedSelectedNode && selectedNode ? (
                  <>
                    <div className="draft-inspector-rail">
                      <div>
                        <span>节点编码</span>
                        <strong>{normalizedSelectedNode.node_code}</strong>
                      </div>
                      <div>
                        <span>排序</span>
                        <strong>{normalizedSelectedNode.sort_order}</strong>
                      </div>
                    </div>

                    <div className="form-grid draft-inspector-form">
                      <label className="full-width">
                        节点名称 *
                        <input
                          value={selectedNode.node_name}
                          onChange={(event) => updateNode(selectedNodeIndex, (item) => ({ ...item, node_name: event.target.value }))}
                          disabled={!canEdit}
                        />
                      </label>
                      <label>
                        节点编码 *
                        <input
                          value={selectedNode.node_code}
                          onChange={(event) => updateNode(selectedNodeIndex, (item) => ({ ...item, node_code: slugify(event.target.value) }))}
                          disabled={!canEdit}
                        />
                      </label>
                      <label>
                        节点类型 *
                        <select
                          value={normalizedSelectedNode.node_type}
                          onChange={(event) =>
                            updateNode(selectedNodeIndex, (item) => {
                              const nextNodeType = event.target.value;
                              return {
                                ...item,
                                node_type: nextNodeType,
                                executor_type: isAgentNodeType(nextNodeType) ? "agent" : "human",
                                task_type: isAgentNodeType(nextNodeType)
                                  ? item.task_type || (nextNodeType === "agent_export" ? "export" : "query")
                                  : ""
                              };
                            })
                          }
                          disabled={!canEdit}
                        >
                          <option value="human_input">人工录入</option>
                          <option value="human_review">人工审核</option>
                          <option value="agent_execute">龙虾执行</option>
                          <option value="agent_export">龙虾导出</option>
                          <option value="human_acceptance">最终签收</option>
                        </select>
                      </label>
                      <label>
                        执行主体 *
                        <select
                          value={normalizedSelectedNode.executor_type}
                          onChange={(event) =>
                            updateNode(selectedNodeIndex, (item) => {
                              const nextExecutorType = event.target.value as DraftNode["executor_type"];
                              return {
                                ...item,
                                executor_type: nextExecutorType,
                                node_type:
                                  nextExecutorType === "agent"
                                    ? item.node_type === "agent_export"
                                      ? "agent_export"
                                      : "agent_execute"
                                    : item.node_type === "human_acceptance"
                                      ? "human_acceptance"
                                      : item.node_type === "human_input"
                                        ? "human_input"
                                        : "human_review"
                              };
                            })
                          }
                          disabled={!canEdit}
                        >
                          <option value="human">人</option>
                          <option value="agent">龙虾</option>
                        </select>
                      </label>

                      <label>
                        执行责任规则 *
                        <select
                          value={selectedNode.owner_rule}
                          onChange={(event) => updateNode(selectedNodeIndex, (item) => ({ ...item, owner_rule: event.target.value }))}
                          disabled={!canEdit}
                        >
                          <option value="initiator">发起人</option>
                          <option value="specified_person">指定人员</option>
                          <option value="middle_office">中台角色</option>
                          <option value="operation">运营角色</option>
                          <option value="current_owner">沿用当前责任人</option>
                        </select>
                      </label>
                      <label>
                        执行责任人
                        <PersonSelect
                          value={selectedNode.owner_person_id}
                          onChange={(value) => updateNode(selectedNodeIndex, (item) => ({ ...item, owner_person_id: value }))}
                          disabled={!canEdit}
                        />
                      </label>

                      <label>
                        结果责任规则 *
                        <select
                          value={selectedNode.result_owner_rule}
                          onChange={(event) => updateNode(selectedNodeIndex, (item) => ({ ...item, result_owner_rule: event.target.value }))}
                          disabled={!canEdit}
                        >
                          <option value="initiator">发起人</option>
                          <option value="specified_person">指定人员</option>
                          <option value="node_owner">节点执行责任人</option>
                          <option value="middle_office">中台角色</option>
                          <option value="operation">运营角色</option>
                          <option value="current_owner">沿用当前责任人</option>
                        </select>
                      </label>
                      <label>
                        结果责任人
                        <PersonSelect
                          value={selectedNode.result_owner_person_id}
                          onChange={(value) => updateNode(selectedNodeIndex, (item) => ({ ...item, result_owner_person_id: value }))}
                          disabled={!canEdit}
                        />
                      </label>

                      {normalizedSelectedNode.executor_type === "agent" ? (
                        <>
                          <label>
                            执行龙虾 *
                            <select
                              value={selectedNode.executor_agent_code ?? ""}
                              onChange={(event) => updateNode(selectedNodeIndex, (item) => ({ ...item, executor_agent_code: event.target.value }))}
                              disabled={!canEdit}
                            >
                              <option value="">{agents.loading ? "加载中..." : "请选择执行龙虾"}</option>
                              {(agents.data?.items ?? []).map((agent) => (
                                <option key={agent.id} value={agent.code}>
                                  {agent.name}（{agent.code}）
                                </option>
                              ))}
                            </select>
                          </label>
                          <label>
                            task_type *
                            <select
                              value={selectedNode.task_type || "query"}
                              onChange={(event) => updateNode(selectedNodeIndex, (item) => ({ ...item, task_type: event.target.value }))}
                              disabled={!canEdit}
                            >
                              <option value="query">query</option>
                              <option value="batch_operation">batch_operation</option>
                              <option value="export">export</option>
                            </select>
                          </label>
                        </>
                      ) : null}

                      <label className="full-width">
                        完成条件
                        <textarea
                          rows={3}
                          value={selectedNode.completion_condition ?? ""}
                          onChange={(event) => updateNode(selectedNodeIndex, (item) => ({ ...item, completion_condition: event.target.value }))}
                          disabled={!canEdit}
                        />
                      </label>
                      <label className="full-width">
                        失败条件 / 升级说明
                        <textarea
                          rows={3}
                          value={[selectedNode.failure_condition ?? "", selectedNode.escalation_rule ?? ""].filter(Boolean).join("\n")}
                          onChange={(event) =>
                            updateNode(selectedNodeIndex, (item) => ({
                              ...item,
                              failure_condition: event.target.value.trim(),
                              escalation_rule: event.target.value.trim()
                            }))
                          }
                          disabled={!canEdit}
                        />
                      </label>
                    </div>

                    {canEdit ? (
                      <div className="draft-inspector-actions">
                        <button
                          type="button"
                          className="btn btn-text danger"
                          onClick={() => {
                            setNodes((prev) => prev.filter((_, index) => index !== selectedNodeIndex));
                          }}
                        >
                          删除当前节点
                        </button>
                      </div>
                    ) : null}
                  </>
                ) : (
                  <div className="draft-canvas-empty">
                    <strong>请选择节点</strong>
                    <p>点击画布中的任意节点后，这里会展示该节点的详细属性。</p>
                  </div>
                )}

                <div className="draft-json-block">
                  <div className="page-title">
                    <div>
                      <span className="section-kicker">plan preview</span>
                      <h3>结构化预览</h3>
                    </div>
                  </div>
                  <details className="draft-json-collapse">
                    <summary>展开结构化 JSON</summary>
                    <pre className="draft-json-preview">{JSON.stringify(planPreview, null, 2)}</pre>
                  </details>
                </div>
              </aside>
            </div>
          </div>
        ) : null}
      </article>
    </section>
  );
}

export default DraftConfirmPage;
