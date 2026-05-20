<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import apiClient from '@/api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent } from '@/components/ui/card'
import {
  Dialog, DialogContent, DialogClose,
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { Search, ChevronLeft, ChevronRight, X } from 'lucide-vue-next'
import type { Image, PaginatedResponse } from '@/types'

const images = ref<Image[]>([])
const loading = ref(false)
const currentPage = ref(1)
const lastPage = ref(1)
const total = ref(0)
const perPage = 20
const keyword = ref('')

const previewImage = ref<Image | null>(null)
const previewIndex = ref(0)
const showPreview = ref(false)

function openPreview(img: Image, idx: number) {
  previewImage.value = img
  previewIndex.value = idx
  showPreview.value = true
}

function prevImage() {
  if (previewIndex.value > 0) {
    previewIndex.value--
    previewImage.value = images.value[previewIndex.value] || null
  }
}

function nextImage() {
  if (previewIndex.value < images.value.length - 1) {
    previewIndex.value++
    previewImage.value = images.value[previewIndex.value] || null
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (!showPreview.value) return
  if (e.key === 'ArrowLeft') prevImage()
  else if (e.key === 'ArrowRight') nextImage()
}

function formatSize(kb: number): string {
  if (kb <= 0) return '0 KB'
  if (kb >= 1048576) return (kb / 1048576).toFixed(2) + ' GB'
  if (kb >= 1024) return (kb / 1024).toFixed(2) + ' MB'
  return Math.round(kb) + ' KB'
}

async function loadImages(page = 1) {
  loading.value = true
  try {
    const params: Record<string, any> = { page, per_page: perPage }
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    const res = await apiClient.get<PaginatedResponse<Image>>('/gallery', { params })
    images.value = res.data || []
    currentPage.value = res.current_page || 1
    lastPage.value = res.last_page || 1
    total.value = res.total || 0
  } catch {
    toast.error('加载画廊失败')
    images.value = []
  } finally {
    loading.value = false
  }
}

function goPage(page: number) {
  if (page < 1 || page > lastPage.value || page === currentPage.value) return
  loadImages(page)
}

function search() {
  loadImages(1)
}

const pageNumbers = (): number[] => {
  const pages: number[] = []
  const start = Math.max(1, currentPage.value - 2)
  const end = Math.min(lastPage.value, currentPage.value + 2)
  for (let i = start; i <= end; i++) pages.push(i)
  return pages
}

onMounted(() => {
  loadImages()
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">画廊</h1>
      <div class="flex items-center gap-2">
        <div class="relative">
          <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            v-model="keyword"
            placeholder="搜索图片..."
            class="pl-9 w-56"
            @keyup.enter="search"
          />
        </div>
        <Button variant="outline" size="sm" @click="search">搜索</Button>
      </div>
    </div>

    <p class="text-sm text-muted-foreground mb-4">
      共 {{ total }} 张公开图片
      <span v-if="lastPage > 1"> · 第 {{ currentPage }}/{{ lastPage }} 页</span>
    </p>

    <!-- Loading -->
    <div v-if="loading" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
      <div v-for="i in 10" :key="i" class="aspect-square bg-muted animate-pulse rounded-lg" />
    </div>

    <!-- Image Grid -->
    <div v-else-if="images.length > 0" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
      <Card
        v-for="(img, idx) in images"
        :key="img.key"
        class="overflow-hidden cursor-pointer group"
        @click="openPreview(img, idx)"
      >
        <div class="aspect-square overflow-hidden bg-muted">
          <img
            :src="img.url"
            :alt="img.name"
            class="w-full h-full object-cover transition-transform duration-300 group-hover:scale-110"
            loading="lazy"
          />
        </div>
        <CardContent class="p-3">
          <p class="text-sm truncate font-medium">{{ img.name || img.origin_name }}</p>
          <p class="text-xs text-muted-foreground flex items-center justify-between mt-0.5">
            <span>{{ img.width }}x{{ img.height }}</span>
            <span>{{ formatSize(img.size) }}</span>
          </p>
        </CardContent>
      </Card>
    </div>

    <!-- Empty -->
    <div v-else class="text-center py-16 text-muted-foreground">
      <p class="text-lg mb-1">暂无公开图片</p>
      <p class="text-sm">目前还没有用户分享过图片</p>
    </div>

    <!-- Pagination -->
    <div v-if="lastPage > 1" class="flex items-center justify-center gap-1 mt-6">
      <Button variant="outline" size="sm" :disabled="currentPage <= 1" @click="goPage(currentPage - 1)">
        <ChevronLeft class="h-4 w-4" />
      </Button>
      <Button
        v-for="p in pageNumbers()"
        :key="p"
        :variant="p === currentPage ? 'default' : 'outline'"
        size="sm"
        @click="goPage(p)"
      >
        {{ p }}
      </Button>
      <Button variant="outline" size="sm" :disabled="currentPage >= lastPage" @click="goPage(currentPage + 1)">
        <ChevronRight class="h-4 w-4" />
      </Button>
    </div>

    <!-- Lightbox -->
    <Dialog :open="showPreview" @update:open="showPreview = $event">
      <DialogContent class="max-w-5xl p-0 overflow-hidden bg-black/95 border-0">
        <DialogClose class="absolute top-3 right-3 z-10 rounded-full bg-black/50 p-1.5 text-white hover:bg-black/70">
          <X class="h-5 w-5" />
        </DialogClose>

        <!-- Prev button -->
        <button
          class="absolute left-3 top-1/2 -translate-y-1/2 z-10 rounded-full bg-black/50 p-2 text-white hover:bg-black/70 transition disabled:opacity-20 disabled:cursor-not-allowed"
          :disabled="previewIndex <= 0"
          @click.stop="prevImage"
        >
          <ChevronLeft class="h-6 w-6" />
        </button>

        <!-- Next button -->
        <button
          class="absolute right-3 top-1/2 -translate-y-1/2 z-10 rounded-full bg-black/50 p-2 text-white hover:bg-black/70 transition disabled:opacity-20 disabled:cursor-not-allowed"
          :disabled="previewIndex >= images.length - 1"
          @click.stop="nextImage"
        >
          <ChevronRight class="h-6 w-6" />
        </button>

        <div v-if="previewImage" class="flex flex-col max-h-[85vh]">
          <div class="flex-1 flex items-center justify-center p-4 min-h-0">
            <img
              :src="previewImage.url"
              :alt="previewImage.name"
              class="max-w-full max-h-[70vh] object-contain rounded"
            />
          </div>
          <div class="bg-black/80 text-white px-6 py-4 space-y-1 border-t border-white/10">
            <div class="flex items-center justify-between">
              <p class="font-medium">{{ previewImage.name || previewImage.origin_name }}</p>
              <span class="text-xs text-white/40">{{ previewIndex + 1 }} / {{ images.length }}</span>
            </div>
            <p class="text-sm text-white/60">
              {{ previewImage.width }} × {{ previewImage.height }} ·
              {{ formatSize(previewImage.size) }} ·
              {{ previewImage.extension?.toUpperCase() }}
            </p>
            <a
              :href="previewImage.url"
              target="_blank"
              class="text-sm text-blue-400 hover:underline inline-block mt-1"
            >
              查看原图
            </a>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
