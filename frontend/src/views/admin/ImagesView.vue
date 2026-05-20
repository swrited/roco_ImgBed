<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { adminApi } from '@/api/admin'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { toast } from 'vue-sonner'
import { Search, Trash2, ChevronLeft, ChevronRight, Image } from 'lucide-vue-next'
import type { PaginatedResponse } from '@/types'

// --------------- Types ---------------
interface ImageItem {
  key: string
  name: string
  alias_name?: string
  path: string
  size: number
  width: number
  height: number
  url: string
  extension: string
  md5: string
  permission: number
  created_at: string
  user_id: number
  user?: {
    id: number
    name: string
    email: string
  }
}

// --------------- Constants ---------------
const PER_PAGE = 20

// --------------- Route ---------------
const route = useRoute()
const router = useRouter()

// --------------- State ---------------
const images = ref<ImageItem[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const lastPage = ref(1)
const selectedKeys = ref<Set<string>>(new Set())

// Confirm dialog
const showDeleteDialog = ref(false)

// Search
const searchQuery = ref((route.query.q as string) || '')

// --------------- Computed ---------------
const allSelected = computed(() => {
  return images.value.length > 0 && images.value.every((img) => selectedKeys.value.has(img.key))
})

const someSelected = computed(() => {
  return images.value.some((img) => selectedKeys.value.has(img.key)) && !allSelected.value
})

const selectedCount = computed(() => selectedKeys.value.size)

const hasPrev = computed(() => currentPage.value > 1)
const hasNext = computed(() => currentPage.value < lastPage.value)

// --------------- Helpers ---------------
function formatSize(kb: number): string {
  if (kb >= 1048576) {
    return (kb / 1048576).toFixed(2) + ' GB'
  }
  if (kb >= 1024) {
    return (kb / 1024).toFixed(2) + ' MB'
  }
  return kb.toFixed(2) + ' KB'
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// --------------- Actions ---------------
async function fetchImages() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: currentPage.value,
      per_page: PER_PAGE,
    }
    if (searchQuery.value.trim()) {
      params.q = searchQuery.value.trim()
    }
    const res = (await adminApi.listImages(params)) as PaginatedResponse<ImageItem>
    images.value = res.data || []
    total.value = res.total || 0
    currentPage.value = res.current_page || 1
    lastPage.value = res.last_page || 1
  } catch (e: any) {
    toast.error(e.message || '加载图片列表失败')
    images.value = []
  } finally {
    loading.value = false
  }
}

function onSearch() {
  currentPage.value = 1
  // Sync query param to URL
  router.replace({
    query: {
      ...route.query,
      q: searchQuery.value.trim() || undefined,
      page: undefined,
    },
  })
  fetchImages()
}

function goToPage(page: number) {
  if (page < 1 || page > lastPage.value) return
  currentPage.value = page
  fetchImages()
}

function toggleSelectAll() {
  if (allSelected.value) {
    selectedKeys.value = new Set()
  } else {
    selectedKeys.value = new Set(images.value.map((img) => img.key))
  }
}

function toggleSelect(key: string) {
  const newSet = new Set(selectedKeys.value)
  if (newSet.has(key)) {
    newSet.delete(key)
  } else {
    newSet.add(key)
  }
  selectedKeys.value = newSet
}

function openDeleteDialog() {
  if (selectedKeys.value.size === 0) {
    toast.error('请先选择要删除的图片')
    return
  }
  showDeleteDialog.value = true
}

async function handleBatchDelete() {
  try {
    await adminApi.deleteImages(Array.from(selectedKeys.value))
    toast.success(`成功删除 ${selectedKeys.value.size} 张图片`)
    selectedKeys.value.clear()
    fetchImages()
  } catch (e: any) {
    toast.error(e.message || '删除失败')
  }
}

// --------------- Lifecycle ---------------
onMounted(() => {
  // Restore page from query if present
  const pageParam = route.query.page
  if (pageParam) {
    currentPage.value = Number(pageParam) || 1
  }
  fetchImages()
})

// Sync search query from URL changes (e.g. browser back)
watch(
  () => route.query.q,
  (newVal) => {
    if (newVal !== searchQuery.value) {
      searchQuery.value = (newVal as string) || ''
      fetchImages()
    }
  },
)
</script>

<template>
  <div>
    <!-- Page Header -->
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">图片管理</h1>
    </div>

    <!-- Search Bar -->
    <div class="relative mb-6">
      <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
      <Input
        v-model="searchQuery"
        placeholder="搜索图片..."
        class="pl-9"
        @keyup.enter="onSearch"
      />
    </div>

    <!-- Batch Actions -->
    <div
      v-if="selectedCount > 0"
      class="flex items-center gap-3 mb-4 px-4 py-2 bg-muted/50 rounded-lg border"
    >
      <span class="text-sm text-muted-foreground">
        已选择 <span class="font-medium text-foreground">{{ selectedCount }}</span> 张图片
      </span>
      <Button variant="destructive" size="sm" @click="openDeleteDialog">
        <Trash2 class="mr-1 h-4 w-4" />
        批量删除
      </Button>
    </div>

    <!-- Table -->
    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead class="w-12">
              <Checkbox
                :checked="allSelected"
                @update:checked="toggleSelectAll"
              />
            </TableHead>
            <TableHead class="w-16">缩略图</TableHead>
            <TableHead>文件名</TableHead>
            <TableHead>上传用户</TableHead>
            <TableHead>文件大小</TableHead>
            <TableHead>尺寸</TableHead>
            <TableHead>上传时间</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <!-- Loading Skeleton -->
          <template v-if="loading">
            <TableRow v-for="i in 6" :key="'skeleton-' + i">
              <TableCell colspan="7">
                <div class="h-10 bg-muted animate-pulse rounded" />
              </TableCell>
            </TableRow>
          </template>

          <!-- Empty State -->
          <template v-else-if="images.length === 0">
            <TableRow>
              <TableCell colspan="7">
                <div class="flex flex-col items-center justify-center py-16 text-muted-foreground">
                  <Image class="h-12 w-12 mb-4 opacity-40" />
                  <p class="text-lg font-medium">暂无图片</p>
                  <p class="text-sm mt-1">系统中还没有任何图片</p>
                </div>
              </TableCell>
            </TableRow>
          </template>

          <!-- Data Rows -->
          <template v-else>
            <TableRow v-for="img in images" :key="img.key">
              <TableCell>
                <Checkbox
                  :checked="selectedKeys.has(img.key)"
                  @update:checked="toggleSelect(img.key)"
                />
              </TableCell>
              <TableCell>
                <img
                  :src="img.url"
                  :alt="img.alias_name || img.name"
                  class="h-10 w-10 rounded object-cover"
                  loading="lazy"
                />
              </TableCell>
              <TableCell class="font-medium max-w-[200px] truncate">
                {{ img.alias_name || img.name }}
              </TableCell>
              <TableCell>{{ img.user?.name || '-' }}</TableCell>
              <TableCell>{{ formatSize(img.size) }}</TableCell>
              <TableCell>{{ img.width }}×{{ img.height }}</TableCell>
              <TableCell class="text-muted-foreground">{{ formatDate(img.created_at) }}</TableCell>
            </TableRow>
          </template>
        </TableBody>
      </Table>
    </div>

    <!-- Pagination -->
    <div class="flex items-center justify-between mt-4">
      <p class="text-sm text-muted-foreground">
        {{ total > 0 ? `共 ${total} 张图片，第 ${currentPage} / ${lastPage} 页` : '' }}
      </p>
      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          :disabled="!hasPrev"
          @click="goToPage(currentPage - 1)"
        >
          <ChevronLeft class="mr-1 h-4 w-4" />
          上一页
        </Button>
        <Button
          variant="outline"
          size="sm"
          :disabled="!hasNext"
          @click="goToPage(currentPage + 1)"
        >
          下一页
          <ChevronRight class="ml-1 h-4 w-4" />
        </Button>
      </div>
    </div>

    <!-- Delete Confirm Dialog -->
    <ConfirmDialog
      v-model:open="showDeleteDialog"
      title="确认删除"
      :description="`确定要删除选中的 ${selectedCount} 张图片吗？此操作不可撤销。`"
      confirm-text="删除"
      cancel-text="取消"
      variant="destructive"
      @confirm="handleBatchDelete"
    />
  </div>
</template>
