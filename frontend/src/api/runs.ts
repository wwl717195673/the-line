import { requestJSON } from "../lib/http";
import type { CreateRunInput, PageData, RunDetail, RunListItem, RunQuery } from "../types/api";

export function createRun(input: CreateRunInput): Promise<RunDetail> {
  return requestJSON<RunDetail>("/api/runs", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function getRunDetail(id: number): Promise<RunDetail> {
  return requestJSON<RunDetail>(`/api/runs/${id}`);
}

export function listRuns(query: RunQuery): Promise<PageData<RunListItem>> {
  return requestJSON<PageData<RunListItem>>("/api/runs", undefined, query);
}

export function cancelRun(id: number, reason: string): Promise<RunDetail> {
  return requestJSON<RunDetail>(`/api/runs/${id}/cancel`, {
    method: "POST",
    body: JSON.stringify({ reason })
  });
}
