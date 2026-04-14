import { useMemo } from "react";
import { usePersons } from "../hooks/usePersons";
import type { Status } from "../types/api";

type PersonSelectProps = {
  value?: number;
  onChange: (value: number | undefined) => void;
  placeholder?: string;
  status?: Status;
  disabled?: boolean;
};

function PersonSelect({ value, onChange, placeholder = "请选择人员", status = 1, disabled = false }: PersonSelectProps) {
  const query = useMemo(
    () => ({
      page: 1,
      page_size: 200,
      status
    }),
    [status]
  );
  const { data, loading } = usePersons(query);

  return (
    <select
      value={value ?? ""}
      disabled={disabled}
      onChange={(event) => {
        const next = event.target.value;
        onChange(next ? Number(next) : undefined);
      }}
    >
      <option value="">{loading ? "加载中..." : placeholder}</option>
      {data?.items.map((person) => (
        <option key={person.id} value={person.id}>
          {person.name}（{person.role_type}）
        </option>
      ))}
    </select>
  );
}

export default PersonSelect;
