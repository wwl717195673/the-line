import type { PropsWithChildren } from "react";

type ModalProps = PropsWithChildren<{
  open: boolean;
  title: string;
  onClose: () => void;
}>;

function Modal({ open, title, onClose, children }: ModalProps) {
  if (!open) {
    return null;
  }

  return (
    <div className="modal-mask" onClick={onClose}>
      <div className="modal-panel" onClick={(event) => event.stopPropagation()}>
        <div className="modal-header">
          <h3>{title}</h3>
          <button type="button" className="btn btn-text" onClick={onClose}>
            关闭
          </button>
        </div>
        <div className="modal-body">{children}</div>
      </div>
    </div>
  );
}

export default Modal;
