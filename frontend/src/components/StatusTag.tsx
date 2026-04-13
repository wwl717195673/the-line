import type { Status } from "../types/api";

type StatusTagProps = {
  value: Status;
};

function StatusTag({ value }: StatusTagProps) {
  return <span className={`status-tag ${value === 1 ? "enabled" : "disabled"}`}>{value === 1 ? "启用" : "停用"}</span>;
}

export default StatusTag;
