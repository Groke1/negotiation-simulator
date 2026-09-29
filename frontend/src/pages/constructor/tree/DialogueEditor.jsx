import { useState } from 'react';
import { semanticActionLabel } from '../../../app/presentation';
import { createId } from '../../../app/id';
import './DialogueEditor.css';

const newId = createId;
const WEIGHT_OPTIONS = [[0.5, 'Низкая значимость'], [1, 'Обычная значимость'], [1.5, 'Высокая значимость']];
const LEVEL_OPTIONS = [[0, '0 · Контрпродуктивно'], [1, '1 · Слабо'], [2, '2 · Базово'], [3, '3 · Эффективно'], [4, '4 · Сильное решение']];

const emptyNode = (roleId = '') => ({
  id: newId(),
  incoming_message: { role_id: roleId, text: '', semantic_action_code: '' },
  assessment_targets: [],
  reply_options: [],
});
const assessmentsFor = (targets, current = []) => targets.map((target) => ({
  cluster_id: target.cluster_id,
  level: current.find((assessment) => assessment.cluster_id === target.cluster_id)?.level ?? 2,
  comment: current.find((assessment) => assessment.cluster_id === target.cluster_id)?.comment || '',
}));
const emptyReply = (roleId = '', targets = []) => ({
  id: newId(),
  text: '',
  role_id: roleId,
  semantic_action_code: '',
  assessments: assessmentsFor(targets),
  next_node_id: '',
  transition_id: '',
});

function SelectAction({ actions, value, onChange, disabled }) {
  return <select value={value || ''} disabled={disabled} onChange={(event) => onChange(event.target.value)}>
    <option value="">Не выбрано</option>
    {actions.map((action) => <option key={action.code} value={action.code}>{semanticActionLabel(action)}</option>)}
  </select>;
}

function AssessmentTargets({ clusters, targets, onChange, disabled }) {
  const toggle = (clusterId) => {
    const exists = targets.some((target) => target.cluster_id === clusterId);
    onChange(exists ? targets.filter((target) => target.cluster_id !== clusterId) : [...targets, { cluster_id: clusterId, weight: 1 }]);
  };
  const setWeight = (clusterId, weight) => onChange(targets.map((target) => target.cluster_id === clusterId ? { ...target, weight: Number(weight) } : target));
  return <section className="assessment-targets">
    <div><h3>Оцениваемые компетенции</h3><p className="muted">Укажите, какие кластеры проверяет именно этот выбор. Это не связано с семантической меткой реплики.</p></div>
    {clusters.length ? clusters.map((cluster) => {
      const target = targets.find((item) => item.cluster_id === cluster.id);
      return <div className={`assessment-target ${target ? 'selected' : ''}`} key={cluster.id}>
        <button className="assessment-target-toggle" type="button" disabled={disabled} aria-pressed={Boolean(target)} onClick={() => toggle(cluster.id)}><span>{cluster.name}</span><small>{target ? 'Оценивается' : 'Не оценивается'}</small></button>
        {target && <label className="field">Значимость<select disabled={disabled} value={target.weight} onChange={(event) => setWeight(cluster.id, event.target.value)}>{WEIGHT_OPTIONS.map(([weight, label]) => <option key={weight} value={weight}>{label}</option>)}</select></label>}
      </div>;
    }) : <p className="muted">Сначала отметьте кластеры развития в настройках сценария.</p>}
  </section>;
}

function ReplyAssessments({ reply, targets, clusters, onChange, disabled }) {
  if (!targets.length) return null;
  return <section className="assessment-list"><h4>Методическая оценка варианта</h4>{targets.map((target) => {
    const assessment = reply.assessments?.find((item) => item.cluster_id === target.cluster_id) || { cluster_id: target.cluster_id, level: 2, comment: '' };
    const cluster = clusters.find((item) => item.id === target.cluster_id);
    return <div className="assessment-reply" key={target.cluster_id}><h4>{cluster?.name || 'Компетенция'}</h4><div className="assessment-reply-grid"><label className="field">Качество решения<select disabled={disabled} value={assessment.level} onChange={(event) => onChange(target.cluster_id, { level: Number(event.target.value) })}>{LEVEL_OPTIONS.map(([level, label]) => <option key={level} value={level}>{label}</option>)}</select></label><label className="field">Комментарий для разбора <textarea disabled={disabled} rows="2" value={assessment.comment || ''} placeholder="Поясните, почему этот вариант сильный или слабый" onChange={(event) => onChange(target.cluster_id, { comment: event.target.value })} /></label></div></div>;
  })}</section>;
}

function ReplyDestination({ reply, nodes, transitions, nodeNames, blockNames, selectedNodeId, disabled, onChange }) {
  const internalNodes = nodes.filter((node) => node.id !== selectedNodeId);
  const savedDestinationType = reply.next_node_id ? 'node' : reply.transition_id ? 'transition' : '';
  const [destinationType, setDestinationType] = useState(savedDestinationType);
  const selectType = (type) => {
    setDestinationType(type);
    if (type !== savedDestinationType) onChange({ next_node_id: '', transition_id: '' });
  };
  const setDestination = (value) => onChange(destinationType === 'node' ? { next_node_id: value, transition_id: '' } : { next_node_id: '', transition_id: value });
  return <section className="reply-destination">
    <span className="reply-destination-label">Куда ведёт ответ</span>
    <div className="reply-destination-types" role="group" aria-label="Тип назначения">
      <button type="button" className={destinationType === 'node' ? 'selected' : ''} disabled={disabled || !internalNodes.length} onClick={() => selectType('node')}><span>Внутри блока</span><small>Следующая реплика</small></button>
      <button type="button" className={destinationType === 'transition' ? 'selected' : ''} disabled={disabled || !transitions.length} onClick={() => selectType('transition')}><span>В другой блок</span><small>По переходу сценария</small></button>
    </div>
    {destinationType && <label className="field reply-destination-select"><span>{destinationType === 'node' ? 'Следующая реплика' : 'Следующий блок'}</span><select disabled={disabled} value={destinationType === 'node' ? reply.next_node_id || '' : reply.transition_id || ''} onChange={(event) => setDestination(event.target.value)}><option value="">Выберите назначение</option>{destinationType === 'node' ? internalNodes.map((node) => <option key={node.id} value={node.id}>{nodeNames.get(node.id)}</option>) : transitions.map((transition) => <option key={transition.id} value={transition.id}>{blockNames.get(transition.target_block_id) || 'Блок не найден'}</option>)}</select></label>}
    {!destinationType && !internalNodes.length && !transitions.length && <small className="muted">Добавьте ещё реплику или переход сценария, чтобы указать продолжение.</small>}
  </section>;
}

export default function DialogueEditor({ block, blocks = [], roles, actions, clusters = [], readOnly, onChange, onClose, learnerRoleId }) {
  const dialogue = {
    ...(block?.deterministic_dialogue || {}),
    learner_role_id: block?.deterministic_dialogue?.learner_role_id || '',
    start_node_id: block?.deterministic_dialogue?.start_node_id || '',
    nodes: Array.isArray(block?.deterministic_dialogue?.nodes) ? block.deterministic_dialogue.nodes : [],
  };
  const availableRoles = Array.isArray(roles) ? roles : [];
  const availableActions = Array.isArray(actions) ? actions : [];
  const [selectedId, setSelectedId] = useState(dialogue.start_node_id || dialogue.nodes[0]?.id || '');
  const selected = dialogue.nodes.find((node) => node.id === selectedId) || null;
  const learnerRole = learnerRoleId || dialogue.learner_role_id || '';
  const partnerRole = availableRoles.find((role) => role.id !== learnerRole)?.id || '';
  const nodeNames = new Map(dialogue.nodes.map((node, index) => [node.id, `Реплика ${index + 1}`]));
  const blockNames = new Map((Array.isArray(blocks) ? blocks : []).map((item) => [item.id, item.name || 'Без названия']));
  const targetClusters = clusters;
  const targets = selected?.assessment_targets || [];

  function update(next) { onChange({ ...block, deterministic_dialogue: next }); }
  function updateNode(nextNode) { update({ ...dialogue, nodes: dialogue.nodes.map((node) => node.id === nextNode.id ? nextNode : node) }); }
  function addNode() {
    const node = emptyNode(partnerRole);
    update({ ...dialogue, learner_role_id: learnerRole, start_node_id: dialogue.start_node_id || node.id, nodes: [...dialogue.nodes, node] });
    setSelectedId(node.id);
  }
  function removeNode(nodeId) {
    const remaining = dialogue.nodes.filter((node) => node.id !== nodeId).map((node) => ({ ...node, reply_options: node.reply_options.filter((reply) => reply.next_node_id !== nodeId) }));
    update({ ...dialogue, start_node_id: dialogue.start_node_id === nodeId ? (remaining[0]?.id || '') : dialogue.start_node_id, nodes: remaining });
    setSelectedId(remaining[0]?.id || '');
  }
  function updateReply(replyId, patch) { updateNode({ ...selected, reply_options: selected.reply_options.map((reply) => reply.id === replyId ? { ...reply, ...patch } : reply) }); }
  function updateTargets(nextTargets) {
    updateNode({ ...selected, assessment_targets: nextTargets, reply_options: selected.reply_options.map((reply) => ({ ...reply, assessments: assessmentsFor(nextTargets, reply.assessments) })) });
  }
  function updateAssessment(replyId, clusterId, patch) {
    const reply = selected.reply_options.find((item) => item.id === replyId);
    const current = reply.assessments?.find((item) => item.cluster_id === clusterId) || { cluster_id: clusterId, level: 2, comment: '' };
    updateReply(replyId, { assessments: assessmentsFor(targets, reply.assessments).map((item) => item.cluster_id === clusterId ? { ...current, ...patch } : item) });
  }

  return <div className="modal-overlay" role="dialog" aria-modal="true" aria-label="Редактор подготовленного диалога">
    <div className="modal dialogue-editor">
      <header className="dialogue-editor-header"><div><span className="eyebrow">Блок: {block.name}</span><h2>Подготовленный диалог</h2><p>Роль ученика: {availableRoles.find((role) => role.id === learnerRole)?.name || 'не выбрана'}. Качество ответа и его семантическая метка задаются независимо.</p></div><button className="secondary" type="button" onClick={onClose}>× Закрыть</button></header>
      <div className="dialogue-editor-grid">
        <aside className="dialogue-outline"><div className="dialogue-outline-title"><strong>Структура</strong>{!readOnly && <button type="button" className="secondary" onClick={addNode}>+ Реплика</button>}</div>{dialogue.nodes.length ? dialogue.nodes.map((node, index) => <button type="button" key={node.id} className={node.id === selectedId ? 'dialogue-outline-node active' : 'dialogue-outline-node'} onClick={() => setSelectedId(node.id)}><span>{node.id === dialogue.start_node_id ? 'Старт · ' : ''}Реплика {index + 1}</span><small>{node.incoming_message?.text || 'Без текста'}</small></button>) : <p className="muted">Диалог пока пуст.</p>}</aside>
        <section className="dialogue-form">{selected ? <>
          <div className="dialogue-form-heading"><h3>{nodeNames.get(selected.id)}</h3>{!readOnly && <div><label className="check-row"><input type="radio" name="dialogue-start" checked={dialogue.start_node_id === selected.id} onChange={() => update({ ...dialogue, start_node_id: selected.id })} /> Начало</label><button className="danger" type="button" onClick={() => removeNode(selected.id)}>Удалить</button></div>}</div>
          <label className="field">Кто говорит<select disabled value={partnerRole}><option>{availableRoles.find((role) => role.id === partnerRole)?.name || 'Собеседник'}</option></select></label>
          <label className="field">Реплика собеседника<textarea disabled={readOnly} rows="4" value={selected.incoming_message?.text || ''} onChange={(event) => updateNode({ ...selected, incoming_message: { ...selected.incoming_message, text: event.target.value } })} /></label>
          <label className="field">Семантическое действие собеседника<SelectAction actions={availableActions} disabled={readOnly} value={selected.incoming_message?.semantic_action_code} onChange={(value) => updateNode({ ...selected, incoming_message: { ...selected.incoming_message, semantic_action_code: value } })} /></label>
          <AssessmentTargets clusters={targetClusters} targets={targets} onChange={updateTargets} disabled={readOnly} />
          <div className="dialogue-replies"><div className="dialogue-outline-title"><h3>Варианты ответа ученика</h3>{!readOnly && <button type="button" className="secondary" onClick={() => updateNode({ ...selected, reply_options: [...selected.reply_options, emptyReply(learnerRole, targets)] })}>+ Вариант</button>}</div>
            {(selected.reply_options || []).map((reply, index) => <article className="panel dialogue-reply" key={reply.id}><div><strong>Вариант {index + 1}</strong>{!readOnly && <button className="text-button" type="button" onClick={() => updateNode({ ...selected, reply_options: selected.reply_options.filter((item) => item.id !== reply.id) })}>Удалить</button>}</div><label className="field">Текст<textarea disabled={readOnly} rows="2" value={reply.text || ''} onChange={(event) => updateReply(reply.id, { text: event.target.value })} /></label><label className="field">Роль ученика<select disabled value={learnerRole}><option>{availableRoles.find((role) => role.id === learnerRole)?.name || 'Роль ученика'}</option></select></label><label className="field">Семантическое действие<SelectAction actions={availableActions} disabled={readOnly} value={reply.semantic_action_code} onChange={(value) => updateReply(reply.id, { semantic_action_code: value })} /></label><ReplyAssessments reply={reply} targets={targets} clusters={targetClusters} disabled={readOnly} onChange={(clusterId, patch) => updateAssessment(reply.id, clusterId, patch)} /><ReplyDestination reply={reply} nodes={dialogue.nodes} transitions={block.transitions || []} nodeNames={nodeNames} blockNames={blockNames} selectedNodeId={selected.id} disabled={readOnly} onChange={(patch) => updateReply(reply.id, patch)} /></article>)}
            {!selected.reply_options.length && <p className="muted">Добавьте хотя бы один вариант ответа.</p>}
          </div>
        </> : <p className="muted">Добавьте первую реплику, чтобы собрать дерево.</p>}</section>
      </div>
    </div>
  </div>;
}
