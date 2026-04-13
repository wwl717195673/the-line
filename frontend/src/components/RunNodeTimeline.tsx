import type { RunNode } from "../types/api";

type RunNodeTimelineProps = {
  nodes: RunNode[];
  selectedNodeID?: number;
  onSelect: (node: RunNode) => void;
};

function nodeStatusClass(status: string): string {
  if (status === "done") {
    return "done";
  }
  if (status === "failed" || status === "rejected") {
    return "error";
  }
  if (status === "waiting_confirm" || status === "waiting_material") {
    return "warning";
  }
  if (status === "running" || status === "ready") {
    return "active";
  }
  return "idle";
}

function RunNodeTimeline({ nodes, selectedNodeID, onSelect }: RunNodeTimelineProps) {
  if (!nodes.length) {
    return <p className="muted">暂无节点</p>;
  }

  return (
    <div className="run-node-timeline">
      {nodes
        .slice()
        .sort((a, b) => a.sort_order - b.sort_order)
        .map((node) => (
          <button
            key={node.id}
            type="button"
            className={`run-node-item ${nodeStatusClass(node.status)} ${selectedNodeID === node.id ? "selected" : ""}`}
            onClick={() => onSelect(node)}
          >
            <span className="order">{node.sort_order}</span>
            <span className="name">{node.node_name}</span>
            <span className="status">{node.status}</span>
          </button>
        ))}
    </div>
  );
}

export default RunNodeTimeline;
