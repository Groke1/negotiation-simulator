export const runtimeLabel = (value) => ({ deterministic: 'Подготовленный диалог', virtual_partner: 'Виртуальный партнёр', arena: 'Arena' }[value] || value || '—');
export const channelLabel = (value) => ({ chat: 'Чат', email: 'Почта', voice_message: 'Голосовое сообщение' }[value] || value || '—');
export const semanticActionLabel = (action) => action?.name || action?.code || action || 'Действие';
