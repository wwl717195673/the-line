import { useState } from "react";

type CommentEditorProps = {
  submitting: boolean;
  placeholder?: string;
  onSubmit: (content: string) => Promise<void>;
};

function CommentEditor({ submitting, placeholder = "输入评论内容", onSubmit }: CommentEditorProps) {
  const [content, setContent] = useState("");
  const [error, setError] = useState("");

  return (
    <div className="inline-form">
      <input value={content} onChange={(event) => setContent(event.target.value)} placeholder={placeholder} />
      <button
        type="button"
        className="btn"
        disabled={submitting}
        onClick={() =>
          void (async () => {
            if (!content.trim()) {
              setError("评论内容不能为空");
              return;
            }
            setError("");
            await onSubmit(content.trim());
            setContent("");
          })().catch((err: unknown) => {
            setError(err instanceof Error ? err.message : "评论发布失败");
          })
        }
      >
        发布评论
      </button>
      {error ? <p className="error-text full-width">{error}</p> : null}
    </div>
  );
}

export default CommentEditor;
