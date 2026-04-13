import { useCallback, useEffect, useState } from "react";
import {
  approveRunNode,
  confirmRunNodeAgentResult,
  completeRunNode,
  failRunNode,
  getRunNodeDetail,
  listRunNodeLogs,
  requestRunNodeMaterial,
  runNodeAgent,
  saveRunNodeInput,
  submitRunNode,
  takeoverRunNode
} from "../api/runNodes";
import { rejectRunNode } from "../api/runNodes";
import type { ConfirmAgentResultInput, RunNodeDetail, RunNodeLog, TakeoverRunNodeInput } from "../types/api";

type UseRunNodeDetailResult = {
  data: RunNodeDetail | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useRunNodeDetail(nodeID?: number): UseRunNodeDetailResult {
  const [data, setData] = useState<RunNodeDetail | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  const fetchData = useCallback(async () => {
    if (!nodeID) {
      setData(null);
      setError("节点 ID 不合法");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const result = await getRunNodeDetail(nodeID);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载节点详情失败");
    } finally {
      setLoading(false);
    }
  }, [nodeID]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

export function useRunNodeActions() {
  const [loading, setLoading] = useState(false);

  const run = useCallback(async <T,>(fn: () => Promise<T>) => {
    setLoading(true);
    try {
      return await fn();
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    loading,
    saveInput: (nodeID: number, inputJSON: unknown) => run(() => saveRunNodeInput(nodeID, inputJSON)),
    submit: (nodeID: number, comment?: string) => run(() => submitRunNode(nodeID, comment)),
    approve: (nodeID: number, payload: { review_comment?: string; final_plan?: string; output_json?: unknown }) =>
      run(() => approveRunNode(nodeID, payload)),
    reject: (nodeID: number, reason: string) => run(() => rejectRunNode(nodeID, reason)),
    requestMaterial: (nodeID: number, requirement: string) => run(() => requestRunNodeMaterial(nodeID, requirement)),
    complete: (nodeID: number, payload: { comment?: string; output_json?: unknown }) => run(() => completeRunNode(nodeID, payload)),
    fail: (nodeID: number, reason: string) => run(() => failRunNode(nodeID, reason)),
    runAgent: (nodeID: number) => run(() => runNodeAgent(nodeID)),
    confirmAgentResult: (nodeID: number, input: ConfirmAgentResultInput) => run(() => confirmRunNodeAgentResult(nodeID, input)),
    takeover: (nodeID: number, input: TakeoverRunNodeInput) => run(() => takeoverRunNode(nodeID, input))
  };
}

type UseRunNodeLogsResult = {
  data: RunNodeLog[];
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useRunNodeLogs(nodeID?: number): UseRunNodeLogsResult {
  const [data, setData] = useState<RunNodeLog[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    if (!nodeID) {
      setData([]);
      setError("");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const logs = await listRunNodeLogs(nodeID);
      setData(logs);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载节点日志失败");
    } finally {
      setLoading(false);
    }
  }, [nodeID]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}
