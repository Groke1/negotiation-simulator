const annotationCodes = (annotations) => (Array.isArray(annotations) ? annotations : []).map((item) => item?.semantic_action_code).filter(Boolean);

function SemanticCodes({ codes, actionNames }) {
  const uniqueCodes = [...new Set(codes.filter(Boolean))];
  if (!uniqueCodes.length) return null;
  if (uniqueCodes.length === 1) return <strong>{actionNames.get(uniqueCodes[0]) || uniqueCodes[0]}</strong>;
  return <div className="semantic-code-list">{uniqueCodes.map((code) => <span key={code}>{actionNames.get(code) || code}</span>)}</div>;
}

/** Shows only the current user's annotations and classification result. */
export default function ArenaMessageSemanticReport({ message, mine, actionNames }) {
  const ownCodes = [message?.semantic_action_code, ...annotationCodes(message?.semantic_annotations)];
  const classification = message?.classification;
  const selectedCodes = [classification?.selected_semantic_action_code, ...annotationCodes(classification?.selected_semantic_annotations)];
  const expectedCodes = [classification?.expected_semantic_action_code, ...annotationCodes(classification?.expected_semantic_annotations)];
  const showAuthoredMark = mine;

  if (!classification && (!showAuthoredMark || !ownCodes.some(Boolean))) return null;
  return <div className="message-semantic-report partner-classification-result">
    {showAuthoredMark && ownCodes.some(Boolean) && <p><span>Ваш вариант:</span> <SemanticCodes codes={ownCodes} actionNames={actionNames} /></p>}
    {!mine && classification && <><p><span>Ваш вариант:</span> <SemanticCodes codes={selectedCodes} actionNames={actionNames} /></p><p className={classification.correct ? 'feedback good' : 'feedback bad'}>{classification.correct ? '✓ Действие распознано верно' : <>Не совпало. Правильная метка: <SemanticCodes codes={expectedCodes} actionNames={actionNames} /></>}</p></>}
  </div>;
}
