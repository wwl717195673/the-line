import { useMemo, useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import DeliverableStatusTag from "../components/DeliverableStatusTag";
import { useDeliverables } from "../hooks/useDeliverables";
import type { DeliverableQuery } from "../types/api";

function DeliverableListPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();
  const statusInQuery = searchParams.get("review_status") ?? "";
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [reviewerInput, setReviewerInput] = useState("");
  const [reviewerID, setReviewerID] = useState<number | undefined>(undefined);

  const query = useMemo<DeliverableQuery>(
    () => ({
      page,
      page_size: pageSize,
      review_status: (statusInQuery || undefined) as DeliverableQuery["review_status"],
      reviewer_person_id: reviewerID
    }),
    [page, pageSize, reviewerID, statusInQuery]
  );

  const { data, loading, error, refetch } = useDeliverables(query);
  const total = data?.total ?? 0;
  const totalPage = Math.max(1, Math.ceil(total / pageSize));
  const hasData = !!data?.items.length;

  return (
    <section className="page-card">
      <div className="page-title">
        <h2>交付中心</h2>
        <div className="toolbar">
          <button type="button" className="btn" onClick={() => navigate("/runs")}>
            去流程中心
          </button>
          <button type="button" className="btn" onClick={() => void refetch()}>
            刷新
          </button>
        </div>
      </div>

      <div className="toolbar">
        <button
          type="button"
          className={`btn ${statusInQuery === "" ? "btn-primary" : ""}`}
          onClick={() => {
            setSearchParams({});
            setPage(1);
          }}
        >
          全部交付
        </button>
        <button
          type="button"
          className={`btn ${statusInQuery === "pending" ? "btn-primary" : ""}`}
          onClick={() => {
            setSearchParams({ review_status: "pending" });
            setPage(1);
          }}
        >
          待验收
        </button>
        <button
          type="button"
          className={`btn ${statusInQuery === "approved" ? "btn-primary" : ""}`}
          onClick={() => {
            setSearchParams({ review_status: "approved" });
            setPage(1);
          }}
        >
          已归档
        </button>
      </div>

      <p className="page-note">支持按验收状态和验收人筛选，默认展示全部交付记录。</p>

      <div className="toolbar">
        <input
          value={reviewerInput}
          onChange={(event) => setReviewerInput(event.target.value)}
          placeholder="按验收人 ID 筛选"
        />
        <button
          type="button"
          className="btn"
          onClick={() => {
            const next = Number(reviewerInput);
            setReviewerID(Number.isFinite(next) && next > 0 ? next : undefined);
            setPage(1);
          }}
        >
          筛选
        </button>
      </div>

      {error ? <p className={hasData ? "warning-text" : "error-text"}>{hasData ? `最新刷新失败，当前展示上次结果：${error}` : error}</p> : null}

      <table className="table">
        <thead>
          <tr>
            <th>交付标题</th>
            <th>关联流程</th>
            <th>发起人</th>
            <th>验收人</th>
            <th>验收状态</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr>
              <td colSpan={7}>加载中...</td>
            </tr>
          ) : data?.items.length ? (
            data.items.map((item) => (
              <tr key={item.id}>
                <td>{item.title}</td>
                <td>{item.run?.title ?? `#${item.run_id}`}</td>
                <td>{item.run?.initiator?.name ?? "-"}</td>
                <td>{item.reviewer?.name ?? `#${item.reviewer_person_id}`}</td>
                <td>
                  <DeliverableStatusTag value={item.review_status} />
                </td>
                <td>{new Date(item.created_at).toLocaleString()}</td>
                <td>
                  <Link className="btn btn-text" to={`/deliverables/${item.id}`}>
                    查看详情
                  </Link>
                </td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan={7}>暂无交付物</td>
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
    </section>
  );
}

export default DeliverableListPage;
