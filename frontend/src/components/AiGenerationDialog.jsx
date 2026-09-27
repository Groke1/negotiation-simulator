export default function AiGenerationDialog() {
  return <div className="modal-overlay" role="status" aria-live="polite" aria-label="Идёт генерация">
    <section className="modal generation-dialog">
      <div className="generation-orbit" aria-hidden="true"><div className="generation-core">✦</div></div>
      <h2>Генерация может занять несколько минут</h2>
      <p>ИИ готовит предложение. Не закрывайте страницу: результат появится в полях редактора, и вы сможете проверить его перед сохранением.</p>
    </section>
  </div>;
}
