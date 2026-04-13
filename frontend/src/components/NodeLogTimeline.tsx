import type { RunNodeLog } from "../types/api";

type NodeLogTimelineProps = {
  logs: RunNodeLog[];
};

function operatorText(operatorType: string): string {
  if (operatorType === "person") {
    return "人工";
  }
  if (operatorType === "agent") {
    return "龙虾";
  }
  if (operatorType === "system") {
    return "系统";
  }
  return operatorType;
}

function NodeLogTimeline({ logs }: NodeLogTimelineProps) {
  if (!logs.length) {
    return <p className="muted">暂无日志</p>;
  }

  return (
    <ul className="plain-list">
      {logs.map((item) => (
        <li key={item.id} className={item.log_type === "error" ? "log-error" : ""}>
          <div className="comment-row">
            <span>
              <span className="pill">{item.log_type}</span> <span className="pill">{operatorText(item.operator_type)}</span> {item.content}
            </span>
            <span className="muted">{new Date(item.created_at).toLocaleString()}</span>
          </div>
        </li>
      ))}
    </ul>
  );
}

export default NodeLogTimeline;
