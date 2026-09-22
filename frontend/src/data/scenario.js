export const scenarios = [
  {
    scenario_id: 'sc-contract-20',
    title: 'Продление контракта — скидка 20%',
    description: 'Клиент требует скидку 20%, ссылаясь на конкурента. Нужно удержать маржу и не потерять клиента.',
    expanded_context: 'Клиент — крупная компания, у которой есть альтернативный поставщик с более низкой ценой. Менеджер может дать максимум 10% скидки, но взамен может предложить расширенный SLA или более длинный срок контракта.',
    status: 'draft',
    difficulty: 'medium',
    tags: ['B2B', 'продажи', 'цена'],
    supported_runtimes: ['deterministic'],
    supported_channels: ['chat'],
    target_cluster_ids: [],
    roles: [
      {
        id: 'role-learner',
        name: 'Менеджер по продажам',
        description: 'Продаёт услугу, отвечает за условия',
        goal: 'Продлить контракт с минимальной скидкой',
        interests: 'Сохранить маржу',
        constraints: 'Максимум 10% скидки',
        private_information: 'Готов предложить расширенный SLA вместо скидки'
      },
      {
        id: 'role-partner',
        name: 'Клиент',
        description: 'Закупщик крупной компании',
        goal: 'Получить скидку 20%',
        interests: 'Снизить издержки',
        constraints: 'Есть альтернативный поставщик',
        private_information: 'Конкурент без расширенной поддержки'
      }
    ],
    blocks: [
      // ============================================
      // БЛОК 1: НАЧАЛО
      // ============================================
      {
        id: 'block-start',
        type: 'start',
        name: 'Начало',
        description: 'Старт переговоров',
        expanded_context: '',
        lead_role_id: null,
        lead_probability: 50,
        role_goals: [],
        behavior: '',
        action_rules: [],
        constraints: [],
        transitions: [
          {
            id: 'tr-start-to-price',
            target_block_id: 'block-price',
            condition: 'Клиент начинает разговор с требования скидки',
            weight: 100,
            forbidden: false
          }
        ]
      },

      // ============================================
      // БЛОК 2: ОБСУЖДЕНИЕ ЦЕНЫ — большой диалог
      // ============================================
      {
        id: 'block-price',
        type: 'regular',
        name: 'Обсуждение цены',
        description: 'Клиент требует скидку 20%',
        expanded_context: 'Клиент давит на цену, ссылаясь на конкурента. Важно понять, что за этим стоит — реальная цена или желание долгосрочных условий.',
        lead_role_id: 'role-partner',
        lead_probability: 70,
        role_goals: [],
        behavior: 'Давит, не уступает сразу',
        action_rules: [],
        constraints: ['Не соглашаться на скидку без выяснения причин'],
        position: { x: 300, y: 100 },
        deterministic_dialogue: {
          start_node_id: 'dn-1',
          nodes: [
            // ===== УЗЕЛ 1 =====
            {
              id: 'dn-1',
              incoming_message: {
                role_id: 'role-partner',
                text: 'Готовы продлить, но нам нужна скидка 20%. У конкурентов дешевле.',
                semantic_action_code: 'PRESSURE'
              },
              reply_options: [
                {
                  id: 'ro-1-1',
                  role_id: 'role-learner',
                  text: 'Хорошо, давайте 20%, лишь бы вы остались.',
                  semantic_action_code: 'MAKE_CONCESSION',
                  transition_id: 'tr-price-to-fail'
                },
                {
                  id: 'ro-1-2',
                  role_id: 'role-learner',
                  text: 'А что для вас важнее — цена или условия поддержки?',
                  semantic_action_code: 'ASK_QUESTION',
                  next_node_id: 'dn-2'
                },
                {
                  id: 'ro-1-3',
                  role_id: 'role-learner',
                  text: 'Скидку 20% дать не можем. Давайте обсудим, что ещё важно.',
                  semantic_action_code: 'HOLD_BOUNDARY',
                  next_node_id: 'dn-2'
                }
              ]
            },

            // ===== УЗЕЛ 2 =====
            {
              id: 'dn-2',
              incoming_message: {
                role_id: 'role-partner',
                text: 'Нам важно уложиться в бюджет. Но и поддержка важна — у конкурента она слабее.',
                semantic_action_code: 'ASK_QUESTION'
              },
              reply_options: [
                {
                  id: 'ro-2-1',
                  role_id: 'role-learner',
                  text: 'Хорошо, давайте 15%, но это максимум.',
                  semantic_action_code: 'MAKE_CONCESSION',
                  transition_id: 'tr-price-to-fail'
                },
                {
                  id: 'ro-2-2',
                  role_id: 'role-learner',
                  text: 'Понимаю. Давайте обсудим детали — возможно, найдём вариант, который устроит обоих.',
                  semantic_action_code: 'MAKE_PROPOSAL',
                  next_node_id: 'dn-3'
                },
                {
                  id: 'ro-2-3',
                  role_id: 'role-learner',
                  text: 'Что именно для вас критично в поддержке?',
                  semantic_action_code: 'ASK_QUESTION',
                  next_node_id: 'dn-3'
                }
              ]
            },

            // ===== УЗЕЛ 3 =====
            {
              id: 'dn-3',
              incoming_message: {
                role_id: 'role-partner',
                text: 'Нам важно, чтобы поддержка была круглосуточной. А что вы можете предложить взамен?',
                semantic_action_code: 'REQUEST_RECIPROCITY'
              },
              reply_options: [
                {
                  id: 'ro-3-1',
                  role_id: 'role-learner',
                  text: 'Ничего, просто дадим 10% и всё.',
                  semantic_action_code: 'MAKE_CONCESSION',
                  transition_id: 'tr-price-to-fail'
                },
                {
                  id: 'ro-3-2',
                  role_id: 'role-learner',
                  text: 'Предлагаем скидку 10% и расширенный SLA, если вы фиксируете контракт на 12 месяцев.',
                  semantic_action_code: 'MAKE_PROPOSAL',
                  next_node_id: 'dn-4'
                },
                {
                  id: 'ro-3-3',
                  role_id: 'role-learner',
                  text: 'Мы готовы на расширенный SLA. Что вы можете предложить взамен?',
                  semantic_action_code: 'REQUEST_RECIPROCITY',
                  next_node_id: 'dn-4'
                }
              ]
            },

            // ===== УЗЕЛ 4 =====
            {
              id: 'dn-4',
              incoming_message: {
                role_id: 'role-partner',
                text: 'Расширенный SLA — это интересно. А что насчёт срока контракта?',
                semantic_action_code: 'ASK_QUESTION'
              },
              reply_options: [
                {
                  id: 'ro-4-1',
                  role_id: 'role-learner',
                  text: 'Давайте оставим срок как есть.',
                  semantic_action_code: 'MAKE_CONCESSION',
                  transition_id: 'tr-price-to-fail'
                },
                {
                  id: 'ro-4-2',
                  role_id: 'role-learner',
                  text: 'Готовы зафиксировать 12 месяцев, если вы согласны на 10% и расширенный SLA.',
                  semantic_action_code: 'MAKE_PROPOSAL',
                  next_node_id: 'dn-5'
                }
              ]
            },

            // ===== УЗЕЛ 5 — ФИНАЛЬНЫЙ ВЫБОР =====
            {
              id: 'dn-5',
              incoming_message: {
                role_id: 'role-partner',
                text: 'Хорошо, давайте обсудим детали. Что вы предлагаете по итогу?',
                semantic_action_code: 'ASK_QUESTION'
              },
              reply_options: [
                {
                  id: 'ro-5-1',
                  role_id: 'role-learner',
                  text: 'Хорошо, финальное предложение: скидка 10%, расширенный SLA, контракт на 12 месяцев. Подписываем сегодня.',
                  semantic_action_code: 'SUMMARIZE',
                  transition_id: 'tr-price-to-deal'
                },
                {
                  id: 'ro-5-2',
                  role_id: 'role-learner',
                  text: 'Ладно, давайте 15% без дополнительных условий.',
                  semantic_action_code: 'MAKE_CONCESSION',
                  transition_id: 'tr-price-to-fail'
                }
              ]
            }
          ]
        },
        transitions: [
          {
            id: 'tr-price-to-deal',
            target_block_id: 'block-deal',
            condition: 'Удалось договориться об обмене',
            weight: 70,
            forbidden: false
          },
          {
            id: 'tr-price-to-fail',
            target_block_id: 'block-fail',
            condition: 'Уступка без обмена или отказ клиента',
            weight: 30,
            forbidden: false
          }
        ]
      },

      // ============================================
      // БЛОК 3: ФИНАЛ — УСПЕХ
      // ============================================
      {
        id: 'block-deal',
        type: 'end',
        name: 'Сделка закрыта',
        description: 'Контракт продлён на 12 месяцев. Скидка 10%, расширенный SLA.',
        expanded_context: '',
        lead_role_id: null,
        lead_probability: 50,
        role_goals: [],
        behavior: '',
        action_rules: [],
        constraints: [],
        position: { x: 900, y: 50 },
        transitions: []
      },

      // ============================================
      // БЛОК 4: ФИНАЛ — ПРОВАЛ
      // ============================================
      {
        id: 'block-fail',
        type: 'end',
        name: 'Сделка не состоялась',
        description: 'Клиент ушёл к конкуренту. Скидка 20% не была обоснована, обмен не предложен.',
        expanded_context: '',
        lead_role_id: null,
        lead_probability: 50,
        role_goals: [],
        behavior: '',
        action_rules: [],
        constraints: [],
        position: { x: 900, y: 250 },
        transitions: []
      }
    ]
  }
];

export function getScenario(id) {
  return scenarios.find(s => s.scenario_id === id);
}