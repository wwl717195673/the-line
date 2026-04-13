import { useCallback, useEffect, useMemo, useState } from "react";
import { getLatestAgentTaskReceipt, listAgentTasks, type AgentTaskQuery } from "../api/agentTasks";
import type { AgentTask, AgentTaskReceipt, PageData } from "../types/api";

type UseAgentTasksResult = {
  data: PageData<AgentTask> | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useAgentTasks(query: AgentTaskQuery, enabled = true): UseAgentTasksResult {
  const [data, setData] = useState<PageData<AgentTask> | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    if (!enabled) {
      setData(null);
      setError("");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const result = await listAgentTasks(query);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载龙虾任务失败");
    } finally {
      setLoading(false);
    }
  }, [enabled, query]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

type UseLatestAgentReceiptResult = {
  data: AgentTaskReceipt | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useLatestAgentReceipt(taskID?: number, enabled = true): UseLatestAgentReceiptResult {
  const [data, setData] = useState<AgentTaskReceipt | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    if (!enabled || !taskID) {
      setData(null);
      setError("");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const result = await getLatestAgentTaskReceipt(taskID);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载回执失败");
    } finally {
      setLoading(false);
    }
  }, [enabled, taskID]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

type UseAgentTaskSnapshotResult = {
  task: AgentTask | null;
  receipt: AgentTaskReceipt | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useAgentTaskSnapshot(runNodeID?: number, enabled = true): UseAgentTaskSnapshotResult {
  const query = useMemo<AgentTaskQuery>(
    () => ({
      page: 1,
      page_size: 5,
      run_node_id: runNodeID
    }),
    [runNodeID]
  );
  const tasksQuery = useAgentTasks(query, enabled && !!runNodeID);
  const task = tasksQuery.data?.items[0] ?? null;
  const receiptQuery = useLatestAgentReceipt(task?.id, enabled && !!task?.id);

  return {
    task,
    receipt: receiptQuery.data,
    loading: tasksQuery.loading || receiptQuery.loading,
    error: tasksQuery.error || receiptQuery.error,
    refetch: async () => {
      await tasksQuery.refetch();
      if (task?.id) {
        await receiptQuery.refetch();
      }
    }
  };
}
