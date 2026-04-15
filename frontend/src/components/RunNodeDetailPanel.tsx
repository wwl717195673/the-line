import type { RunNode } from "../types/api";

type RunNodeDetailPanelProps = {
  node?: RunNode;
};

function stringifyJSON(value: unknown): string {
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return "{}";
  }
}

function RunNodeDetailPanel({ node }: RunNodeDetailPanelProps) {
  if (!node) {
    return <p className="muted">请选择节点查看详情</p>;
  }

  const inputRecord = typeof node.input_json === "object" && node.input_json && !Array.isArray(node.input_json) ? (node.input_json as Record<string, unknown>) : {};
  const outputRecord = typeof node.output_json === "object" && node.output_json && !Array.isArray(node.output_json) ? (node.output_json as Record<string, unknown>) : {};
  const matter = typeof inputRecord.matter === "string" ? inputRecord.matter : typeof outputRecord.summary === "string" ? outputRecord.summary : node.node_name;
  const lobsterStatus = node.status === "done" || node.status === "waiting_confirm" ? "已完成" : "未完成";
  const reviewStatus = node.status === "done" ? "已审核" : "未审核";

  return (
    <section className="run-node-detail-panel">
      <h3>节点详情</h3>
      <p>
        <b>哪个龙虾：</b>
        {node.bound_agent?.name ?? "未指定"}
      </p>
      <p>
        <b>哪个人：</b>
        {node.owner_person?.name ?? node.result_owner_person?.name ?? "-"}
      </p>
      <p>
        <b>事项：</b>
        {matter}
      </p>
      <p>
        <b>龙虾状态：</b>
        {lobsterStatus}
      </p>
      <p>
        <b>审核状态：</b>
        {reviewStatus}
      </p>
      <p>
        <b>节点编码：</b>
        {node.node_code}
      </p>
      <p>
        <b>开始时间：</b>
        {node.started_at ? new Date(node.started_at).toLocaleString() : "-"}
      </p>
      <p>
        <b>完成时间：</b>
        {node.completed_at ? new Date(node.completed_at).toLocaleString() : "-"}
      </p>
      <details>
        <summary>节点原始数据</summary>
        <pre>{stringifyJSON({ input_json: node.input_json, output_json: node.output_json })}</pre>
      </details>
    </section>
  );
}

export default RunNodeDetailPanel;
