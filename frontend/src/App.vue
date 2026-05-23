<script setup lang="ts">
import { RouterView } from 'vue-router'
import { Toaster } from '@/components/ui/sonner'
</script>

<template>
  <RouterView v-slot="{ Component, route }">
    <Transition name="route-fade" mode="out-in">
      <component :is="Component" :key="route.matched[0]?.name || route.matched[0]?.path" />
    </Transition>
  </RouterView>
  <Toaster />
</template>
