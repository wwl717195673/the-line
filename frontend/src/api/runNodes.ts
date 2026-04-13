import { requestJSON } from "../lib/http";
import type { ConfirmAgentResultInput, RunNodeDetail, RunNodeLog, TakeoverRunNodeInput } from "../types/api";

export function getRunNodeDetail(id: number): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}`);
}

export function saveRunNodeInput(id: number, inputJSON: unknown): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}/input`, {
    method: "PUT",
    body: JSON.stringify({ input_json: inputJSON })
  });
}

export function submitRunNode(id: number, comment?: string): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}/submit`, {
    method: "POST",
    body: JSON.stringify({ comment: comment ?? "" })
  });
}

export function approveRunNode(id: number, payload: { review_comment?: string; final_plan?: string; output_json?: unknown }): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}/approve`, {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function rejectRunNode(id: number, reason: string): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}/reject`, {
    method: "POST",
    body: JSON.stringify({ reason })
  });
}

export function requestRunNodeMaterial(id: number, requirement: string): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}/request-material`, {
    method: "POST",
    body: JSON.stringify({ requirement })
  });
}

export function completeRunNode(id: number, payload: { comment?: string; output_json?: unknown }): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}/complete`, {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function failRunNode(id: number, reason: string): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}/fail`, {
    method: "POST",
    body: JSON.stringify({ reason })
  });
}

export function runNodeAgent(id: number): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}/run-agent`, {
    method: "POST",
    body: JSON.stringify({})
  });
}

export function confirmRunNodeAgentResult(id: number, input: ConfirmAgentResultInput): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}/confirm-agent-result`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function takeoverRunNode(id: number, input: TakeoverRunNodeInput): Promise<RunNodeDetail> {
  return requestJSON<RunNodeDetail>(`/api/run-nodes/${id}/takeover`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function listRunNodeLogs(id: number): Promise<RunNodeLog[]> {
  return requestJSON<RunNodeLog[]>(`/api/run-nodes/${id}/logs`);
}
