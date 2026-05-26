<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { imagesApi } from '@/api/images'
import { albumsApi } from '@/api/albums'
import { tagsApi } from '@/api/tags'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent } from '@/components/ui/card'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import {
  Pagination, PaginationContent, PaginationItem, PaginationLink,
  PaginationNext, PaginationPrevious,
} from '@/components/ui/pagination'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { toast } from 'vue-sonner'
import { MoreHorizontal, Copy, Trash2, Pencil, FolderInput, Search, Link2, Check, Download, X, Hash, ArrowLeft, Plus, Info, RefreshCw } from 'lucide-vue-next'
import type { Album, Image, Tag } from '@/types'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { copyToClipboard } from '@/utils/clipboard'

const route = useRoute()

const images = ref<Image[]>([])
const albums = ref<Album[]>([])
const tags = ref<Tag[]>([])
const currentPage = ref(1)
const lastPage = ref(1)
const total = ref(0)
const loading = ref(false)
const selectedKeys = ref<string[]>([])

const keyword = ref('')
const filterAlbumId = ref('__all__')
const sortOrder = ref('newest')
const filterTagId = ref('__all__')

const showMoveDialog = ref(false)
const showRenameDialog = ref(false)
const showTagsDialog = ref(false)
const showLinksDialog = ref(false)
const showPreviewDialog = ref(false)
const previewImage = ref<Image | null>(null)
const renameKey = ref('')
const renameName = ref('')
const editingKey = ref('')
const editingTags = ref<string[]>([])
const newTagInput = ref('')
const isEditingPreviewTags = ref(false)
const newPreviewTag = ref('')
const moveAlbumId = ref<string>('__none__')
const moveKeys = ref<string[]>([])
const quickAlbumName = ref('')
const quickAlbumPermission = ref('0')
const creatingAlbum = ref(false)
const linkFormat = ref<'url' | 'markdown' | 'html' | 'bbcode'>('url')
const previewLinkFormat = ref<'url' | 'markdown' | 'html' | 'bbcode'>('url')
const resettingPrivateLink = ref(false)

const selectedImages = computed(() => {
  const selected = new Set(selectedKeys.value)
  return images.value.filter((image) => selected.has(image.key))
})

const currentAlbumTitle = computed(() => {
  if (filterAlbumId.value === '0') return '未分类图片'
  if (filterAlbumId.value !== '__all__') {
    return albums.value.find((album) => String(album.id) === filterAlbumId.value)?.name || '相册图片'
  }
  return '所有图片'
})

const generatedLinks = computed(() => {
  return selectedImages.value
    .map((image) => formatImageLink(image, linkFormat.value))
    .join('\n')
})

async function loadImages(page = 1) {
  loading.value = true
  selectedKeys.value = []
  try {
    const params: Record<string, any> = { page }
    if (keyword.value.trim()) params.q = keyword.value.trim()
    if (filterAlbumId.value && filterAlbumId.value !== '__all__') params.album_id = filterAlbumId.value
    if (filterTagId.value && filterTagId.value !== '__all__') params.tag_id = filterTagId.value
    params.sort = sortOrder.value

    const res = await imagesApi.list(params)
    images.value = res?.data ?? []
    currentPage.value = res.current_page
    lastPage.value = res.last_page
    total.value = res.total
  } catch {
    toast.error('加载图片失败')
  } finally {
    loading.value = false
  }
}

async function loadAlbums() {
  try {
    albums.value = await albumsApi.list()
  } catch { /**/ }
}

async function loadTags() {
  try {
    tags.value = await tagsApi.list()
  } catch { /**/ }
}

function toggleSelect(key: string) {
  const idx = selectedKeys.value.indexOf(key)
  if (idx >= 0) selectedKeys.value.splice(idx, 1)
  else selectedKeys.value.push(key)
}

function openPreview(image: Image) {
  previewImage.value = image
  previewLinkFormat.value = 'url'
  showPreviewDialog.value = true
}

function copyLink(url: string) {
  copyToClipboard(url)
  toast.success('链接已复制')
}

function getImageAlt(image: Image): string {
  return imageDisplayName(image)
}

function imageDisplayName(image: Image): string {
  const alias = String(image.alias_name || '').trim()
  const origin = String(image.origin_name || '').trim()
  const name = String(image.name || '').trim()
  return alias || origin || name || image.key
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

function formatImageLink(image: Image, format: typeof linkFormat.value): string {
  const url = image.url
  const alt = getImageAlt(image)

  if (format === 'markdown') return `![${alt}](${url})`
  if (format === 'html') return `<img src="${url}" alt="${escapeHtml(alt)}" />`
  if (format === 'bbcode') return `[img]${url}[/img]`
  return url
}

function openLinksDialog() {
  if (selectedImages.value.length === 0) {
    toast.error('请先选择图片')
    return
  }
  showLinksDialog.value = true
}

function copyGeneratedLinks() {
  if (!generatedLinks.value) {
    toast.error('没有可复制的链接')
    return
  }
  copyToClipboard(generatedLinks.value)
  toast.success('链接已复制到剪贴板')
}

function copyFormattedLink(img: Image, format: 'url' | 'markdown' | 'html' | 'bbcode') {
  let text = ''
  if (format === 'url') {
    text = img.url
  } else if (format === 'markdown') {
    text = `![${img.alias_name || img.origin_name}](${img.url})`
  } else if (format === 'html') {
    text = `<img src="${img.url}" alt="${img.alias_name || img.origin_name}" />`
  } else if (format === 'bbcode') {
    text = `[img]${img.url}[/img]`
  }
  copyToClipboard(text)
  toast.success('链接已复制到剪贴板')
}

async function resetPrivateLink() {
  if (!previewImage.value || previewImage.value.permission === 1) return
  if (!confirm('重置后，之前复制的私密图片链接将立即失效。确定继续吗？')) return
  resettingPrivateLink.value = true
  try {
    const updated = await imagesApi.resetPrivateLink(previewImage.value.key)
    previewImage.value = { ...previewImage.value, ...updated }
    const index = images.value.findIndex((image) => image.key === updated.key)
    if (index >= 0) images.value[index] = { ...images.value[index], ...updated }
    toast.success('私密图片链接已重置')
  } catch (e: any) {
    toast.error(e.message || '重置私密链接失败')
  } finally {
    resettingPrivateLink.value = false
  }
}

const showDeleteConfirmDialog = ref(false)
const deleteConfirmKeys = ref<string[]>([])

function confirmDeletePreview(img: Image) {
  deleteConfirmKeys.value = [img.key]
  showDeleteConfirmDialog.value = true
}

async function executeDelete() {
  if (deleteConfirmKeys.value.length === 0) return
  try {
    await imagesApi.delete(deleteConfirmKeys.value)
    toast.success('已移至回收站')
    showDeleteConfirmDialog.value = false
    showPreviewDialog.value = false
    selectedKeys.value = []
    loadImages(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '删除失败')
  }
}

async function deleteImages(keys: string[]) {
  deleteConfirmKeys.value = keys
  showDeleteConfirmDialog.value = true
}

function openRename(key: string, name: string) {
  renameKey.value = key
  renameName.value = name
  showRenameDialog.value = true
}

async function handleRename() {
  if (!renameName.value.trim()) {
    toast.error('请输入新名称')
    return
  }
  try {
    await imagesApi.rename(renameKey.value, renameName.value.trim())
    toast.success('重命名成功')
    showRenameDialog.value = false
    if (previewImage.value?.key === renameKey.value) {
      previewImage.value.alias_name = renameName.value.trim()
    }
    loadImages(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '重命名失败')
  }
}

function openTags(key: string, tags: any[]) {
  editingKey.value = key
  editingTags.value = (tags || []).map(t => t.name)
  newTagInput.value = ''
  showTagsDialog.value = true
}

function addTag() {
  const val = newTagInput.value.trim()
  if (val && editingTags.value.length < 5 && !editingTags.value.includes(val)) {
    editingTags.value.push(val)
  }
  newTagInput.value = ''
}

async function addTagAndSave() {
  addTag()
  await updateTags({ close: false })
}

function removeTag(idx: number) {
  editingTags.value.splice(idx, 1)
}

async function removeTagAndSave(idx: number) {
  removeTag(idx)
  await updateTags({ close: false })
}

async function updateTags(options: { close?: boolean } = {}) {
  if (!editingKey.value) return
  try {
    await imagesApi.updateTags(editingKey.value, editingTags.value)
    toast.success('标签更新成功')
    if (options.close !== false) showTagsDialog.value = false
    loadImages(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '更新标签失败')
  }
}

function startPreviewTagEdit() {
  if (!previewImage.value) return
  editingTags.value = (previewImage.value.tags || []).map(t => t.name)
  newPreviewTag.value = ''
  isEditingPreviewTags.value = true
}

function addPreviewTag() {
  const val = newPreviewTag.value.trim()
  if (val && editingTags.value.length < 5 && !editingTags.value.includes(val)) {
    editingTags.value.push(val)
  }
  newPreviewTag.value = ''
}

async function addPreviewTagAndSave() {
  addPreviewTag()
  await savePreviewTags({ keepEditing: true })
}

async function removePreviewTagAndSave(idx: number) {
  removeTag(idx)
  await savePreviewTags({ keepEditing: true })
}

async function savePreviewTags(options: { keepEditing?: boolean } = {}) {
  if (!previewImage.value) return
  try {
    await imagesApi.updateTags(previewImage.value.key, editingTags.value)
    toast.success('标签更新成功')
    if (!options.keepEditing) isEditingPreviewTags.value = false
    // Refresh the list in the background
    loadImages(currentPage.value)
    // We ideally want to update previewImage tags without reloading the dialog.
    previewImage.value.tags = editingTags.value.map((t, id) => ({ id, name: t, image_num: 0, created_at: '' }))
  } catch (e: any) {
    toast.error(e.message || '更新标签失败')
  }
}

const previewAlbumName = computed(() => {
  if (!previewImage.value?.album_id) return '未分类图片'
  return albums.value.find((album) => album.id === previewImage.value?.album_id)?.name || '未知相册'
})

function openMove(keys?: string[]) {
  moveKeys.value = keys && keys.length > 0 ? keys : [...selectedKeys.value]
  moveAlbumId.value = '__none__'
  quickAlbumName.value = ''
  quickAlbumPermission.value = '0'
  showMoveDialog.value = true
}

async function createAlbumForMove() {
  const name = quickAlbumName.value.trim()
  if (!name) {
    toast.error('请输入相册名称')
    return
  }
  creatingAlbum.value = true
  try {
    const album = await albumsApi.create({
      name,
      intro: '',
      permission: Number(quickAlbumPermission.value),
    })
    toast.success('相册已创建')
    await loadAlbums()
    moveAlbumId.value = String(album.id)
    quickAlbumName.value = ''
  } catch (e: any) {
    toast.error(e.message || '创建相册失败')
  } finally {
    creatingAlbum.value = false
  }
}

async function handleMove() {
  if (moveKeys.value.length === 0) {
    toast.error('请先选择要移动的图片')
    return
  }
  if (moveAlbumId.value === '__create__') {
    toast.error('请先创建相册，或选择已有目标相册')
    return
  }
  try {
    const albumId = moveAlbumId.value && moveAlbumId.value !== '__none__' ? Number(moveAlbumId.value) : null
    await imagesApi.move(moveKeys.value, albumId)
    toast.success('移动成功')
    showMoveDialog.value = false
    if (previewImage.value && moveKeys.value.includes(previewImage.value.key)) {
      previewImage.value.album_id = albumId ?? undefined
    }
    selectedKeys.value = []
    moveKeys.value = []
    loadImages(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '移动失败')
  }
}

function downloadImage(url: string, name: string) {
  const a = document.createElement('a')
  a.href = url
  a.download = name
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

function formatSize(kb: number): string {
  if (kb >= 1048576) return (kb / 1048576).toFixed(2) + ' GB'
  if (kb >= 1024) return (kb / 1024).toFixed(2) + ' MB'
  return kb.toFixed(2) + ' KB'
}

onMounted(() => {
  if (route.query.album_id) {
    filterAlbumId.value = String(route.query.album_id)
  }
  if (route.query.tag_id) {
    filterTagId.value = String(route.query.tag_id)
  }
  loadImages()
  loadAlbums()
  loadTags()
})

watch(
  () => route.query,
  (query) => {
    filterAlbumId.value = query.album_id ? String(query.album_id) : '__all__'
    filterTagId.value = query.tag_id ? String(query.tag_id) : '__all__'
    loadImages()
  },
)
</script>

<template>
  <div>
    <div class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <Button variant="ghost" class="-ml-2 mb-2" @click="$router.push('/library')">
          <ArrowLeft class="mr-2 h-4 w-4" /> 返回相册
        </Button>
        <h1 class="text-2xl font-semibold">{{ currentAlbumTitle }}</h1>
        <p class="mt-1 text-sm text-muted-foreground">可以通过标签继续筛选当前图片列表。</p>
      </div>
      <div class="flex items-center gap-2" v-if="selectedKeys.length > 0">
        <Badge variant="secondary">{{ selectedKeys.length }} 张已选</Badge>
        <Button variant="outline" size="sm" @click="openLinksDialog">
          <Link2 class="mr-1 h-4 w-4" /> 生成链接
        </Button>
        <Button variant="outline" size="sm" @click="openMove">
          <FolderInput class="mr-1 h-4 w-4" /> 移动
        </Button>
        <Button variant="destructive" size="sm" @click="deleteImages(selectedKeys)">
          <Trash2 class="mr-1 h-4 w-4" /> 删除
        </Button>
      </div>
    </div>

    <!-- Filter bar -->
    <div class="flex flex-wrap gap-2 mb-5">
      <div class="relative flex-1 min-w-48">
        <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
        <Input
          v-model="keyword"
          placeholder="搜索图片名称..."
          class="pl-9"
          @keyup.enter="loadImages()"
        />
      </div>
      <Select v-model="filterAlbumId" @update:model-value="loadImages()">
        <SelectTrigger class="w-40">
          <SelectValue placeholder="全部相册" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="__all__">全部相册</SelectItem>
          <SelectItem value="0">未分类图片</SelectItem>
          <SelectItem v-for="a in albums" :key="a.id" :value="String(a.id)">
            {{ a.name }}
          </SelectItem>
        </SelectContent>
      </Select>
      <Select v-model="filterTagId" @update:model-value="loadImages()">
        <SelectTrigger class="w-36">
          <SelectValue placeholder="全部标签" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="__all__">全部标签</SelectItem>
          <SelectItem v-for="tag in tags" :key="tag.id" :value="String(tag.id)">
            #{{ tag.name }}
          </SelectItem>
        </SelectContent>
      </Select>
      <Select v-model="sortOrder" @update:model-value="loadImages()">
        <SelectTrigger class="w-32">
          <SelectValue placeholder="最新上传" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="newest">最新上传</SelectItem>
          <SelectItem value="earliest">最早上传</SelectItem>
          <SelectItem value="utmost">文件最大</SelectItem>
          <SelectItem value="least">文件最小</SelectItem>
        </SelectContent>
      </Select>
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
        class="cursor-zoom-in group overflow-hidden transition-all !ring-transparent hover:!ring-white/20 hover:!ring-1"
        :class="selectedKeys.includes(img.key) ? 'scale-[0.985] shadow-2xl shadow-primary/25 bg-primary/10 !ring-primary/50' : ''"
        @click="openPreview(img)"
      >
        <div class="relative aspect-square overflow-hidden rounded-t-lg">
          <img :src="img.url" :alt="imageDisplayName(img)" class="h-full w-full object-cover" loading="lazy" />
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
          <div class="absolute top-2 right-2 z-30 opacity-0 group-hover:opacity-100 transition-opacity">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="secondary" size="icon" class="h-8 w-8" @click.stop>
                  <MoreHorizontal class="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" @click.stop>
                <DropdownMenuItem @click.stop="copyLink(img.url)">
                  <Copy class="mr-2 h-4 w-4" /> 复制链接
                </DropdownMenuItem>
                <DropdownMenuItem @click.stop="openRename(img.key, imageDisplayName(img))">
                  <Pencil class="mr-2 h-4 w-4" /> 重命名
                </DropdownMenuItem>
                <DropdownMenuItem @click.stop="openTags(img.key, img.tags || [])">
                  <Hash class="mr-2 h-4 w-4" /> 编辑标签
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem class="text-destructive" @click.stop="deleteImages([img.key])">
                  <Trash2 class="mr-2 h-4 w-4" /> 删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
        <CardContent class="p-3 transition-colors" :class="selectedKeys.includes(img.key) ? 'bg-primary/15' : ''">
          <p class="text-sm truncate font-medium">{{ imageDisplayName(img) }}</p>
          <p class="text-xs text-muted-foreground">
            {{ img.width }}x{{ img.height }} · {{ formatSize(img.size) }}
          </p>
          <div v-if="img.tags && img.tags.length > 0" class="flex flex-wrap gap-1 mt-1.5 h-[18px] overflow-hidden">
            <span v-for="tag in img.tags" :key="tag.id" class="px-1.5 py-[1px] rounded bg-violet-500/20 text-violet-300 text-[10px] leading-none whitespace-nowrap flex items-center">
              {{ tag.name }}
            </span>
          </div>
        </CardContent>
      </Card>
    </div>

    <div v-if="images.length === 0 && !loading" class="text-center py-12 text-muted-foreground">
      暂无图片
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

    <!-- Tags dialog -->
    <Dialog :open="showTagsDialog" @update:open="showTagsDialog = $event">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>编辑标签</DialogTitle>
          <DialogDescription>为该图片设置标签，多个标签请用逗号分隔（最多 5 个）。</DialogDescription>
        </DialogHeader>
        <div class="py-4 space-y-4">
          <div class="flex flex-wrap gap-2">
            <div v-for="(tag, idx) in editingTags" :key="idx" class="flex items-center gap-1 px-3 py-1 rounded-full bg-violet-500/20 text-violet-300 text-sm">
              #{{ tag }}
              <X class="h-3 w-3 cursor-pointer hover:text-red-400 ml-1" @click="removeTagAndSave(idx)" />
            </div>
            <div v-if="editingTags.length === 0" class="text-sm text-muted-foreground py-1">暂无标签</div>
          </div>
          <Input 
            v-model="newTagInput" 
            placeholder="输入标签后按回车添加..." 
            @keyup.enter="addTagAndSave"
            :disabled="editingTags.length >= 5"
          />
          <p class="text-xs text-muted-foreground" v-if="editingTags.length >= 5">最多只能添加 5 个标签</p>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showTagsDialog = false">关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Link Dialog -->
    <Dialog v-model:open="showLinksDialog">
      <DialogContent class="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>生成图片链接</DialogTitle>
          <DialogDescription>
            已选 {{ selectedImages.length }} 张图片，选择格式后复制即可使用。
          </DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="link-format">链接格式</Label>
            <Select v-model="linkFormat">
              <SelectTrigger id="link-format">
                <SelectValue placeholder="选择链接格式" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="url">直链 URL</SelectItem>
                <SelectItem value="markdown">Markdown</SelectItem>
                <SelectItem value="html">HTML</SelectItem>
                <SelectItem value="bbcode">BBCode</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <Textarea
            class="min-h-56 font-mono text-xs"
            readonly
            :model-value="generatedLinks"
          />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showLinksDialog = false">关闭</Button>
          <Button @click="copyGeneratedLinks">
            <Copy class="mr-2 h-4 w-4" /> 复制链接
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Preview Dialog -->
    <Dialog v-model:open="showPreviewDialog">
      <DialogContent class="sm:max-w-6xl">
        <DialogHeader class="min-w-0">
          <DialogTitle class="break-all leading-6">{{ previewImage ? imageDisplayName(previewImage) : '图片详情' }}</DialogTitle>
          <DialogDescription>查看大图、基础信息、标签和常用操作。</DialogDescription>
        </DialogHeader>
        <div v-if="previewImage" class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_340px]">
          <div class="overflow-hidden rounded-2xl border border-white/10 bg-black">
            <img
              :src="previewImage.url"
              :alt="getImageAlt(previewImage)"
              class="max-h-[72vh] w-full object-contain"
            />
          </div>

          <aside class="min-w-0 space-y-4">
            <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
              <div class="mb-3 flex items-center gap-2 text-sm font-semibold text-slate-200">
                <Info class="h-4 w-4 text-violet-300" />
                图片信息
              </div>
              <div class="space-y-3 text-sm">
                <div>
                  <p class="text-xs text-muted-foreground">显示名称</p>
                  <p class="mt-1 break-all font-medium leading-6">{{ imageDisplayName(previewImage) }}</p>
                </div>
                <div class="grid grid-cols-2 gap-3">
                  <div class="rounded-xl bg-black/20 p-3">
                    <p class="text-xs text-muted-foreground">尺寸</p>
                    <p class="mt-1 font-medium">{{ previewImage.width }}x{{ previewImage.height }}</p>
                  </div>
                  <div class="rounded-xl bg-black/20 p-3">
                    <p class="text-xs text-muted-foreground">大小</p>
                    <p class="mt-1 font-medium">{{ formatSize(previewImage.size) }}</p>
                  </div>
                </div>
                <div class="rounded-xl bg-black/20 p-3">
                  <p class="text-xs text-muted-foreground">所在相册</p>
                  <p class="mt-1 font-medium">{{ previewAlbumName }}</p>
                </div>
              </div>
            </div>

            <div class="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
              <div class="mb-3 flex items-center justify-between">
                <p class="text-sm font-semibold text-slate-200">标签</p>
                <Button v-if="!isEditingPreviewTags" variant="link" class="h-6 px-0 text-xs text-violet-400" @click="startPreviewTagEdit">
                  编辑标签
                </Button>
              </div>
              <div v-if="isEditingPreviewTags" class="space-y-3">
                <div class="flex flex-wrap gap-2">
                  <div v-for="(tag, idx) in editingTags" :key="idx" class="flex items-center gap-1 rounded-full bg-violet-500/30 px-2 py-0.5 text-xs text-violet-100">
                    #{{ tag }}
                    <X class="ml-0.5 h-3 w-3 cursor-pointer hover:text-red-300" @click="removePreviewTagAndSave(idx)" />
                  </div>
                </div>
                <Input 
                  v-model="newPreviewTag" 
                  placeholder="输入标签后按回车..." 
                  @keyup.enter="addPreviewTagAndSave"
                  class="h-8 text-xs bg-black/20 border-white/10"
                  :disabled="editingTags.length >= 5"
                />
                <div class="flex gap-2">
                  <Button size="sm" variant="ghost" class="h-8 text-xs" @click="isEditingPreviewTags = false">完成</Button>
                </div>
              </div>
              <div v-else-if="previewImage.tags && previewImage.tags.length > 0" class="flex flex-wrap gap-2">
                <button
                  v-for="tag in previewImage.tags"
                  :key="tag.id"
                  type="button"
                  class="rounded-full bg-violet-500/20 px-2 py-0.5 text-xs text-violet-200 transition hover:bg-violet-500/30"
                  @click="filterTagId = String(tag.id); showPreviewDialog = false; loadImages()"
                >
                  #{{ tag.name }}
                </button>
              </div>
              <p v-else class="text-sm text-muted-foreground">暂无标签</p>
            </div>

            <div class="grid grid-cols-2 gap-2">
              <Button variant="outline" @click="openRename(previewImage.key, imageDisplayName(previewImage))">
                <Pencil class="mr-2 h-4 w-4" /> 重命名
              </Button>
              <Button variant="outline" @click="openMove([previewImage.key])">
                <FolderInput class="mr-2 h-4 w-4" /> 移动
              </Button>
            </div>
            <Button class="w-full" @click="downloadImage(previewImage.url, imageDisplayName(previewImage))">
              <Download class="mr-2 h-4 w-4" /> 下载图片
            </Button>
            <div class="grid grid-cols-[minmax(0,1fr)_auto] gap-2">
              <Select v-model="previewLinkFormat">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="url">直链 URL</SelectItem>
                  <SelectItem value="markdown">Markdown</SelectItem>
                  <SelectItem value="html">HTML</SelectItem>
                  <SelectItem value="bbcode">BBCode</SelectItem>
                </SelectContent>
              </Select>
              <Button variant="outline" @click="copyFormattedLink(previewImage!, previewLinkFormat)">
                <Copy class="mr-2 h-4 w-4" /> 复制链接
              </Button>
            </div>
            <div v-if="previewImage.permission !== 1" class="rounded-xl border border-white/10 bg-white/[0.03] p-3">
              <p class="mb-3 text-xs leading-5 text-muted-foreground">
                私密图片链接可直接展示图片。链接泄露时可重置，之前的链接会立即失效。
              </p>
              <Button variant="outline" class="w-full" :disabled="resettingPrivateLink" @click="resetPrivateLink">
                <RefreshCw class="mr-2 h-4 w-4" :class="{ 'animate-spin': resettingPrivateLink }" />
                {{ resettingPrivateLink ? '正在重置...' : '重置私密链接' }}
              </Button>
            </div>
            <Button variant="destructive" class="w-full mt-2" @click="confirmDeletePreview(previewImage!)">
              <Trash2 class="mr-2 h-4 w-4" /> 移至回收站
            </Button>
          </aside>
        </div>
      </DialogContent>
    </Dialog>

    <!-- Rename Dialog -->
    <Dialog v-model:open="showRenameDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>重命名</DialogTitle>
        </DialogHeader>
        <div class="space-y-2">
          <Label for="rename">新名称</Label>
          <Input id="rename" v-model="renameName" @keyup.enter="handleRename" />
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showRenameDialog = false">取消</Button>
          <Button @click="handleRename">确认</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Move Dialog -->
    <Dialog v-model:open="showMoveDialog" :modal="false">
      <DialogContent class="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>移动到相册</DialogTitle>
          <DialogDescription>将 {{ moveKeys.length }} 张图片移动到目标相册。需要新相册时，在目标相册里选择创建新相册。</DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <Label>目标相册</Label>
            <Select v-model="moveAlbumId">
              <SelectTrigger>
                <SelectValue placeholder="未分类图片" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__none__">未分类图片</SelectItem>
                <SelectItem value="__create__">+ 创建新相册</SelectItem>
                <SelectItem v-for="album in albums" :key="album.id" :value="String(album.id)">
                  {{ album.name }} · {{ album.permission === 1 ? '公开' : '私密' }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div v-if="moveAlbumId === '__create__'" class="rounded-xl border border-white/10 bg-white/[0.03] p-3">
            <div class="mb-3 flex items-center gap-2 text-sm font-medium">
              <Plus class="h-4 w-4 text-violet-300" />
              创建新相册
            </div>
            <div class="grid gap-3 sm:grid-cols-[1fr_140px]">
              <Input v-model="quickAlbumName" placeholder="新相册名称" @keyup.enter="createAlbumForMove" />
              <Select v-model="quickAlbumPermission">
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">私密</SelectItem>
                  <SelectItem value="1">公开</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <Button variant="outline" size="sm" class="mt-3" :disabled="creatingAlbum" @click="createAlbumForMove">
              <Plus class="mr-1.5 h-4 w-4" />
              {{ creatingAlbum ? '创建中...' : '创建并选中' }}
            </Button>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showMoveDialog = false">取消</Button>
          <Button @click="handleMove">确认</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete Confirm Dialog -->
    <Dialog v-model:open="showDeleteConfirmDialog">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>删除图片</DialogTitle>
          <DialogDescription>
            确定要删除这 {{ deleteConfirmKeys.length }} 张图片吗？删除后将进入回收站，30天后自动清理。
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" @click="showDeleteConfirmDialog = false">取消</Button>
          <Button variant="destructive" @click="executeDelete">确认删除</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
