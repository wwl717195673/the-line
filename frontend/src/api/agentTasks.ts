import { requestJSON } from "../lib/http";
import type { AgentTask, AgentTaskReceipt, PageData } from "../types/api";

export type AgentTaskQuery = {
  page: number;
  page_size: number;
  run_id?: number;
  run_node_id?: number;
  status?: string;
};

export function listAgentTasks(query: AgentTaskQuery): Promise<PageData<AgentTask>> {
  return requestJSON<PageData<AgentTask>>("/api/agent-tasks", undefined, query);
}

export function getAgentTaskDetail(id: number): Promise<AgentTask> {
  return requestJSON<AgentTask>(`/api/agent-tasks/${id}`);
}

export function getLatestAgentTaskReceipt(id: number): Promise<AgentTaskReceipt> {
  return requestJSON<AgentTaskReceipt>(`/api/agent-tasks/${id}/receipt`);
}
