const timestamp = (message) => {
  const value = message?.created_at || message?.createdAt || message?.timestamp;
  const parsed = value ? Date.parse(value) : Number.NaN;
  return Number.isFinite(parsed) ? parsed : 0;
};

export const sortMessagesChronologically = (messages = []) => [...messages]
  .map((message, index) => ({ message, index }))
  // A backend response is already ordered by its stable creation order. UUIDs are
  // deliberately random, so using one as a tie-breaker scrambles messages created
  // in the same timestamp tick. Keep the server order for equal timestamps.
  .sort((left, right) => timestamp(left.message) - timestamp(right.message) || left.index - right.index)
  .map(({ message }) => message);
