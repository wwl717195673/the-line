import { useMemo, useState } from "react";
import { NavLink } from "react-router-dom";
import { useFeedback } from "../components/FeedbackProvider";
import PersonFormModal from "../components/PersonFormModal";
import StatusTag from "../components/StatusTag";
import { useCreatePerson, useDisablePerson, usePersons, useUpdatePerson } from "../hooks/usePersons";
import type { Person, Status } from "../types/api";

function PersonListPage() {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [status, setStatus] = useState<string>("all");
  const [keyword, setKeyword] = useState("");
  const [keywordInput, setKeywordInput] = useState("");
  const [editing, setEditing] = useState<Person | undefined>(undefined);
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

  const { data, loading, error, refetch } = usePersons(query);
  const createMutation = useCreatePerson();
  const updateMutation = useUpdatePerson();
  const disableMutation = useDisablePerson();
  const { confirm, notify } = useFeedback();

  const total = data?.total ?? 0;
  const totalPage = Math.max(1, Math.ceil(total / pageSize));

  const openCreate = () => {
    setEditing(undefined);
    setActionError("");
    setOpenForm(true);
  };

  const openEdit = (person: Person) => {
    setEditing(person);
    setActionError("");
    setOpenForm(true);
  };

  const onDisable = async (person: Person) => {
    if (person.status === 0) {
      return;
    }
    const confirmed = await confirm({
      title: "停用人员",
      message: `确认停用人员「${person.name}」吗？停用后该人员将不能继续参与新流程。`,
      confirmText: "确认停用",
      tone: "danger"
    });
    if (!confirmed) {
      return;
    }
    setActionError("");
    try {
      await disableMutation.run(person.id);
      notify({
        title: "人员已停用",
        message: `${person.name} 已从启用列表移除。`,
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
        <h2>人员管理</h2>
        <button type="button" className="btn btn-primary" onClick={openCreate}>
          新建人员
        </button>
      </div>
      <div className="sub-nav">
        <NavLink to="/resources/persons">人员管理</NavLink>
        <NavLink to="/resources/agents">龙虾管理</NavLink>
      </div>

      <div className="toolbar">
        <input value={keywordInput} onChange={(event) => setKeywordInput(event.target.value)} placeholder="按姓名或邮箱搜索" />
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
        <button
          type="button"
          className="btn"
          onClick={() => {
            void refetch();
          }}
        >
          刷新
        </button>
      </div>

      {error ? <p className="error-text">{error}</p> : null}
      {actionError ? <p className="error-text">{actionError}</p> : null}

      <table className="table">
        <thead>
          <tr>
            <th>ID</th>
            <th>姓名</th>
            <th>邮箱</th>
            <th>角色</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          {loading ? (
            <tr>
              <td colSpan={6}>加载中...</td>
            </tr>
          ) : data?.items.length ? (
            data.items.map((person) => (
              <tr key={person.id}>
                <td>{person.id}</td>
                <td>{person.name}</td>
                <td>{person.email}</td>
                <td>{person.role_type}</td>
                <td>
                  <StatusTag value={person.status} />
                </td>
                <td>
                  <button type="button" className="btn btn-text" onClick={() => openEdit(person)}>
                    编辑
                  </button>
                  <button type="button" className="btn btn-text danger" onClick={() => void onDisable(person)} disabled={person.status === 0}>
                    停用
                  </button>
                </td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan={6}>暂无人员</td>
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

      <PersonFormModal
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
              email: values.email,
              role_type: values.role_type
            });
          }
          setOpenForm(false);
          await refetch();
        }}
      />
    </section>
  );
}

export default PersonListPage;
