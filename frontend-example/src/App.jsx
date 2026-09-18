import { useCallback, useEffect, useState } from "react";
import { api } from "./api/client.js";
import { AppShell } from "./components/AppShell.jsx";
import { AuthScreen } from "./components/AuthScreen.jsx";
import { Loading } from "./components/Feedback.jsx";
import { Onboarding } from "./components/Onboarding.jsx";
import { CuratorWorkspace } from "./features/curator/CuratorWorkspace.jsx";
import { LearnerWorkspace } from "./features/learner/LearnerWorkspace.jsx";
import { MethodistWorkspace } from "./features/methodist/MethodistWorkspace.jsx";

export default function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(Boolean(api.token()));

  const loadCurrentUser = useCallback(async () => {
    if (!api.token()) {
      setUser(null);
      return;
    }

    try {
      setUser(await api.me());
    } catch {
      api.clearToken();
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadCurrentUser();
  }, [loadCurrentUser]);

  const authenticate = async (mode, form) => {
    if (mode === "login") {
      await api.login({ email: form.email, password: form.password });
    } else {
      await api.register(form);
    }
    setLoading(true);
    await loadCurrentUser();
  };

  const switchRole = async (role) => {
    await api.switchRole(role);
    setLoading(true);
    await loadCurrentUser();
  };

  const logout = async () => {
    await api.logout();
    setUser(null);
  };

  if (loading) {
    return (
      <main className="app-loading">
        <Loading />
      </main>
    );
  }

  if (!user) {
    return <AuthScreen onSubmit={authenticate} />;
  }

  if (!user.organization) {
    return <Onboarding api={api} onComplete={loadCurrentUser} />;
  }

  return (
    <AppShell
      user={user}
      active={user.active_role}
      onRole={switchRole}
      onLogout={logout}
    >
      {user.active_role === "methodist" && <MethodistWorkspace />}
      {user.active_role === "curator" && <CuratorWorkspace />}
      {user.active_role === "learner" && <LearnerWorkspace user={user} />}
    </AppShell>
  );
}
