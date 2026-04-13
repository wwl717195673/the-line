import { requestJSON } from "../lib/http";

export type RecentActivity = {
  id: number;
  run_id: number;
  run_title: string;
  run_node_id: number;
  node_name: string;
  log_type: string;
  operator_type: string;
  operator_id: number;
  operator_name: string;
  content: string;
  created_at: string;
};

export function listRecentActivities(limit = 20): Promise<RecentActivity[]> {
  return requestJSON<RecentActivity[]>("/api/activities/recent", undefined, { limit });
}
