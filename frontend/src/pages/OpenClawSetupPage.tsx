import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { useCreateRegistrationCode, useOpenClawIntegrations, useRegisterOpenClaw } from "../hooks/useOpenClaw";

function OpenClawSetupPage() {
  const [expiresInMinutes, setExpiresInMinutes] = useState(30);
  const [registrationCode, setRegistrationCode] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [bridgeVersion, setBridgeVersion] = useState("0.2.0");
  const [openClawVersion, setOpenClawVersion] = useState("1.0.0");
  const [instanceFingerprint, setInstanceFingerprint] = useState("");
  const [callbackURL, setCallbackURL] = useState("http://127.0.0.1:9999/callback");
  const [ownerName, setOwnerName] = useState("");
  const [ownerEmail, setOwnerEmail] = useState("");
  const [ownerRoleType, setOwnerRoleType] = useState("operation");
  const [ownerExternalID, setOwnerExternalID] = useState("");
  const [agentName, setAgentName] = useState("");
  const [agentCode, setAgentCode] = useState("");
  const [capabilitiesText, setCapabilitiesText] = useState('{"task_types":["query"]}');
  const [actionError, setActionError] = useState("");
  const [registerResult, setRegisterResult] = useState<{
    integrationID: number;
    callbackSecret: string;
    status: string;
  } | null>(null);

  const createCodeMutation = useCreateRegistrationCode();
  const registerMutation = useRegisterOpenClaw();
  const integrationsQuery = useMemo(
    () => ({
      page: 1,
      page_size: 20
    }),
    []
  );
  const integrations = useOpenClawIntegrations(integrationsQuery);

  const validate = (): string => {
    if (!registrationCode.trim()) {
      return "请先生成或填写注册码";
    }
    if (!instanceFingerprint.trim()) {
      return "实例指纹不能为空";
    }
    if (!callbackURL.trim()) {
      return "回调地址不能为空";
    }
    if (!ownerName.trim()) {
      return "拥有者姓名不能为空";
    }
    if (!ownerEmail.trim()) {
      return "拥有者邮箱不能为空";
    }
    if (!agentName.trim()) {
      return "龙虾显示名不能为空";
    }
    if (!agentCode.trim()) {
      return "龙虾编码不能为空";
    }
    if (!bridgeVersion.trim()) {
      return "Bridge 版本不能为空";
    }
    if (!capabilitiesText.trim()) {
      return "能力描述不能为空";
    }
    try {
      JSON.parse(capabilitiesText);
    } catch {
      return "能力描述必须是合法 JSON";
    }
    return "";
  };

  return (
    <section className="setup-page-grid">
      <article className="page-card">
        <div className="page-title">
          <div>
            <span className="section-kicker">openclaw onboarding</span>
            <h2>接入我的 OpenClaw</h2>
          </div>
          <div className="toolbar">
            <Link className="btn" to="/">
              返回工作台
            </Link>
            <Link className="btn" to="/resources/agents">
              龙虾管理
            </Link>
          </div>
        </div>

        <p className="page-note">
          这页负责把 OpenClaw Bridge 接入当前平台，并同步注册拥有者和龙虾归属。生成注册码后，把相关字段交给 Bridge setup wizard 或你自己的接入流程使用。
        </p>

        {actionError ? <p className="error-text">{actionError}</p> : null}
        {registerResult ? (
          <div className="warning-text">
            接入成功：Integration #{registerResult.integrationID}，状态 {registerResult.status}，回调密钥 {registerResult.callbackSecret}
          </div>
        ) : null}

        <div className="setup-block">
          <div className="page-title">
            <div>
              <span className="section-kicker">step 1</span>
              <h3>生成注册码</h3>
            </div>
          </div>
          <div className="form-grid">
            <label>
              有效期（分钟）
              <input
                type="number"
                min={5}
                max={1440}
                value={expiresInMinutes}
                onChange={(event) => setExpiresInMinutes(Number(event.target.value) || 30)}
              />
            </label>
            <label>
              当前注册码
              <input value={registrationCode} onChange={(event) => setRegistrationCode(event.target.value)} placeholder="点击按钮生成注册码" />
            </label>
          </div>
          <div className="toolbar">
            <button
              type="button"
              className="btn btn-primary"
              disabled={createCodeMutation.loading}
              onClick={async () => {
                setActionError("");
                setRegisterResult(null);
                try {
                  const result = await createCodeMutation.run({ expires_in_minutes: expiresInMinutes });
                  setRegistrationCode(result.code);
                } catch (err) {
                  setActionError(err instanceof Error ? err.message : "生成注册码失败");
                }
              }}
            >
              {createCodeMutation.loading ? "生成中..." : "生成注册码"}
            </button>
          </div>
        </div>

        <div className="setup-block">
          <div className="page-title">
            <div>
              <span className="section-kicker">step 2</span>
              <h3>填写接入信息</h3>
            </div>
          </div>
          <div className="form-grid">
            <label>
              Integration 显示名
              <input value={displayName} onChange={(event) => setDisplayName(event.target.value)} placeholder="例如：Alice's OpenClaw" />
            </label>
            <label>
              Bridge 版本 *
              <input value={bridgeVersion} onChange={(event) => setBridgeVersion(event.target.value)} placeholder="例如：0.2.0" />
            </label>
            <label>
              OpenClaw 版本
              <input value={openClawVersion} onChange={(event) => setOpenClawVersion(event.target.value)} placeholder="例如：1.0.0" />
            </label>
            <label>
              实例指纹 *
              <input value={instanceFingerprint} onChange={(event) => setInstanceFingerprint(event.target.value)} placeholder="例如：hostname-user-device" />
            </label>
            <label className="full-width">
              回调地址 *
              <input value={callbackURL} onChange={(event) => setCallbackURL(event.target.value)} placeholder="例如：http://127.0.0.1:9999/callback" />
            </label>
            <label>
              拥有者姓名 *
              <input value={ownerName} onChange={(event) => setOwnerName(event.target.value)} placeholder="例如：Alice" />
            </label>
            <label>
              拥有者邮箱 *
              <input value={ownerEmail} onChange={(event) => setOwnerEmail(event.target.value)} placeholder="例如：alice@example.com" />
            </label>
            <label>
              拥有者角色
              <input value={ownerRoleType} onChange={(event) => setOwnerRoleType(event.target.value)} placeholder="例如：operation" />
            </label>
            <label>
              拥有者外部 ID
              <input value={ownerExternalID} onChange={(event) => setOwnerExternalID(event.target.value)} placeholder="例如：alice-openclaw-001" />
            </label>
            <label>
              龙虾显示名 *
              <input value={agentName} onChange={(event) => setAgentName(event.target.value)} placeholder="例如：Alice Planner" />
            </label>
            <label>
              龙虾编码 *
              <input value={agentCode} onChange={(event) => setAgentCode(event.target.value)} placeholder="例如：alice_planner_001" />
            </label>
            <label className="full-width">
              能力 JSON *
              <textarea rows={8} value={capabilitiesText} onChange={(event) => setCapabilitiesText(event.target.value)} />
            </label>
          </div>

          <div className="toolbar">
            <button
              type="button"
              className="btn btn-primary"
              disabled={registerMutation.loading}
              onClick={async () => {
                const message = validate();
                if (message) {
                  setActionError(message);
                  return;
                }
                setActionError("");
                setRegisterResult(null);
                try {
                  const result = await registerMutation.run({
                    protocol_version: 1,
                    registration_code: registrationCode.trim(),
                    bridge_version: bridgeVersion.trim(),
                    openclaw_version: openClawVersion.trim() || undefined,
                    instance_fingerprint: instanceFingerprint.trim(),
                    display_name: displayName.trim() || undefined,
                    callback_url: callbackURL.trim(),
                    owner_name: ownerName.trim(),
                    owner_email: ownerEmail.trim(),
                    owner_role_type: ownerRoleType.trim() || undefined,
                    owner_external_id: ownerExternalID.trim() || undefined,
                    agent_name: agentName.trim(),
                    agent_code: agentCode.trim(),
                    capabilities: JSON.parse(capabilitiesText)
                  });
                  setRegisterResult({
                    integrationID: result.integration_id,
                    callbackSecret: result.callback_secret,
                    status: result.status
                  });
                  await integrations.refetch();
                } catch (err) {
                  setActionError(err instanceof Error ? err.message : "接入失败");
                }
              }}
            >
              {registerMutation.loading ? "接入中..." : "提交接入"}
            </button>
          </div>
        </div>
      </article>

      <aside className="page-card setup-side-panel">
        <div className="page-title">
          <div>
            <span className="section-kicker">recent integrations</span>
            <h2>最近接入实例</h2>
          </div>
          <button type="button" className="btn btn-text" onClick={() => void integrations.refetch()}>
            刷新
          </button>
        </div>

        {integrations.loading ? <p>加载中...</p> : null}
        {integrations.error ? <p className="error-text">{integrations.error}</p> : null}

        <ul className="plain-list draft-mini-list">
          {(integrations.data?.items ?? []).map((integration) => (
            <li key={integration.id}>
              <div className="comment-row">
                <strong>{integration.display_name}</strong>
                <span className={`pill ${integration.status === "active" ? "" : "draft-status-discarded"}`}>{integration.status}</span>
              </div>
              <p className="muted">指纹：{integration.instance_fingerprint}</p>
              <p className="muted">绑定龙虾 ID：#{integration.bound_agent_id}</p>
              <p className="muted">Bridge：{integration.bridge_version} / OpenClaw：{integration.openclaw_version || "-"}</p>
            </li>
          ))}
        </ul>

        {!integrations.loading && !integrations.error && !(integrations.data?.items.length ?? 0) ? (
          <p className="muted">还没有接入实例，可以先生成注册码并完成第一次接入。</p>
        ) : null}
      </aside>
    </section>
  );
}

export default OpenClawSetupPage;
