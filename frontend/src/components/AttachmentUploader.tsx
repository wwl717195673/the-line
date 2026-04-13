import { useState } from "react";

type AttachmentUploaderProps = {
  disabled?: boolean;
  loading: boolean;
  onUploadByURL: (payload: { file_name: string; file_url: string; file_size: number; file_type: string }) => Promise<void>;
  onUploadFile: (file: File) => Promise<void>;
};

function AttachmentUploader({ disabled, loading, onUploadByURL, onUploadFile }: AttachmentUploaderProps) {
  const [name, setName] = useState("");
  const [url, setURL] = useState("");
  const [error, setError] = useState("");

  return (
    <div className="attachment-uploader">
      <div className="inline-form">
        <input value={name} onChange={(event) => setName(event.target.value)} placeholder="文件名" disabled={disabled || loading} />
        <input value={url} onChange={(event) => setURL(event.target.value)} placeholder="文件 URL（MVP）" disabled={disabled || loading} />
        <button
          type="button"
          className="btn"
          disabled={disabled || loading}
          onClick={() =>
            void (async () => {
              if (!name.trim() || !url.trim()) {
                setError("附件文件名和 URL 不能为空");
                return;
              }
              setError("");
              await onUploadByURL({
                file_name: name.trim(),
                file_url: url.trim(),
                file_size: 0,
                file_type: "url"
              });
              setName("");
              setURL("");
            })().catch((err: unknown) => {
              setError(err instanceof Error ? err.message : "附件上传失败");
            })
          }
        >
          添加 URL 附件
        </button>
      </div>
      <div className="inline-form">
        <input
          type="file"
          disabled={disabled || loading}
          onChange={(event) => {
            const file = event.target.files?.[0];
            if (!file) {
              return;
            }
            void onUploadFile(file).catch((err: unknown) => {
              setError(err instanceof Error ? err.message : "文件上传失败");
            });
            event.currentTarget.value = "";
          }}
        />
      </div>
      {error ? <p className="error-text">{error}</p> : null}
    </div>
  );
}

export default AttachmentUploader;
