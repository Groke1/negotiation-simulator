const annotationCodes = (annotations) => (Array.isArray(annotations) ? annotations : []).map((item) => item?.semantic_action_code).filter(Boolean);

function SemanticCodes({ codes, actionNames }) {
  const uniqueCodes = [...new Set(codes.filter(Boolean))];
  if (!uniqueCodes.length) return null;
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
    {showAuthoredMark && ownCodes.some(Boolean) && <><span>Метка сообщения:</span><SemanticCodes codes={ownCodes} actionNames={actionNames} /></>}
    {!mine && classification && <><p><span>Ваш вариант:</span></p><SemanticCodes codes={selectedCodes} actionNames={actionNames} /><div className={classification.correct ? 'feedback good' : 'feedback bad'}>{classification.correct ? '✓ Действие распознано верно' : <><span>Не совпало. Правильная метка:</span><SemanticCodes codes={expectedCodes} actionNames={actionNames} /></>}</div></>}
  </div>;
}
