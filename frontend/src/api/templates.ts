import { requestJSON } from "../lib/http";
import type { PageData, Template, TemplateDetail, TemplateQuery } from "../types/api";

export function listTemplates(query: TemplateQuery): Promise<PageData<Template>> {
  return requestJSON<PageData<Template>>("/api/templates", undefined, query);
}

export function getTemplateDetail(id: number): Promise<TemplateDetail> {
  return requestJSON<TemplateDetail>(`/api/templates/${id}`);
}

export function deleteTemplate(id: number): Promise<{ success: boolean }> {
  return requestJSON<{ success: boolean }>(`/api/templates/${id}`, {
    method: "DELETE"
  });
}
