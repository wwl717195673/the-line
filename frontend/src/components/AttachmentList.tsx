import type { Attachment } from "../types/api";

type AttachmentListProps = {
  attachments: Attachment[];
};

function formatSize(size: number): string {
  if (!size) {
    return "-";
  }
  if (size < 1024) {
    return `${size} B`;
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  }
  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function AttachmentList({ attachments }: AttachmentListProps) {
  if (!attachments.length) {
    return <p className="muted">暂无附件</p>;
  }

  return (
    <ul className="plain-list">
      {attachments.map((item) => (
        <li key={item.id}>
          <div className="comment-row">
            <a href={item.file_url} target="_blank" rel="noreferrer">
              {item.file_name}
            </a>
            <span className="muted">{new Date(item.created_at).toLocaleString()}</span>
          </div>
          <div className="comment-row muted">
            <span>类型：{item.file_type || "-"}</span>
            <span>大小：{formatSize(item.file_size)}</span>
            <span>上传人：#{item.uploaded_by}</span>
          </div>
        </li>
      ))}
    </ul>
  );
}

export default AttachmentList;
