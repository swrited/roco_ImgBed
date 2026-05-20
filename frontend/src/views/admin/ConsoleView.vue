<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { adminApi } from '@/api/admin'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Users, Image, FolderOpen, HardDrive, TrendingUp } from 'lucide-vue-next'

interface ConsoleData {
  stats: { users: number; images: number; albums: number; recent_uploads: number }
  daily: { date: string; count: number }[]
  strategies_count: number
}

const loading = ref(true)
const data = ref<ConsoleData>({
  stats: { users: 0, images: 0, albums: 0, recent_uploads: 0 },
  daily: [],
  strategies_count: 0,
})

const maxDaily = computed(() => Math.max(1, ...data.value.daily.map(d => d.count)))

onMounted(async () => {
  try {
    const res = await adminApi.getConsole()
    if (res?.stats) {
      data.value.stats = {
        users: res.stats.users || 0,
        images: res.stats.images || 0,
        albums: res.stats.albums || 0,
        recent_uploads: res.stats.recent_uploads || 0,
      }
    }
    if (res?.daily) {
      data.value.daily = res.daily
    }
    // Also fetch strategy count
    try {
      const strategies = await adminApi.listStrategies()
      data.value.strategies_count = strategies.length
    } catch { /* use 0 */ }
  } catch { /* use defaults */ }
  finally { loading.value = false }
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold mb-6">管理控制台</h1>

    <!-- Stats Cards -->
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4 mb-8">
      <template v-if="loading">
        <Card v-for="i in 4" :key="i">
          <CardHeader class="flex flex-row items-center justify-between pb-2">
            <div class="h-4 bg-muted animate-pulse rounded w-16" />
            <div class="h-4 w-4 bg-muted animate-pulse rounded" />
          </CardHeader>
          <CardContent>
            <div class="h-8 bg-muted animate-pulse rounded w-12" />
          </CardContent>
        </Card>
      </template>
      <template v-else>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">总用户数</CardTitle>
          <Users class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <p class="text-2xl font-bold">{{ data.stats.users }}</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">总图片数</CardTitle>
          <Image class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <p class="text-2xl font-bold">{{ data.stats.images }}</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">总相册数</CardTitle>
          <FolderOpen class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <p class="text-2xl font-bold">{{ data.stats.albums }}</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">存储策略</CardTitle>
          <HardDrive class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <p class="text-2xl font-bold">{{ data.strategies_count }}</p>
        </CardContent>
      </Card>
      </template>
    </div>

    <!-- Daily Upload Trend -->
    <Card class="mb-8">
      <CardHeader>
        <CardTitle class="text-base flex items-center gap-2">
          <TrendingUp class="h-4 w-4" /> 近30天上传统计
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="data.daily.length > 0" class="flex items-end gap-1 h-40">
          <div
            v-for="(d, i) in data.daily"
            :key="i"
            class="flex-1 flex flex-col items-center justify-end gap-1 group"
          >
            <span class="text-[10px] text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity">
              {{ d.count }}
            </span>
            <div
              class="w-full bg-primary/60 hover:bg-primary rounded-t-sm transition-all min-h-[2px]"
              :style="{ height: (d.count / maxDaily * 100) + '%' }"
            />
            <span v-if="(i % 5 === 0) || i === data.daily.length - 1" class="text-[10px] text-muted-foreground mt-1">
              {{ d.date }}
            </span>
          </div>
        </div>
        <p v-else class="text-sm text-muted-foreground text-center py-8">暂无上传统计数据</p>
        <div class="mt-4 flex items-center justify-between text-sm text-muted-foreground">
          <div class="flex items-center gap-4">
            <span>近30天上传: <strong class="text-foreground">{{ data.stats.recent_uploads }}</strong></span>
            <span>日均: <strong class="text-foreground">{{ data.daily.length > 0 ? Math.round(data.stats.recent_uploads / data.daily.length) : 0 }}</strong></span>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- System Summary -->
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Card>
        <CardHeader class="pb-2">
          <CardTitle class="text-sm font-medium">人均图片</CardTitle>
        </CardHeader>
        <CardContent>
          <p class="text-2xl font-bold">
            {{ data.stats.users > 0 ? (data.stats.images / data.stats.users).toFixed(1) : '0' }}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardTitle class="text-sm font-medium">人均相册</CardTitle>
        </CardHeader>
        <CardContent>
          <p class="text-2xl font-bold">
            {{ data.stats.users > 0 ? (data.stats.albums / data.stats.users).toFixed(1) : '0' }}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="pb-2">
          <CardTitle class="text-sm font-medium">可用存储策略</CardTitle>
        </CardHeader>
        <CardContent>
          <p class="text-2xl font-bold">{{ data.strategies_count }} 种</p>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
