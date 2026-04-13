import { useCallback, useEffect, useState } from "react";
import { deleteTemplate, getTemplateDetail, listTemplates } from "../api/templates";
import type { PageData, Template, TemplateDetail, TemplateQuery } from "../types/api";

type UseTemplatesResult = {
  data: PageData<Template> | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useTemplates(query: TemplateQuery): UseTemplatesResult {
  const [data, setData] = useState<PageData<Template> | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await listTemplates(query);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载模板失败");
    } finally {
      setLoading(false);
    }
  }, [query]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

type UseTemplateDetailResult = {
  data: TemplateDetail | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useTemplateDetail(templateId?: number): UseTemplateDetailResult {
  const [data, setData] = useState<TemplateDetail | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  const fetchData = useCallback(async () => {
    if (!templateId) {
      setData(null);
      setError("模板 ID 不合法");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const result = await getTemplateDetail(templateId);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载模板详情失败");
    } finally {
      setLoading(false);
    }
  }, [templateId]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

export function useDeleteTemplate() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (id: number) => {
    setLoading(true);
    try {
      return await deleteTemplate(id);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}
