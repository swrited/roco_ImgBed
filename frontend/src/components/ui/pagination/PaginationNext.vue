<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import type { ButtonVariants } from '@/components/ui/button'
import { ChevronRightIcon } from 'lucide-vue-next'
import { cn } from '@/lib/utils'
import { buttonVariants } from '@/components/ui/button'

const props = withDefaults(defineProps<{
  size?: ButtonVariants['size']
  disabled?: boolean
  class?: HTMLAttributes['class']
}>(), {
  size: 'default',
  disabled: false,
})
</script>

<template>
  <button
    type="button"
    data-slot="pagination-next"
    :disabled="disabled"
    :class="cn(buttonVariants({ variant: 'ghost', size }), 'pr-1.5!', disabled && 'pointer-events-none opacity-50', props.class)"
  >
    <slot>
      <span class="hidden sm:block">下一页</span>
      <ChevronRightIcon data-icon="inline-end" class="cn-rtl-flip" />
    </slot>
  </button>
</template>
