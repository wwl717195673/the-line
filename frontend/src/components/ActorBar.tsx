import { useEffect, useState } from "react";
import { getActor, setActor } from "../lib/actor";

function ActorBar() {
  const [personIDInput, setPersonIDInput] = useState("");
  const [roleType, setRoleType] = useState("");
  const [isExpanded, setIsExpanded] = useState(false);

  useEffect(() => {
    const actor = getActor();
    setPersonIDInput(actor.personId ? String(actor.personId) : "");
    setRoleType(actor.roleType ?? "");
  }, []);

  const hasActor = !!personIDInput.trim() || !!roleType;
  const actorLabel = personIDInput.trim() ? `P-${personIDInput.trim()}` : "访客";
  const roleLabel = roleType || "普通用户";

  return (
    <section className="actor-console">
      <button
        type="button"
        className={`actor-summary ${isExpanded ? "active" : ""}`}
        onClick={() => setIsExpanded((value) => !value)}
        aria-expanded={isExpanded}
      >
        <span className={`actor-summary-dot ${hasActor ? "ready" : ""}`} aria-hidden="true" />
        <div>
          <strong>{actorLabel}</strong>
          <p>{roleLabel}</p>
        </div>
        <span className="actor-summary-action">{isExpanded ? "收起" : "切换身份"}</span>
      </button>
      {isExpanded ? (
        <div className="actor-bar">
          <input value={personIDInput} onChange={(event) => setPersonIDInput(event.target.value)} placeholder="输入 Person ID" />
          <select value={roleType} onChange={(event) => setRoleType(event.target.value)}>
            <option value="">普通用户</option>
            <option value="admin">admin</option>
          </select>
          <button
            type="button"
            className="btn"
            onClick={() => {
              const parsed = Number(personIDInput);
              setActor({
                personId: Number.isFinite(parsed) && parsed > 0 ? parsed : undefined,
                roleType: roleType || undefined
              });
              window.location.reload();
            }}
          >
            应用身份
          </button>
        </div>
      ) : null}
    </section>
  );
}

export default ActorBar;
