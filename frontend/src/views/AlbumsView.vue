<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { albumsApi } from '@/api/albums'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { toast } from 'vue-sonner'
import { Plus, FolderOpen, Pencil, Trash2, ChevronRight, Inbox, Lock, Globe2, Images, CalendarDays } from 'lucide-vue-next'
import type { Album } from '@/types'

const router = useRouter()
const albums = ref<Album[]>([])
const showDialog = ref(false)
const editingAlbum = ref<Album | null>(null)
const albumName = ref('')
const albumIntro = ref('')
const albumPermission = ref('0')

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
  showDialog.value = true
}

function openEdit(album: Album) {
  editingAlbum.value = album
  albumName.value = album.name
  albumIntro.value = album.intro || ''
  albumPermission.value = String(album.permission || '0')
  showDialog.value = true
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

function goToImages(albumId?: number) {
  router.push({
    name: 'library.images',
    query: albumId === undefined ? {} : { album_id: albumId },
  })
}

function formatCreatedAt(value: string) {
  if (!value) return '创建时间未知'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '创建时间未知'
  return `创建于 ${date.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })}`
}

onMounted(loadAlbums)
</script>

<template>
  <div>
    <div class="mb-6 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <p class="text-xs font-semibold uppercase text-violet-400">Library</p>
        <h1 class="mt-1.5 text-3xl font-semibold tracking-tight">我的图库</h1>
        <p class="mt-1 text-sm text-muted-foreground">选择相册进入图片浏览，未指定相册的图片会单独归档。</p>
      </div>
      <Button @click="openCreate">
        <Plus class="mr-2 h-4 w-4" /> 新建相册
      </Button>
    </div>

    <div class="mb-6 grid gap-3 sm:grid-cols-2">
      <Card
        class="group cursor-pointer border-violet-400/40 bg-violet-500/[0.1] transition-colors hover:border-violet-300/70 hover:bg-violet-500/15"
        role="button"
        tabindex="0"
        @click="goToImages()"
        @keydown.enter.prevent="goToImages()"
        @keydown.space.prevent="goToImages()"
      >
        <CardHeader class="py-5">
          <div class="flex items-center justify-between gap-3">
            <div class="flex min-w-0 items-center gap-3">
              <div class="rounded-lg bg-violet-500/20 p-2.5 text-violet-200">
                <Images class="h-6 w-6" />
              </div>
              <div class="min-w-0">
                <CardTitle class="text-base">全部图片</CardTitle>
                <p class="mt-1 text-sm text-muted-foreground">按上传顺序查看所有图片</p>
              </div>
            </div>
            <ChevronRight class="h-4 w-4 shrink-0 text-violet-300" />
          </div>
        </CardHeader>
      </Card>

      <Card
        class="group cursor-pointer border-dashed border-white/15 bg-white/[0.03] transition-colors hover:border-violet-400/45 hover:bg-white/[0.06]"
        role="button"
        tabindex="0"
        @click="goToImages(0)"
        @keydown.enter.prevent="goToImages(0)"
        @keydown.space.prevent="goToImages(0)"
      >
        <CardHeader class="py-5">
          <div class="flex items-center justify-between gap-3">
            <div class="flex min-w-0 items-center gap-3">
              <div class="rounded-lg bg-white/[0.06] p-2.5 text-slate-200">
                <Inbox class="h-6 w-6" />
              </div>
              <div class="min-w-0">
                <CardTitle class="text-base">未分类图片</CardTitle>
                <p class="mt-1 text-sm text-muted-foreground">未选择相册的上传图片</p>
              </div>
            </div>
            <ChevronRight class="h-4 w-4 shrink-0 text-muted-foreground group-hover:text-primary" />
          </div>
        </CardHeader>
      </Card>
    </div>

    <div class="mb-3 flex items-center justify-between">
      <h2 class="text-base font-semibold">相册管理</h2>
      <p class="text-xs text-muted-foreground">{{ albums.length }} 个相册</p>
    </div>

    <div class="grid gap-3 lg:grid-cols-2">
      <Card
        v-for="album in albums"
        :key="album.id"
        class="group cursor-pointer transition-colors hover:border-primary/50 hover:bg-white/[0.025]"
        role="button"
        tabindex="0"
        @click="goToImages(album.id)"
        @keydown.enter.prevent="goToImages(album.id)"
        @keydown.space.prevent="goToImages(album.id)"
      >
        <CardHeader class="gap-4 py-4">
          <div class="flex items-start justify-between gap-3">
            <div class="flex min-w-0 items-center gap-3">
              <div class="rounded-lg bg-violet-500/10 p-2.5">
                <FolderOpen class="h-6 w-6 shrink-0 text-primary" />
              </div>
              <div class="min-w-0">
                <div class="flex min-w-0 flex-wrap items-center gap-2">
                  <CardTitle class="truncate text-base">{{ album.name }}</CardTitle>
                  <Badge :class="album.permission === 1 ? 'border-emerald-400/40 bg-emerald-500/15 text-emerald-100' : 'border-violet-300/40 bg-violet-500/20 text-violet-100'">
                    <component :is="album.permission === 1 ? Globe2 : Lock" class="mr-1 h-3 w-3" />
                    {{ album.permission === 1 ? '公开' : '私密' }}
                  </Badge>
                </div>
                <p class="mt-1 line-clamp-1 text-sm text-muted-foreground">
                  {{ album.intro || '暂无相册简介' }}
                </p>
              </div>
            </div>
            <ChevronRight class="mt-3 h-4 w-4 shrink-0 text-muted-foreground transition-colors group-hover:text-primary" />
          </div>
          <div class="flex flex-wrap items-center justify-between gap-3 border-t border-white/[0.06] pt-3 text-xs text-muted-foreground">
            <div class="flex flex-wrap items-center gap-x-4 gap-y-2">
              <span class="flex items-center gap-1.5">
                <Images class="h-3.5 w-3.5" /> {{ album.image_num }} 张图片
              </span>
              <span class="flex items-center gap-1.5">
                <CalendarDays class="h-3.5 w-3.5" /> {{ formatCreatedAt(album.created_at) }}
              </span>
            </div>
            <div class="flex shrink-0 items-center gap-1">
              <Button variant="ghost" size="icon" class="h-8 w-8" title="编辑相册" @click.stop="openEdit(album)">
                <Pencil class="h-4 w-4" />
              </Button>
              <Button variant="ghost" size="icon" class="h-8 w-8" title="删除相册" @click.stop="deleteAlbum(album.id)">
                <Trash2 class="h-4 w-4 text-destructive" />
              </Button>
            </div>
          </div>
        </CardHeader>
      </Card>
    </div>

    <div v-if="albums.length === 0" class="py-10 text-center text-muted-foreground">
      <FolderOpen class="mx-auto h-12 w-12 mb-4 opacity-50" />
      <p>还没有自定义相册，可以新建相册进行分类。</p>
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
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showDialog = false">取消</Button>
          <Button @click="handleSubmit">确认</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
