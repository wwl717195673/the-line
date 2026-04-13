import { useEffect, useMemo, useState } from "react";
import Modal from "./Modal";
import PersonSelect from "./PersonSelect";
import type { Agent, Status } from "../types/api";

type AgentFormModalProps = {
  open: boolean;
  initial?: Agent;
  submitting: boolean;
  onCancel: () => void;
  onSubmit: (values: {
    name: string;
    code: string;
    provider: string;
    version: string;
    owner_person_id: number;
    config_json: unknown;
    status: Status;
  }) => Promise<void>;
};

function AgentFormModal({ open, initial, submitting, onCancel, onSubmit }: AgentFormModalProps) {
  const [name, setName] = useState("");
  const [code, setCode] = useState("");
  const [provider, setProvider] = useState("openclaw");
  const [version, setVersion] = useState("v1");
  const [ownerPersonID, setOwnerPersonID] = useState<number | undefined>(undefined);
  const [configText, setConfigText] = useState("{}");
  const [status, setStatus] = useState<Status>(1);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }
    setName(initial?.name ?? "");
    setCode(initial?.code ?? "");
    setProvider(initial?.provider ?? "openclaw");
    setVersion(initial?.version ?? "v1");
    setOwnerPersonID(initial?.owner_person_id);
    setConfigText(initial?.config_json ? JSON.stringify(initial.config_json, null, 2) : "{}");
    setStatus(initial?.status ?? 1);
    setError("");
  }, [initial, open]);

  const title = useMemo(() => (initial ? "编辑龙虾" : "新建龙虾"), [initial]);

  const validate = (): string => {
    if (!name.trim()) {
      return "名称不能为空";
    }
    if (!code.trim()) {
      return "编码不能为空";
    }
    if (!provider.trim()) {
      return "来源不能为空";
    }
    if (!version.trim()) {
      return "版本不能为空";
    }
    if (!ownerPersonID) {
      return "维护人不能为空";
    }
    if (!configText.trim()) {
      return "配置快照不能为空";
    }
    try {
      JSON.parse(configText);
    } catch {
      return "配置快照必须是合法 JSON";
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
          const parsedConfig = JSON.parse(configText);
          void onSubmit({
            name: name.trim(),
            code: code.trim(),
            provider: provider.trim(),
            version: version.trim(),
            owner_person_id: ownerPersonID!,
            config_json: parsedConfig,
            status
          }).catch((err: unknown) => {
            setError(err instanceof Error ? err.message : "提交失败");
          });
        }}
      >
        <label>
          名称
          <input value={name} onChange={(event) => setName(event.target.value)} placeholder="请输入龙虾名称" />
        </label>
        <label>
          编码
          <input value={code} onChange={(event) => setCode(event.target.value)} placeholder="请输入唯一编码" />
        </label>
        <label>
          来源
          <input value={provider} onChange={(event) => setProvider(event.target.value)} placeholder="openclaw" />
        </label>
        <label>
          版本
          <input value={version} onChange={(event) => setVersion(event.target.value)} placeholder="v1" />
        </label>
        <label>
          维护人
          <PersonSelect value={ownerPersonID} onChange={setOwnerPersonID} placeholder="请选择维护人" />
        </label>
        <label>
          状态
          <select value={status} onChange={(event) => setStatus(Number(event.target.value) as Status)} disabled={!initial}>
            <option value={1}>启用</option>
            <option value={0}>停用</option>
          </select>
        </label>
        <label className="full-width">
          配置快照 JSON
          <textarea rows={8} value={configText} onChange={(event) => setConfigText(event.target.value)} />
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

export default AgentFormModal;
