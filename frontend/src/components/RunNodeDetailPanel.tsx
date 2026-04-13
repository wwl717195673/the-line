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

  return (
    <section className="run-node-detail-panel">
      <h3>节点详情</h3>
      <p>
        <b>名称：</b>
        {node.node_name}
      </p>
      <p>
        <b>编码：</b>
        {node.node_code}
      </p>
      <p>
        <b>类型：</b>
        {node.node_type}
      </p>
      <p>
        <b>状态：</b>
        {node.status}
      </p>
      <p>
        <b>责任人：</b>
        {node.owner_person?.name ?? "-"}
      </p>
      <p>
        <b>审核人：</b>
        {node.reviewer_person?.name ?? "-"}
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
        <summary>输入 JSON</summary>
        <pre>{stringifyJSON(node.input_json)}</pre>
      </details>
      <details>
        <summary>输出 JSON</summary>
        <pre>{stringifyJSON(node.output_json)}</pre>
      </details>
    </section>
  );
}

export default RunNodeDetailPanel;
