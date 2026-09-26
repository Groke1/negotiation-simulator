const timestamp = (message) => {
  const value = message?.created_at || message?.createdAt || message?.timestamp;
  const parsed = value ? Date.parse(value) : Number.NaN;
  return Number.isFinite(parsed) ? parsed : 0;
};

const identifier = (message) => String(message?.message_id || message?.id || '');

export const sortMessagesChronologically = (messages = []) => [...messages]
  .map((message, index) => ({ message, index }))
  .sort((left, right) => timestamp(left.message) - timestamp(right.message) || identifier(left.message).localeCompare(identifier(right.message)) || left.index - right.index)
  .map(({ message }) => message);
