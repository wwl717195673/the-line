import { getActor } from "../lib/actor";
import { getAPIBaseURL, requestJSON } from "../lib/http";
import type { Attachment, Comment } from "../types/api";

export function listComments(targetType: string, targetID: number): Promise<Comment[]> {
  return requestJSON<Comment[]>("/api/comments", undefined, {
    target_type: targetType,
    target_id: targetID
  });
}

export function createComment(payload: { target_type: string; target_id: number; content: string }): Promise<Comment> {
  return requestJSON<Comment>("/api/comments", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export function resolveComment(commentID: number): Promise<Comment> {
  return requestJSON<Comment>(`/api/comments/${commentID}/resolve`, {
    method: "POST",
    body: JSON.stringify({})
  });
}

export function listAttachments(targetType: string, targetID: number): Promise<Attachment[]> {
  return requestJSON<Attachment[]>("/api/attachments", undefined, {
    target_type: targetType,
    target_id: targetID
  });
}

export function createAttachment(payload: {
  target_type: string;
  target_id: number;
  file_name: string;
  file_url: string;
  file_size: number;
  file_type: string;
}): Promise<Attachment> {
  return requestJSON<Attachment>("/api/attachments", {
    method: "POST",
    body: JSON.stringify(payload)
  });
}

export async function uploadAttachmentFile(payload: {
  target_type: string;
  target_id: number;
  file: File;
  file_type?: string;
}): Promise<Attachment> {
  const actor = getActor();
  const headers: Record<string, string> = {};
  if (actor.personId) {
    headers["X-Person-ID"] = String(actor.personId);
  }
  if (actor.roleType) {
    headers["X-Role-Type"] = actor.roleType;
  }

  const formData = new FormData();
  formData.set("target_type", payload.target_type);
  formData.set("target_id", String(payload.target_id));
  if (payload.file_type) {
    formData.set("file_type", payload.file_type);
  }
  formData.set("file", payload.file);

  const response = await fetch(`${getAPIBaseURL()}/api/attachments`, {
    method: "POST",
    headers,
    body: formData
  });

  if (!response.ok) {
    let message = `上传失败（${response.status}）`;
    try {
      const err = (await response.json()) as { message?: string };
      if (err.message) {
        message = err.message;
      }
    } catch {
      // ignore
    }
    throw new Error(message);
  }

  return (await response.json()) as Attachment;
}
