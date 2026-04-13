import { Navigate, NavLink, Route, Routes } from "react-router-dom";
import PersonListPage from "./pages/PersonListPage";
import AgentListPage from "./pages/AgentListPage";
import TemplateListPage from "./pages/TemplateListPage";
import TemplateDetailPage from "./pages/TemplateDetailPage";
import RunStartPage from "./pages/RunStartPage";
import RunDetailPage from "./pages/RunDetailPage";
import RunListPage from "./pages/RunListPage";
import DeliverableListPage from "./pages/DeliverableListPage";
import DeliverableDetailPage from "./pages/DeliverableDetailPage";
import DashboardPage from "./pages/DashboardPage";
import DraftCreatePage from "./pages/DraftCreatePage";
import DraftConfirmPage from "./pages/DraftConfirmPage";
import DraftListPage from "./pages/DraftListPage";

function App() {
  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">
        跳到主内容
      </a>
      <header className="app-header">
        <div className="app-header-inner">
          <div className="app-header-top">
            <div className="brand-block">
              <span className="brand-kicker">ai-native workflow console</span>
              <h1>虾线工作台</h1>
              <p>面向流程协作、节点执行和交付收口的一体化控制台。</p>
            </div>
          </div>
        </div>
      </header>
      <div className="app-body">
        <div className="app-layout">
          <aside className="app-sidebar" aria-label="主导航区域">
            <div className="sidebar-panel">
              <span className="section-kicker">控制面板</span>
              <h2>导航</h2>
              <p>在工作台、流程、模板和交付模块之间切换。</p>
              <nav className="app-nav app-side-nav" aria-label="主导航">
                <NavLink to="/">工作台</NavLink>
                <NavLink to="/drafts">流程草案</NavLink>
                <NavLink to="/templates">模板</NavLink>
                <NavLink to="/runs">流程</NavLink>
              </nav>
            </div>
          </aside>
          <main className="app-main" id="main-content">
            <Routes>
              <Route path="/" element={<DashboardPage />} />
              <Route path="/drafts" element={<DraftListPage />} />
              <Route path="/drafts/create" element={<DraftCreatePage />} />
              <Route path="/drafts/:id/confirm" element={<DraftConfirmPage />} />
              <Route path="/runs" element={<RunListPage scope="all" title="全部流程" />} />
              <Route path="/runs/mine" element={<RunListPage scope="initiated_by_me" title="我发起的流程" />} />
              <Route path="/runs/todo" element={<RunListPage scope="todo" title="待我处理流程" />} />
              <Route path="/templates" element={<TemplateListPage />} />
              <Route path="/templates/:templateId" element={<TemplateDetailPage />} />
              <Route path="/templates/:templateId/start" element={<RunStartPage />} />
              <Route path="/runs/:runId" element={<RunDetailPage />} />
              <Route path="/deliverables" element={<DeliverableListPage />} />
              <Route path="/deliverables/:deliverableId" element={<DeliverableDetailPage />} />
              <Route path="/resources/persons" element={<PersonListPage />} />
              <Route path="/resources/agents" element={<AgentListPage />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </main>
        </div>
      </div>
    </div>
  );
}

export default App;
