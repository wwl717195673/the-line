import { useCallback, useEffect, useState } from "react";
import { createPerson, disablePerson, listPersons, updatePerson } from "../api/persons";
import type { CreatePersonInput, PageData, Person, PersonQuery, UpdatePersonInput } from "../types/api";

type UsePersonsResult = {
  data: PageData<Person> | null;
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function usePersons(query: PersonQuery): UsePersonsResult {
  const [data, setData] = useState<PageData<Person> | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string>("");

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await listPersons(query);
      setData(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载人员失败");
    } finally {
      setLoading(false);
    }
  }, [query]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

export function useCreatePerson() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (input: CreatePersonInput) => {
    setLoading(true);
    try {
      return await createPerson(input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}

export function useUpdatePerson() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (id: number, input: UpdatePersonInput) => {
    setLoading(true);
    try {
      return await updatePerson(id, input);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}

export function useDisablePerson() {
  const [loading, setLoading] = useState(false);
  const run = useCallback(async (id: number) => {
    setLoading(true);
    try {
      return await disablePerson(id);
    } finally {
      setLoading(false);
    }
  }, []);
  return { run, loading };
}
