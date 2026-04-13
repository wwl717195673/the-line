import { useCallback, useEffect, useState } from "react";
import { createDeliverable, getDeliverableDetail, listDeliverables, reviewDeliverable } from "../api/deliverables";
import type {
  CreateDeliverableInput,
  Deliverable,
  DeliverableDetail,
  DeliverableQuery,
  PageData,
  ReviewDeliverableInput
} from "../types/api";

type UseDeliverablesResult = {
  data: PageData<Deliverable> | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useDeliverables(query: DeliverableQuery): UseDeliverablesResult {
  const [data, setData] = useState<PageData<Deliverable> | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await listDeliverables(query);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载交付列表失败");
    } finally {
      setLoading(false);
    }
  }, [query]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

type UseDeliverableDetailResult = {
  data: DeliverableDetail | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useDeliverableDetail(deliverableID?: number): UseDeliverableDetailResult {
  const [data, setData] = useState<DeliverableDetail | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  const fetchData = useCallback(async () => {
    if (!deliverableID) {
      setData(null);
      setError("交付物 ID 不合法");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const result = await getDeliverableDetail(deliverableID);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载交付详情失败");
    } finally {
      setLoading(false);
    }
  }, [deliverableID]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

export function useCreateDeliverable() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (input: CreateDeliverableInput) => {
    setLoading(true);
    try {
      return await createDeliverable(input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}

export function useReviewDeliverable() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (deliverableID: number, input: ReviewDeliverableInput) => {
    setLoading(true);
    try {
      return await reviewDeliverable(deliverableID, input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}
