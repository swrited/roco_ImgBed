<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { apiUsageApi, type ApiUsageStats } from '@/api/apiUsage'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import SparkLine from '@/components/SparkLine.vue'
import { toast } from 'vue-sonner'
import { Activity, CalendarDays, Clock3, Route } from 'lucide-vue-next'

const stats = ref<ApiUsageStats>({
  total: 0,
  today: 0,
  last_7_days: 0,
  daily: [],
  endpoints: [],
})
const loading = ref(true)
const start = ref('')
const end = ref('')

function todayDate() {
  const d = new Date()
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

async function loadStats() {
  loading.value = true
  try {
    stats.value = await apiUsageApi.mine({ start: start.value, end: end.value })
  } catch (e: any) {
    toast.error(e.message || '加载 API 使用统计失败')
  } finally {
    loading.value = false
  }
}

function applyToday() {
  const today = todayDate()
  start.value = today
  end.value = today
  loadStats()
}

onMounted(() => {
  start.value = todayDate()
  end.value = todayDate()
  loadStats()
})
</script>

<template>
  <div class="space-y-6">
    <div>
      <p class="text-sm font-medium text-purple-400">API Usage</p>
      <h1 class="text-2xl font-semibold">我的 API 使用统计</h1>
      <p class="text-sm text-muted-foreground">只统计你自己的 API Key 调用，不包含其他用户数据。</p>
    </div>

    <div class="flex flex-wrap items-end gap-3 rounded-xl border bg-card p-4">
      <div class="space-y-1.5">
        <label class="text-xs text-muted-foreground">开始日期</label>
        <input v-model="start" type="date" class="date-input" />
      </div>
      <div class="space-y-1.5">
        <label class="text-xs text-muted-foreground">结束日期</label>
        <input v-model="end" type="date" class="date-input" />
      </div>
      <Button @click="loadStats">
        查询
      </Button>
      <Button variant="outline" @click="applyToday">
        今天
      </Button>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <Card>
        <CardContent class="flex items-center gap-3 p-5">
          <Activity class="h-5 w-5 text-purple-400" />
          <div>
            <p class="text-sm text-muted-foreground">累计调用</p>
            <p class="text-xs text-muted-foreground">当前筛选范围</p>
            <p class="text-2xl font-semibold">{{ stats.total }}</p>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="flex items-center gap-3 p-5">
          <Clock3 class="h-5 w-5 text-purple-400" />
          <div>
            <p class="text-sm text-muted-foreground">今日调用</p>
            <p class="text-2xl font-semibold">{{ stats.today }}</p>
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardContent class="flex items-center gap-3 p-5">
          <CalendarDays class="h-5 w-5 text-purple-400" />
          <div>
            <p class="text-sm text-muted-foreground">近 7 天</p>
            <p class="text-2xl font-semibold">{{ stats.last_7_days }}</p>
          </div>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>近 7 天调用趋势</CardTitle>
        <CardDescription>按自然日统计 API 请求次数</CardDescription>
      </CardHeader>
      <CardContent>
        <SparkLine v-if="stats.daily.length" :data="stats.daily" :height="180" />
        <p v-else class="py-8 text-center text-sm text-muted-foreground">暂无调用数据</p>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>常用接口</CardTitle>
        <CardDescription>按调用次数排序的前 10 个接口</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>接口</TableHead>
              <TableHead class="w-28 text-right">调用次数</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-for="item in stats.endpoints" :key="`${item.method}-${item.path}`">
              <TableCell>
                <div class="flex items-center gap-2">
                  <Badge variant="secondary">{{ item.method }}</Badge>
                  <span class="font-mono text-sm">{{ item.path }}</span>
                </div>
              </TableCell>
              <TableCell class="text-right font-medium">{{ item.count }}</TableCell>
            </TableRow>
            <TableRow v-if="!loading && stats.endpoints.length === 0">
              <TableCell colspan="2" class="py-8 text-center text-muted-foreground">
                <Route class="mx-auto mb-2 h-8 w-8 opacity-40" />
                暂无接口调用记录
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  </div>
</template>
