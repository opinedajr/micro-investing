import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import DashboardKpiCards from './DashboardKpiCards.vue'

describe('DashboardKpiCards', () => {
  it('renders the three kpi cards with values formatted as BRL', () => {
    const wrapper = mount(DashboardKpiCards, {
      props: {
        summary: {
          current_patrimony: 1050,
          yearly_dividends: 2050,
          stocks_invested: 3050,
        },
      },
    })

    const cards = wrapper.findAll('.kpi-card')

    expect(cards).toHaveLength(3)
    expect(wrapper.text()).toContain('Patrimônio')
    expect(wrapper.text()).toContain('R$ 10,50')
    expect(wrapper.text()).toContain('Dividendo')
    expect(wrapper.text()).toContain('R$ 20,50')
    expect(wrapper.text()).toContain('Ações')
    expect(wrapper.text()).toContain('R$ 30,50')
  })

  it('renders zero values when summary is not loaded yet', () => {
    const wrapper = mount(DashboardKpiCards, {
      props: { summary: null },
    })

    expect(wrapper.findAll('.kpi-card')).toHaveLength(3)
    expect(wrapper.text()).toContain('R$ 0,00')
  })
})
