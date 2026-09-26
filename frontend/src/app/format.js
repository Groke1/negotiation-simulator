export const formatNumber = (value) => {
  const number = Number(value);
  if (!Number.isFinite(number)) return '—';
  const rounded = Math.round(number * 10) / 10;
  return Number.isInteger(rounded) ? String(rounded) : rounded.toFixed(1);
};

export const formatDelta = (value) => {
  const number = Number(value);
  if (!Number.isFinite(number)) return '—';
  return `${number > 0 ? '+' : ''}${formatNumber(number)}`;
};
