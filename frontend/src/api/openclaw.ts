import { requestJSON } from "../lib/http";
import type {
  CreateRegistrationCodeInput,
  OpenClawIntegration,
  OpenClawIntegrationQuery,
  PageData,
  RegisterOpenClawInput,
  RegisterOpenClawResult,
  RegistrationCode
} from "../types/api";

export function createRegistrationCode(input: CreateRegistrationCodeInput): Promise<RegistrationCode> {
  return requestJSON<RegistrationCode>("/api/integrations/openclaw/registration-codes", {
    method: "POST",
    body: JSON.stringify(input)
  });
}

export function listOpenClawIntegrations(query: OpenClawIntegrationQuery): Promise<PageData<OpenClawIntegration>> {
  return requestJSON<PageData<OpenClawIntegration>>("/api/integrations/openclaw", undefined, query);
}

export function registerOpenClaw(input: RegisterOpenClawInput): Promise<RegisterOpenClawResult> {
  return requestJSON<RegisterOpenClawResult>("/api/integrations/openclaw/register", {
    method: "POST",
    body: JSON.stringify(input)
  });
}
