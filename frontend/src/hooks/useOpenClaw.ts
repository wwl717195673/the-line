import { useCallback, useEffect, useState } from "react";
import { createRegistrationCode, listOpenClawIntegrations, registerOpenClaw } from "../api/openclaw";
import type {
  CreateRegistrationCodeInput,
  OpenClawIntegration,
  OpenClawIntegrationQuery,
  PageData,
  RegisterOpenClawInput,
  RegisterOpenClawResult,
  RegistrationCode
} from "../types/api";

type UseOpenClawIntegrationsResult = {
  data: PageData<OpenClawIntegration> | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useOpenClawIntegrations(query: OpenClawIntegrationQuery): UseOpenClawIntegrationsResult {
  const [data, setData] = useState<PageData<OpenClawIntegration> | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await listOpenClawIntegrations(query);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载接入实例失败");
    } finally {
      setLoading(false);
    }
  }, [query]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

export function useCreateRegistrationCode() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (input: CreateRegistrationCodeInput): Promise<RegistrationCode> => {
    setLoading(true);
    try {
      return await createRegistrationCode(input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}

export function useRegisterOpenClaw() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (input: RegisterOpenClawInput): Promise<RegisterOpenClawResult> => {
    setLoading(true);
    try {
      return await registerOpenClaw(input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}
