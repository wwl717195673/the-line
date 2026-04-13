import type { AppErrorPayload } from "../types/api";
import { getActor } from "./actor";

const API_BASE_URL = ((import.meta.env.VITE_API_BASE_URL as string | undefined) ?? "").replace(/\/$/, "");

export class ApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(message: string, code: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }
}

function buildQuery(params?: Record<string, string | number | undefined>): string {
  if (!params) {
    return "";
  }
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === "") {
      return;
    }
    search.set(key, String(value));
  });
  const text = search.toString();
  return text ? `?${text}` : "";
}

async function parseError(response: Response): Promise<never> {
  let payload: AppErrorPayload | null = null;
  try {
    payload = (await response.json()) as AppErrorPayload;
  } catch {
    payload = null;
  }
  const message = payload?.message ?? `请求失败（${response.status}）`;
  const code = payload?.code ?? "HTTP_ERROR";
  throw new ApiError(message, code, response.status);
}

export async function requestJSON<T>(
  path: string,
  init?: RequestInit,
  query?: Record<string, string | number | undefined>
): Promise<T> {
  const actor = getActor();
  const actorHeaders: Record<string, string> = {};
  if (actor.personId) {
    actorHeaders["X-Person-ID"] = String(actor.personId);
  }
  if (actor.roleType) {
    actorHeaders["X-Role-Type"] = actor.roleType;
  }

  const response = await fetch(`${API_BASE_URL}${path}${buildQuery(query)}`, {
    ...init,
    headers: {
      ...(init?.body instanceof FormData ? {} : { "Content-Type": "application/json" }),
      ...actorHeaders,
      ...(init?.headers ?? {})
    }
  });
  if (!response.ok) {
    await parseError(response);
  }
  return (await response.json()) as T;
}

export function getAPIBaseURL(): string {
  return API_BASE_URL;
}
