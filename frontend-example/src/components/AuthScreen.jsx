import { useState } from "react";
import { ErrorMessage } from "./Feedback.jsx";

export function AuthScreen({ onSubmit }) {
  const [mode, setMode] = useState("login");
  const [form, setForm] = useState({ username: "", email: "", password: "" });
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const submit = async (event) => {
    event.preventDefault();
    setBusy(true);
    setError("");
    try {
      await onSubmit(mode, form);
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  };
  return (
    <main className="auth-layout">
      <section className="auth-copy">
        <div className="brand">
          <span className="brand-mark">N</span>
          <span>Negotiator</span>
        </div>
        <p className="eyebrow">Практика переговоров</p>
        <h1>Репетируйте важные разговоры до того, как они случатся.</h1>
        <p>
          Индивидуальные сценарии, живая Арена и видимый прогресс одной команды.
        </p>
      </section>
      <section className="auth-card">
        <div className="switcher">
          <button
            className={mode === "login" ? "selected" : ""}
            onClick={() => setMode("login")}
          >
            Войти
          </button>
          <button
            className={mode === "register" ? "selected" : ""}
            onClick={() => setMode("register")}
          >
            Регистрация
          </button>
        </div>
        <form onSubmit={submit}>
          <h2>{mode === "login" ? "С возвращением" : "Создайте аккаунт"}</h2>
          {mode === "register" && (
            <label>
              Имя пользователя
              <input
                required
                minLength="3"
                value={form.username}
                onChange={(event) =>
                  setForm({ ...form, username: event.target.value })
                }
              />
            </label>
          )}
          <label>
            Почта
            <input
              type="email"
              required
              value={form.email}
              onChange={(event) =>
                setForm({ ...form, email: event.target.value })
              }
            />
          </label>
          <label>
            Пароль
            <input
              type="password"
              required
              minLength="8"
              value={form.password}
              onChange={(event) =>
                setForm({ ...form, password: event.target.value })
              }
            />
          </label>
          <ErrorMessage>{error}</ErrorMessage>
          <button className="primary wide" disabled={busy}>
            {busy
              ? "Подождите…"
              : mode === "login"
                ? "Войти"
                : "Зарегистрироваться"}
          </button>
        </form>
      </section>
    </main>
  );
}
