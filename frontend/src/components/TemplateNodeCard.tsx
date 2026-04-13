import type { TemplateNode } from "../types/api";

type TemplateNodeCardProps = {
  node: TemplateNode;
};

function stringifySchema(value: unknown): string {
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return "{}";
  }
}

function TemplateNodeCard({ node }: TemplateNodeCardProps) {
  return (
    <article className="template-node-card">
      <div className="template-node-title">
        <strong>
          {node.sort_order}. {node.node_name}
        </strong>
        <span className="pill">{node.node_type}</span>
      </div>
      <p>
        <b>节点编码：</b>
        {node.node_code}
      </p>
      <p>
        <b>默认责任人规则：</b>
        {node.default_owner_rule || "-"}
      </p>
      <p>
        <b>默认龙虾：</b>
        {node.default_agent ? `${node.default_agent.name} (${node.default_agent.code})` : "-"}
      </p>
      <details>
        <summary>输入结构摘要</summary>
        <pre>{stringifySchema(node.input_schema_json)}</pre>
      </details>
      <details>
        <summary>输出结构摘要</summary>
        <pre>{stringifySchema(node.output_schema_json)}</pre>
      </details>
    </article>
  );
}

export default TemplateNodeCard;
