import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import AttachmentList from "../components/AttachmentList";
import DeliverableStatusTag from "../components/DeliverableStatusTag";
import NodeLogTimeline from "../components/NodeLogTimeline";
import ReviewDeliverableModal from "../components/ReviewDeliverableModal";
import { useDeliverableDetail, useReviewDeliverable } from "../hooks/useDeliverables";
import { getActor } from "../lib/actor";
import type { RunNodeLog } from "../types/api";

function parseID(value?: string): number | undefined {
  if (!value) {
    return undefined;
  }
  const next = Number(value);
  return Number.isInteger(next) && next > 0 ? next : undefined;
}

function parseNodeSummary(resultJSON: unknown): Array<{ node_name?: string; status?: string; completed_at?: string }> {
  if (!resultJSON || typeof resultJSON !== "object") {
    return [];
  }
  const obj = resultJSON as { node_summary?: unknown };
  if (!Array.isArray(obj.node_summary)) {
    return [];
  }
  return obj.node_summary as Array<{ node_name?: string; status?: string; completed_at?: string }>;
}

function canReview(reviewerPersonID: number, reviewStatus: string): boolean {
  if (reviewStatus !== "pending") {
    return false;
  }
  const actor = getActor();
  if (actor.roleType === "admin") {
    return true;
  }
  return !!actor.personId && actor.personId === reviewerPersonID;
}

function DeliverableDetailPage() {
  const params = useParams();
  const deliverableID = parseID(params.deliverableId);
  const { data, loading, error, refetch } = useDeliverableDetail(deliverableID);
  const reviewMutation = useReviewDeliverable();
  const [reviewModal, setReviewModal] = useState<"approved" | "rejected" | "">("");
  const [actionError, setActionError] = useState("");

  const nodeSummary = useMemo(() => parseNodeSummary(data?.result_json), [data?.result_json]);
  const nodeLogs = useMemo<RunNodeLog[]>(
    () =>
      (data?.nodes ?? []).map((node) => ({
        id: node.id,
        run_id: node.run_id,
        run_node_id: node.id,
        log_type: node.status,
        operator_type: "system",
        operator_id: 0,
        content: `${node.node_name} - ${node.status}`,
        extra_json: {},
        created_at: node.updated_at
      })),
    [data?.nodes]
  );

  if (!deliverableID) {
    return (
      <section className="page-card">
        <p className="error-text">交付物 ID 不合法</p>
      </section>
    );
  }

  return (
    <section className="page-card">
      <div className="page-title">
        <h2>交付详情</h2>
        <div className="toolbar">
          <Link className="btn" to="/deliverables">
            返回交付中心
          </Link>
          <button type="button" className="btn" onClick={() => void refetch()}>
            刷新
          </button>
          {data && canReview(data.reviewer_person_id, data.review_status) ? (
            <>
              <button type="button" className="btn btn-primary" onClick={() => setReviewModal("approved")}>
                验收通过
              </button>
              <button type="button" className="btn danger" onClick={() => setReviewModal("rejected")}>
                验收驳回
              </button>
            </>
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
              <b>交付标题：</b>
              {data.title}
            </p>
            <p>
              <b>验收状态：</b>
              <DeliverableStatusTag value={data.review_status} />
            </p>
            <p>
              <b>关联流程：</b>
              {data.run?.title ?? `#${data.run_id}`}
            </p>
            <p>
              <b>流程发起人：</b>
              {data.run?.initiator?.name ?? "-"}
            </p>
            <p>
              <b>验收人：</b>
              {data.reviewer?.name ?? `#${data.reviewer_person_id}`}
            </p>
            <p>
              <b>验收时间：</b>
              {data.reviewed_at ? new Date(data.reviewed_at).toLocaleString() : "-"}
            </p>
            <p className="full-width">
              <b>交付摘要：</b>
              {data.summary}
            </p>
            <p className="full-width">
              <b>验收意见：</b>
              {typeof data.result_json === "object" && data.result_json && "review_comment" in data.result_json
                ? String((data.result_json as Record<string, unknown>).review_comment ?? "-")
                : "-"}
            </p>
          </section>

          <section className="flow-comment-section">
            <h3>节点完成情况</h3>
            {nodeSummary.length ? (
              <ul className="plain-list">
                {nodeSummary.map((item, index) => (
                  <li key={`${item.node_name ?? "node"}_${index}`}>
                    <span>{item.node_name ?? "-"}</span>
                    <span className="muted">状态：{item.status ?? "-"}</span>
                    <span className="muted">完成时间：{item.completed_at ? new Date(item.completed_at).toLocaleString() : "-"}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="muted">暂无节点摘要，展示流程节点状态列表。</p>
            )}
            {!nodeSummary.length ? <NodeLogTimeline logs={nodeLogs} /> : null}
          </section>

          <section className="flow-comment-section">
            <h3>关键附件</h3>
            <AttachmentList attachments={data.attachments} />
          </section>
        </>
      ) : null}

      <ReviewDeliverableModal
        open={reviewModal !== ""}
        submitting={reviewMutation.loading}
        reviewStatus={reviewModal === "approved" ? "approved" : "rejected"}
        onCancel={() => setReviewModal("")}
        onSubmit={async (payload) => {
          if (!deliverableID) {
            return;
          }
          setActionError("");
          await reviewMutation.run(deliverableID, payload);
          setReviewModal("");
          await refetch();
        }}
      />
    </section>
  );
}

export default DeliverableDetailPage;
