<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import apiClient from '@/api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog, DialogContent,
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { Search, ChevronLeft, ChevronRight, X, ExternalLink, Images } from 'lucide-vue-next'
import type { Image, PaginatedResponse } from '@/types'

const images       = ref<Image[]>([])
const loading      = ref(false)
const currentPage  = ref(1)
const lastPage     = ref(1)
const total        = ref(0)
const perPage      = 20
const keyword      = ref('')

const previewImage = ref<Image | null>(null)
const previewIndex = ref(0)
const showPreview  = ref(false)

function openPreview(img: Image, idx: number) {
  previewImage.value = img
  previewIndex.value = idx
  showPreview.value  = true
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
  if (e.key === 'ArrowLeft')  prevImage()
  else if (e.key === 'ArrowRight') nextImage()
  else if (e.key === 'Escape') showPreview.value = false
}

function formatSize(kb: number): string {
  if (kb <= 0) return '0 KB'
  if (kb >= 1048576) return (kb / 1048576).toFixed(2) + ' GB'
  if (kb >= 1024)    return (kb / 1024).toFixed(2) + ' MB'
  return Math.round(kb) + ' KB'
}

async function loadImages(page = 1) {
  loading.value = true
  try {
    const params: Record<string, any> = { page, per_page: perPage }
    if (keyword.value.trim()) params.keyword = keyword.value.trim()
    const res = await apiClient.get<PaginatedResponse<Image>>('/gallery', { params })
    images.value      = res.data || []
    currentPage.value = res.current_page || 1
    lastPage.value    = res.last_page || 1
    total.value       = res.total || 0
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

function search() { loadImages(1) }

const pageNumbers = (): number[] => {
  const pages: number[] = []
  const start = Math.max(1, currentPage.value - 2)
  const end   = Math.min(lastPage.value, currentPage.value + 2)
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
    <!-- ── Header ─────────────────────────────────────────────────────── -->
    <div class="animate-fade-in-up mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-xs font-semibold uppercase tracking-widest text-violet-400">Gallery</p>
        <h1 class="mt-1.5 text-3xl font-semibold tracking-tight">画廊</h1>
        <p class="mt-2 text-sm text-muted-foreground">
          共 <span class="font-medium text-slate-200 tabular-nums">{{ total }}</span> 张公开图片
          <span v-if="lastPage > 1"> · 第 {{ currentPage }}/{{ lastPage }} 页</span>
        </p>
      </div>
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

    <!-- ── Loading skeleton ─────────────────────────────────────────────── -->
    <div v-if="loading" class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
      <div
        v-for="i in 15"
        :key="i"
        class="skeleton aspect-square rounded-xl"
      />
    </div>

    <!-- ── Image grid ────────────────────────────────────────────────────── -->
    <div
      v-else-if="images.length > 0"
      class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5"
    >
      <div
        v-for="(img, idx) in images"
        :key="img.key"
        class="group relative cursor-pointer overflow-hidden rounded-xl bg-[oklch(9%_0.008_270)] ring-1 ring-white/5 transition-all duration-300 hover:-translate-y-0.5 hover:ring-violet-500/25 hover:shadow-xl hover:shadow-black/40"
        @click="openPreview(img, idx)"
      >
        <!-- Image -->
        <div class="aspect-square overflow-hidden">
          <img
            :src="img.url"
            :alt="img.alias_name || img.origin_name"
            class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-110"
            loading="lazy"
          />
        </div>

        <!-- Hover overlay with metadata -->
        <div class="absolute inset-0 flex flex-col justify-end bg-gradient-to-t from-black/80 via-black/20 to-transparent opacity-0 transition-opacity duration-300 group-hover:opacity-100">
          <div class="p-3">
            <p class="truncate text-xs font-medium text-white">{{ img.alias_name || img.origin_name }}</p>
            <p class="mt-0.5 flex items-center justify-between text-xs text-white/60">
              <span>{{ img.width }}×{{ img.height }}</span>
              <span>{{ formatSize(img.size) }}</span>
            </p>
          </div>
        </div>

        <!-- Top-right zoom indicator -->
        <div class="absolute right-2 top-2 flex h-7 w-7 items-center justify-center rounded-full bg-black/50 opacity-0 backdrop-blur-sm transition-opacity duration-300 group-hover:opacity-100">
          <ExternalLink class="h-3.5 w-3.5 text-white" />
        </div>
      </div>
    </div>

    <!-- ── Empty state ────────────────────────────────────────────────────── -->
    <div v-else class="flex flex-col items-center justify-center py-24 text-muted-foreground">
      <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-3xl bg-white/5">
        <Images class="h-8 w-8 text-muted-foreground/50" />
      </div>
      <p class="text-lg font-medium">暂无公开图片</p>
      <p class="mt-1 text-sm">目前还没有用户分享过图片</p>
    </div>

    <!-- ── Pagination ────────────────────────────────────────────────────── -->
    <div v-if="lastPage > 1" class="mt-8 flex items-center justify-center gap-1">
      <Button variant="outline" size="sm" :disabled="currentPage <= 1" @click="goPage(currentPage - 1)">
        <ChevronLeft class="h-4 w-4" />
      </Button>
      <Button
        v-for="p in pageNumbers()"
        :key="p"
        :variant="p === currentPage ? 'default' : 'outline'"
        size="sm"
        class="w-9"
        @click="goPage(p)"
      >
        {{ p }}
      </Button>
      <Button variant="outline" size="sm" :disabled="currentPage >= lastPage" @click="goPage(currentPage + 1)">
        <ChevronRight class="h-4 w-4" />
      </Button>
    </div>

    <!-- ── Lightbox ─────────────────────────────────────────────────────── -->
    <Dialog :open="showPreview" @update:open="showPreview = $event">
      <!--
        :show-close-button="false" 禁用 DialogContent 自带的 X 按钮，
        避免与我们自定义的关闭按钮重复出现。
      -->
      <DialogContent
        :show-close-button="false"
        class="sm:max-w-5xl gap-0 overflow-hidden border-white/8 bg-[oklch(5%_0.005_270)]/98 p-0 backdrop-blur-2xl"
      >
        <!-- Custom close button (top-right, above the image) -->
        <button
          class="absolute right-3 top-3 z-20 flex h-8 w-8 items-center justify-center rounded-full bg-white/8 text-white/70 transition-all hover:bg-white/15 hover:text-white"
          @click="showPreview = false"
        >
          <X class="h-4 w-4" />
        </button>

        <!-- Prev -->
        <button
          class="absolute left-3 top-1/2 z-20 -translate-y-1/2 flex h-9 w-9 items-center justify-center rounded-full bg-white/8 text-white/70 backdrop-blur-sm transition-all hover:bg-white/15 hover:text-white disabled:pointer-events-none disabled:opacity-20"
          :disabled="previewIndex <= 0"
          @click.stop="prevImage"
        >
          <ChevronLeft class="h-5 w-5" />
        </button>

        <!-- Next -->
        <button
          class="absolute right-3 top-1/2 z-20 -translate-y-1/2 flex h-9 w-9 items-center justify-center rounded-full bg-white/8 text-white/70 backdrop-blur-sm transition-all hover:bg-white/15 hover:text-white disabled:pointer-events-none disabled:opacity-20"
          :disabled="previewIndex >= images.length - 1"
          @click.stop="nextImage"
        >
          <ChevronRight class="h-5 w-5" />
        </button>

        <div v-if="previewImage" class="flex max-h-[88vh] flex-col">
          <!-- Image -->
          <div class="flex flex-1 items-center justify-center p-6 min-h-0">
            <img
              :src="previewImage.url"
              :alt="previewImage.alias_name || previewImage.origin_name"
              class="max-h-[70vh] max-w-full rounded-lg object-contain shadow-2xl shadow-black/60"
            />
          </div>

          <!-- Meta footer — px-14 给左右导航箭头留出空间，避免遮挡内容 -->
          <div class="border-t border-white/8 bg-white/[0.03] px-14 py-4">
            <div class="flex items-center justify-between gap-4">
              <div class="min-w-0">
                <p class="truncate font-medium text-white">
                  {{ previewImage.alias_name || previewImage.origin_name }}
                </p>
                <p class="mt-0.5 text-sm text-white/50">
                  {{ previewImage.width }} × {{ previewImage.height }} ·
                  {{ formatSize(previewImage.size) }} ·
                  {{ previewImage.extension?.toUpperCase() }}
                </p>
              </div>
              <div class="flex shrink-0 items-center gap-3">
                <span class="text-xs text-white/30 tabular-nums">
                  {{ previewIndex + 1 }} / {{ images.length }}
                </span>
                <a
                  :href="previewImage.url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="flex items-center gap-1.5 rounded-lg border border-white/10 bg-white/5 px-3 py-1.5 text-xs text-white/70 transition-all hover:border-white/20 hover:bg-white/10 hover:text-white"
                  @click.stop
                >
                  <ExternalLink class="h-3.5 w-3.5" />
                  查看原图
                </a>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
