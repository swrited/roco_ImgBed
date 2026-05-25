import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { settingsApi } from '@/api/settings'

export const useSettingsStore = defineStore('settings', () => {
  const raw = ref<Record<string, any>>({})
  const loaded = ref(false)

  const bgImage = computed(() => raw.value.site_bg_image || '')
  const bgOpacity = computed(() => {
    const v = parseInt(raw.value.site_bg_opacity || '85', 10)
    return isNaN(v) ? 85 : Math.max(0, Math.min(100, v))
  })

  async function fetchPublicSettings() {
    try {
      const data = await settingsApi.getPublicSettings()
      raw.value = data as Record<string, any>
      loaded.value = true
    } catch {
      // 静默失败，使用默认值
    }
  }

  return { raw, loaded, bgImage, bgOpacity, fetchPublicSettings }
})
