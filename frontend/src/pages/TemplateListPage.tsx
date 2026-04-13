import { useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useFeedback } from "../components/FeedbackProvider";
import { useDeleteTemplate, useTemplates } from "../hooks/useTemplates";

function TemplateListPage() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [keywordInput, setKeywordInput] = useState("");
  const [keyword, setKeyword] = useState("");

  const query = useMemo(
    () => ({
      page,
      page_size: pageSize,
      keyword
    }),
    [keyword, page, pageSize]
  );

  const { data, loading, error, refetch } = useTemplates(query);
  const deleteTemplateMutation = useDeleteTemplate();
  const { confirm, notify } = useFeedback();
  const total = data?.total ?? 0;
  const totalPage = Math.max(1, Math.ceil(total / pageSize));

  return (
    <section className="page-card">
      <div className="page-title">
        <h2>模板中心</h2>
      </div>
      <div className="toolbar">
        <input value={keywordInput} onChange={(event) => setKeywordInput(event.target.value)} placeholder="按模板名称或编码搜索" />
        <button
          type="button"
          className="btn"
          onClick={() => {
            setPage(1);
            setKeyword(keywordInput.trim());
          }}
        >
          查询
        </button>
        <button type="button" className="btn" onClick={() => void refetch()}>
          刷新
        </button>
      </div>

      {error ? <p className="error-text">{error}</p> : null}

      <table className="table">
        <thead>
          <tr>
            <th>模板名称</th>
            <th>编码</th>
            <th>版本</th>
            <th>分类</th>
            <th>状态</th>
            <th>说明</th>
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
                <td>{item.name}</td>
                <td>{item.code}</td>
                <td>{item.version}</td>
                <td>{item.category}</td>
                <td>{item.status}</td>
                <td>{item.description || "-"}</td>
                <td>{new Date(item.updated_at).toLocaleString()}</td>
                <td>
                  <Link className="btn btn-text" to={`/templates/${item.id}`}>
                    查看详情
                  </Link>
                  {item.status === "published" ? (
                    <button type="button" className="btn btn-text" onClick={() => navigate(`/templates/${item.id}/start`)}>
                      使用模板
                    </button>
                  ) : null}
                  <button
                    type="button"
                    className="btn btn-text danger"
                    disabled={deleteTemplateMutation.loading}
                    onClick={async () => {
                      const confirmed = await confirm({
                        title: "删除模板",
                        message: `确认删除模板「${item.name}」吗？如果它已经被流程实例引用，系统会阻止删除。`,
                        confirmText: "确认删除",
                        tone: "danger"
                      });
                      if (!confirmed) {
                        return;
                      }
                      try {
                        await deleteTemplateMutation.run(item.id);
                        notify({
                          title: "模板已删除",
                          message: `${item.name} 已从模板中心移除。`,
                          tone: "success"
                        });
                        await refetch();
                      } catch (err) {
                        notify({
                          title: "删除模板失败",
                          message: err instanceof Error ? err.message : "删除模板失败",
                          tone: "danger"
                        });
                      }
                    }}
                  >
                    删除
                  </button>
                </td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan={8}>暂无模板</td>
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

      {deleteTemplateMutation.loading ? <p className="muted">正在删除模板...</p> : null}
    </section>
  );
}

export default TemplateListPage;
