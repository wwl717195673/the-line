import { Link, useParams } from "react-router-dom";
import TemplateNodeTimeline from "../components/TemplateNodeTimeline";
import { useTemplateDetail } from "../hooks/useTemplates";

function parseID(value?: string): number | undefined {
  if (!value) {
    return undefined;
  }
  const next = Number(value);
  return Number.isInteger(next) && next > 0 ? next : undefined;
}

function TemplateDetailPage() {
  const params = useParams();
  const templateID = parseID(params.templateId);
  const { data, loading, error, refetch } = useTemplateDetail(templateID);

  if (!templateID) {
    return (
      <section className="page-card">
        <p className="error-text">模板 ID 不合法</p>
      </section>
    );
  }

  return (
    <section className="page-card">
      <div className="page-title">
        <h2>模板详情</h2>
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

      {loading ? <p>加载中...</p> : null}
      {error ? (
        <div>
          <p className="error-text">模板不存在或已下线：{error}</p>
          <button type="button" className="btn" onClick={() => void refetch()}>
            重试
          </button>
        </div>
      ) : null}

      {data ? (
        <>
          <section className="kv-grid">
            <p>
              <b>模板名称：</b>
              {data.name}
            </p>
            <p>
              <b>模板编码：</b>
              {data.code}
            </p>
            <p>
              <b>版本：</b>
              {data.version}
            </p>
            <p>
              <b>分类：</b>
              {data.category}
            </p>
            <p>
              <b>状态：</b>
              {data.status}
            </p>
            <p>
              <b>更新时间：</b>
              {new Date(data.updated_at).toLocaleString()}
            </p>
            <p className="full-width">
              <b>模板说明：</b>
              {data.description || "-"}
            </p>
          </section>
          <h3>节点时间线（只读）</h3>
          <TemplateNodeTimeline nodes={data.nodes} />
        </>
      ) : null}
    </section>
  );
}

export default TemplateDetailPage;
