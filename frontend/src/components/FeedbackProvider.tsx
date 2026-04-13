import { createContext, useCallback, useContext, useMemo, useState, type PropsWithChildren } from "react";

type ConfirmTone = "default" | "danger";
type ToastTone = "info" | "success" | "danger";

type ConfirmOptions = {
  title: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  tone?: ConfirmTone;
};

type ToastOptions = {
  title: string;
  message?: string;
  tone?: ToastTone;
};

type ConfirmState = ConfirmOptions & {
  open: boolean;
  resolve: (value: boolean) => void;
};

type ToastItem = ToastOptions & {
  id: number;
};

type FeedbackContextValue = {
  confirm: (options: ConfirmOptions) => Promise<boolean>;
  notify: (options: ToastOptions) => void;
};

const FeedbackContext = createContext<FeedbackContextValue | null>(null);

function FeedbackProvider({ children }: PropsWithChildren) {
  const [confirmState, setConfirmState] = useState<ConfirmState | null>(null);
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const confirm = useCallback((options: ConfirmOptions) => {
    return new Promise<boolean>((resolve) => {
      setConfirmState({
        ...options,
        open: true,
        resolve
      });
    });
  }, []);

  const notify = useCallback((options: ToastOptions) => {
    const id = window.setTimeout(() => undefined, 0);
    window.clearTimeout(id);
    const nextToast: ToastItem = {
      id,
      tone: "info",
      ...options
    };
    setToasts((current) => [...current, nextToast]);
    window.setTimeout(() => {
      setToasts((current) => current.filter((toast) => toast.id !== nextToast.id));
    }, 3200);
  }, []);

  const handleConfirmClose = useCallback(
    (value: boolean) => {
      setConfirmState((current) => {
        current?.resolve(value);
        return null;
      });
    },
    []
  );

  const contextValue = useMemo<FeedbackContextValue>(
    () => ({
      confirm,
      notify
    }),
    [confirm, notify]
  );

  return (
    <FeedbackContext.Provider value={contextValue}>
      {children}
      {confirmState?.open ? (
        <div className="modal-mask" onClick={() => handleConfirmClose(false)}>
          <div className="modal-panel confirm-panel" onClick={(event) => event.stopPropagation()}>
            <div className="modal-header">
              <h3>{confirmState.title}</h3>
              <button type="button" className="btn btn-text" onClick={() => handleConfirmClose(false)}>
                关闭
              </button>
            </div>
            <div className="modal-body confirm-panel-body">
              <p>{confirmState.message}</p>
              <div className="modal-actions">
                <button type="button" className="btn" onClick={() => handleConfirmClose(false)}>
                  {confirmState.cancelText ?? "取消"}
                </button>
                <button
                  type="button"
                  className={`btn ${confirmState.tone === "danger" ? "danger" : "btn-primary"}`}
                  onClick={() => handleConfirmClose(true)}
                >
                  {confirmState.confirmText ?? "确认"}
                </button>
              </div>
            </div>
          </div>
        </div>
      ) : null}
      <div className="toast-viewport" aria-live="polite" aria-atomic="true">
        {toasts.map((toast) => (
          <article key={toast.id} className={`toast-card ${toast.tone ?? "info"}`}>
            <strong>{toast.title}</strong>
            {toast.message ? <p>{toast.message}</p> : null}
          </article>
        ))}
      </div>
    </FeedbackContext.Provider>
  );
}

export function useFeedback() {
  const context = useContext(FeedbackContext);
  if (!context) {
    throw new Error("useFeedback must be used within FeedbackProvider");
  }
  return context;
}

export default FeedbackProvider;
