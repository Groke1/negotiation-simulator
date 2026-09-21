export const mockScenario = {
  scenario_id: 'sc-mock-001',
  title: 'Продление контракта — скидка 15%',
  description: 'Клиент требует скидку 15%, ссылаясь на конкурента. Менеджер может дать максимум 7%, но предложить расширенный SLA.',
  expanded_context: 'Клиент — крупная компания, у которой есть альтернативный поставщик с более низкой ценой. Реальная причина требования скидки — не только цена, но и желание зафиксировать долгосрочные условия.',
  status: 'draft',
  difficulty: 'medium',
  tags: ['B2B', 'продажи'],
  supported_runtimes: ['deterministic', 'virtual_partner'],
  supported_channels: ['chat'],
  target_cluster_ids: [],
  roles: [
    {
      id: 'role-learner',
      name: 'Менеджер по продажам',
      description: 'Продаёт услугу, отвечает за условия',
      goal: 'Продлить контракт на 12 месяцев',
      interests: 'Сохранить маржу',
      constraints: 'Максимум 7% скидки',
      private_information: 'Готов предложить расширенный SLA вместо скидки'
    },
    {
      id: 'role-partner',
      name: 'Клиент',
      description: 'Закупщик крупной компании',
      goal: 'Получить скидку 15%',
      interests: 'Снизить издержки',
      constraints: 'Есть альтернативный поставщик',
      private_information: 'Конкурент без расширенной поддержки'
    }
  ],
  blocks: [
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
          id: 'tr-1',
          target_block_id: 'block-price',
          condition: 'Клиент начинает с требования скидки',
          weight: 100,
          forbidden: false
        }
      ]
    },
    {
      id: 'block-price',
      type: 'regular',
      name: 'Обсуждение цены',
      description: 'Клиент требует скидку 15%',
      expanded_context: 'Клиент давит на цену, ссылаясь на конкурента. Важно понять, что за этим стоит — реальная цена или желание долгосрочных условий.',
      lead_role_id: 'role-partner',
      lead_probability: 70,
      role_goals: [],
      behavior: 'Давит, не уступает сразу',
      action_rules: [
        { semantic_action_code: 'ASK_QUESTION', reaction: 'Уточняет причину требования' },
        { semantic_action_code: 'MAKE_PROPOSAL', reaction: 'Обсуждает условия обмена' }
      ],
      constraints: ['Не соглашаться на скидку без выяснения причин'],
      transitions: [
        {
          id: 'tr-2',
          target_block_id: 'block-end-success',
          condition: 'Договорились с обменом',
          weight: 70,
          forbidden: false
        },
        {
          id: 'tr-3',
          target_block_id: 'block-end-fail',
          condition: 'Уступили без обмена',
          weight: 30,
          forbidden: false
        }
      ]
    },
    {
      id: 'block-end-success',
      type: 'end',
      name: 'Сделка закрыта',
      description: 'Контракт продлён на 12 месяцев',
      expanded_context: '',
      lead_role_id: null,
      lead_probability: 50,
      role_goals: [],
      behavior: '',
      action_rules: [],
      constraints: [],
      transitions: []
    },
    {
      id: 'block-end-fail',
      type: 'end',
      name: 'Сделка не состоялась',
      description: 'Клиент ушёл к конкуренту',
      expanded_context: '',
      lead_role_id: null,
      lead_probability: 50,
      role_goals: [],
      behavior: '',
      action_rules: [],
      constraints: [],
      transitions: []
    }
  ]
};