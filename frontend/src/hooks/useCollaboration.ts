import { useCallback, useEffect, useState } from "react";
import {
  createAttachment,
  createComment,
  listAttachments,
  listComments,
  resolveComment,
  uploadAttachmentFile
} from "../api/collaboration";
import type { Attachment, Comment } from "../types/api";

type UseCommentsResult = {
  data: Comment[];
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useComments(targetType: string, targetID?: number): UseCommentsResult {
  const [data, setData] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    if (!targetID) {
      setData([]);
      setError("");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const next = await listComments(targetType, targetID);
      setData(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载评论失败");
    } finally {
      setLoading(false);
    }
  }, [targetID, targetType]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

type UseAttachmentsResult = {
  data: Attachment[];
  loading: boolean;
  error: string;
  refetch: () => Promise<void>;
};

export function useAttachments(targetType: string, targetID?: number): UseAttachmentsResult {
  const [data, setData] = useState<Attachment[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const fetchData = useCallback(async () => {
    if (!targetID) {
      setData([]);
      setError("");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const next = await listAttachments(targetType, targetID);
      setData(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载附件失败");
    } finally {
      setLoading(false);
    }
  }, [targetID, targetType]);

  useEffect(() => {
    void fetchData();
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

export function useCollaborationActions() {
  const [loading, setLoading] = useState(false);

  const run = useCallback(async <T,>(fn: () => Promise<T>) => {
    setLoading(true);
    try {
      return await fn();
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    loading,
    createComment: (targetType: string, targetID: number, content: string) =>
      run(() => createComment({ target_type: targetType, target_id: targetID, content })),
    resolveComment: (commentID: number) => run(() => resolveComment(commentID)),
    createAttachment: (
      targetType: string,
      targetID: number,
      payload: { file_name: string; file_url: string; file_size: number; file_type: string }
    ) => run(() => createAttachment({ target_type: targetType, target_id: targetID, ...payload })),
    uploadAttachmentFile: (targetType: string, targetID: number, file: File) =>
      run(() =>
        uploadAttachmentFile({
          target_type: targetType,
          target_id: targetID,
          file
        })
      )
  };
}
