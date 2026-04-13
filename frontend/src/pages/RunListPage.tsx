import { useMemo, useState } from "react";
import { Link, NavLink, useNavigate } from "react-router-dom";
import RunStatusTag from "../components/RunStatusTag";
import { useRuns } from "../hooks/useRuns";
import { useTemplates } from "../hooks/useTemplates";
import type { RunQuery } from "../types/api";

type RunListPageProps = {
  scope: RunQuery["scope"];
  title: string;
};

function RunListPage({ scope, title }: RunListPageProps) {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [status, setStatus] = useState("");
  const [ownerPersonIDInput, setOwnerPersonIDInput] = useState("");
  const [initiatorPersonIDInput, setInitiatorPersonIDInput] = useState("");
  const [ownerPersonID, setOwnerPersonID] = useState<number | undefined>(undefined);
  const [initiatorPersonID, setInitiatorPersonID] = useState<number | undefined>(undefined);

  const query = useMemo(
    () => ({
      page,
      page_size: pageSize,
      scope,
      status: status || undefined,
      owner_person_id: ownerPersonID,
      initiator_person_id: initiatorPersonID
    }),
    [initiatorPersonID, ownerPersonID, page, pageSize, scope, status]
  );

  const { data, loading, error, refetch } = useRuns(query);
  const { data: templatesData } = useTemplates({ page: 1, page_size: 200 });
  const templateNameMap = useMemo(() => {
    const map = new Map<number, string>();
    templatesData?.items.forEach((template) => map.set(template.id, template.name));
    return map;
  }, [templatesData]);

  const total = data?.total ?? 0;
  const totalPage = Math.max(1, Math.ceil(total / pageSize));

  return (
    <section className="page-card">
      <div className="page-title">
        <h2>{title}</h2>
        <button type="button" className="btn btn-primary" onClick={() => navigate("/templates")}>
          发起流程
        </button>
      </div>
      <div className="sub-nav">
        <NavLink to="/runs" end>
          全部流程
        </NavLink>
        <NavLink to="/runs/mine">我发起的</NavLink>
        <NavLink to="/runs/todo">待我处理</NavLink>
      </div>
      <div className="toolbar">
        <select value={status} onChange={(event) => setStatus(event.target.value)}>
          <option value="">全部状态</option>
          <option value="running">进行中</option>
          <option value="waiting">等待中</option>
          <option value="blocked">阻塞</option>
          <option value="completed">已完成</option>
          <option value="cancelled">已取消</option>
        </select>
        <input value={ownerPersonIDInput} onChange={(event) => setOwnerPersonIDInput(event.target.value)} placeholder="负责人 ID" />
        <input value={initiatorPersonIDInput} onChange={(event) => setInitiatorPersonIDInput(event.target.value)} placeholder="发起人 ID" />
        <button
          type="button"
          className="btn"
          onClick={() => {
            setPage(1);
            const owner = Number(ownerPersonIDInput);
            const initiator = Number(initiatorPersonIDInput);
            setOwnerPersonID(Number.isFinite(owner) && owner > 0 ? owner : undefined);
            setInitiatorPersonID(Number.isFinite(initiator) && initiator > 0 ? initiator : undefined);
          }}
        >
          筛选
        </button>
        <button type="button" className="btn" onClick={() => void refetch()}>
          刷新
        </button>
      </div>
      {error ? <p className="error-text">{error}</p> : null}

      <table className="table">
        <thead>
          <tr>
            <th>实例名称</th>
            <th>模板名称</th>
            <th>当前节点</th>
            <th>当前责任人</th>
            <th>状态</th>
            <th>发起人</th>
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
            data.items.map((item) => (
              <tr key={item.id}>
                <td>{item.title}</td>
                <td>{templateNameMap.get(item.template_id) ?? `#${item.template_id}`}</td>
                <td>{(item.current_node?.node_name ?? item.current_node_code) || "-"}</td>
                <td>{item.current_node?.owner_person?.name ?? "-"}</td>
                <td>
                  <RunStatusTag value={item.current_status} />
                </td>
                <td>{item.initiator?.name ?? `#${item.initiator_person_id}`}</td>
                <td>{new Date(item.updated_at).toLocaleString()}</td>
                <td>
                  <Link className="btn btn-text" to={`/runs/${item.id}`}>
                    查看详情
                  </Link>
                </td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan={8}>暂无流程</td>
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

export default RunListPage;
