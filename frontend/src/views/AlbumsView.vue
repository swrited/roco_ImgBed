<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { albumsApi } from '@/api/albums'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { toast } from 'vue-sonner'
import { Plus, FolderOpen, Pencil, Trash2, ChevronRight } from 'lucide-vue-next'
import type { Album } from '@/types'

const router = useRouter()
const albums = ref<Album[]>([])
const showDialog = ref(false)
const editingAlbum = ref<Album | null>(null)
const albumName = ref('')
const albumIntro = ref('')

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
  showDialog.value = true
}

function openEdit(album: Album) {
  editingAlbum.value = album
  albumName.value = album.name
  albumIntro.value = album.intro || ''
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
      })
      toast.success('更新成功')
    } else {
      await albumsApi.create({
        name: albumName.value,
        intro: albumIntro.value,
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
  router.push({ name: 'images', query: { album_id: albumId } })
}

onMounted(loadAlbums)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">相册管理</h1>
      <Button @click="openCreate">
        <Plus class="mr-2 h-4 w-4" /> 新建相册
      </Button>
    </div>

    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Card
        v-for="album in albums"
        :key="album.id"
        class="cursor-pointer hover:border-primary/50 transition-colors group"
        @click="goToImages(album.id)"
      >
        <CardHeader>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3 min-w-0">
              <FolderOpen class="h-8 w-8 text-primary shrink-0" />
              <div class="min-w-0">
                <CardTitle class="text-lg truncate">{{ album.name }}</CardTitle>
                <p class="text-sm text-muted-foreground">{{ album.image_num }} 张图片</p>
              </div>
            </div>
            <div class="flex items-center gap-1 shrink-0">
              <Button variant="ghost" size="icon" @click.stop="openEdit(album)">
                <Pencil class="h-4 w-4" />
              </Button>
              <Button variant="ghost" size="icon" @click.stop="deleteAlbum(album.id)">
                <Trash2 class="h-4 w-4 text-destructive" />
              </Button>
              <ChevronRight class="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors ml-1" />
            </div>
          </div>
        </CardHeader>
        <CardContent v-if="album.intro">
          <p class="text-sm text-muted-foreground">{{ album.intro }}</p>
        </CardContent>
      </Card>
    </div>

    <div v-if="albums.length === 0" class="text-center py-12 text-muted-foreground">
      <FolderOpen class="mx-auto h-12 w-12 mb-4 opacity-50" />
      <p>暂无相册，点击上方按钮创建</p>
    </div>

    <!-- Create/Edit Dialog -->
    <Dialog v-model:open="showDialog">
      <DialogContent>
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
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showDialog = false">取消</Button>
          <Button @click="handleSubmit">确认</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
