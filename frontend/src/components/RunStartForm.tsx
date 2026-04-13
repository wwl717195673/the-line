import { useEffect, useState } from "react";
import PersonSelect from "./PersonSelect";

type RunStartFormValue = {
  title: string;
  reason: string;
  class_info: string;
  current_teacher: string;
  expected_time: string;
  extra_note: string;
  initiator_person_id?: number;
};

type RunStartFormProps = {
  defaultTitle: string;
  submitting: boolean;
  onSubmit: (value: RunStartFormValue) => Promise<void>;
};

function RunStartForm({ defaultTitle, submitting, onSubmit }: RunStartFormProps) {
  const [title, setTitle] = useState(defaultTitle);
  const [reason, setReason] = useState("");
  const [classInfo, setClassInfo] = useState("");
  const [currentTeacher, setCurrentTeacher] = useState("");
  const [expectedTime, setExpectedTime] = useState("");
  const [extraNote, setExtraNote] = useState("");
  const [initiatorPersonID, setInitiatorPersonID] = useState<number | undefined>(undefined);
  const [error, setError] = useState("");

  useEffect(() => {
    setTitle(defaultTitle);
  }, [defaultTitle]);

  const validate = (): string => {
    if (!title.trim()) {
      return "实例标题不能为空";
    }
    if (!reason.trim()) {
      return "申请原因不能为空";
    }
    if (!classInfo.trim()) {
      return "涉及班级不能为空";
    }
    if (!currentTeacher.trim()) {
      return "当前班主任不能为空";
    }
    if (!expectedTime.trim()) {
      return "期望处理时间不能为空";
    }
    return "";
  };

  return (
    <form
      className="form-grid"
      onSubmit={(event) => {
        event.preventDefault();
        const validationError = validate();
        if (validationError) {
          setError(validationError);
          return;
        }
        setError("");
        void onSubmit({
          title: title.trim(),
          reason: reason.trim(),
          class_info: classInfo.trim(),
          current_teacher: currentTeacher.trim(),
          expected_time: expectedTime.trim(),
          extra_note: extraNote.trim(),
          initiator_person_id: initiatorPersonID
        }).catch((err: unknown) => {
          setError(err instanceof Error ? err.message : "发起流程失败");
        });
      }}
    >
      <label className="full-width">
        实例标题
        <input value={title} onChange={(event) => setTitle(event.target.value)} placeholder="请输入实例标题" />
      </label>
      <label>
        申请原因
        <input value={reason} onChange={(event) => setReason(event.target.value)} placeholder="请输入申请原因" />
      </label>
      <label>
        涉及班级
        <input value={classInfo} onChange={(event) => setClassInfo(event.target.value)} placeholder="请输入涉及班级" />
      </label>
      <label>
        当前班主任
        <input value={currentTeacher} onChange={(event) => setCurrentTeacher(event.target.value)} placeholder="请输入当前班主任" />
      </label>
      <label>
        期望处理时间
        <input value={expectedTime} onChange={(event) => setExpectedTime(event.target.value)} placeholder="例如 2026-04-08 18:00 前" />
      </label>
      <label>
        发起人（可选）
        <PersonSelect value={initiatorPersonID} onChange={setInitiatorPersonID} placeholder="不填则依赖请求头 X-Person-ID" />
      </label>
      <label className="full-width">
        补充说明
        <textarea rows={5} value={extraNote} onChange={(event) => setExtraNote(event.target.value)} placeholder="可选补充说明" />
      </label>
      <label className="full-width">
        附件（MVP 占位）
        <input disabled value="后续接入附件上传模块（05）" />
      </label>
      {error ? <p className="error-text">{error}</p> : null}
      <div className="modal-actions">
        <button type="submit" className="btn btn-primary" disabled={submitting}>
          {submitting ? "提交中..." : "发起流程"}
        </button>
      </div>
    </form>
  );
}

export type { RunStartFormValue };
export default RunStartForm;
