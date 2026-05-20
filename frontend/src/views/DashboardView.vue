<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usersApi } from '@/api/users'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Upload, Image, FolderOpen, HardDrive, BarChart3 } from 'lucide-vue-next'

interface DashboardData {
  user: { name: string; image_num: number; album_num: number; capacity: number }
  image_count: number
  album_count: number
  used_capacity: number
}

const data = ref<DashboardData>({
  user: { name: '', image_num: 0, album_num: 0, capacity: 0 },
  image_count: 0,
  album_count: 0,
  used_capacity: 0,
})
const loading = ref(true)

function formatSize(kb: number): string {
  if (kb <= 0) return '0 KB'
  if (kb >= 1048576) return (kb / 1048576).toFixed(2) + ' GB'
  if (kb >= 1024) return (kb / 1024).toFixed(2) + ' MB'
  return Math.round(kb) + ' KB'
}

function usagePercent(): number {
  if (!data.value.user.capacity || data.value.user.capacity <= 0) return 0
  return Math.min(100, Math.round((data.value.used_capacity / data.value.user.capacity) * 100))
}

onMounted(async () => {
  try {
    const res = await usersApi.profile()
    // profile returns user object which we combine with dashboard data
    data.value = {
      user: {
        name: res.name || '',
        image_num: res.image_num || 0,
        album_num: res.album_num || 0,
        capacity: res.capacity || 0,
      },
      image_count: res.image_num || 0,
      album_count: res.album_num || 0,
      used_capacity: res.used_capacity || 0,
    }
  } catch { /* use defaults */ } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="mb-6 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
      <div>
        <p class="text-sm font-medium text-primary">Dashboard</p>
        <h1 class="mt-1 text-3xl font-semibold">控制面板</h1>
        <p class="mt-2 text-sm text-muted-foreground">欢迎回来，{{ data.user.name || '用户' }}。这里汇总你的图片、相册和容量使用情况。</p>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4 mb-6">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">图片数量</CardTitle>
          <span class="flex h-9 w-9 items-center justify-center rounded-2xl bg-primary/10 text-primary">
            <Image class="h-4 w-4" />
          </span>
        </CardHeader>
        <CardContent>
          <p class="text-3xl font-semibold">{{ data.image_count }}</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">相册数量</CardTitle>
          <span class="flex h-9 w-9 items-center justify-center rounded-2xl bg-accent text-accent-foreground">
            <FolderOpen class="h-4 w-4" />
          </span>
        </CardHeader>
        <CardContent>
          <p class="text-3xl font-semibold">{{ data.album_count }}</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">已用容量</CardTitle>
          <span class="flex h-9 w-9 items-center justify-center rounded-2xl bg-slate-100 text-slate-700">
            <HardDrive class="h-4 w-4" />
          </span>
        </CardHeader>
        <CardContent>
          <p class="text-3xl font-semibold">{{ formatSize(data.used_capacity) }}</p>
          <p class="text-xs text-muted-foreground mt-1">
            总容量 {{ formatSize(data.user.capacity) }}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-sm font-medium">使用率</CardTitle>
          <span class="flex h-9 w-9 items-center justify-center rounded-2xl bg-emerald-100 text-emerald-700">
            <BarChart3 class="h-4 w-4" />
          </span>
        </CardHeader>
        <CardContent>
          <p class="text-3xl font-semibold">{{ usagePercent() }}%</p>
          <Progress :model-value="usagePercent()" class="mt-2 h-2" />
        </CardContent>
      </Card>
    </div>

    <!-- Storage Usage Details -->
    <Card>
      <CardHeader>
        <CardTitle class="text-base">存储空间</CardTitle>
      </CardHeader>
      <CardContent>
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-sm text-muted-foreground">已使用</span>
            <span class="text-sm font-medium">{{ formatSize(data.used_capacity) }}</span>
          </div>
          <Progress :model-value="usagePercent()" class="h-3" />
          <div class="flex items-center justify-between">
            <span class="text-sm text-muted-foreground">剩余空间</span>
            <span class="text-sm font-medium">
              {{ data.user.capacity > 0 ? formatSize(data.user.capacity - data.used_capacity) : '无限制' }}
            </span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-sm text-muted-foreground">总容量</span>
            <span class="text-sm font-medium">{{ data.user.capacity > 0 ? formatSize(data.user.capacity) : '无限制' }}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
