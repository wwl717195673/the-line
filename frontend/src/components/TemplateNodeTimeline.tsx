import type { TemplateNode } from "../types/api";
import TemplateNodeCard from "./TemplateNodeCard";

type TemplateNodeTimelineProps = {
  nodes: TemplateNode[];
};

function TemplateNodeTimeline({ nodes }: TemplateNodeTimelineProps) {
  if (!nodes.length) {
    return <p className="muted">暂无节点配置</p>;
  }
  return (
    <div className="template-node-timeline">
      {nodes
        .slice()
        .sort((a, b) => a.sort_order - b.sort_order)
        .map((node) => (
          <TemplateNodeCard key={node.id} node={node} />
        ))}
    </div>
  );
}

export default TemplateNodeTimeline;
