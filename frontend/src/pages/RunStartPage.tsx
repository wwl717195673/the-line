import { useMemo } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import RunStartForm from "../components/RunStartForm";
import { useStartRun } from "../hooks/useRuns";
import { getActor, setActor } from "../lib/actor";
import { useTemplateDetail } from "../hooks/useTemplates";

function parseID(value?: string): number | undefined {
  if (!value) {
    return undefined;
  }
  const next = Number(value);
  return Number.isInteger(next) && next > 0 ? next : undefined;
}

function RunStartPage() {
  const params = useParams();
  const navigate = useNavigate();
  const templateID = parseID(params.templateId);
  const { data, loading, error, refetch } = useTemplateDetail(templateID);
  const startRunMutation = useStartRun();

  const defaultTitle = useMemo(() => {
    if (!data) {
      return "";
    }
    return `${data.name}-${new Date().toLocaleDateString()}`;
  }, [data]);

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
        <div>
          <span className="section-kicker">run start</span>
          <h2>发起流程</h2>
        </div>
        <div className="toolbar">
          <Link className="btn btn-primary" to="/drafts/create">
            让龙虾生成新草案
          </Link>
          <Link className="btn" to={`/templates/${templateID}`}>
            返回模板详情
          </Link>
          <button type="button" className="btn" onClick={() => void refetch()}>
            刷新模板
          </button>
        </div>
      </div>

      {loading ? <p>加载模板中...</p> : null}
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
          <p className="muted">
            当前模板：{data.name}（{data.code}）
          </p>
          <p className="page-note">
            如果这份模板还不够贴合当前业务场景，可以先回到 <Link to="/drafts/create">龙虾草案入口</Link> 重新生成一版再确认。
          </p>
          <RunStartForm
            defaultTitle={defaultTitle}
            submitting={startRunMutation.loading}
            onSubmit={async (values) => {
              const detail = await startRunMutation.run({
                template_id: templateID,
                title: values.title,
                initiator_person_id: values.initiator_person_id,
                input_payload_json: {
                  form_data: {
                    reason: values.reason,
                    class_info: values.class_info,
                    current_teacher: values.current_teacher,
                    expected_time: values.expected_time,
                    extra_note: values.extra_note
                  },
                  attachments: [],
                  comments_context: []
                }
              });
              const actor = getActor();
              if (!actor.personId && detail.initiator_person_id) {
                setActor({
                  personId: detail.initiator_person_id,
                  roleType: detail.initiator?.role_type || undefined
                });
              }
              navigate(`/runs/${detail.id}`);
            }}
          />
        </>
      ) : null}
    </section>
  );
}

export default RunStartPage;
