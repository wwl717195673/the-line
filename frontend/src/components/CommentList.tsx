import type { Comment } from "../types/api";

type CommentListProps = {
  comments: Comment[];
  resolving: boolean;
  onResolve?: (commentID: number) => Promise<void>;
};

function CommentList({ comments, resolving, onResolve }: CommentListProps) {
  if (!comments.length) {
    return <p className="muted">暂无评论</p>;
  }

  return (
    <ul className="plain-list">
      {comments.map((item) => (
        <li key={item.id}>
          <div className="comment-row">
            <span>
              <b>{item.author?.name ?? `#${item.author_person_id}`}</b>：{item.content}
            </span>
            <span className="muted">{new Date(item.created_at).toLocaleString()}</span>
          </div>
          <div className="comment-row">
            <span className={item.is_resolved ? "status-tag enabled" : "status-tag disabled"}>{item.is_resolved ? "已解决" : "未解决"}</span>
            {!item.is_resolved && onResolve ? (
              <button type="button" className="btn btn-text" disabled={resolving} onClick={() => void onResolve(item.id)}>
                标记已解决
              </button>
            ) : null}
          </div>
        </li>
      ))}
    </ul>
  );
}

export default CommentList;
