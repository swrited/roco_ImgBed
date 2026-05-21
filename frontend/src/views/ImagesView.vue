<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { imagesApi } from '@/api/images'
import { albumsApi } from '@/api/albums'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
import { MoreHorizontal, Copy, Trash2, Pencil, FolderInput, Search, Eye, EyeOff } from 'lucide-vue-next'
import type { Album, Image } from '@/types'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'

const route = useRoute()

const images = ref<Image[]>([])
const albums = ref<Album[]>([])
const currentPage = ref(1)
const lastPage = ref(1)
const total = ref(0)
const loading = ref(false)
const selectedKeys = ref<string[]>([])

const keyword = ref('')
const filterAlbumId = ref('__all__')
const sortOrder = ref('newest')
const filterPermission = ref('__all__')

const showMoveDialog = ref(false)
const showRenameDialog = ref(false)
const renameKey = ref('')
const renameName = ref('')
const moveAlbumId = ref<string>('__none__')

async function loadImages(page = 1) {
  loading.value = true
  selectedKeys.value = []
  try {
    const params: Record<string, any> = { page }
    if (keyword.value.trim()) params.q = keyword.value.trim()
    if (filterAlbumId.value && filterAlbumId.value !== '__all__') params.album_id = filterAlbumId.value
    if (filterPermission.value !== '' && filterPermission.value !== '__all__') params.permission = filterPermission.value
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

function toggleSelect(key: string) {
  const idx = selectedKeys.value.indexOf(key)
  if (idx >= 0) selectedKeys.value.splice(idx, 1)
  else selectedKeys.value.push(key)
}

function copyLink(url: string) {
  navigator.clipboard.writeText(url)
  toast.success('链接已复制')
}

async function deleteImages(keys: string[]) {
  try {
    await imagesApi.delete(keys)
    toast.success('删除成功')
    selectedKeys.value = []
    loadImages(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '删除失败')
  }
}

function openRename(key: string, name: string) {
  renameKey.value = key
  renameName.value = name
  showRenameDialog.value = true
}

async function handleRename() {
  try {
    await imagesApi.rename(renameKey.value, renameName.value)
    toast.success('重命名成功')
    showRenameDialog.value = false
    loadImages(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '重命名失败')
  }
}

function openMove() {
  moveAlbumId.value = '__none__'
  showMoveDialog.value = true
}

async function handleMove() {
  try {
    const albumId = moveAlbumId.value && moveAlbumId.value !== '__none__' ? Number(moveAlbumId.value) : null
    await imagesApi.move(selectedKeys.value, albumId)
    toast.success('移动成功')
    showMoveDialog.value = false
    selectedKeys.value = []
    loadImages(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '移动失败')
  }
}

async function setPermission(keys: string[], permission: number) {
  try {
    await imagesApi.setPermission(keys, permission)
    toast.success('权限已更新')
    selectedKeys.value = []
    loadImages(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '更新权限失败')
  }
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
  loadImages()
  loadAlbums()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h1 class="text-2xl font-bold">图片管理</h1>
      <div class="flex items-center gap-2" v-if="selectedKeys.length > 0">
        <Badge variant="secondary">{{ selectedKeys.length }} 张已选</Badge>
        <Button variant="outline" size="sm" @click="setPermission(selectedKeys, 1)">
          <Eye class="mr-1 h-4 w-4" /> 设为公开
        </Button>
        <Button variant="outline" size="sm" @click="setPermission(selectedKeys, 0)">
          <EyeOff class="mr-1 h-4 w-4" /> 设为私密
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
        <SelectTrigger class="w-36">
          <SelectValue placeholder="全部相册" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="__all__">全部相册</SelectItem>
          <SelectItem v-for="a in albums" :key="a.id" :value="String(a.id)">
            {{ a.name }}
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
      <Select v-model="filterPermission" @update:model-value="loadImages()">
        <SelectTrigger class="w-28">
          <SelectValue placeholder="全部权限" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="__all__">全部</SelectItem>
          <SelectItem value="1">公开</SelectItem>
          <SelectItem value="0">私密</SelectItem>
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
        class="cursor-pointer group overflow-hidden"
        :class="selectedKeys.includes(img.key) ? 'ring-2 ring-primary' : ''"
        @click="toggleSelect(img.key)"
      >
        <div class="relative aspect-square overflow-hidden rounded-t-lg">
          <img :src="img.url" :alt="img.name" class="h-full w-full object-cover" loading="lazy" />
          <div class="absolute top-2 left-2">
            <Badge
              v-if="img.permission === 1"
              class="text-xs py-0 bg-emerald-500/80 hover:bg-emerald-500/80 text-white border-0"
            >公开</Badge>
            <Badge
              v-else
              class="text-xs py-0 bg-black/50 hover:bg-black/50 text-white border-0"
            >私密</Badge>
          </div>
          <div class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity">
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
                <DropdownMenuItem @click.stop="openRename(img.key, img.name || img.origin_name)">
                  <Pencil class="mr-2 h-4 w-4" /> 重命名
                </DropdownMenuItem>
                <DropdownMenuItem @click.stop="setPermission([img.key], img.permission === 1 ? 0 : 1)">
                  <Eye v-if="img.permission === 0" class="mr-2 h-4 w-4" />
                  <EyeOff v-else class="mr-2 h-4 w-4" />
                  {{ img.permission === 1 ? '设为私密' : '设为公开' }}
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem class="text-destructive" @click.stop="deleteImages([img.key])">
                  <Trash2 class="mr-2 h-4 w-4" /> 删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
        <CardContent class="p-3">
          <p class="text-sm truncate font-medium">{{ img.name || img.origin_name }}</p>
          <p class="text-xs text-muted-foreground">
            {{ img.width }}x{{ img.height }} · {{ formatSize(img.size) }}
          </p>
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
      <DialogContent>
        <DialogHeader>
          <DialogTitle>移动到相册</DialogTitle>
          <DialogDescription>选择目标相册，留空则从相册中移出</DialogDescription>
        </DialogHeader>
        <Select v-model="moveAlbumId">
          <SelectTrigger>
            <SelectValue placeholder="不归属相册" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">(不归属相册)</SelectItem>
            <SelectItem v-for="album in albums" :key="album.id" :value="String(album.id)">
              {{ album.name }}
            </SelectItem>
          </SelectContent>
        </Select>
        <DialogFooter>
          <Button variant="outline" @click="showMoveDialog = false">取消</Button>
          <Button @click="handleMove">确认</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
