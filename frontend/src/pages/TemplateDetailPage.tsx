import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useTemplateDetail } from "../hooks/useTemplates";
import type { TemplateNode } from "../types/api";

function parseID(value?: string): number | undefined {
  if (!value) {
    return undefined;
  }
  const next = Number(value);
  return Number.isInteger(next) && next > 0 ? next : undefined;
}

function readRecord(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }
  return value as Record<string, unknown>;
}

function toTextValue(value: unknown): string {
  if (typeof value === "string") {
    return value;
  }
  if (value === undefined || value === null) {
    return "";
  }
  return String(value);
}

function summarizeOwner(rule: string, personID?: number): string {
  if (personID) {
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
    case "specified_person":
      return "指定人员";
    default:
      return rule || "-";
  }
}

function deriveMatter(node: TemplateNode): string {
  const config = readRecord(node.config_json);
  return toTextValue(config.completion_condition).trim() || node.node_name;
}

function normalizeNode(node: TemplateNode) {
  return {
    ...node,
    matter: deriveMatter(node),
    owner: summarizeOwner(node.default_owner_rule, node.default_owner_person_id),
    agent: node.default_agent ? `${node.default_agent.name}（${node.default_agent.code}）` : "未指定"
  };
}

function TemplateDetailPage() {
  const params = useParams();
  const templateID = parseID(params.templateId);
  const { data, loading, error, refetch } = useTemplateDetail(templateID);
  const [selectedNodeIndex, setSelectedNodeIndex] = useState(0);

  const nodes = useMemo(() => (data?.nodes ?? []).slice().sort((a, b) => a.sort_order - b.sort_order).map(normalizeNode), [data?.nodes]);
  const selectedNode = selectedNodeIndex >= 0 ? nodes[selectedNodeIndex] : undefined;

  useEffect(() => {
    if (!nodes.length) {
      setSelectedNodeIndex(-1);
      return;
    }
    if (selectedNodeIndex < 0 || selectedNodeIndex >= nodes.length) {
      setSelectedNodeIndex(0);
    }
  }, [nodes, selectedNodeIndex]);

  if (!templateID) {
    return (
      <section className="page-card">
        <p className="error-text">模板 ID 不合法</p>
      </section>
    );
  }

  return (
    <section className="draft-studio">
      <article className="draft-studio-shell">
        <div className="page-title draft-studio-header">
          <div>
            <span className="section-kicker">template canvas</span>
            <h2>查看模板画布</h2>
            <p className="page-note">模板详情和流程草案保持同一套无限画布结构，节点字段对齐到“哪个龙虾 / 哪个人 / 事项”。</p>
          </div>
          <div className="toolbar">
            <Link to="/templates" className="btn">
              返回列表
            </Link>
            <button type="button" className="btn" onClick={() => void refetch()}>
              刷新
            </button>
            {data?.status === "published" ? (
              <Link to={`/templates/${templateID}/start`} className="btn btn-primary">
                使用模板
              </Link>
            ) : null}
          </div>
        </div>

        {loading ? (
          <div className="draft-canvas-empty">
            <strong>加载模板中...</strong>
            <p>正在同步模板节点画布。</p>
          </div>
        ) : null}
        {error ? (
          <div className="draft-canvas-empty">
            <strong>模板详情加载失败</strong>
            <p className="error-text">{error}</p>
          </div>
        ) : null}

        {data ? (
          <div className="draft-studio-grid">
            <div className="draft-main-column">
              <aside className="draft-summary-banner">
                <div className="draft-summary-floating">
                  <div className="draft-summary-inline-bar">
                    <div className="draft-summary-head">
                      <div>
                        <span className="section-kicker">overview</span>
                        <h3>{data.name}</h3>
                      </div>
                      <span className={`pill draft-status-${data.status}`}>{data.status}</span>
                    </div>

                    <div className="draft-summary-inline-metrics">
                      <span>{nodes.length} 个节点</span>
                      <span>版本 {data.version}</span>
                      <span>{data.category || "未分类"}</span>
                    </div>

                    <div className="draft-summary-inline-meta">
                      <span>模板编码 {data.code}</span>
                      <span>更新时间 {new Date(data.updated_at).toLocaleString()}</span>
                    </div>
                  </div>

                  <details className="draft-summary-editor" open>
                    <summary>展开说明</summary>
                    <p className="draft-summary-brief">{data.description || "模板说明未填写。"}</p>
                  </details>
                </div>
              </aside>

              <div className="page-card draft-canvas-panel">
                <div className="draft-canvas-toolbar">
                  <div>
                    <span className="section-kicker">canvas</span>
                    <h3>模板节点画布</h3>
                  </div>
                </div>

                {nodes.length ? (
                  <div className="draft-canvas-scroller">
                    <div className="draft-canvas-board">
                      <div className="draft-flow-origin">
                        <span className="draft-flow-origin-label">Template</span>
                        <strong>{data.name}</strong>
                        <p>{data.description || "从模板定义进入第一个节点。"}</p>
                      </div>

                      <div className="draft-flow-lane">
                        {nodes.map((node, index) => (
                          <div key={node.id} className={`draft-flow-step draft-flow-step-${index % 3}`}>
                            {index > 0 ? <div className="draft-flow-connector" aria-hidden="true" /> : null}
                            <button
                              type="button"
                              className={`draft-flow-node draft-flow-node-${node.default_agent ? "blue" : "amber"} ${selectedNodeIndex === index ? "active" : ""}`}
                              onClick={() => setSelectedNodeIndex(index)}
                            >
                              <span className="draft-flow-node-index">{String(index + 1).padStart(2, "0")}</span>
                              <span className="draft-flow-node-type">{node.default_agent ? "龙虾节点" : "人工节点"}</span>
                              <strong>{node.matter}</strong>
                              <p>{node.agent === "未指定" ? "未指定龙虾，默认由人处理" : `龙虾：${node.agent}`}</p>
                              <dl>
                                <div>
                                  <dt>哪个人</dt>
                                  <dd>{node.owner}</dd>
                                </div>
                                <div>
                                  <dt>哪个龙虾</dt>
                                  <dd>{node.agent}</dd>
                                </div>
                              </dl>
                            </button>
                          </div>
                        ))}

                        <div className="draft-flow-endcap">
                          <span className="draft-flow-origin-label">Done</span>
                          <strong>模板终点</strong>
                          <p>节点完成后进入流程交付或归档。</p>
                        </div>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="draft-canvas-empty">
                    <strong>当前没有节点</strong>
                    <p>这个模板还没有节点配置。</p>
                  </div>
                )}

                <div className="draft-canvas-footer">
                  <p className="muted">当前是只读模板画布，用同一套空间结构阅读节点定义，不在这里改模板内容。</p>
                </div>
              </div>

              <div className="page-card draft-inspector-dock">
                <div className="draft-inspector-head">
                  <div>
                    <span className="section-kicker">inspector</span>
                    <h3>{selectedNode ? selectedNode.matter : "节点属性"}</h3>
                  </div>
                  {selectedNode ? (
                    <span className={`pill draft-node-pill-${selectedNode.default_agent ? "blue" : "amber"}`}>
                      {selectedNode.default_agent ? "龙虾节点" : "人工节点"}
                    </span>
                  ) : null}
                </div>
              </div>

              <aside className="page-card draft-inspector-panel">
                {selectedNode ? (
                  <>
                    <div className="draft-inspector-rail">
                      <div>
                        <span>节点编码</span>
                        <strong>{selectedNode.node_code}</strong>
                      </div>
                      <div>
                        <span>排序</span>
                        <strong>{selectedNode.sort_order}</strong>
                      </div>
                    </div>

                    <div className="form-grid draft-inspector-form">
                      <label className="full-width">
                        事项
                        <input value={selectedNode.matter} disabled />
                      </label>
                      <label>
                        哪个人
                        <input value={selectedNode.owner} disabled />
                      </label>
                      <label>
                        哪个龙虾
                        <input value={selectedNode.agent} disabled />
                      </label>
                      <label className="full-width">
                        节点名称
                        <input value={selectedNode.node_name} disabled />
                      </label>
                    </div>
                  </>
                ) : (
                  <div className="draft-canvas-empty">
                    <strong>请选择节点</strong>
                    <p>点击画布中的任意节点后，这里会展示该节点的对齐字段。</p>
                  </div>
                )}
              </aside>
            </div>
          </div>
        ) : null}
      </article>
    </section>
  );
}

export default TemplateDetailPage;
