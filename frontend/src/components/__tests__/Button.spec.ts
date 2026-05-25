import { describe, it, expect } from 'vitest'

import { mount } from '@vue/test-utils'
import { Button } from '../ui/button'

describe('Button', () => {
  it('renders slot content', () => {
    const wrapper = mount(Button, { slots: { default: '上传图片' } })
    expect(wrapper.text()).toContain('上传图片')
  })
})
