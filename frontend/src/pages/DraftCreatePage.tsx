import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import AgentSelect from "../components/AgentSelect";
import PersonSelect from "../components/PersonSelect";
import { useAgents } from "../hooks/useAgents";
import { useCreateDraft, useDrafts } from "../hooks/useDrafts";
import { getActor } from "../lib/actor";

function DraftCreatePage() {
  const navigate = useNavigate();
  const actor = getActor();
  const agentQuery = useMemo(
    () => ({
      page: 1,
      page_size: 100,
      status: 1 as const
    }),
    []
  );
  const recentDraftsQuery = useMemo(
    () => ({
      page: 1,
      page_size: 6,
      creator_person_id: actor.personId
    }),
    [actor.personId]
  );
  const agents = useAgents(agentQuery);
  const recentDrafts = useDrafts(recentDraftsQuery, !!actor.personId);
  const createDraftMutation = useCreateDraft();
  const [creatorPersonID, setCreatorPersonID] = useState<number | undefined>(actor.personId);
  const [plannerAgentID, setPlannerAgentID] = useState<number | undefined>(undefined);
  const [sourcePrompt, setSourcePrompt] = useState(
    "帮我创建一个视频绑定的工作流程。\n第一个自动节点是收集当前距离开课不足 2 天的课程场次数据。\n第二个节点是人工审核确认，由我负责确认。\n第三个节点是龙虾执行绑定操作，把录播课资源绑定到课程场次上。\n第四个节点是龙虾导出绑定结果，生成核查清单。\n第五个节点是最终结果签收，由指定同事确认完成。"
  );
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [submitError, setSubmitError] = useState("");

  useEffect(() => {
    if (plannerAgentID || !agents.data?.items.length) {
      return;
    }
    setPlannerAgentID(agents.data.items[0].id);
  }, [agents.data?.items, plannerAgentID]);

  return (
    <section className="draft-page-grid">
      <article className="page-card">
        <div className="page-title">
          <div>
            <span className="section-kicker">lobster planner</span>
            <h2>让龙虾生成流程草案</h2>
          </div>
          <div className="toolbar">
            <Link className="btn" to="/">
              返回工作台
            </Link>
            <Link className="btn" to="/drafts">
              草案列表
            </Link>
            <Link className="btn" to="/templates">
              浏览已有模板
            </Link>
          </div>
        </div>

        <p className="page-note">
          用自然语言把目标、执行动作、人工确认和最终签收说清楚。龙虾会先生成一份可编辑的流程草案，确认后再创建模板。
        </p>
        {submitError ? <p className="error-text">{submitError}</p> : null}

        <div className="form-grid">
          <label className="full-width">
            流程标题（可选）
            <input value={title} onChange={(event) => setTitle(event.target.value)} placeholder="例如：视频资源自动绑定流程" />
          </label>
          <label className="full-width">
            补充说明（可选）
            <textarea
              rows={3}
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              placeholder="补充范围、验收要求、失败后的处理方式。"
            />
          </label>
          <label>
            草案发起人 *
            <PersonSelect value={creatorPersonID} onChange={setCreatorPersonID} placeholder="请选择草案发起人" />
          </label>
          <label>
            规划龙虾 *
            <AgentSelect value={plannerAgentID} onChange={setPlannerAgentID} placeholder="请选择规划龙虾" />
          </label>
          <label className="full-width">
            需求描述 *
            <textarea
              rows={12}
              value={sourcePrompt}
              onChange={(event) => setSourcePrompt(event.target.value)}
              placeholder="描述你想让龙虾生成的流程，包括自动节点、人工确认节点、交付物和签收人。"
            />
          </label>
        </div>

        <div className="toolbar">
          <button
            type="button"
            className="btn btn-primary"
            disabled={createDraftMutation.loading}
            onClick={async () => {
              if (!creatorPersonID) {
                setSubmitError("请先选择草案发起人");
                return;
              }
              if (!plannerAgentID) {
                setSubmitError("请先选择规划龙虾");
                return;
              }
              if (!sourcePrompt.trim()) {
                setSubmitError("请先输入流程需求");
                return;
              }
              setSubmitError("");
              try {
                const draft = await createDraftMutation.run({
                  title: title.trim() || undefined,
                  description: description.trim() || undefined,
                  source_prompt: sourcePrompt.trim(),
                  creator_person_id: creatorPersonID,
                  planner_agent_id: plannerAgentID
                });
                navigate(`/drafts/${draft.id}/confirm`);
              } catch (err) {
                setSubmitError(err instanceof Error ? err.message : "生成草案失败");
              }
            }}
          >
            {createDraftMutation.loading ? "龙虾规划中..." : "生成流程草案"}
          </button>
        </div>
      </article>

      <aside className="page-card draft-side-panel">
        <div className="page-title">
          <div>
            <span className="section-kicker">recent drafts</span>
            <h2>最近草案</h2>
          </div>
        </div>
        {!actor.personId ? <p className="muted">设置发起人后，这里会显示 ta 最近创建的草案。</p> : null}
        {recentDrafts.loading ? <p>加载中...</p> : null}
        {recentDrafts.error ? <p className="error-text">{recentDrafts.error}</p> : null}
        <ul className="plain-list draft-mini-list">
          {(recentDrafts.data?.items ?? []).map((draft) => (
            <li key={draft.id}>
              <div className="comment-row">
                <Link to={`/drafts/${draft.id}/confirm`}>{draft.title || `草案 #${draft.id}`}</Link>
                <span className={`pill draft-status-${draft.status}`}>{draft.status}</span>
              </div>
              <p className="muted">{draft.description || draft.source_prompt.slice(0, 72) || "暂无描述"}</p>
            </li>
          ))}
        </ul>
        {!recentDrafts.loading && !recentDrafts.error && !(recentDrafts.data?.items.length ?? 0) ? (
          <p className="muted">还没有草案，可以先生成第一份。</p>
        ) : null}
        <div className="toolbar">
          <Link className="btn" to="/drafts">
            打开完整草案列表
          </Link>
        </div>
      </aside>
    </section>
  );
}

export default DraftCreatePage;
