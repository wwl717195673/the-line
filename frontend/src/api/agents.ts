import { requestJSON } from "../lib/http";
import type { Agent, AgentQuery, CreateAgentInput, PageData, UpdateAgentInput } from "../types/api";

export function listAgents(query: AgentQuery): Promise<PageData<Agent>> {
  return requestJSON<PageData<Agent>>("/api/agents", undefined, query);
}

export function createAgent(input: CreateAgentInput): Promise<Agent> {
  return requestJSON<Agent>("/api/agents", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function updateAgent(id: number, input: UpdateAgentInput): Promise<Agent> {
  return requestJSON<Agent>(`/api/agents/${id}`, {
    method: "PUT",
    body: JSON.stringify(input)
  });
}

export function disableAgent(id: number): Promise<Agent> {
  return requestJSON<Agent>(`/api/agents/${id}/disable`, {
    method: "POST",
    body: JSON.stringify({})
  });
}
