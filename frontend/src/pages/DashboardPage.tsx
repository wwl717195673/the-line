import { useMemo } from "react";
import { Link } from "react-router-dom";
import { useRecentActivities } from "../hooks/useActivities";
import { useRuns } from "../hooks/useRuns";
import { getActor } from "../lib/actor";
import RunStatusTag from "../components/RunStatusTag";

function formatTime(value: string) {
  return new Date(value).toLocaleString("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}

function DashboardPage() {
  const actor = getActor();
  const hasActor = !!actor.personId;
  const todoQuery = useMemo(() => ({
    page: 1,
    page_size: 8,
    scope: "todo" as const
  }), []);
  const runningQuery = useMemo(() => ({
    page: 1,
    page_size: 8,
    scope: "all" as const,
    status: "running"
  }), []);
  const todoRuns = useRuns(todoQuery, {
    enabled: hasActor
  });
  const runningRuns = useRuns(runningQuery);
  const recentActivities = useRecentActivities(20);

  const todoCount = todoRuns.data?.items.length ?? 0;
  const runningCount = runningRuns.data?.items.length ?? 0;
  const activityCount = recentActivities.data.length;
  const latestRunning = runningRuns.data?.items[0];
  const latestActivity = recentActivities.data[0];
  const blockedCount = (runningRuns.data?.items ?? []).filter((run) => run.current_status === "blocked").length;
  const waitingCount = (runningRuns.data?.items ?? []).filter((run) => run.current_status === "waiting").length;
  const inFlightRate = runningCount ? Math.round(((runningCount - blockedCount) / runningCount) * 100) : 0;
  const spotlightItems = [
    {
      label: "队列状态",
      value: hasActor ? `${todoCount} 项待处理` : "未识别身份",
      note: hasActor ? "当前角色的输入队列已同步" : "设置身份后开始读取个人工作流"
    },
    {
      label: "运行健康度",
      value: runningCount ? `${inFlightRate}%` : "0%",
      note: runningCount ? `${blockedCount} 个阻塞 / ${waitingCount} 个等待` : "当前没有活跃实例"
    },
    {
      label: "最后动作",
      value: latestActivity ? formatTime(latestActivity.created_at) : "--",
      note: latestActivity ? `${latestActivity.operator_name || `#${latestActivity.operator_id}`} 执行了 ${latestActivity.log_type}` : "活动流暂时为空"
    }
  ];

  return (
    <section className="dashboard-grid">
      <article className="page-card dashboard-hero dashboard-wide">
        <div className="dashboard-hero-copy">
          <span className="section-kicker">协作控制台</span>
          <h2>把流程、责任人和交付物收进同一条工作流里。</h2>
          <p>
            首屏现在直接提供系统态势、运行健康度和下一步动作，而不是把用户扔进一组平均分配的后台卡片里。重点信息应该先被扫到，而不是靠点击再找。
          </p>
          <div className="hero-actions">
            <Link className="btn btn-primary" to={hasActor ? "/runs/todo" : "/templates"}>
              {hasActor ? "继续处理待办" : "发起新流程"}
            </Link>
            <Link className="btn" to="/drafts/create">
              让龙虾生成草案
            </Link>
            <Link className="btn" to="/integrations/openclaw/setup">
              接入 OpenClaw
            </Link>
            <Link className="btn" to="/runs">
              打开流程总览
            </Link>
          </div>
          <div className="spotlight-grid">
            {spotlightItems.map((item) => (
              <article key={item.label} className="spotlight-card">
                <span className="metric-label">{item.label}</span>
                <strong>{item.value}</strong>
                <p>{item.note}</p>
              </article>
            ))}
          </div>
        </div>
        <div className="dashboard-hero-panel">
          <div className="hero-control-panel">
            <div className="panel-header">
              <div>
                <span className="section-kicker">运行面板</span>
                <h3>workspace pulse</h3>
              </div>
              <span className="panel-chip">{runningCount ? "monitoring" : "idle"}</span>
            </div>
            <div className="hero-metric-grid">
              <article className="metric-card">
                <span className="metric-label">运行中流程</span>
                <strong>{runningCount}</strong>
                <p>当前取样为最近 8 条运行中的实例。</p>
              </article>
              <article className="metric-card">
                <span className="metric-label">阻塞节点</span>
                <strong>{blockedCount}</strong>
                <p>需要优先清理的流程瓶颈。</p>
              </article>
              <article className="metric-card">
                <span className="metric-label">等待节点</span>
                <strong>{waitingCount}</strong>
                <p>等待上游输入或人工确认。</p>
              </article>
              <article className="metric-card">
                <span className="metric-label">动态流量</span>
                <strong>{activityCount}</strong>
                <p>最近活动流刷新后立即更新。</p>
              </article>
            </div>
          </div>
          <div className="signal-rail">
            <div className="signal-rail-head">
              <span className="metric-label">节点链路</span>
              <span className="muted">realtime</span>
            </div>
            <div className="signal-track" aria-hidden="true">
              <span className="signal-node active" />
              <span className="signal-node" />
              <span className="signal-node warning" />
              <span className="signal-node" />
              <span className="signal-node success" />
            </div>
            <div className="signal-strip">
              <div>
                <span className="metric-label">最新运行实例</span>
                <p>{latestRunning?.title ?? "暂无运行中的流程"}</p>
              </div>
              <div>
                <span className="metric-label">最新事件</span>
                <p>{latestActivity ? `${latestActivity.log_type} / ${formatTime(latestActivity.created_at)}` : "活动流暂时为空"}</p>
              </div>
            </div>
          </div>
        </div>
      </article>

      <article className="page-card signal-card">
        <div className="page-title">
          <div>
            <span className="section-kicker">我的输入队列</span>
            <h2>待办节点</h2>
          </div>
          <Link className="btn btn-text" to="/runs/todo">
            查看全部
          </Link>
        </div>
        {!hasActor ? <p className="muted">请先在右上角设置身份后查看待办流程。</p> : null}
        {hasActor && todoRuns.loading ? <p>加载中...</p> : null}
        {hasActor && todoRuns.error ? (
          <>
            <p className="error-text">{todoRuns.error}</p>
            <button type="button" className="btn" onClick={() => void todoRuns.refetch()}>
              重试
            </button>
          </>
        ) : null}
        <ul className="plain-list flow-list">
          {(todoRuns.data?.items ?? []).map((run) => (
            <li key={run.id} className="flow-list-item">
              <div className="comment-row">
                <Link to={`/runs/${run.id}`}>{run.title}</Link>
                <RunStatusTag value={run.current_status} />
              </div>
              <div className="flow-meta-grid">
                <p className="muted">当前节点：{(run.current_node?.node_name ?? run.current_node_code) || "-"}</p>
                <p className="muted">发起人：{run.initiator?.name ?? `#${run.initiator_person_id}`}</p>
              </div>
            </li>
          ))}
        </ul>
        {hasActor && !todoRuns.loading && !todoRuns.error && !(todoRuns.data?.items.length ?? 0) ? <p className="muted">暂无待办流程</p> : null}
      </article>

      <article className="page-card signal-card">
        <div className="page-title">
          <div>
            <span className="section-kicker">系统脉冲</span>
            <h2>进行中流程</h2>
          </div>
          <Link className="btn btn-text" to="/runs">
            查看全部
          </Link>
        </div>
        {runningRuns.loading ? <p>加载中...</p> : null}
        {runningRuns.error ? (
          <>
            <p className="error-text">{runningRuns.error}</p>
            <button type="button" className="btn" onClick={() => void runningRuns.refetch()}>
              重试
            </button>
          </>
        ) : null}
        <ul className="plain-list flow-list">
          {(runningRuns.data?.items ?? []).map((run) => (
            <li key={run.id} className="flow-list-item">
              <div className="comment-row">
                <Link to={`/runs/${run.id}`}>{run.title}</Link>
                <RunStatusTag value={run.current_status} />
              </div>
              <div className="flow-meta-grid">
                <p className="muted">发起人：{run.initiator?.name ?? `#${run.initiator_person_id}`}</p>
                <p className="muted">当前节点：{(run.current_node?.node_name ?? run.current_node_code) || "-"}</p>
              </div>
            </li>
          ))}
        </ul>
        {!runningRuns.loading && !runningRuns.error && !(runningRuns.data?.items.length ?? 0) ? <p className="muted">暂无进行中流程</p> : null}
      </article>

      <article className="page-card dashboard-wide">
        <div className="page-title">
          <div>
            <span className="section-kicker">实时事件流</span>
            <h2>最近动态</h2>
          </div>
          <button type="button" className="btn btn-text" onClick={() => void recentActivities.refetch()}>
            刷新
          </button>
        </div>
        {recentActivities.loading ? <p>加载中...</p> : null}
        {recentActivities.error ? (
          <>
            <p className="error-text">{recentActivities.error}</p>
            <button type="button" className="btn" onClick={() => void recentActivities.refetch()}>
              重试
            </button>
          </>
        ) : null}
        <ul className="plain-list activity-feed">
          {recentActivities.data.map((item) => (
            <li key={item.id} className="activity-feed-item">
              <div className="comment-row">
                <span>
                  <span className="pill">{item.log_type}</span> {item.run_title || `#${item.run_id}`} / {item.node_name || `#${item.run_node_id}`}
                </span>
                <span className="muted">{formatTime(item.created_at)}</span>
              </div>
              <div className="activity-operator-row">
                <span className="activity-operator">{item.operator_name || `#${item.operator_id}`}</span>
                <span className="muted">{item.operator_type}</span>
              </div>
              <p className="muted">{item.content}</p>
            </li>
          ))}
        </ul>
        {!recentActivities.loading && !recentActivities.error && !recentActivities.data.length ? <p className="muted">暂无最近动态</p> : null}
      </article>

      <article className="page-card dashboard-wide">
        <div className="page-title">
          <div>
            <span className="section-kicker">guided entry</span>
            <h2>快速入口</h2>
          </div>
        </div>
        <div className="template-shortcuts">
          <Link className="shortcut-card" to="/drafts/create">
            <span className="section-kicker">draft</span>
            <strong>龙虾生成草案</strong>
            <p>先用自然语言生成流程草案，再人工确认创建模板。</p>
          </Link>
          <Link className="shortcut-card" to="/drafts">
            <span className="section-kicker">drafts</span>
            <strong>管理历史草案</strong>
            <p>查看草案池里的待确认方案，继续编辑或直接发起流程。</p>
          </Link>
          <Link className="shortcut-card" to="/templates">
            <span className="section-kicker">template</span>
            <strong>已有模板发起</strong>
            <p>适合高频、已经固化的流程，直接进入模板库挑选。</p>
          </Link>
        </div>
      </article>
    </section>
  );
}

export default DashboardPage;
