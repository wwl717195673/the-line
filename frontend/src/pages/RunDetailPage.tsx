import { useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import CancelRunModal from "../components/CancelRunModal";
import CommentEditor from "../components/CommentEditor";
import CommentList from "../components/CommentList";
import CreateDeliverableModal from "../components/CreateDeliverableModal";
import RunNodeTimeline from "../components/RunNodeTimeline";
import RunNodeWorkbench from "../components/RunNodeWorkbench";
import RunStatusTag from "../components/RunStatusTag";
import { useCollaborationActions, useComments } from "../hooks/useCollaboration";
import { useCreateDeliverable } from "../hooks/useDeliverables";
import { useCancelRun, useRunDetail } from "../hooks/useRuns";
import { getActor } from "../lib/actor";
import type { RunNode } from "../types/api";

function parseID(value?: string): number | undefined {
  if (!value) {
    return undefined;
  }
  const next = Number(value);
  return Number.isInteger(next) && next > 0 ? next : undefined;
}

function canShowCancel(runInitiatorID: number, runStatus: string): boolean {
  if (runStatus === "completed" || runStatus === "cancelled") {
    return false;
  }
  const actor = getActor();
  if (actor.roleType === "admin") {
    return true;
  }
  return !!actor.personId && actor.personId === runInitiatorID;
}

function RunDetailPage() {
  const params = useParams();
  const navigate = useNavigate();
  const runID = parseID(params.runId);
  const { data, loading, error, refetch } = useRunDetail(runID);
  const cancelMutation = useCancelRun();
  const createDeliverableMutation = useCreateDeliverable();
  const flowComments = useComments("flow_run", runID);
  const collaborationActions = useCollaborationActions();
  const [openCancel, setOpenCancel] = useState(false);
  const [openCreateDeliverable, setOpenCreateDeliverable] = useState(false);
  const [selectedNodeID, setSelectedNodeID] = useState<number | undefined>(undefined);
  const [actionError, setActionError] = useState("");

  const refreshAll = async () => {
    await refetch();
    await flowComments.refetch();
  };

  const selectedNode = useMemo<RunNode | undefined>(() => {
    if (!data?.nodes.length) {
      return undefined;
    }
    const nextID = selectedNodeID ?? data.current_node?.id ?? data.nodes.find((node) => node.is_current)?.id ?? data.nodes[0].id;
    return data.nodes.find((node) => node.id === nextID) ?? data.nodes[0];
  }, [data, selectedNodeID]);

  const actor = getActor();
  const canCreateDeliverable =
    !!data &&
    data.current_status === "completed" &&
    !data.has_deliverable &&
    (actor.roleType === "admin" ||
      actor.personId === data.initiator_person_id ||
      actor.personId === data.nodes[data.nodes.length - 1]?.owner_person_id);

  if (!runID) {
    return (
      <section className="page-card">
        <p className="error-text">流程 ID 不合法</p>
      </section>
    );
  }

  return (
    <section className="page-card">
      <div className="page-title">
        <h2>流程详情</h2>
        <div className="toolbar">
          <Link className="btn" to="/runs">
            返回流程列表
          </Link>
          <button type="button" className="btn" onClick={() => void refetch()}>
            刷新
          </button>
          {data && canShowCancel(data.initiator_person_id, data.current_status) ? (
            <button type="button" className="btn danger" onClick={() => setOpenCancel(true)}>
              取消流程
            </button>
          ) : null}
        </div>
      </div>

      {loading ? <p>加载中...</p> : null}
      {error ? <p className="error-text">{error}</p> : null}
      {actionError ? <p className="error-text">{actionError}</p> : null}

      {data ? (
        <>
          <section className="run-summary-grid">
            <p>
              <b>流程标题：</b>
              {data.title}
            </p>
            <p>
              <b>状态：</b>
              <RunStatusTag value={data.current_status} />
            </p>
            <p>
              <b>发起人：</b>
              {data.initiator?.name ?? `#${data.initiator_person_id}`}
            </p>
            <p>
              <b>当前节点：</b>
              {(data.current_node?.node_name ?? data.current_node_code) || "-"}
            </p>
            <p>
              <b>当前责任人：</b>
              {data.current_node?.owner_person?.name ?? "-"}
            </p>
            <p>
              <b>开始时间：</b>
              {data.started_at ? new Date(data.started_at).toLocaleString() : "-"}
            </p>
            <p className="full-width">
              <b>模板：</b>
              {data.template?.name ?? `#${data.template_id}`}
            </p>
          </section>

          <div className="run-detail-layout">
            <section>
              <h3>节点时间线</h3>
              <RunNodeTimeline
                nodes={data.nodes}
                selectedNodeID={selectedNode?.id}
                onSelect={(node) => {
                  setSelectedNodeID(node.id);
                }}
              />
            </section>
            <RunNodeWorkbench nodeID={selectedNode?.id} runStatus={data.current_status} onMutated={refreshAll} />
          </div>

          <section className="flow-comment-section">
            <h3>流程评论</h3>
            {data.current_status !== "cancelled" ? (
              <CommentEditor
                submitting={collaborationActions.loading}
                placeholder="在流程上下文中发布评论（支持输入 @张三 文本）"
                onSubmit={async (content) => {
                  await collaborationActions.createComment("flow_run", data.id, content);
                  await flowComments.refetch();
                }}
              />
            ) : (
              <p className="muted">流程已取消，评论区只读。</p>
            )}
            {flowComments.error ? <p className="error-text">{flowComments.error}</p> : null}
            <CommentList
              comments={flowComments.data}
              resolving={collaborationActions.loading}
              onResolve={
                data.current_status === "cancelled"
                  ? undefined
                  : async (commentID) => {
                      await collaborationActions.resolveComment(commentID);
                      await flowComments.refetch();
                    }
              }
            />
          </section>

          {data.current_status === "completed" ? (
            <div className="deliverable-entry">
              {data.has_deliverable && data.deliverable_id ? (
                <>
                  <p className="muted">流程已完成，交付物已生成。</p>
                  <Link to={`/deliverables/${data.deliverable_id}`} className="btn btn-primary">
                    查看交付物
                  </Link>
                </>
              ) : (
                <>
                  <p className="muted">流程已完成，可生成交付物。</p>
                  {canCreateDeliverable ? (
                    <button type="button" className="btn btn-primary" onClick={() => setOpenCreateDeliverable(true)}>
                      生成交付物
                    </button>
                  ) : (
                    <Link to="/deliverables" className="btn">
                      进入交付中心
                    </Link>
                  )}
                </>
              )}
            </div>
          ) : null}
        </>
      ) : null}

      <CancelRunModal
        open={openCancel}
        submitting={cancelMutation.loading}
        onCancel={() => setOpenCancel(false)}
        onSubmit={async (reason) => {
          if (!runID) {
            return;
          }
          setActionError("");
          await cancelMutation.run(runID, reason);
          setOpenCancel(false);
          await refreshAll();
        }}
      />

      {data ? (
        <CreateDeliverableModal
          open={openCreateDeliverable}
          runID={data.id}
          runTitle={data.title}
          runNodes={data.nodes}
          submitting={createDeliverableMutation.loading}
          onCancel={() => setOpenCreateDeliverable(false)}
          onSubmit={async (payload) => {
            setActionError("");
            const detail = await createDeliverableMutation.run({
              run_id: data.id,
              title: payload.title,
              summary: payload.summary,
              reviewer_person_id: payload.reviewer_person_id,
              result_json: payload.result_json,
              attachment_ids: payload.attachment_ids
            });
            setOpenCreateDeliverable(false);
            await refreshAll();
            navigate(`/deliverables/${detail.id}`);
          }}
        />
      ) : null}
    </section>
  );
}

export default RunDetailPage;
