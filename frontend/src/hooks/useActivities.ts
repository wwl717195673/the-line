import { useCallback, useEffect, useState } from "react";
import { listRecentActivities, type RecentActivity } from "../api/activities";

type UseRecentActivitiesResult = {
  data: RecentActivity[];
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useRecentActivities(limit = 20, options?: { enabled?: boolean }): UseRecentActivitiesResult {
  const [data, setData] = useState<RecentActivity[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const enabled = options?.enabled ?? true;

  const fetchData = useCallback(async () => {
    if (!enabled) {
      setLoading(false);
      setError("");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const next = await listRecentActivities(limit);
      setData(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载最近动态失败");
    } finally {
      setLoading(false);
    }
  }, [enabled, limit]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}
