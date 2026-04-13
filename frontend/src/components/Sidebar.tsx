import { NavLink } from "react-router-dom";

function Sidebar() {
  return (
    <aside className="app-sidebar">
      <section className="menu-group">
        <h4>工作台</h4>
        <NavLink to="/" end>
          首页工作台
        </NavLink>
      </section>

      <section className="menu-group">
        <h4>流程中心</h4>
        <NavLink to="/runs">全部流程</NavLink>
        <NavLink to="/runs/mine">我发起的</NavLink>
        <NavLink to="/runs/todo">待我处理</NavLink>
      </section>

      <section className="menu-group">
        <h4>模板中心</h4>
        <NavLink to="/templates">模板列表</NavLink>
      </section>

      <section className="menu-group">
        <h4>交付中心</h4>
        <NavLink to="/deliverables">全部交付</NavLink>
        <NavLink to="/deliverables?review_status=pending">待验收</NavLink>
        <NavLink to="/deliverables?review_status=approved">已归档</NavLink>
      </section>

      <section className="menu-group">
        <h4>资源中心</h4>
        <NavLink to="/resources/persons">人员管理</NavLink>
        <NavLink to="/resources/agents">龙虾管理</NavLink>
      </section>
    </aside>
  );
}

export default Sidebar;
