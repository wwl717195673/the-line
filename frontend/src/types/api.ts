export type Status = 0 | 1;

export type PageData<T> = {
  items: T[];
  total: number;
  page: number;
  page_size: number;
};

export type AppErrorPayload = {
  code?: string;
  message?: string;
};

export type Person = {
  id: number;
  name: string;
  email: string;
  role_type: string;
  status: Status;
  created_at: string;
  updated_at: string;
};

export type Agent = {
  id: number;
  name: string;
  code: string;
  provider: string;
  version: string;
  owner_person_id: number;
  config_json: unknown;
  status: Status;
  created_at: string;
  updated_at: string;
};

export type PersonQuery = {
  page: number;
  page_size: number;
  status?: Status;
  keyword?: string;
};

export type AgentQuery = {
  page: number;
  page_size: number;
  status?: Status;
  keyword?: string;
};

export type CreatePersonInput = {
  name: string;
  email: string;
  role_type: string;
};

export type UpdatePersonInput = {
  name: string;
  email: string;
  role_type: string;
  status: Status;
};

export type CreateAgentInput = {
  name: string;
  code: string;
  provider: string;
  version: string;
  owner_person_id: number;
  config_json: unknown;
};

export type UpdateAgentInput = {
  name: string;
  code: string;
  provider: string;
  version: string;
  owner_person_id: number;
  config_json: unknown;
  status: Status;
};

export type Template = {
  id: number;
  name: string;
  code: string;
  version: number;
  category: string;
  description: string;
  status: string;
  created_by: number;
  created_at: string;
  updated_at: string;
};

export type TemplateNode = {
  id: number;
  template_id: number;
  node_code: string;
  node_name: string;
  node_type: string;
  sort_order: number;
  default_owner_rule: string;
  default_owner_person_id?: number;
  result_owner_rule: string;
  result_owner_person_id?: number;
  default_agent_id?: number;
  default_agent?: {
    id: number;
    name: string;
    code: string;
    provider: string;
    version: string;
    status: Status;
  };
  input_schema_json: unknown;
  output_schema_json: unknown;
  config_json: unknown;
  created_at: string;
  updated_at: string;
};

export type TemplateDetail = Template & {
  nodes: TemplateNode[];
};

export type TemplateQuery = {
  page: number;
  page_size: number;
  keyword?: string;
};

export type CreateRunInput = {
  template_id: number;
  title: string;
  biz_key?: string;
  initiator_person_id?: number;
  input_payload_json: unknown;
};

export type RunDetail = {
  id: number;
  template_id: number;
  template_version: number;
  title: string;
  biz_key: string;
  initiator_person_id: number;
  initiator?: {
    id: number;
    name: string;
    email: string;
    role_type: string;
    status: Status;
  };
  current_status: string;
  current_node_code: string;
  current_node?: RunNode;
  input_payload_json: unknown;
  output_payload_json: unknown;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  template?: Template;
  has_deliverable: boolean;
  deliverable_id?: number;
  nodes: RunNode[];
  logs?: RunNodeLog[];
};

export type RunNode = {
  id: number;
  run_id: number;
  template_node_id: number;
  node_code: string;
  node_name: string;
  node_type: string;
  sort_order: number;
  owner_person_id?: number;
  owner_person?: {
    id: number;
    name: string;
    email: string;
    role_type: string;
    status: Status;
  };
  reviewer_person_id?: number;
  reviewer_person?: {
    id: number;
    name: string;
    email: string;
    role_type: string;
    status: Status;
  };
  result_owner_person_id?: number;
  result_owner_person?: {
    id: number;
    name: string;
    email: string;
    role_type: string;
    status: Status;
  };
  bound_agent_id?: number;
  bound_agent?: {
    id: number;
    name: string;
    code: string;
    provider: string;
    version: string;
    status: Status;
  };
  status: string;
  input_json: unknown;
  output_json: unknown;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
  is_current: boolean;
};

export type RunNodeLog = {
  id: number;
  run_id: number;
  run_node_id: number;
  log_type: string;
  operator_type: string;
  operator_id: number;
  content: string;
  extra_json: unknown;
  created_at: string;
};

export type Attachment = {
  id: number;
  target_type: string;
  target_id: number;
  file_name: string;
  file_url: string;
  file_size: number;
  file_type: string;
  uploaded_by: number;
  created_at: string;
};

export type Comment = {
  id: number;
  target_type: string;
  target_id: number;
  author_person_id: number;
  author?: {
    id: number;
    name: string;
    email: string;
    role_type: string;
    status: Status;
  };
  content: string;
  is_resolved: boolean;
  created_at: string;
  updated_at: string;
};

export type RunNodeDetail = RunNode & {
  run?: RunListItem;
  attachments: Attachment[];
  comments: Comment[];
  logs: RunNodeLog[];
  available_actions: string[];
};

export type RunQuery = {
  page: number;
  page_size: number;
  scope: "all" | "initiated_by_me" | "todo";
  status?: string;
  owner_person_id?: number;
  initiator_person_id?: number;
};

export type RunListItem = {
  id: number;
  template_id: number;
  template_version: number;
  title: string;
  biz_key: string;
  initiator_person_id: number;
  initiator?: {
    id: number;
    name: string;
    email: string;
    role_type: string;
    status: Status;
  };
  current_status: string;
  current_node_code: string;
  current_node?: RunNode;
  input_payload_json: unknown;
  output_payload_json: unknown;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
};

export type Deliverable = {
  id: number;
  run_id: number;
  run?: RunListItem;
  title: string;
  summary: string;
  result_json: unknown;
  reviewer_person_id: number;
  reviewer?: {
    id: number;
    name: string;
    email: string;
    role_type: string;
    status: Status;
  };
  review_status: "pending" | "approved" | "rejected";
  reviewed_at?: string;
  created_at: string;
  updated_at: string;
};

export type DeliverableDetail = Deliverable & {
  nodes: RunNode[];
  attachments: Attachment[];
};

export type DraftNode = {
  node_code: string;
  node_name: string;
  node_type: string;
  sort_order: number;
  executor_type: "agent" | "human";
  owner_rule: string;
  owner_person_id?: number;
  executor_agent_code?: string;
  result_owner_rule: string;
  result_owner_person_id?: number;
  task_type?: string;
  input_schema: Record<string, unknown>;
  output_schema: Record<string, unknown>;
  completion_condition?: string;
  failure_condition?: string;
  escalation_rule?: string;
};

export type DraftPlan = {
  title: string;
  description: string;
  nodes: DraftNode[];
  final_deliverable: string;
};

export type FlowDraft = {
  id: number;
  title: string;
  description: string;
  source_prompt: string;
  creator_person_id: number;
  planner_agent_id?: number;
  status: "draft" | "confirmed" | "discarded";
  structured_plan_json: DraftPlan;
  confirmed_template_id?: number;
  created_at: string;
  updated_at: string;
  confirmed_at?: string;
};

export type FlowDraftQuery = {
  page: number;
  page_size: number;
  creator_person_id?: number;
  status?: FlowDraft["status"];
};

export type CreateDraftInput = {
  title?: string;
  description?: string;
  source_prompt: string;
  creator_person_id: number;
  planner_agent_id: number;
  structured_plan_json?: DraftPlan;
};

export type UpdateDraftInput = {
  title?: string;
  description?: string;
  planner_agent_id?: number;
  structured_plan_json?: DraftPlan;
};

export type ConfirmDraftResponse = {
  draft_id: number;
  template_id: number;
  message: string;
};

export type AgentArtifact = {
  name: string;
  url: string;
  type: string;
};

export type AgentTask = {
  id: number;
  run_id: number;
  run_node_id: number;
  agent_id: number;
  task_type: string;
  input_json: Record<string, unknown>;
  status: "queued" | "running" | "completed" | "needs_review" | "failed" | "blocked" | "cancelled";
  started_at?: string;
  finished_at?: string;
  error_message?: string;
  result_json: Record<string, unknown>;
  artifacts_json: AgentArtifact[];
  created_at: string;
  updated_at: string;
};

export type AgentTaskReceipt = {
  id: number;
  agent_task_id: number;
  run_id: number;
  run_node_id: number;
  agent_id: number;
  receipt_status: "completed" | "needs_review" | "failed" | "blocked";
  payload_json: Record<string, unknown>;
  received_at: string;
};

export type ConfirmAgentResultInput = {
  action: "confirm" | "reject";
  comment?: string;
};

export type TakeoverRunNodeInput = {
  action: "complete" | "retry";
  comment?: string;
  manual_result?: Record<string, unknown>;
};

export type DeliverableQuery = {
  page: number;
  page_size: number;
  review_status?: "pending" | "approved" | "rejected";
  reviewer_person_id?: number;
};

export type CreateDeliverableInput = {
  run_id: number;
  title: string;
  summary: string;
  result_json: unknown;
  reviewer_person_id: number;
  attachment_ids: number[];
};

export type ReviewDeliverableInput = {
  review_status: "approved" | "rejected";
  review_comment: string;
};
