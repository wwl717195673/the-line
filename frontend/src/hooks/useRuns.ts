import { useCallback, useEffect, useState } from "react";
import { cancelRun, createRun, getRunDetail, listRuns } from "../api/runs";
import type { CreateRunInput, PageData, RunDetail, RunListItem, RunQuery } from "../types/api";

export function useStartRun() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (input: CreateRunInput) => {
    setLoading(true);
    try {
      return await createRun(input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}

type UseRunDetailResult = {
  data: RunDetail | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useRunDetail(runID?: number): UseRunDetailResult {
  const [data, setData] = useState<RunDetail | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  const fetchData = useCallback(async () => {
    if (!runID) {
      setData(null);
      setError("流程 ID 不合法");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const result = await getRunDetail(runID);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载流程失败");
    } finally {
      setLoading(false);
    }
  }, [runID]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

type UseRunsResult = {
  data: PageData<RunListItem> | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useRuns(query: RunQuery, options?: { enabled?: boolean }): UseRunsResult {
  const [data, setData] = useState<PageData<RunListItem> | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");
  const enabled = options?.enabled ?? true;
  const queryKey = JSON.stringify(query);

  const fetchData = useCallback(async () => {
    if (!enabled) {
      setLoading(false);
      setError("");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const result = await listRuns(query);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载流程列表失败");
    } finally {
      setLoading(false);
    }
  }, [enabled, queryKey]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

export function useCancelRun() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (runID: number, reason: string) => {
    setLoading(true);
    try {
      return await cancelRun(runID, reason);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}
