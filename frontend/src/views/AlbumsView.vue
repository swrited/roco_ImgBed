<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { albumsApi } from '@/api/albums'
import { imagesApi } from '@/api/images'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { toast } from 'vue-sonner'
import { Plus, FolderOpen, Pencil, Trash2, ChevronRight, Inbox, Image, List, Lock, Globe2 } from 'lucide-vue-next'
import type { Album, Image as UserImage } from '@/types'

const router = useRouter()
const albums = ref<Album[]>([])
const showDialog = ref(false)
const editingAlbum = ref<Album | null>(null)
const albumName = ref('')
const albumIntro = ref('')
const albumPermission = ref('0')
const displayMode = ref<'cover' | 'name'>('cover')
const albumCoverId = ref<string>('__auto__')
const coverImages = ref<UserImage[]>([])
const coverLoading = ref(false)

async function loadAlbums() {
  try {
    const res = await albumsApi.list()
    albums.value = res
  } catch {
    toast.error('加载相册失败')
  }
}

function openCreate() {
  editingAlbum.value = null
  albumName.value = ''
  albumIntro.value = ''
  albumPermission.value = '0'
  albumCoverId.value = '__auto__'
  coverImages.value = []
  showDialog.value = true
}

async function openEdit(album: Album) {
  editingAlbum.value = album
  albumName.value = album.name
  albumIntro.value = album.intro || ''
  albumPermission.value = String(album.permission || '0')
  albumCoverId.value = album.cover_image_id ? String(album.cover_image_id) : '__auto__'
  await loadAlbumCoverImages(album.id)
  showDialog.value = true
}

async function loadAlbumCoverImages(albumId: number) {
  coverLoading.value = true
  try {
    const res = await imagesApi.list({ album_id: albumId, per_page: 100 })
    coverImages.value = res?.data ?? []
  } catch {
    coverImages.value = []
    toast.error('加载封面图片失败')
  } finally {
    coverLoading.value = false
  }
}

async function handleSubmit() {
  if (!albumName.value.trim()) {
    toast.error('请输入相册名称')
    return
  }
  try {
    if (editingAlbum.value) {
      await albumsApi.update(editingAlbum.value.id, {
        name: albumName.value,
        intro: albumIntro.value,
        permission: Number(albumPermission.value),
        cover_image_id: albumCoverId.value === '__auto__' ? null : Number(albumCoverId.value),
      })
      toast.success('更新成功')
    } else {
      await albumsApi.create({
        name: albumName.value,
        intro: albumIntro.value,
        permission: Number(albumPermission.value),
      })
      toast.success('创建成功')
    }
    showDialog.value = false
    loadAlbums()
  } catch (e: any) {
    toast.error(e.message || '操作失败')
  }
}

async function deleteAlbum(id: number) {
  if (!confirm('确定删除此相册？相册内图片不会被删除。')) return
  try {
    await albumsApi.delete(id)
    toast.success('删除成功')
    loadAlbums()
  } catch (e: any) {
    toast.error(e.message || '删除失败')
  }
}

function goToImages(albumId: number) {
  router.push({ name: 'library.images', query: { album_id: albumId } })
}

function permissionLabel(album: Album) {
  return album.permission === 1 ? '公开相册' : '私密相册'
}

onMounted(loadAlbums)
</script>

<template>
  <div>
    <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="inline-flex w-fit rounded-xl border border-white/10 bg-white/[0.03] p-1">
        <Button
          variant="ghost"
          size="sm"
          class="h-8 gap-1.5"
          :class="displayMode === 'cover' ? 'bg-violet-500/15 text-violet-200' : 'text-muted-foreground'"
          @click="displayMode = 'cover'"
        >
          <Image class="h-3.5 w-3.5" /> 封面
        </Button>
        <Button
          variant="ghost"
          size="sm"
          class="h-8 gap-1.5"
          :class="displayMode === 'name' ? 'bg-violet-500/15 text-violet-200' : 'text-muted-foreground'"
          @click="displayMode = 'name'"
        >
          <List class="h-3.5 w-3.5" /> 名称
        </Button>
      </div>
      <Button @click="openCreate">
        <Plus class="mr-2 h-4 w-4" /> 新建相册
      </Button>
    </div>

    <div v-if="displayMode === 'cover'" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Card
        class="group cursor-pointer overflow-hidden border-dashed border-violet-400/25 bg-violet-500/[0.04] transition-colors hover:border-violet-400/50 hover:bg-violet-500/10"
        @click="goToImages(0)"
      >
        <div class="flex aspect-[4/3] items-center justify-center bg-violet-500/10">
          <Inbox class="h-14 w-14 text-violet-300" />
        </div>
        <CardHeader class="py-4">
          <div class="flex items-center justify-between gap-3">
            <div class="min-w-0">
              <CardTitle class="truncate text-lg">未分类图片</CardTitle>
              <p class="text-sm text-muted-foreground">未选择相册的上传默认在这里</p>
            </div>
            <ChevronRight class="ml-1 h-4 w-4 shrink-0 text-muted-foreground transition-colors group-hover:text-primary" />
          </div>
        </CardHeader>
      </Card>

      <Card
        v-for="album in albums"
        :key="album.id"
        class="group flex h-full cursor-pointer flex-col overflow-hidden transition-colors hover:border-primary/50"
        @click="goToImages(album.id)"
      >
        <div class="relative aspect-[4/3] shrink-0 overflow-hidden bg-white/[0.04]">
          <img v-if="album.cover_url" :src="album.cover_url" :alt="album.name" class="absolute inset-0 h-full w-full object-cover transition-transform duration-500 group-hover:scale-105" />
          <div v-else class="flex h-full w-full items-center justify-center">
            <FolderOpen class="h-14 w-14 text-primary" />
          </div>
          <div class="absolute left-3 top-3">
            <Badge :class="album.permission === 1 ? 'border-emerald-400/40 bg-emerald-500/20 text-emerald-100' : 'border-violet-300/40 bg-violet-500/25 text-violet-100'">
              <component :is="album.permission === 1 ? Globe2 : Lock" class="mr-1 h-3 w-3" />
              {{ permissionLabel(album) }}
            </Badge>
          </div>
        </div>
        <CardHeader class="py-4">
          <div class="flex items-center justify-between gap-3">
            <div class="min-w-0">
              <CardTitle class="truncate text-lg">{{ album.name }}</CardTitle>
              <p class="text-sm text-muted-foreground">{{ album.image_num }} 张图片</p>
            </div>
            <div class="flex shrink-0 items-center gap-1">
              <Button variant="ghost" size="icon" @click.stop="openEdit(album)">
                <Pencil class="h-4 w-4" />
              </Button>
              <Button variant="ghost" size="icon" @click.stop="deleteAlbum(album.id)">
                <Trash2 class="h-4 w-4 text-destructive" />
              </Button>
              <ChevronRight class="ml-1 h-4 w-4 text-muted-foreground transition-colors group-hover:text-primary" />
            </div>
          </div>
        </CardHeader>
        <CardContent v-if="album.intro" class="pt-0">
          <p class="line-clamp-2 text-sm text-muted-foreground">{{ album.intro }}</p>
        </CardContent>
      </Card>
    </div>

    <div v-else class="space-y-2">
      <Card
        class="group cursor-pointer border-dashed border-violet-400/25 bg-violet-500/[0.04] transition-colors hover:border-violet-400/50 hover:bg-violet-500/10"
        @click="goToImages(0)"
      >
        <CardHeader class="py-4">
          <div class="flex items-center justify-between gap-3">
            <div class="flex min-w-0 items-center gap-3">
              <Inbox class="h-7 w-7 shrink-0 text-violet-300" />
              <div class="min-w-0">
                <CardTitle class="truncate text-base">未分类图片</CardTitle>
                <p class="text-sm text-muted-foreground">未选择相册的上传默认在这里</p>
              </div>
            </div>
            <ChevronRight class="h-4 w-4 text-muted-foreground group-hover:text-primary" />
          </div>
        </CardHeader>
      </Card>

      <Card
        v-for="album in albums"
        :key="album.id"
        class="group cursor-pointer transition-colors hover:border-primary/50"
        @click="goToImages(album.id)"
      >
        <CardHeader class="py-4">
          <div class="flex items-center justify-between gap-3">
            <div class="flex min-w-0 items-center gap-3">
              <FolderOpen class="h-7 w-7 shrink-0 text-primary" />
              <div class="min-w-0">
                <div class="flex min-w-0 items-center gap-2">
                  <CardTitle class="truncate text-base">{{ album.name }}</CardTitle>
                  <Badge :class="album.permission === 1 ? 'border-emerald-400/40 bg-emerald-500/15 text-emerald-100' : 'border-violet-300/40 bg-violet-500/20 text-violet-100'">
                    <component :is="album.permission === 1 ? Globe2 : Lock" class="mr-1 h-3 w-3" />
                    {{ album.permission === 1 ? '公开' : '私密' }}
                  </Badge>
                </div>
                <p class="text-sm text-muted-foreground">{{ album.image_num }} 张图片</p>
              </div>
            </div>
            <div class="flex shrink-0 items-center gap-1">
              <Button variant="ghost" size="icon" @click.stop="openEdit(album)">
                <Pencil class="h-4 w-4" />
              </Button>
              <Button variant="ghost" size="icon" @click.stop="deleteAlbum(album.id)">
                <Trash2 class="h-4 w-4 text-destructive" />
              </Button>
              <ChevronRight class="ml-1 h-4 w-4 text-muted-foreground transition-colors group-hover:text-primary" />
            </div>
          </div>
        </CardHeader>
      </Card>
    </div>

    <div v-if="albums.length === 0" class="text-center py-12 text-muted-foreground">
      <FolderOpen class="mx-auto h-12 w-12 mb-4 opacity-50" />
      <p>暂无相册，点击上方按钮创建</p>
    </div>

    <!-- Create/Edit Dialog -->
    <Dialog v-model:open="showDialog">
      <DialogContent class="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{{ editingAlbum ? '编辑相册' : '新建相册' }}</DialogTitle>
          <DialogDescription>{{ editingAlbum ? '修改相册信息' : '创建一个新相册来分类管理图片' }}</DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="name">相册名称</Label>
            <Input id="name" v-model="albumName" placeholder="输入相册名称" @keyup.enter="handleSubmit" />
          </div>
          <div class="space-y-2">
            <Label for="intro">相册简介</Label>
            <Textarea id="intro" v-model="albumIntro" placeholder="可选，简要描述相册内容" />
          </div>
          <div class="space-y-2">
            <Label>公开权限</Label>
            <div class="grid gap-2 sm:grid-cols-2">
              <label
                class="flex cursor-pointer items-center gap-3 rounded-xl border p-3 transition"
                :class="albumPermission === '0' ? 'border-violet-400/50 bg-violet-500/15 text-violet-100' : 'border-white/10 bg-white/[0.03] text-muted-foreground hover:bg-white/[0.06]'"
              >
                <input type="radio" v-model="albumPermission" value="0" class="sr-only" />
                <Lock class="h-4 w-4" />
                <span class="text-sm font-medium">私密相册</span>
              </label>
              <label
                class="flex cursor-pointer items-center gap-3 rounded-xl border p-3 transition"
                :class="albumPermission === '1' ? 'border-emerald-400/50 bg-emerald-500/15 text-emerald-100' : 'border-white/10 bg-white/[0.03] text-muted-foreground hover:bg-white/[0.06]'"
              >
                <input type="radio" v-model="albumPermission" value="1" class="sr-only" />
                <Globe2 class="h-4 w-4" />
                <span class="text-sm font-medium">公开相册</span>
              </label>
            </div>
          </div>
          <div v-if="editingAlbum" class="space-y-3">
            <div class="flex items-center justify-between gap-3">
              <Label>相册封面</Label>
              <Select v-model="albumCoverId">
                <SelectTrigger class="w-44">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="__auto__">自动使用最新图片</SelectItem>
                  <SelectItem v-for="img in coverImages" :key="img.id" :value="String(img.id)">
                    {{ img.alias_name || img.origin_name || img.name }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div v-if="coverLoading" class="grid grid-cols-4 gap-2">
              <div v-for="i in 4" :key="i" class="aspect-[4/3] rounded-xl bg-white/[0.06] animate-pulse" />
            </div>
            <div v-else-if="coverImages.length > 0" class="grid max-h-56 grid-cols-3 gap-2 overflow-y-auto pr-1 sm:grid-cols-4">
              <button
                type="button"
                class="group relative overflow-hidden rounded-xl border bg-black/30 text-left transition"
                :class="albumCoverId === '__auto__' ? 'border-violet-400/50' : 'border-white/10 hover:border-violet-400/45'"
                @click="albumCoverId = '__auto__'"
              >
                <div class="flex aspect-[4/3] items-center justify-center bg-violet-500/10">
                  <Image class="h-7 w-7 text-violet-300" />
                </div>
                <div class="absolute inset-x-0 bottom-0 bg-black/70 px-2 py-1 text-xs text-white">自动</div>
              </button>
              <button
                v-for="img in coverImages"
                :key="img.key"
                type="button"
                class="group relative aspect-[4/3] overflow-hidden rounded-xl border bg-black/30 transition"
                :class="albumCoverId === String(img.id) ? 'border-violet-300 ring-2 ring-violet-400/45' : 'border-white/10 hover:border-violet-400/45'"
                @click="albumCoverId = String(img.id)"
              >
                <img :src="img.url" :alt="img.alias_name || img.origin_name" class="absolute inset-0 h-full w-full object-cover transition duration-300 group-hover:scale-105" />
                <div class="absolute inset-x-0 bottom-0 truncate bg-black/70 px-2 py-1 text-xs text-white">
                  {{ img.alias_name || img.origin_name || img.name }}
                </div>
              </button>
            </div>
            <div v-else class="rounded-xl border border-dashed border-white/10 bg-white/[0.03] p-4 text-sm text-muted-foreground">
              这个相册还没有图片，上传图片后可以回来选择封面。
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showDialog = false">取消</Button>
          <Button @click="handleSubmit">确认</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
