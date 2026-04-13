import { useEffect, useMemo, useState } from "react";
import Modal from "./Modal";
import type { Person, Status } from "../types/api";

type PersonFormModalProps = {
  open: boolean;
  initial?: Person;
  submitting: boolean;
  onCancel: () => void;
  onSubmit: (values: { name: string; email: string; role_type: string; status: Status }) => Promise<void>;
};

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

function PersonFormModal({ open, initial, submitting, onCancel, onSubmit }: PersonFormModalProps) {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [roleType, setRoleType] = useState("teacher_leader");
  const [status, setStatus] = useState<Status>(1);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }
    setName(initial?.name ?? "");
    setEmail(initial?.email ?? "");
    setRoleType(initial?.role_type ?? "teacher_leader");
    setStatus(initial?.status ?? 1);
    setError("");
  }, [initial, open]);

  const title = useMemo(() => (initial ? "编辑人员" : "新建人员"), [initial]);

  const validate = () => {
    if (!name.trim()) {
      return "姓名不能为空";
    }
    if (!email.trim()) {
      return "邮箱不能为空";
    }
    if (!EMAIL_PATTERN.test(email.trim())) {
      return "邮箱格式不合法";
    }
    if (!roleType.trim()) {
      return "角色不能为空";
    }
    return "";
  };

  return (
    <Modal open={open} title={title} onClose={onCancel}>
      <form
        className="form-grid"
        onSubmit={(event) => {
          event.preventDefault();
          const message = validate();
          if (message) {
            setError(message);
            return;
          }
          void onSubmit({
            name: name.trim(),
            email: email.trim(),
            role_type: roleType.trim(),
            status
          }).catch((err: unknown) => {
            setError(err instanceof Error ? err.message : "提交失败");
          });
        }}
      >
        <label>
          姓名
          <input value={name} onChange={(event) => setName(event.target.value)} placeholder="请输入姓名" />
        </label>
        <label>
          邮箱
          <input value={email} onChange={(event) => setEmail(event.target.value)} placeholder="请输入邮箱" />
        </label>
        <label>
          角色
          <input value={roleType} onChange={(event) => setRoleType(event.target.value)} placeholder="例如 middle_office" />
        </label>
        <label>
          状态
          <select value={status} onChange={(event) => setStatus(Number(event.target.value) as Status)} disabled={!initial}>
            <option value={1}>启用</option>
            <option value={0}>停用</option>
          </select>
        </label>
        {!initial ? <p className="muted">新建时状态默认为启用。</p> : null}
        {error ? <p className="error-text">{error}</p> : null}
        <div className="modal-actions">
          <button type="button" className="btn" onClick={onCancel} disabled={submitting}>
            取消
          </button>
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? "提交中..." : "提交"}
          </button>
        </div>
      </form>
    </Modal>
  );
}

export default PersonFormModal;
