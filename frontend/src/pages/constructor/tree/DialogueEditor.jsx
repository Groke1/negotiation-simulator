import { useMemo, useState } from 'react';
import { semanticActionLabel } from '../../../app/presentation';

const newId = () => crypto.randomUUID();

const emptyNode = (roleId = '') => ({
  id: newId(),
  incoming_message: { role_id: roleId, text: '', semantic_action_code: '' },
  reply_options: [],
});

const emptyReply = (roleId = '') => ({
  id: newId(),
  text: '',
  role_id: roleId,
  semantic_action_code: '',
  next_node_id: '',
  transition_id: '',
});

function SelectAction({ actions, value, onChange, disabled }) {
  return <select value={value || ''} disabled={disabled} onChange={(event) => onChange(event.target.value)}>
    <option value="">Не выбрано</option>
    {actions.map((action) => <option key={action.code} value={action.code}>{semanticActionLabel(action)}</option>)}
  </select>;
}

export default function DialogueEditor({ block, roles, actions, readOnly, onChange, onClose, learnerRoleId }) {
  const dialogue = block.deterministic_dialogue || { learner_role_id: '', start_node_id: '', nodes: [] };
  const [selectedId, setSelectedId] = useState(dialogue.start_node_id || dialogue.nodes[0]?.id || '');
  const selected = dialogue.nodes.find((node) => node.id === selectedId) || null;
  const learnerRole = learnerRoleId || dialogue.learner_role_id || '';
  const partnerRole = roles.find((role) => role.id !== learnerRole)?.id || '';
  const nodeNames = useMemo(() => new Map(dialogue.nodes.map((node, index) => [node.id, `Реплика ${index + 1}`])), [dialogue.nodes]);

  function update(next) { onChange({ ...block, deterministic_dialogue: next }); }
  function updateNode(nextNode) {
    update({ ...dialogue, nodes: dialogue.nodes.map((node) => node.id === nextNode.id ? nextNode : node) });
  }
  function addNode() {
    const node = emptyNode(partnerRole);
    update({ ...dialogue, learner_role_id: learnerRole, start_node_id: dialogue.start_node_id || node.id, nodes: [...dialogue.nodes, node] });
    setSelectedId(node.id);
  }
  function removeNode(nodeId) {
    const remaining = dialogue.nodes.filter((node) => node.id !== nodeId).map((node) => ({
      ...node,
      reply_options: node.reply_options.filter((reply) => reply.next_node_id !== nodeId),
    }));
    update({ ...dialogue, start_node_id: dialogue.start_node_id === nodeId ? (remaining[0]?.id || '') : dialogue.start_node_id, nodes: remaining });
    setSelectedId(remaining[0]?.id || '');
  }
  function updateReply(replyId, patch) {
    updateNode({ ...selected, reply_options: selected.reply_options.map((reply) => reply.id === replyId ? { ...reply, ...patch } : reply) });
  }

  return <div className="modal-overlay" role="dialog" aria-modal="true" aria-label="Редактор подготовленного диалога">
    <div className="modal dialogue-editor">
      <header className="dialogue-editor-header">
        <div><span className="eyebrow">Блок: {block.name}</span><h2>Подготовленный диалог</h2><p>Роль ученика: {roles.find((role) => role.id === learnerRole)?.name || 'не выбрана'}. Реплики собеседника и варианты ответа фиксируются автоматически.</p></div>
        <button className="secondary" type="button" onClick={onClose}>× Закрыть</button>
      </header>
      <div className="dialogue-editor-grid">
        <aside className="dialogue-outline">
          <div className="dialogue-outline-title"><strong>Структура</strong>{!readOnly && <button type="button" className="secondary" onClick={addNode}>+ Реплика</button>}</div>
          {dialogue.nodes.length ? dialogue.nodes.map((node, index) => <button type="button" key={node.id} className={node.id === selectedId ? 'dialogue-outline-node active' : 'dialogue-outline-node'} onClick={() => setSelectedId(node.id)}>
            <span>{node.id === dialogue.start_node_id ? 'Старт · ' : ''}Реплика {index + 1}</span>
            <small>{node.incoming_message?.text || 'Без текста'}</small>
          </button>) : <p className="muted">Диалог пока пуст.</p>}
        </aside>
        <section className="dialogue-form">
          {selected ? <>
            <div className="dialogue-form-heading"><h3>{nodeNames.get(selected.id)}</h3>{!readOnly && <div><label className="check-row"><input type="radio" name="dialogue-start" checked={dialogue.start_node_id === selected.id} onChange={() => update({ ...dialogue, start_node_id: selected.id })} /> Начало</label><button className="danger" type="button" onClick={() => removeNode(selected.id)}>Удалить</button></div>}</div>
            <label className="field">Кто говорит<select disabled value={partnerRole}><option>{roles.find((role) => role.id === partnerRole)?.name || 'Собеседник'}</option></select></label>
            <label className="field">Реплика собеседника<textarea disabled={readOnly} rows="4" value={selected.incoming_message?.text || ''} onChange={(event) => updateNode({ ...selected, incoming_message: { ...selected.incoming_message, text: event.target.value } })} /></label>
            <label className="field">Ожидаемое действие собеседника<SelectAction actions={actions} disabled={readOnly} value={selected.incoming_message?.semantic_action_code} onChange={(value) => updateNode({ ...selected, incoming_message: { ...selected.incoming_message, semantic_action_code: value } })} /></label>
            <div className="dialogue-replies"><div className="dialogue-outline-title"><h3>Варианты ответа пользователя</h3>{!readOnly && <button type="button" className="secondary" onClick={() => updateNode({ ...selected, reply_options: [...selected.reply_options, emptyReply(learnerRole)] })}>+ Вариант</button>}</div>
              {selected.reply_options.map((reply, index) => <article className="panel dialogue-reply" key={reply.id}>
                <div><strong>Вариант {index + 1}</strong>{!readOnly && <button className="text-button" type="button" onClick={() => updateNode({ ...selected, reply_options: selected.reply_options.filter((item) => item.id !== reply.id) })}>Удалить</button>}</div>
                <label className="field">Текст<textarea disabled={readOnly} rows="2" value={reply.text || ''} onChange={(event) => updateReply(reply.id, { text: event.target.value })} /></label>
                <label className="field">Роль ученика<select disabled value={learnerRole}><option>{roles.find((role) => role.id === learnerRole)?.name || 'Роль ученика'}</option></select></label>
                <label className="field">Семантическое действие<SelectAction actions={actions} disabled={readOnly} value={reply.semantic_action_code} onChange={(value) => updateReply(reply.id, { semantic_action_code: value })} /></label>
                <label className="field">Куда ведёт<select disabled={readOnly} value={reply.next_node_id ? `node:${reply.next_node_id}` : reply.transition_id ? `transition:${reply.transition_id}` : ''} onChange={(event) => { const [kind, value] = event.target.value.split(':'); updateReply(reply.id, kind === 'node' ? { next_node_id: value, transition_id: '' } : kind === 'transition' ? { next_node_id: '', transition_id: value } : { next_node_id: '', transition_id: '' }); }}><option value="">Выберите назначение</option><optgroup label="Внутри этого блока">{dialogue.nodes.filter((node) => node.id !== selected.id).map((node) => <option key={node.id} value={`node:${node.id}`}>{nodeNames.get(node.id)}</option>)}</optgroup><optgroup label="Переход в другой блок">{(block.transitions || []).map((transition) => <option key={transition.id} value={`transition:${transition.id}`}>{transition.condition || 'Без условия'}</option>)}</optgroup></select></label>
              </article>)}
              {!selected.reply_options.length && <p className="muted">Добавьте хотя бы один вариант ответа.</p>}
            </div>
          </> : <p className="muted">Добавьте первую реплику, чтобы собрать дерево.</p>}
        </section>
      </div>
    </div>
  </div>;
}
