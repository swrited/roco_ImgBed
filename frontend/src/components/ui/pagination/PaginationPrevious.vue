<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import type { ButtonVariants } from '@/components/ui/button'
import { ChevronLeftIcon } from 'lucide-vue-next'
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
    data-slot="pagination-previous"
    :disabled="disabled"
    :class="cn(buttonVariants({ variant: 'ghost', size }), 'pl-1.5!', disabled && 'pointer-events-none opacity-50', props.class)"
  >
    <slot>
      <ChevronLeftIcon data-icon="inline-start" class="cn-rtl-flip" />
      <span class="hidden sm:block">上一页</span>
    </slot>
  </button>
</template>
