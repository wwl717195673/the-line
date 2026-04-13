import { useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useFeedback } from "../components/FeedbackProvider";
import { useAgents } from "../hooks/useAgents";
import { useDeleteDraft, useDrafts } from "../hooks/useDrafts";
import { getActor } from "../lib/actor";
import type { FlowDraft } from "../types/api";

function DraftListPage() {
  const navigate = useNavigate();
  const actor = getActor();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [status, setStatus] = useState<FlowDraft["status"] | "">("");
  const [scope, setScope] = useState<"all" | "mine">(actor.personId ? "mine" : "all");

  const query = useMemo(
    () => ({
      page,
      page_size: pageSize,
      status: status || undefined,
      creator_person_id: scope === "mine" ? actor.personId : undefined
    }),
    [actor.personId, page, pageSize, scope, status]
  );
  const agentQuery = useMemo(
    () => ({
      page: 1,
      page_size: 100,
      status: 1 as const
    }),
    []
  );
  const { data, loading, error, refetch } = useDrafts(query);
  const deleteDraftMutation = useDeleteDraft();
  const { confirm, notify } = useFeedback();
  const { data: agentsData } = useAgents(agentQuery);
  const plannerNameMap = useMemo(() => {
    const map = new Map<number, string>();
    agentsData?.items.forEach((agent) => map.set(agent.id, `${agent.name}（${agent.code}）`));
    return map;
  }, [agentsData]);

  const total = data?.total ?? 0;
  const totalPage = Math.max(1, Math.ceil(total / pageSize));

  return (
    <section className="page-card">
      <div className="page-title">
        <div>
          <span className="section-kicker">draft center</span>
          <h2>流程草案</h2>
        </div>
        <div className="toolbar">
          <button type="button" className="btn btn-primary" onClick={() => navigate("/drafts/create")}>
            新建草案
          </button>
          <button type="button" className="btn" onClick={() => void refetch()}>
            刷新
          </button>
        </div>
      </div>

      <p className="page-note">这里集中查看龙虾生成的流程草案，继续编辑、确认成模板，或回到历史草案复用已有编排。</p>

      <div className="toolbar">
        <select
          value={scope}
          onChange={(event) => {
            setPage(1);
            setScope(event.target.value as "all" | "mine");
          }}
        >
          <option value="all">全部草案</option>
          <option value="mine" disabled={!actor.personId}>
            只看我的草案
          </option>
        </select>
        <select
          value={status}
          onChange={(event) => {
            setPage(1);
            setStatus(event.target.value as FlowDraft["status"] | "");
          }}
        >
          <option value="">全部状态</option>
          <option value="draft">草稿中</option>
          <option value="confirmed">已确认</option>
          <option value="discarded">已废弃</option>
        </select>
      </div>

      {error ? <p className="error-text">{error}</p> : null}

      <table className="table">
        <thead>
          <tr>
            <th>草案标题</th>
            <th>状态</th>
            <th>发起人</th>
            <th>规划龙虾</th>
            <th>节点数</th>
            <th>最终交付</th>
            <th>更新时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr>
              <td colSpan={8}>加载中...</td>
            </tr>
          ) : data?.items.length ? (
            data.items.map((draft) => (
              <tr key={draft.id}>
                <td>
                  <div className="table-primary">
                    <strong>{draft.title || `草案 #${draft.id}`}</strong>
                    <p>{draft.description || draft.source_prompt.slice(0, 72) || "-"}</p>
                  </div>
                </td>
                <td>
                  <span className={`pill draft-status-${draft.status}`}>{draft.status}</span>
                </td>
                <td>#{draft.creator_person_id}</td>
                <td>{draft.planner_agent_id ? plannerNameMap.get(draft.planner_agent_id) ?? `#${draft.planner_agent_id}` : "-"}</td>
                <td>{draft.structured_plan_json.nodes.length}</td>
                <td>{draft.structured_plan_json.final_deliverable || "-"}</td>
                <td>{new Date(draft.updated_at).toLocaleString()}</td>
                <td>
                  <Link className="btn btn-text" to={`/drafts/${draft.id}/confirm`}>
                    {draft.status === "draft" ? "继续编辑" : "查看详情"}
                  </Link>
                  {draft.status !== "confirmed" ? (
                    <button
                      type="button"
                      className="btn btn-text danger"
                      disabled={deleteDraftMutation.loading}
                      onClick={async () => {
                        const confirmed = await confirm({
                          title: "删除流程草案",
                          message: `确认删除草案「${draft.title || `#${draft.id}`}」吗？删除后将无法继续编辑这份草案。`,
                          confirmText: "确认删除",
                          tone: "danger"
                        });
                        if (!confirmed) {
                          return;
                        }
                        try {
                          await deleteDraftMutation.run(draft.id);
                          notify({
                            title: "草案已删除",
                            message: `${draft.title || `草案 #${draft.id}`} 已从草案池移除。`,
                            tone: "success"
                          });
                          await refetch();
                        } catch (err) {
                          notify({
                            title: "删除草案失败",
                            message: err instanceof Error ? err.message : "删除草案失败",
                            tone: "danger"
                          });
                        }
                      }}
                    >
                      删除
                    </button>
                  ) : null}
                  {draft.confirmed_template_id ? (
                    <Link className="btn btn-text" to={`/templates/${draft.confirmed_template_id}/start`}>
                      发起流程
                    </Link>
                  ) : null}
                </td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan={8}>暂无草案</td>
            </tr>
          )}
        </tbody>
      </table>

      <div className="pager">
        <span>
          共 {total} 条 / 第 {page} 页
        </span>
        <select
          value={pageSize}
          onChange={(event) => {
            setPageSize(Number(event.target.value));
            setPage(1);
          }}
        >
          <option value={10}>10 / 页</option>
          <option value={20}>20 / 页</option>
          <option value={50}>50 / 页</option>
        </select>
        <button type="button" className="btn" onClick={() => setPage((value) => Math.max(1, value - 1))} disabled={page <= 1}>
          上一页
        </button>
        <button type="button" className="btn" onClick={() => setPage((value) => Math.min(totalPage, value + 1))} disabled={page >= totalPage}>
          下一页
        </button>
      </div>

      {deleteDraftMutation.loading ? <p className="muted">正在删除草案...</p> : null}
    </section>
  );
}

export default DraftListPage;
