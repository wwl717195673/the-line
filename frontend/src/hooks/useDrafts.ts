import { useCallback, useEffect, useState } from "react";
import { confirmDraft, createDraft, deleteDraft, discardDraft, getDraftDetail, listDrafts, updateDraft } from "../api/drafts";
import type { ConfirmDraftResponse, CreateDraftInput, FlowDraft, FlowDraftQuery, PageData, UpdateDraftInput } from "../types/api";

type UseDraftsResult = {
  data: PageData<FlowDraft> | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useDrafts(query: FlowDraftQuery, enabled = true): UseDraftsResult {
  const [data, setData] = useState<PageData<FlowDraft> | null>(null);
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
      const result = await listDrafts(query);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载草案失败");
    } finally {
      setLoading(false);
    }
  }, [enabled, query]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

type UseDraftDetailResult = {
  data: FlowDraft | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useDraftDetail(draftID?: number): UseDraftDetailResult {
  const [data, setData] = useState<FlowDraft | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    if (!draftID) {
      setData(null);
      setError("草案 ID 不合法");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const result = await getDraftDetail(draftID);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载草案详情失败");
    } finally {
      setLoading(false);
    }
  }, [draftID]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

function useMutation<Args extends unknown[], Result>(fn: (...args: Args) => Promise<Result>) {
  const [loading, setLoading] = useState(false);
  const run = useCallback(
    async (...args: Args) => {
      setLoading(true);
      try {
        return await fn(...args);
      } finally {
        setLoading(false);
      }
    },
    [fn]
  );
  return { run, loading };
}

export function useCreateDraft() {
  return useMutation<[CreateDraftInput], FlowDraft>((input) => createDraft(input));
}

export function useUpdateDraft() {
  return useMutation<[number, UpdateDraftInput], FlowDraft>((id, input) => updateDraft(id, input));
}

export function useConfirmDraft() {
  return useMutation<[number, number], ConfirmDraftResponse>((id, confirmedBy) => confirmDraft(id, confirmedBy));
}

export function useDiscardDraft() {
  return useMutation<[number, number, string?], FlowDraft>((id, discardedBy, reason) => discardDraft(id, discardedBy, reason));
}

export function useDeleteDraft() {
  return useMutation<[number], { success: boolean }>((id) => deleteDraft(id));
}
