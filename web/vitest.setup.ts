import { vi } from 'vitest'

vi.mock('primevue/chart', async () => {
  const { defineComponent, h } = await import('vue')

  return {
    default: defineComponent({
      name: 'Chart',
      props: {
        type: { type: String, required: true },
        data: { type: Object, default: () => ({}) },
        options: { type: Object, default: () => ({}) },
      },
      setup(props) {
        return () =>
          h('div', {
            class: 'chart-stub',
            'data-testid': 'chart-stub',
            'data-chart-type': props.type,
            'data-chart-data': JSON.stringify(props.data),
            'data-chart-options': JSON.stringify(props.options),
          })
      },
    }),
  }
})
