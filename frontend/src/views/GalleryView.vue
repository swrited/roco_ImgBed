<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import apiClient from '@/api/client'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog, DialogContent,
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { Search, ChevronLeft, ChevronRight, X, ExternalLink, Images, Folder, Download, Copy, RefreshCw } from 'lucide-vue-next'
import type { Image, Album, PaginatedResponse } from '@/types'

// View mode: 'albums' or 'images'
const viewMode = ref<'albums' | 'images'>('albums')
const activeAlbum = ref<Album | null>(null)

// Data
const albums = ref<Album[]>([])
const images = ref<Image[]>([])
const loading = ref(false)

// Pagination
const currentPage = ref(1)
const lastPage = ref(1)
const total = ref(0)
const perPage = 20
const keyword = ref('')

// Lightbox
const previewImage = ref<Image | null>(null)
const previewIndex = ref(0)
const showPreview = ref(false)

async function loadAlbums(page = 1) {
  loading.value = true
  viewMode.value = 'albums'
  activeAlbum.value = null
  try {
    const params: Record<string, any> = { page, per_page: perPage }
    if (keyword.value.trim()) params.q = keyword.value.trim()
    const res = await apiClient.get<PaginatedResponse<Album>>('/gallery', { params })
    albums.value = res.data || []
    currentPage.value = res.current_page || 1
    lastPage.value = res.last_page || 1
    total.value = res.total || 0
  } catch (e: any) {
    toast.error('加载画廊失败')
    albums.value = []
  } finally {
    loading.value = false
  }
}

async function loadImages(album: Album, page = 1) {
  loading.value = true
  viewMode.value = 'images'
  activeAlbum.value = album
  try {
    const params: Record<string, any> = { page, per_page: perPage }
    const res = await apiClient.get<any>(`/gallery/albums/${album.id}`, { params })
    images.value = res.data || []
    currentPage.value = res.current_page || 1
    lastPage.value = res.last_page || 1
    total.value = res.total || 0
  } catch (e: any) {
    toast.error('加载图片失败')
    images.value = []
  } finally {
    loading.value = false
  }
}

function goPage(page: number) {
  if (page < 1 || page > lastPage.value || page === currentPage.value) return
  if (viewMode.value === 'albums') {
    loadAlbums(page)
  } else {
    loadImages(activeAlbum.value!, page)
  }
}

function search() {
  if (viewMode.value === 'albums') {
    loadAlbums(1)
  }
}

const pageNumbers = (): number[] => {
  const pages: number[] = []
  const start = Math.max(1, currentPage.value - 2)
  const end = Math.min(lastPage.value, currentPage.value + 2)
  for (let i = start; i <= end; i++) pages.push(i)
  return pages
}

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
  else if (e.key === 'Escape') showPreview.value = false
}

function formatSize(kb: number): string {
  if (kb <= 0) return '0 KB'
  if (kb >= 1048576) return (kb / 1048576).toFixed(2) + ' GB'
  if (kb >= 1024) return (kb / 1024).toFixed(2) + ' MB'
  return Math.round(kb) + ' KB'
}

function downloadImage() {
  if (!previewImage.value) return
  const url = previewImage.value.links.url
  const a = document.createElement('a')
  a.href = url
  a.download = previewImage.value.alias_name || previewImage.value.origin_name
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

function copyLink(type: 'url' | 'markdown' | 'html' | 'bbcode') {
  if (!previewImage.value) return
  let text = ''
  switch(type) {
    case 'url': text = previewImage.value.links.url; break
    case 'markdown': text = previewImage.value.links.markdown; break
    case 'html': text = previewImage.value.links.html; break
    case 'bbcode': text = previewImage.value.links.bbcode; break
  }
  navigator.clipboard.writeText(text)
  toast.success('复制成功')
}

onMounted(() => {
  loadAlbums()
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div>
    <!-- Header -->
    <div class="animate-fade-in-up mb-8 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
      <div v-if="viewMode === 'albums'">
        <p class="text-xs font-semibold uppercase tracking-widest text-violet-400">Discover</p>
        <h1 class="mt-1.5 text-3xl font-semibold tracking-tight">探索图库</h1>
        <p class="mt-2 text-sm text-muted-foreground">
          共 <span class="font-medium text-slate-200 tabular-nums">{{ total }}</span> 个公开相册
          <span v-if="lastPage > 1"> · 第 {{ currentPage }}/{{ lastPage }} 页</span>
        </p>
      </div>
      <div v-else class="flex flex-col gap-2">
        <Button variant="ghost" size="sm" class="w-fit -ml-3 text-muted-foreground" @click="loadAlbums(1)">
          <ChevronLeft class="h-4 w-4 mr-1" />
          返回探索
        </Button>
        <h1 class="text-3xl font-semibold tracking-tight">{{ activeAlbum?.name }}</h1>
        <p class="mt-2 text-sm text-muted-foreground">
          共 <span class="font-medium text-slate-200 tabular-nums">{{ total }}</span> 张图片
          <span v-if="activeAlbum?.user_name"> · 作者: {{ activeAlbum.user_name }}</span>
        </p>
      </div>

      <div class="flex items-center gap-2" v-if="viewMode === 'albums'">
        <div class="relative">
          <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            v-model="keyword"
            placeholder="搜索相册..."
            class="pl-9 w-56"
            @keyup.enter="search"
          />
        </div>
        <Button variant="outline" size="sm" @click="search">搜索</Button>
      </div>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
      <div v-for="i in 12" :key="i" class="skeleton aspect-square rounded-xl" />
    </div>

    <!-- Albums grid -->
    <div v-else-if="viewMode === 'albums'">
      <div v-if="albums.length > 0" class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
        <div
          v-for="album in albums"
          :key="album.id"
          class="group relative cursor-pointer overflow-hidden rounded-xl bg-white/5 ring-1 ring-white/10 transition-all duration-300 hover:-translate-y-1 hover:ring-violet-500/50 hover:shadow-xl hover:shadow-violet-900/20"
          @click="loadImages(album, 1)"
        >
          <div class="aspect-square bg-black/40 overflow-hidden flex items-center justify-center">
            <img
              v-if="album.cover_url"
              :src="album.cover_url"
              class="w-full h-full object-cover opacity-80 transition-transform duration-500 group-hover:scale-105"
            />
            <Folder v-else class="h-16 w-16 text-white/20" />
          </div>
          <div class="absolute inset-x-0 bottom-0 bg-gradient-to-t from-black/90 to-transparent p-4 pt-12">
            <h3 class="font-semibold text-white truncate">{{ album.name }}</h3>
            <div class="mt-1 flex items-center gap-2 text-xs text-white/60">
              <span v-if="album.user_name" class="truncate max-w-[100px]">{{ album.user_name }}</span>
              <span class="w-1 h-1 rounded-full bg-white/30" v-if="album.user_name" />
              <span>{{ album.image_num }} 张</span>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="flex flex-col items-center justify-center py-24 text-muted-foreground">
        <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-3xl bg-white/5">
          <Folder class="h-8 w-8 text-muted-foreground/50" />
        </div>
        <p class="text-lg font-medium">暂无公开相册</p>
      </div>
    </div>

    <!-- Images grid -->
    <div v-else>
      <div v-if="images.length > 0" class="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        <div
          v-for="(img, idx) in images"
          :key="img.key"
          class="group relative cursor-pointer overflow-hidden rounded-xl bg-white/5 ring-1 ring-white/5 transition-all duration-300 hover:-translate-y-0.5 hover:ring-violet-500/25 hover:shadow-xl hover:shadow-black/40"
          @click="openPreview(img, idx)"
        >
          <div class="aspect-square overflow-hidden">
            <img
              :src="img.links.url"
              :alt="img.alias_name || img.origin_name"
              class="h-full w-full object-cover transition-transform duration-500 group-hover:scale-110"
              loading="lazy"
            />
          </div>
          <div class="absolute inset-0 flex flex-col justify-end bg-gradient-to-t from-black/80 via-black/20 to-transparent opacity-0 transition-opacity duration-300 group-hover:opacity-100">
            <div class="p-3">
              <p class="truncate text-xs font-medium text-white">{{ img.alias_name || img.origin_name }}</p>
            </div>
          </div>
        </div>
      </div>
      <div v-else class="flex flex-col items-center justify-center py-24 text-muted-foreground">
        <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-3xl bg-white/5">
          <Images class="h-8 w-8 text-muted-foreground/50" />
        </div>
        <p class="text-lg font-medium">相册为空</p>
      </div>
    </div>

    <!-- Pagination -->
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

    <!-- Lightbox -->
    <Dialog :open="showPreview" @update:open="showPreview = $event">
      <DialogContent :show-close-button="false" class="sm:max-w-5xl gap-0 overflow-hidden border-white/8 bg-[oklch(5%_0.005_270)]/98 p-0 backdrop-blur-2xl">
        <button class="absolute right-3 top-3 z-20 flex h-8 w-8 items-center justify-center rounded-full bg-white/8 text-white/70 transition-all hover:bg-white/15 hover:text-white" @click="showPreview = false">
          <X class="h-4 w-4" />
        </button>

        <button class="absolute left-3 top-1/2 z-20 -translate-y-1/2 flex h-9 w-9 items-center justify-center rounded-full bg-white/8 text-white/70 backdrop-blur-sm transition-all hover:bg-white/15 hover:text-white disabled:pointer-events-none disabled:opacity-20" :disabled="previewIndex <= 0" @click.stop="prevImage">
          <ChevronLeft class="h-5 w-5" />
        </button>
        <button class="absolute right-3 top-1/2 z-20 -translate-y-1/2 flex h-9 w-9 items-center justify-center rounded-full bg-white/8 text-white/70 backdrop-blur-sm transition-all hover:bg-white/15 hover:text-white disabled:pointer-events-none disabled:opacity-20" :disabled="previewIndex >= images.length - 1" @click.stop="nextImage">
          <ChevronRight class="h-5 w-5" />
        </button>

        <div v-if="previewImage" class="flex max-h-[88vh] flex-col">
          <div class="flex flex-1 items-center justify-center p-6 min-h-0">
            <img :src="previewImage.links.url" class="max-h-[70vh] max-w-full rounded-lg object-contain shadow-2xl shadow-black/60" />
          </div>

          <div class="border-t border-white/8 bg-white/[0.03] px-14 py-4">
            <div class="flex items-center justify-between gap-4">
              <div class="min-w-0">
                <p class="truncate font-medium text-white">{{ previewImage.alias_name || previewImage.origin_name }}</p>
                <div class="mt-2 flex flex-wrap gap-2">
                  <span v-for="tag in previewImage.tags" :key="tag.id" class="px-2 py-0.5 rounded-full bg-violet-500/20 text-violet-300 text-[10px]">
                    #{{ tag.name }}
                  </span>
                </div>
              </div>
              <div class="flex shrink-0 items-center gap-2">
                <Button variant="outline" size="sm" class="h-8 gap-1 border-white/10 bg-white/5 hover:bg-white/10" @click="downloadImage">
                  <Download class="h-3.5 w-3.5" />
                  下载图片
                </Button>
                <div class="flex items-center rounded-lg border border-white/10 bg-white/5 p-1">
                  <Button variant="ghost" size="sm" class="h-6 px-2 text-xs hover:bg-white/10" @click="copyLink('url')">URL</Button>
                  <Button variant="ghost" size="sm" class="h-6 px-2 text-xs hover:bg-white/10" @click="copyLink('markdown')">MD</Button>
                  <Button variant="ghost" size="sm" class="h-6 px-2 text-xs hover:bg-white/10" @click="copyLink('html')">HTML</Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>
