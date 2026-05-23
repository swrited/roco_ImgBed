<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { computed, useAttrs } from 'vue'
import { cn } from '@/lib/utils'

defineOptions({
  inheritAttrs: false,
})

const props = withDefaults(defineProps<{
  modelValue?: boolean | number | string
  checked?: boolean | number | string
  disabled?: boolean
  class?: HTMLAttributes['class']
  size?: 'sm' | 'default'
}>(), {
  modelValue: undefined,
  checked: undefined,
  disabled: false,
  size: 'default',
})

const emits = defineEmits<{
  'update:modelValue': [value: boolean]
  'update:checked': [value: boolean]
}>()

const attrs = useAttrs()

const isChecked = computed(() => {
  const value = props.modelValue ?? props.checked
  return value === true || value === 1 || value === '1' || value === 'true'
})

function toggle() {
  if (props.disabled) return
  const next = !isChecked.value
  emits('update:modelValue', next)
  emits('update:checked', next)
}
</script>

<template>
  <button
    v-bind="attrs"
    type="button"
    role="switch"
    :aria-checked="isChecked"
    :disabled="disabled"
    :data-state="isChecked ? 'checked' : 'unchecked'"
    :data-size="size"
    :class="cn(
      'relative inline-flex shrink-0 items-center rounded-full border transition-all outline-none',
      'focus-visible:ring-2 focus-visible:ring-purple-500/45 focus-visible:ring-offset-2 focus-visible:ring-offset-background',
      'disabled:cursor-not-allowed disabled:opacity-50',
      size === 'sm' ? 'h-[14px] w-[24px]' : 'h-5 w-9',
      isChecked
        ? 'border-purple-500 bg-purple-500'
        : 'border-white/10 bg-slate-700',
      props.class,
    )"
    @click="toggle"
  >
    <span
      :class="cn(
        'pointer-events-none block rounded-full bg-white shadow-sm transition-transform',
        size === 'sm' ? 'h-3 w-3' : 'h-4 w-4',
        isChecked
          ? size === 'sm' ? 'translate-x-[10px]' : 'translate-x-[18px]'
          : 'translate-x-0.5',
      )"
    />
  </button>
</template>
