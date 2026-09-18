const labels = { learner: "Ученик", methodist: "Методист", curator: "Куратор" };

export function AppShell({ user, active, onRole, onLogout, children }) {
  return (
    <main className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-mark">N</span>
          <span>Negotiator</span>
        </div>
        <div className="workspace-name">
          {user.organization?.name || "Без организации"}
        </div>
        <nav>
          {user.roles.includes("methodist") && (
            <button
              className={active === "methodist" ? "nav-active" : ""}
              onClick={() => onRole("methodist")}
            >
              Сценарии
            </button>
          )}
          {user.roles.includes("curator") && (
            <button
              className={active === "curator" ? "nav-active" : ""}
              onClick={() => onRole("curator")}
            >
              Команда
            </button>
          )}
          <button
            className={active === "learner" ? "nav-active" : ""}
            onClick={() => onRole("learner")}
          >
            Практика
          </button>
        </nav>
        <div className="account">
          <span className="avatar">
            {user.username.slice(0, 2).toUpperCase()}
          </span>
          <div>
            <strong>{user.username}</strong>
            <small>{labels[user.active_role]}</small>
          </div>
          <button className="logout" onClick={onLogout} title="Выйти">
            ↗
          </button>
        </div>
      </aside>
      <section className="workspace">{children}</section>
    </main>
  );
}
