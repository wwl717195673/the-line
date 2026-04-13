type RunStatusTagProps = {
  value: string;
};

const STATUS_TEXT_MAP: Record<string, string> = {
  running: "进行中",
  waiting: "等待中",
  blocked: "阻塞",
  completed: "已完成",
  cancelled: "已取消"
};

function RunStatusTag({ value }: RunStatusTagProps) {
  const text = STATUS_TEXT_MAP[value] ?? value;
  return <span className={`run-status-tag ${value}`}>{text}</span>;
}

export default RunStatusTag;
