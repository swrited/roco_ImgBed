<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{ class?: HTMLAttributes['class'] }>()
</script>

<template>
  <div
    data-slot="pagination-content"
    :class="cn('gap-0.5 flex items-center', props.class)"
  >
    <slot />
  </div>
</template>
