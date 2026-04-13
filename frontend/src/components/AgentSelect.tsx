import { useMemo } from "react";
import { useAgents } from "../hooks/useAgents";
import type { Status } from "../types/api";

type AgentSelectProps = {
  value?: number;
  onChange: (value: number | undefined) => void;
  placeholder?: string;
  status?: Status;
};

function AgentSelect({ value, onChange, placeholder = "请选择龙虾", status = 1 }: AgentSelectProps) {
  const query = useMemo(
    () => ({
      page: 1,
      page_size: 200,
      status
    }),
    [status]
  );
  const { data, loading } = useAgents(query);

  return (
    <select
      value={value ?? ""}
      onChange={(event) => {
        const next = event.target.value;
        onChange(next ? Number(next) : undefined);
      }}
    >
      <option value="">{loading ? "加载中..." : placeholder}</option>
      {data?.items.map((agent) => (
        <option key={agent.id} value={agent.id}>
          {agent.name}（{agent.code}）
        </option>
      ))}
    </select>
  );
}

export default AgentSelect;
