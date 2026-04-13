import { requestJSON } from "../lib/http";
import type {
  CreateDeliverableInput,
  Deliverable,
  DeliverableDetail,
  DeliverableQuery,
  PageData,
  ReviewDeliverableInput
} from "../types/api";

export function listDeliverables(query: DeliverableQuery): Promise<PageData<Deliverable>> {
  return requestJSON<PageData<Deliverable>>("/api/deliverables", undefined, query);
}

export function getDeliverableDetail(id: number): Promise<DeliverableDetail> {
  return requestJSON<DeliverableDetail>(`/api/deliverables/${id}`);
}

export function createDeliverable(input: CreateDeliverableInput): Promise<DeliverableDetail> {
  return requestJSON<DeliverableDetail>("/api/deliverables", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function reviewDeliverable(id: number, input: ReviewDeliverableInput): Promise<DeliverableDetail> {
  return requestJSON<DeliverableDetail>(`/api/deliverables/${id}/review`, {
    method: "POST",
    body: JSON.stringify(input)
  });
}
