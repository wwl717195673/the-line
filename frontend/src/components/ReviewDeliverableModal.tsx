import { useEffect, useState } from "react";
import Modal from "./Modal";

type ReviewDeliverableModalProps = {
  open: boolean;
  submitting: boolean;
  reviewStatus: "approved" | "rejected";
  onCancel: () => void;
  onSubmit: (payload: { review_status: "approved" | "rejected"; review_comment: string }) => Promise<void>;
};

function ReviewDeliverableModal({ open, submitting, reviewStatus, onCancel, onSubmit }: ReviewDeliverableModalProps) {
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }
    setComment("");
    setError("");
  }, [open, reviewStatus]);

  return (
    <Modal open={open} title={reviewStatus === "approved" ? "验收通过" : "验收驳回"} onClose={onCancel}>
      <form
        className="form-grid"
        onSubmit={(event) => {
          event.preventDefault();
          setError("");
          void onSubmit({ review_status: reviewStatus, review_comment: comment.trim() }).catch((err: unknown) => {
            setError(err instanceof Error ? err.message : "验收操作失败");
          });
        }}
      >
        <label className="full-width">
          验收意见
          <textarea rows={5} value={comment} onChange={(event) => setComment(event.target.value)} placeholder="可选" />
        </label>
        {error ? <p className="error-text">{error}</p> : null}
        <div className="modal-actions">
          <button type="button" className="btn" disabled={submitting} onClick={onCancel}>
            取消
          </button>
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? "提交中..." : reviewStatus === "approved" ? "确认通过" : "确认驳回"}
          </button>
        </div>
      </form>
    </Modal>
  );
}

export default ReviewDeliverableModal;
