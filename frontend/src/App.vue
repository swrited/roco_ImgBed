<script setup lang="ts">
import { RouterView } from 'vue-router'
import { Toaster } from '@/components/ui/sonner'
import { useSettingsStore } from '@/stores/settings'
import { onMounted, computed, ref, watch } from 'vue'

const settings = useSettingsStore()
const bgLoaded = ref(false)

onMounted(() => {
  settings.fetchPublicSettings()
})

const hasBg = computed(() => !!settings.bgImage)

// 预加载背景图片，加载完毕后再显示
watch(() => settings.bgImage, (url) => {
  bgLoaded.value = false
  if (!url) return
  const img = new Image()
  img.onload = () => { bgLoaded.value = true }
  img.onerror = () => { bgLoaded.value = false }
  img.src = url
}, { immediate: true })

const overlayOpacity = computed(() => settings.bgOpacity / 100)
</script>

<template>
  <!-- 全局背景图层 -->
  <Transition name="bg-fade">
    <div
      v-if="hasBg && bgLoaded"
      class="site-bg-layer"
      :style="{ backgroundImage: `url(${settings.bgImage})` }"
      aria-hidden="true"
    >
      <div class="site-bg-overlay" :style="{ opacity: overlayOpacity }" />
    </div>
  </Transition>

  <RouterView v-slot="{ Component, route }">
    <Transition name="route-fade" mode="out-in">
      <component :is="Component" :key="route.matched[0]?.name || route.matched[0]?.path" />
    </Transition>
  </RouterView>
  <Toaster />
</template>

<style scoped>
.site-bg-layer {
  position: fixed;
  inset: 0;
  z-index: -1;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}

.site-bg-overlay {
  position: absolute;
  inset: 0;
  background: oklch(8% 0.01 270);
}

@media (max-width: 1023px) {
  .site-bg-layer {
    display: none;
  }
}

/* 背景图淡入动画 */
.bg-fade-enter-active {
  transition: opacity 0.8s ease;
}
.bg-fade-leave-active {
  transition: opacity 0.4s ease;
}
.bg-fade-enter-from,
.bg-fade-leave-to {
  opacity: 0;
}
</style>
