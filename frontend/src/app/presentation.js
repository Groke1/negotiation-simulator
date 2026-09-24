export const runtimeLabel = (value) => ({ deterministic: 'Подготовленный диалог', virtual_partner: 'Виртуальный партнёр', arena: 'Arena' }[value] || value || '—');
export const channelLabel = (value) => ({ chat: 'Чат', email: 'Почта', voice_message: 'Голосовое сообщение' }[value] || value || '—');
const semanticLabels = { ASK_QUESTION: 'Задать вопрос', MAKE_PROPOSAL: 'Сделать предложение', MAKE_CONCESSION: 'Сделать уступку', REQUEST_RECIPROCITY: 'Запросить встречный шаг', HOLD_BOUNDARY: 'Обозначить границу', CHANGE_TOPIC: 'Сменить тему', PAUSE: 'Взять паузу', SUMMARIZE: 'Подвести итог', CLOSE_NEGOTIATION: 'Завершить переговоры' };
export const semanticActionLabel = (action) => semanticLabels[action?.code || action] || action?.name || action?.code || 'Действие';
