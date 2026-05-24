<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { imagesApi } from '@/api/images'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Pagination, PaginationContent, PaginationItem, PaginationLink, PaginationNext, PaginationPrevious,
} from '@/components/ui/pagination'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { Check, Trash2, RotateCcw } from 'lucide-vue-next'
import type { Image } from '@/types'

const images = ref<Image[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const lastPage = ref(1)

const selectedKeys = ref<string[]>([])
const showForceDeleteDialog = ref(false)
const keysToDelete = ref<string[]>([])

async function loadImages(page = 1) {
  loading.value = true
  try {
    const res = await imagesApi.getTrash({ page }) as any
    images.value = res.data
    total.value = res.total
    currentPage.value = res.current_page
    lastPage.value = res.last_page
  } catch (e: any) {
    toast.error(e.message || '获取回收站失败')
  } finally {
    loading.value = false
  }
}

function toggleSelect(key: string) {
  const index = selectedKeys.value.indexOf(key)
  if (index === -1) {
    selectedKeys.value.push(key)
  } else {
    selectedKeys.value.splice(index, 1)
  }
}

async function restoreImages(keys: string[]) {
  try {
    await imagesApi.restore(keys)
    toast.success('恢复成功')
    selectedKeys.value = []
    loadImages(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '恢复失败')
  }
}

function confirmForceDelete(keys: string[]) {
  keysToDelete.value = keys
  showForceDeleteDialog.value = true
}

async function forceDeleteImages() {
  try {
    await imagesApi.forceDelete(keysToDelete.value)
    toast.success('彻底删除成功')
    showForceDeleteDialog.value = false
    selectedKeys.value = []
    loadImages(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '彻底删除失败')
  }
}

function formatSize(kb: number): string {
  if (kb >= 1048576) return (kb / 1048576).toFixed(2) + ' GB'
  if (kb >= 1024) return (kb / 1024).toFixed(2) + ' MB'
  return kb.toFixed(2) + ' KB'
}

function calculateDaysLeft(deletedAt: string): number {
  if (!deletedAt) return 30
  const deletedTime = new Date(deletedAt).getTime()
  const expiryTime = deletedTime + 30 * 24 * 60 * 60 * 1000
  const now = new Date().getTime()
  const days = Math.ceil((expiryTime - now) / (1000 * 60 * 60 * 24))
  return days > 0 ? days : 0
}

onMounted(() => {
  loadImages()
})
</script>

<template>
  <div>
    <div class="mb-6">
      <p class="text-xs font-semibold uppercase tracking-widest text-violet-400">Trash</p>
      <h1 class="mt-1.5 text-3xl font-semibold tracking-tight">回收站</h1>
      <p class="mt-2 text-sm text-muted-foreground">删除的图片会在这里保留 30 天，之后自动清理。</p>
    </div>

    <div class="flex items-center justify-between mb-4">
      <div class="flex items-center gap-2" v-if="selectedKeys.length > 0">
        <Badge variant="secondary">{{ selectedKeys.length }} 张已选</Badge>
        <Button variant="outline" size="sm" @click="restoreImages(selectedKeys)">
          <RotateCcw class="mr-1 h-4 w-4" /> 恢复
        </Button>
        <Button variant="destructive" size="sm" @click="confirmForceDelete(selectedKeys)">
          <Trash2 class="mr-1 h-4 w-4" /> 彻底删除
        </Button>
      </div>
      <div v-else>
        <p class="text-sm text-muted-foreground">回收站中的文件将在 30 天后自动清理。</p>
      </div>
    </div>

    <p class="text-sm text-muted-foreground mb-4">共 {{ total }} 张图片</p>

    <!-- Loading skeleton -->
    <div v-if="loading" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <div v-for="i in 8" :key="i" class="aspect-square bg-muted animate-pulse rounded-lg" />
    </div>

    <!-- Image grid -->
    <div v-else class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      <Card
        v-for="img in images" :key="img.key"
        class="cursor-pointer group overflow-hidden transition-all !ring-transparent hover:!ring-white/20 hover:!ring-1"
        :class="selectedKeys.includes(img.key) ? 'scale-[0.985] shadow-2xl shadow-primary/25 bg-primary/10 !ring-primary/50' : ''"
        @click="toggleSelect(img.key)"
      >
        <div class="relative aspect-square overflow-hidden rounded-t-lg">
          <img :src="img.url" :alt="img.alias_name || img.origin_name" class="h-full w-full object-cover" loading="lazy" />
          <div
            v-if="selectedKeys.includes(img.key)"
            class="pointer-events-none absolute inset-0 z-10 bg-black/35"
          />
          <div
            v-if="selectedKeys.includes(img.key)"
            class="pointer-events-none absolute inset-0 z-20 flex items-center justify-center"
          >
            <div class="flex h-16 w-16 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-2xl shadow-primary/50">
              <Check class="h-8 w-8" />
            </div>
          </div>
          <div class="absolute left-2 top-2 z-30 flex items-center gap-2">
            <button
              type="button"
              class="flex h-10 w-10 items-center justify-center rounded-lg border border-white/30 bg-black/65 text-white shadow-lg backdrop-blur transition-all hover:border-primary hover:bg-primary/35"
              :class="selectedKeys.includes(img.key) ? 'border-primary bg-primary text-primary-foreground shadow-primary/40' : ''"
              @click.stop="toggleSelect(img.key)"
            >
              <Check v-if="selectedKeys.includes(img.key)" class="h-4 w-4" />
            </button>
            <Badge
              v-if="selectedKeys.includes(img.key)"
              class="border-primary bg-primary text-primary-foreground shadow-lg shadow-primary/40 hover:bg-primary"
            >已选</Badge>
          </div>
          
          <div class="absolute bottom-0 inset-x-0 p-2 bg-gradient-to-t from-black/80 to-transparent pointer-events-none">
            <p class="text-xs text-red-300 font-medium text-right">{{ calculateDaysLeft(img.deleted_at || '') }} 天后清除</p>
          </div>
        </div>
        <CardContent class="p-3 transition-colors" :class="selectedKeys.includes(img.key) ? 'bg-primary/15' : ''">
          <p class="text-sm truncate font-medium">{{ img.alias_name || img.origin_name }}</p>
          <p class="text-xs text-muted-foreground">
            {{ img.width }}x{{ img.height }} · {{ formatSize(img.size) }}
          </p>
        </CardContent>
      </Card>
    </div>

    <div v-if="images.length === 0 && !loading" class="text-center py-12 text-muted-foreground">
      回收站为空
    </div>

    <!-- Pagination -->
    <div v-if="lastPage > 1" class="mt-6 flex justify-center">
      <Pagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious :disabled="currentPage <= 1" @click="loadImages(currentPage - 1)" />
          </PaginationItem>
          <PaginationItem
            v-for="page in Array.from({ length: lastPage }, (_, i) => i + 1)"
            :key="page"
          >
            <PaginationLink :is-active="page === currentPage" @click="loadImages(page)">
              {{ page }}
            </PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationNext :disabled="currentPage >= lastPage" @click="loadImages(currentPage + 1)" />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>

    <Dialog v-model:open="showForceDeleteDialog">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>彻底删除图片</DialogTitle>
          <DialogDescription>
            警告：彻底删除后图片将被从服务器上永久抹除，无法恢复！确认要继续吗？
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="showForceDeleteDialog = false">取消</Button>
          <Button variant="destructive" @click="forceDeleteImages">彻底删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
