import { useMemo, useState } from "react";
import { NavLink } from "react-router-dom";
import { useFeedback } from "../components/FeedbackProvider";
import AgentFormModal from "../components/AgentFormModal";
import StatusTag from "../components/StatusTag";
import { useAgents, useCreateAgent, useDisableAgent, useUpdateAgent } from "../hooks/useAgents";
import { usePersons } from "../hooks/usePersons";
import type { Agent, Status } from "../types/api";

function AgentListPage() {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [status, setStatus] = useState<string>("all");
  const [keyword, setKeyword] = useState("");
  const [keywordInput, setKeywordInput] = useState("");
  const [editing, setEditing] = useState<Agent | undefined>(undefined);
  const [openForm, setOpenForm] = useState(false);
  const [actionError, setActionError] = useState("");

  const query = useMemo(
    () => ({
      page,
      page_size: pageSize,
      status: status === "all" ? undefined : (Number(status) as Status),
      keyword
    }),
    [keyword, page, pageSize, status]
  );

  const { data, loading, error, refetch } = useAgents(query);
  const createMutation = useCreateAgent();
  const updateMutation = useUpdateAgent();
  const disableMutation = useDisableAgent();
  const { confirm, notify } = useFeedback();

  const { data: personsData } = usePersons({
    page: 1,
    page_size: 200
  });
  const personNameMap = useMemo(() => {
    const map = new Map<number, string>();
    personsData?.items.forEach((person) => map.set(person.id, person.name));
    return map;
  }, [personsData]);

  const total = data?.total ?? 0;
  const totalPage = Math.max(1, Math.ceil(total / pageSize));

  const openCreate = () => {
    setEditing(undefined);
    setActionError("");
    setOpenForm(true);
  };

  const openEdit = (agent: Agent) => {
    setEditing(agent);
    setActionError("");
    setOpenForm(true);
  };

  const onDisable = async (agent: Agent) => {
    if (agent.status === 0) {
      return;
    }
    const confirmed = await confirm({
      title: "停用龙虾",
      message: `确认停用龙虾「${agent.name}」吗？停用后自动节点将不能继续分配给它。`,
      confirmText: "确认停用",
      tone: "danger"
    });
    if (!confirmed) {
      return;
    }
    setActionError("");
    try {
      await disableMutation.run(agent.id);
      notify({
        title: "龙虾已停用",
        message: `${agent.name} 已从可执行列表中移除。`,
        tone: "success"
      });
      await refetch();
    } catch (err) {
      const message = err instanceof Error ? err.message : "停用失败";
      setActionError(message);
      notify({
        title: "停用失败",
        message,
        tone: "danger"
      });
    }
  };

  return (
    <section className="page-card">
      <div className="page-title">
        <h2>龙虾管理</h2>
        <button type="button" className="btn btn-primary" onClick={openCreate}>
          新建龙虾
        </button>
      </div>
      <div className="sub-nav">
        <NavLink to="/resources/persons">人员管理</NavLink>
        <NavLink to="/resources/agents">龙虾管理</NavLink>
      </div>

      <div className="toolbar">
        <input value={keywordInput} onChange={(event) => setKeywordInput(event.target.value)} placeholder="按名称或编码搜索" />
        <select value={status} onChange={(event) => setStatus(event.target.value)}>
          <option value="all">全部状态</option>
          <option value="1">仅启用</option>
          <option value="0">仅停用</option>
        </select>
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
      {actionError ? <p className="error-text">{actionError}</p> : null}

      <table className="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>名称</th>
            <th>编码</th>
            <th>来源</th>
            <th>版本</th>
            <th>维护人</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr>
              <td colSpan={8}>加载中...</td>
            </tr>
          ) : data?.items.length ? (
            data.items.map((agent) => (
              <tr key={agent.id}>
                <td>{agent.id}</td>
                <td>{agent.name}</td>
                <td>{agent.code}</td>
                <td>{agent.provider}</td>
                <td>{agent.version}</td>
                <td>{personNameMap.get(agent.owner_person_id) ?? `#${agent.owner_person_id}`}</td>
                <td>
                  <StatusTag value={agent.status} />
                </td>
                <td>
                  <button type="button" className="btn btn-text" onClick={() => openEdit(agent)}>
                    编辑
                  </button>
                  <button type="button" className="btn btn-text danger" onClick={() => void onDisable(agent)} disabled={agent.status === 0}>
                    停用
                  </button>
                </td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan={8}>暂无龙虾</td>
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

      <AgentFormModal
        open={openForm}
        initial={editing}
        submitting={createMutation.loading || updateMutation.loading}
        onCancel={() => setOpenForm(false)}
        onSubmit={async (values) => {
          setActionError("");
          if (editing) {
            await updateMutation.run(editing.id, values);
          } else {
            await createMutation.run({
              name: values.name,
              code: values.code,
              provider: values.provider,
              version: values.version,
              owner_person_id: values.owner_person_id,
              config_json: values.config_json
            });
          }
          setOpenForm(false);
          await refetch();
        }}
      />
    </section>
  );
}

export default AgentListPage;
