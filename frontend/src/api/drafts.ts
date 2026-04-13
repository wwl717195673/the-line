import { requestJSON } from "../lib/http";
import type { ConfirmDraftResponse, CreateDraftInput, FlowDraft, FlowDraftQuery, PageData, UpdateDraftInput } from "../types/api";

export function listDrafts(query: FlowDraftQuery): Promise<PageData<FlowDraft>> {
  return requestJSON<PageData<FlowDraft>>("/api/flow-drafts", undefined, query);
}

export function getDraftDetail(id: number): Promise<FlowDraft> {
  return requestJSON<FlowDraft>(`/api/flow-drafts/${id}`);
}

export function createDraft(input: CreateDraftInput): Promise<FlowDraft> {
  return requestJSON<FlowDraft>("/api/flow-drafts", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function updateDraft(id: number, input: UpdateDraftInput): Promise<FlowDraft> {
  return requestJSON<FlowDraft>(`/api/flow-drafts/${id}`, {
    method: "PUT",
    body: JSON.stringify(input)
  });
}

export function confirmDraft(id: number, confirmedBy: number): Promise<ConfirmDraftResponse> {
  return requestJSON<ConfirmDraftResponse>(`/api/flow-drafts/${id}/confirm`, {
    method: "POST",
    body: JSON.stringify({ confirmed_by: confirmedBy })
  });
}

export function discardDraft(id: number, discardedBy: number, reason?: string): Promise<FlowDraft> {
  return requestJSON<FlowDraft>(`/api/flow-drafts/${id}/discard`, {
    method: "POST",
    body: JSON.stringify({
      discarded_by: discardedBy,
      reason: reason ?? ""
    })
  });
}

export function deleteDraft(id: number): Promise<{ success: boolean }> {
  return requestJSON<{ success: boolean }>(`/api/flow-drafts/${id}`, {
    method: "DELETE"
  });
}
