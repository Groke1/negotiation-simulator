export function ErrorMessage({ children }) {
  return children ? <p className="error-message">{children}</p> : null;
}
export function EmptyState({ title, detail }) {
  return (
    <div className="empty-state">
      <strong>{title}</strong>
      <span>{detail}</span>
    </div>
  );
}
export function Loading() {
  return <div className="loading">Загружаем данные…</div>;
}
