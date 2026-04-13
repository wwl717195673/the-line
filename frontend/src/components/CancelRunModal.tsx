import { useEffect, useState } from "react";
import Modal from "./Modal";

type CancelRunModalProps = {
  open: boolean;
  submitting: boolean;
  onCancel: () => void;
  onSubmit: (reason: string) => Promise<void>;
};

function CancelRunModal({ open, submitting, onCancel, onSubmit }: CancelRunModalProps) {
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }
    setReason("");
    setError("");
  }, [open]);

  return (
    <Modal open={open} title="取消流程" onClose={onCancel}>
      <form
        className="form-grid"
        onSubmit={(event) => {
          event.preventDefault();
          if (!reason.trim()) {
            setError("取消原因不能为空");
            return;
          }
          setError("");
          void onSubmit(reason.trim()).catch((err: unknown) => {
            setError(err instanceof Error ? err.message : "取消流程失败");
          });
        }}
      >
        <label className="full-width">
          取消原因
          <textarea rows={5} value={reason} onChange={(event) => setReason(event.target.value)} placeholder="请输入取消原因" />
        </label>
        {error ? <p className="error-text">{error}</p> : null}
        <div className="modal-actions">
          <button type="button" className="btn" onClick={onCancel} disabled={submitting}>
            关闭
          </button>
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? "提交中..." : "确认取消"}
          </button>
        </div>
      </form>
    </Modal>
  );
}

export default CancelRunModal;
