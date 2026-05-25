<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usersApi, type DashboardData } from '@/api/users'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import SparkLine from '@/components/SparkLine.vue'
import { Image, FolderOpen, HardDrive, TrendingUp, CalendarDays } from 'lucide-vue-next'

const data = ref<DashboardData>({
  user: { id: 0, name: '', email: '', image_num: 0, album_num: 0, capacity: 0, is_adminer: false, group_id: 0, created_at: '' },
  image_count: 0,
  album_count: 0,
  used_capacity: 0,
  today_count: 0,
  month_count: 0,
  daily_stats: [],
})
const loading = ref(true)

function formatSize(kb: number): string {
  if (kb <= 0) return '0 KB'
  if (kb >= 1048576) return (kb / 1048576).toFixed(2) + ' GB'
  if (kb >= 1024)    return (kb / 1024).toFixed(2) + ' MB'
  return Math.round(kb) + ' KB'
}

function usagePercent(): number {
  if (!data.value.user.capacity || data.value.user.capacity <= 0) return 0
  return Math.min(100, Math.round((data.value.used_capacity / data.value.user.capacity) * 100))
}

onMounted(async () => {
  try {
    data.value = await usersApi.dashboard()
  } catch { /* use defaults */ } finally {
    loading.value = false
  }
})

const statCards = [
  {
    label: '图片数量',
    icon: Image,
    value: () => data.value.image_count,
    sub: null,
    iconBg: 'bg-violet-500/10 text-violet-400',
    glow:   'shadow-violet-500/10',
    accent: 'border-t-violet-500/40',
  },
  {
    label: '今日上传',
    icon: TrendingUp,
    value: () => data.value.today_count,
    sub:   () => `本月 ${data.value.month_count} 张`,
    iconBg: 'bg-emerald-500/10 text-emerald-400',
    glow:   'shadow-emerald-500/10',
    accent: 'border-t-emerald-500/40',
  },
  {
    label: '相册数量',
    icon: FolderOpen,
    value: () => data.value.album_count,
    sub: null,
    iconBg: 'bg-indigo-500/10 text-indigo-400',
    glow:   'shadow-indigo-500/10',
    accent: 'border-t-indigo-500/40',
  },
  {
    label: '已用容量',
    icon: HardDrive,
    value: () => formatSize(data.value.used_capacity),
    sub:   () => `使用率 ${usagePercent()}%`,
    iconBg: 'bg-amber-500/10 text-amber-400',
    glow:   'shadow-amber-500/10',
    accent: 'border-t-amber-500/40',
  },
]
</script>

<template>
  <div>
    <!-- ── Page header ─────────────────────────────────────────────────── -->
    <div class="animate-fade-in-up mb-8 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
      <div>
        <p class="text-xs font-semibold uppercase tracking-widest text-violet-400">Dashboard</p>
        <h1 class="mt-1.5 text-3xl font-semibold tracking-tight">控制面板</h1>
        <p class="mt-2 text-sm leading-6 text-muted-foreground">
          欢迎回来，<span class="font-medium text-slate-200">{{ data.user.name || '用户' }}</span>。
          这里汇总你的图片、相册和容量使用情况。
        </p>
      </div>
    </div>

    <!-- ── Stat cards ──────────────────────────────────────────────────── -->
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4 mb-6">
      <Card
        v-for="(card, i) in statCards"
        :key="card.label"
        class="animate-fade-in-up overflow-hidden border-t-2"
        :class="[card.accent, `delay-${i * 75}`]"
        :style="`--tw-shadow: 0 8px 32px var(--tw-shadow-color); box-shadow: 0 8px 32px oklch(0 0 0 / 0.35)`"
      >
        <CardHeader class="flex flex-row items-center justify-between pb-2 pt-4">
          <CardTitle class="text-sm font-medium text-muted-foreground">{{ card.label }}</CardTitle>
          <span
            :class="card.iconBg"
            class="flex h-9 w-9 items-center justify-center rounded-2xl shadow-sm"
          >
            <component :is="card.icon" class="h-4 w-4" />
          </span>
        </CardHeader>
        <CardContent class="pb-4">
          <p v-if="loading" class="skeleton h-8 w-20 rounded-lg" />
          <p v-else class="text-3xl font-semibold tabular-nums tracking-tight">{{ card.value() }}</p>
          <p v-if="card.sub" class="mt-1 text-xs text-muted-foreground">
            <template v-if="!loading">{{ card.sub() }}</template>
            <span v-else class="skeleton inline-block h-3 w-24 rounded" />
          </p>
        </CardContent>
      </Card>
    </div>

    <!-- ── Upload trend chart ──────────────────────────────────────────── -->
    <Card class="animate-fade-in-up delay-300 mb-6 border-t-2 border-t-violet-500/30">
      <CardHeader class="pb-2">
        <CardTitle class="flex items-center gap-2 text-sm font-semibold text-muted-foreground uppercase tracking-widest">
          <CalendarDays class="h-4 w-4 text-violet-400" />
          近 30 天上传统计
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="skeleton h-40 w-full rounded-xl" />
        <div v-else-if="data.daily_stats.length > 0">
          <SparkLine :data="data.daily_stats" :height="160" />
        </div>
        <div v-else class="flex h-40 items-center justify-center text-sm text-muted-foreground">
          暂无上传数据
        </div>
      </CardContent>
    </Card>

    <!-- ── Storage usage ───────────────────────────────────────────────── -->
    <Card class="animate-fade-in-up delay-450 border-t-2 border-t-sky-500/30">
      <CardHeader class="pb-3">
        <CardTitle class="flex items-center gap-2 text-sm font-semibold text-muted-foreground uppercase tracking-widest">
          <HardDrive class="h-4 w-4 text-sky-400" />
          存储空间
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div class="space-y-3">
          <div class="flex items-center justify-between text-sm">
            <span class="text-muted-foreground">已使用</span>
            <span class="font-medium tabular-nums">{{ formatSize(data.used_capacity) }}</span>
          </div>
          <Progress :model-value="usagePercent()" class="h-2" />
          <div class="flex items-center justify-between text-sm">
            <span class="text-muted-foreground">剩余空间</span>
            <span class="font-medium tabular-nums">
              {{ data.user.capacity > 0 ? formatSize(data.user.capacity - data.used_capacity) : '无限制' }}
            </span>
          </div>
          <div class="flex items-center justify-between text-sm">
            <span class="text-muted-foreground">总容量</span>
            <span class="font-medium tabular-nums">
              {{ data.user.capacity > 0 ? formatSize(data.user.capacity) : '无限制' }}
            </span>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
