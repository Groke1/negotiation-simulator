export const patterns = [
  { id: 'ASK_QUESTION', label: 'Уточнить', hint: 'Задать вопрос' },
  { id: 'MAKE_PROPOSAL', label: 'Предложить', hint: 'Внести вариант решения' },
  { id: 'MAKE_CONCESSION', label: 'Уступить', hint: 'Согласиться на что-то' },
  { id: 'REQUEST_RECIPROCITY', label: 'Запросить взаимность', hint: 'Попросить что-то взамен' },
  { id: 'HOLD_BOUNDARY', label: 'Удержать границу', hint: 'Не идти на уступку' },
  { id: 'PAUSE', label: 'Взять паузу', hint: 'Отложить решение' },
  { id: 'SUMMARIZE', label: 'Резюмировать', hint: 'Зафиксировать договорённости' }
];

const markers = {
  ASK_QUESTION: ['что для вас', 'что важно', 'уточните', 'расскажите', 'почему', 'как вы'],
  MAKE_PROPOSAL: ['предлагаю', 'давайте', 'можем предложить', 'вариант'],
  MAKE_CONCESSION: ['согласны', 'уступим', 'готовы снизить', 'идём навстречу'],
  REQUEST_RECIPROCITY: ['а что вы', 'взамен', 'с вашей стороны', 'если вы'],
  HOLD_BOUNDARY: ['не можем', 'это окончательно', 'наше условие', 'ниже не пойдём'],
  PAUSE: ['пауза', 'завтра', 'вернёмся', 'нужно время'],
  SUMMARIZE: ['итак', 'резюмирую', 'договорились', 'фиксируем']
};

export function detectPattern(text) {
  const lower = text.toLowerCase();
  for (const [pattern, words] of Object.entries(markers)) {
    if (words.some(w => lower.includes(w))) return pattern;
  }
  return null;
}

export function getPatternLabel(id) {
  const p = patterns.find(x => x.id === id);
  return p ? p.label : id;
}