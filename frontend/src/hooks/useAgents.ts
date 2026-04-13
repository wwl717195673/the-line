import { useCallback, useEffect, useState } from "react";
import { createAgent, disableAgent, listAgents, updateAgent } from "../api/agents";
import type { Agent, AgentQuery, CreateAgentInput, PageData, UpdateAgentInput } from "../types/api";

type UseAgentsResult = {
  data: PageData<Agent> | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useAgents(query: AgentQuery): UseAgentsResult {
  const [data, setData] = useState<PageData<Agent> | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await listAgents(query);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载龙虾失败");
    } finally {
      setLoading(false);
    }
  }, [query]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

export function useCreateAgent() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (input: CreateAgentInput) => {
    setLoading(true);
    try {
      return await createAgent(input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}

export function useUpdateAgent() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (id: number, input: UpdateAgentInput) => {
    setLoading(true);
    try {
      return await updateAgent(id, input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}

export function useDisableAgent() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (id: number) => {
    setLoading(true);
    try {
      return await disableAgent(id);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}
