import { useState } from "react";
import { useResource } from "../hooks/useResource.js";
import { ErrorMessage } from "./Feedback.jsx";

export function Onboarding({ api, onComplete }) {
  const [name, setName] = useState("");
  const [organizationMode, setOrganizationMode] = useState("new");
  const [organizationID, setOrganizationID] = useState("");
  const [roles, setRoles] = useState(["learner"]);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const organizations = useResource(api.organizations, []);

  const submit = async (event) => {
    event.preventDefault();
    setBusy(true);
    setError("");

    try {
      const selectedOrganizationID =
        organizationMode === "new"
          ? (await api.createOrganization(name)).id
          : organizationID;
      await api.setOrganization(selectedOrganizationID);

      for (const role of roles.filter((item) => item !== "learner")) {
        await api.grantRole(role);
      }
      await api.switchRole(
        roles.includes("methodist")
          ? "methodist"
          : roles.includes("curator")
            ? "curator"
            : "learner",
      );
      await onComplete();
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setBusy(false);
    }
  };

  const toggleRole = (role) => {
    setRoles((current) =>
      current.includes(role)
        ? current.filter((item) => item !== role || item === "learner")
        : [...current, role],
    );
  };

  return (
    <main className="onboarding">
      <form className="onboarding-card" onSubmit={submit}>
        <p className="eyebrow">Первый запуск</p>
        <h1>Соберите рабочее пространство</h1>
        <p>
          Организация объединяет людей, сценарии, назначения и историю
          переговоров.
        </p>

        <div className="switcher">
          <button
            type="button"
            className={organizationMode === "new" ? "selected" : ""}
            onClick={() => setOrganizationMode("new")}
          >
            Новая организация
          </button>
          <button
            type="button"
            className={organizationMode === "existing" ? "selected" : ""}
            onClick={() => setOrganizationMode("existing")}
          >
            Присоединиться
          </button>
        </div>

        {organizationMode === "new" ? (
          <label>
            Название организации
            <input
              required
              minLength="1"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Northwind"
            />
          </label>
        ) : (
          <label>
            Организация
            <select
              required
              value={organizationID}
              disabled={organizations.loading}
              onChange={(event) => setOrganizationID(event.target.value)}
            >
              <option value="">Выберите организацию</option>
              {organizations.data?.organizations?.map((organization) => (
                <option key={organization.id} value={organization.id}>
                  {organization.name}
                </option>
              ))}
            </select>
          </label>
        )}

        <div className="role-grid">
          {[
            ["learner", "Ученик", "Прохожу практику"],
            ["methodist", "Методист", "Создаю сценарии"],
            ["curator", "Куратор", "Назначаю Arena"],
          ].map(([role, title, detail]) => (
            <label className="role-choice" key={role}>
              <input
                type="checkbox"
                checked={roles.includes(role)}
                disabled={role === "learner"}
                onChange={() => toggleRole(role)}
              />
              <span>
                <strong>{title}</strong>
                <small>{detail}</small>
              </span>
            </label>
          ))}
        </div>

        <ErrorMessage>{organizations.error || error}</ErrorMessage>
        <button
          className="primary wide"
          disabled={
            busy || (organizationMode === "existing" && !organizationID)
          }
        >
          {busy ? "Создаём…" : "Продолжить"}
        </button>
      </form>
    </main>
  );
}
