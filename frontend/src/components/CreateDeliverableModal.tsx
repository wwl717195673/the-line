import { useEffect, useMemo, useState } from "react";
import { listAttachments } from "../api/collaboration";
import type { Attachment, RunNode } from "../types/api";
import Modal from "./Modal";
import PersonSelect from "./PersonSelect";

type CreateDeliverableModalProps = {
  open: boolean;
  runID: number;
  runTitle: string;
  runNodes: RunNode[];
  submitting: boolean;
  onCancel: () => void;
  onSubmit: (payload: {
    title: string;
    summary: string;
    reviewer_person_id: number;
    result_json: unknown;
    attachment_ids: number[];
  }) => Promise<void>;
};

type AttachmentOption = Attachment & {
  source: string;
};

function CreateDeliverableModal({ open, runID, runTitle, runNodes, submitting, onCancel, onSubmit }: CreateDeliverableModalProps) {
  const [title, setTitle] = useState("");
  const [summary, setSummary] = useState("");
  const [conclusion, setConclusion] = useState("");
  const [abnormalNote, setAbnormalNote] = useState("");
  const [reviewerPersonID, setReviewerPersonID] = useState<number | undefined>(undefined);
  const [attachmentLoading, setAttachmentLoading] = useState(false);
  const [attachmentError, setAttachmentError] = useState("");
  const [attachmentOptions, setAttachmentOptions] = useState<AttachmentOption[]>([]);
  const [selectedAttachmentIDs, setSelectedAttachmentIDs] = useState<number[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }
    setTitle(`${runTitle} 交付结果`);
    setSummary("");
    setConclusion("");
    setAbnormalNote("");
    setReviewerPersonID(undefined);
    setSelectedAttachmentIDs([]);
    setError("");
    setAttachmentError("");
    setAttachmentLoading(true);

    void (async () => {
      try {
        const flowAttachments = await listAttachments("flow_run", runID);
        const nodeAttachments = await Promise.all(
          runNodes.map(async (node) => {
            const attachments = await listAttachments("flow_run_node", node.id);
            return attachments.map((item) => ({
              ...item,
              source: node.node_name
            }));
          })
        );
        const flowItems = flowAttachments.map((item) => ({ ...item, source: "流程附件" }));
        setAttachmentOptions([...flowItems, ...nodeAttachments.flat()]);
      } catch (err) {
        setAttachmentError(err instanceof Error ? err.message : "加载可选附件失败");
      } finally {
        setAttachmentLoading(false);
      }
    })();
  }, [open, runID, runNodes, runTitle]);

  const optionMap = useMemo(() => {
    const map = new Map<number, AttachmentOption>();
    attachmentOptions.forEach((item) => map.set(item.id, item));
    return map;
  }, [attachmentOptions]);

  return (
    <Modal open={open} title="生成交付物" onClose={onCancel}>
      <form
        className="form-grid"
        onSubmit={(event) => {
          event.preventDefault();
          if (!title.trim()) {
            setError("交付标题不能为空");
            return;
          }
          if (!summary.trim()) {
            setError("交付摘要不能为空");
            return;
          }
          if (!reviewerPersonID) {
            setError("验收人不能为空");
            return;
          }
          setError("");
          void onSubmit({
            title: title.trim(),
            summary: summary.trim(),
            reviewer_person_id: reviewerPersonID,
            result_json: {
              conclusion: conclusion.trim(),
              abnormal_note: abnormalNote.trim()
            },
            attachment_ids: selectedAttachmentIDs
          }).catch((err: unknown) => {
            setError(err instanceof Error ? err.message : "生成交付物失败");
          });
        }}
      >
        <label className="full-width">
          交付标题
          <input value={title} onChange={(event) => setTitle(event.target.value)} placeholder="请输入交付标题" />
        </label>
        <label className="full-width">
          交付摘要
          <textarea rows={3} value={summary} onChange={(event) => setSummary(event.target.value)} placeholder="请输入交付摘要" />
        </label>
        <label className="full-width">
          关键结论
          <textarea rows={3} value={conclusion} onChange={(event) => setConclusion(event.target.value)} placeholder="可选" />
        </label>
        <label className="full-width">
          异常说明
          <textarea rows={3} value={abnormalNote} onChange={(event) => setAbnormalNote(event.target.value)} placeholder="可选" />
        </label>
        <label>
          验收人
          <PersonSelect value={reviewerPersonID} onChange={setReviewerPersonID} placeholder="请选择验收人" />
        </label>
        <div className="full-width">
          <label>关键附件</label>
          {attachmentLoading ? <p className="muted">加载附件中...</p> : null}
          {attachmentError ? <p className="error-text">{attachmentError}</p> : null}
          {!attachmentLoading && !attachmentOptions.length ? <p className="muted">暂无可选附件</p> : null}
          <div className="check-list">
            {attachmentOptions.map((item) => (
              <label key={item.id} className="check-item">
                <input
                  type="checkbox"
                  checked={selectedAttachmentIDs.includes(item.id)}
                  onChange={(event) => {
                    const checked = event.target.checked;
                    setSelectedAttachmentIDs((ids) => {
                      if (checked) {
                        return ids.includes(item.id) ? ids : [...ids, item.id];
                      }
                      return ids.filter((id) => id !== item.id);
                    });
                  }}
                />
                <span>
                  {item.file_name}（{optionMap.get(item.id)?.source ?? "-"}）
                </span>
              </label>
            ))}
          </div>
        </div>
        {error ? <p className="error-text">{error}</p> : null}
        <div className="modal-actions">
          <button type="button" className="btn" disabled={submitting} onClick={onCancel}>
            取消
          </button>
          <button type="submit" className="btn btn-primary" disabled={submitting}>
            {submitting ? "提交中..." : "生成交付物"}
          </button>
        </div>
      </form>
    </Modal>
  );
}

export default CreateDeliverableModal;
