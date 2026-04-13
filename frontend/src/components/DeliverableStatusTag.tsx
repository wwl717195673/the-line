type DeliverableStatusTagProps = {
  value: "pending" | "approved" | "rejected" | string;
};

const STATUS_TEXT: Record<string, string> = {
  pending: "待验收",
  approved: "已通过",
  rejected: "已驳回"
};

function DeliverableStatusTag({ value }: DeliverableStatusTagProps) {
  return <span className={`deliverable-status-tag ${value}`}>{STATUS_TEXT[value] ?? value}</span>;
}

export default DeliverableStatusTag;
