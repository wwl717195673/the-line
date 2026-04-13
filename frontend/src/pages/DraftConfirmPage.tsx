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
  const [actionError, setActionError] = useState("");

  useEffect(() => {
    if (!data) {
      return;
    }
    setTitle(data.structured_plan_json.title || data.title);
    setDescription(data.structured_plan_json.description || data.description);
    setFinalDeliverable(data.structured_plan_json.final_deliverable || "");
    setNodes((data.structured_plan_json.nodes ?? []).map((node, index) => normalizeNode(node, index)));
    setActionError("");
  }, [data]);

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

  if (!draftID) {
    return (
      <section className="page-card">
        <p className="error-text">草案 ID 不合法</p>
      </section>
    );
  }

  return (
    <section className="draft-page-grid">
      <article className="page-card">
        <div className="page-title">
          <div>
            <span className="section-kicker">draft confirm</span>
            <h2>校对并确认流程草案</h2>
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
        {loading ? <p>加载草案中...</p> : null}
        {error ? <p className="error-text">{error}</p> : null}
        {actionError ? <p className="error-text">{actionError}</p> : null}

        {data ? (
          <>
            <div className="kv-grid draft-meta-grid">
              <p>
                <b>草案状态：</b>
                {data.status}
              </p>
              <p>
                <b>草案发起人：</b>
                #{data.creator_person_id}
              </p>
              <p className="full-width">
                <b>原始需求：</b>
                {data.source_prompt}
              </p>
            </div>

            <div className="form-grid">
              <label className="full-width">
                草案标题 *
                <input value={title} onChange={(event) => setTitle(event.target.value)} disabled={!canEdit} />
              </label>
              <label className="full-width">
                草案说明
                <textarea rows={3} value={description} onChange={(event) => setDescription(event.target.value)} disabled={!canEdit} />
              </label>
              <label className="full-width">
                最终交付定义 *
                <textarea rows={3} value={finalDeliverable} onChange={(event) => setFinalDeliverable(event.target.value)} disabled={!canEdit} />
              </label>
            </div>

            <div className="page-title draft-node-title">
              <div>
                <span className="section-kicker">node plan</span>
                <h2>节点编排</h2>
              </div>
              {canEdit ? (
                <button
                  type="button"
                  className="btn"
                  onClick={() => setNodes((prev) => [...prev, createEmptyNode(prev.length + 1)])}
                >
                  新增节点
                </button>
              ) : null}
            </div>

            <div className="draft-node-stack">
              {nodes.map((node, index) => {
                const normalized = normalizeNode(node, index);
                return (
                  <article key={`${node.node_code}-${index}`} className="template-node-card draft-node-card">
                    <div className="template-node-title">
                      <strong>
                        {index + 1}. {normalized.node_name}
                      </strong>
                      {canEdit ? (
                        <button
                          type="button"
                          className="btn btn-text danger"
                          onClick={() => setNodes((prev) => prev.filter((_, currentIndex) => currentIndex !== index))}
                        >
                          删除
                        </button>
                      ) : null}
                    </div>

                    <div className="form-grid">
                      <label>
                        节点名称 *
                        <input
                          value={node.node_name}
                          onChange={(event) =>
                            setNodes((prev) =>
                              prev.map((item, currentIndex) =>
                                currentIndex === index ? { ...item, node_name: event.target.value } : item
                              )
                            )
                          }
                          disabled={!canEdit}
                        />
                      </label>
                      <label>
                        节点编码 *
                        <input
                          value={node.node_code}
                          onChange={(event) =>
                            setNodes((prev) =>
                              prev.map((item, currentIndex) =>
                                currentIndex === index ? { ...item, node_code: slugify(event.target.value) } : item
                              )
                            )
                          }
                          disabled={!canEdit}
                        />
                      </label>
                      <label>
                        节点类型 *
                        <select
                          value={normalized.node_type}
                          onChange={(event) =>
                            setNodes((prev) =>
                              prev.map((item, currentIndex) => {
                                if (currentIndex !== index) {
                                  return item;
                                }
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
                            )
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
                          value={normalized.executor_type}
                          onChange={(event) =>
                            setNodes((prev) =>
                              prev.map((item, currentIndex) => {
                                if (currentIndex !== index) {
                                  return item;
                                }
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
                            )
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
                          value={node.owner_rule}
                          onChange={(event) =>
                            setNodes((prev) =>
                              prev.map((item, currentIndex) =>
                                currentIndex === index ? { ...item, owner_rule: event.target.value } : item
                              )
                            )
                          }
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
                          value={node.owner_person_id}
                          onChange={(value) =>
                            setNodes((prev) =>
                              prev.map((item, currentIndex) =>
                                currentIndex === index ? { ...item, owner_person_id: value } : item
                              )
                            )
                          }
                        />
                      </label>

                      <label>
                        结果责任规则 *
                        <select
                          value={node.result_owner_rule}
                          onChange={(event) =>
                            setNodes((prev) =>
                              prev.map((item, currentIndex) =>
                                currentIndex === index ? { ...item, result_owner_rule: event.target.value } : item
                              )
                            )
                          }
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
                          value={node.result_owner_person_id}
                          onChange={(value) =>
                            setNodes((prev) =>
                              prev.map((item, currentIndex) =>
                                currentIndex === index ? { ...item, result_owner_person_id: value } : item
                              )
                            )
                          }
                        />
                      </label>

                      {normalized.executor_type === "agent" ? (
                        <>
                          <label>
                            执行龙虾 *
                            <select
                              value={node.executor_agent_code ?? ""}
                              onChange={(event) =>
                                setNodes((prev) =>
                                  prev.map((item, currentIndex) =>
                                    currentIndex === index ? { ...item, executor_agent_code: event.target.value } : item
                                  )
                                )
                              }
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
                            executor_agent_code *
                            <input
                              value={node.executor_agent_code ?? ""}
                              onChange={(event) =>
                                setNodes((prev) =>
                                  prev.map((item, currentIndex) =>
                                    currentIndex === index ? { ...item, executor_agent_code: event.target.value } : item
                                  )
                                )
                              }
                              disabled={!canEdit}
                              placeholder="例如：video_binder"
                            />
                          </label>
                          <label>
                            task_type *
                            <select
                              value={node.task_type || "query"}
                              onChange={(event) =>
                                setNodes((prev) =>
                                  prev.map((item, currentIndex) =>
                                    currentIndex === index ? { ...item, task_type: event.target.value } : item
                                  )
                                )
                              }
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
                          rows={2}
                          value={node.completion_condition ?? ""}
                          onChange={(event) =>
                            setNodes((prev) =>
                              prev.map((item, currentIndex) =>
                                currentIndex === index ? { ...item, completion_condition: event.target.value } : item
                              )
                            )
                          }
                          disabled={!canEdit}
                        />
                      </label>
                      <label className="full-width">
                        失败条件 / 升级说明
                        <textarea
                          rows={2}
                          value={[node.failure_condition ?? "", node.escalation_rule ?? ""].filter(Boolean).join("\n")}
                          onChange={(event) =>
                            setNodes((prev) =>
                              prev.map((item, currentIndex) =>
                                currentIndex === index
                                  ? {
                                      ...item,
                                      failure_condition: event.target.value.trim(),
                                      escalation_rule: event.target.value.trim()
                                    }
                                  : item
                              )
                            )
                          }
                          disabled={!canEdit}
                        />
                      </label>
                    </div>
                  </article>
                );
              })}
            </div>

            <div className="toolbar">
              {canEdit ? (
                <>
                  <button
                    type="button"
                    className="btn"
                    disabled={updateDraftMutation.loading}
                    onClick={async () => {
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
                    }}
                  >
                    {updateDraftMutation.loading ? "保存中..." : "保存草案"}
                  </button>
                  <button
                    type="button"
                    className="btn btn-primary"
                    disabled={confirmDraftMutation.loading}
                    onClick={async () => {
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
                    }}
                  >
                    {confirmDraftMutation.loading ? "创建模板中..." : "确认并创建模板"}
                  </button>
                  <button
                    type="button"
                    className="btn danger"
                    disabled={discardDraftMutation.loading}
                    onClick={async () => {
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
                    }}
                  >
                    废弃草案
                  </button>
                </>
              ) : (
                <Link className="btn btn-primary" to={data.confirmed_template_id ? `/templates/${data.confirmed_template_id}/start` : "/templates"}>
                  {data.status === "confirmed" ? "进入模板发起流程" : "返回模板列表"}
                </Link>
              )}
            </div>
          </>
        ) : null}
      </article>

      <aside className="page-card draft-side-panel">
        <div className="page-title">
          <div>
            <span className="section-kicker">plan preview</span>
            <h2>结构化预览</h2>
          </div>
        </div>
        <p className="muted">这里显示将要提交给后端的结构化草案，方便快速核对字段和排序。</p>
        <pre className="draft-json-preview">{JSON.stringify(planPreview, null, 2)}</pre>
      </aside>
    </section>
  );
}

export default DraftConfirmPage;
