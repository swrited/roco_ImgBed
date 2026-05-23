<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { imagesApi } from '@/api/images'
import { usersApi } from '@/api/users'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { toast } from 'vue-sonner'
import { AlertTriangle, Clipboard, ImagePlus, CloudUpload, X, CheckCircle2 } from 'lucide-vue-next'

const uploading        = ref(false)
const progress         = ref(0)
const uploadedImages   = ref<any[]>([])
const dragOver         = ref(false)
const fileInput        = ref<HTMLInputElement>()
const uploadMaxSizeKB  = ref(10240)
const uploadError      = ref('')
const limitLoading     = ref(false)

const uploadLimitLabel = computed(() => {
  if (!uploadMaxSizeKB.value || uploadMaxSizeKB.value <= 0) return '不限制单张图片大小'
  if (uploadMaxSizeKB.value >= 1048576) return `单张不超过 ${(uploadMaxSizeKB.value / 1048576).toFixed(2)} GB`
  if (uploadMaxSizeKB.value >= 1024)    return `单张不超过 ${(uploadMaxSizeKB.value / 1024).toFixed(2)} MB`
  return `单张不超过 ${uploadMaxSizeKB.value} KB`
})

async function loadUploadLimit() {
  limitLoading.value = true
  try {
    const settings = await usersApi.settings()
    const limit = Number(settings.upload_max_size)
    if (!Number.isNaN(limit)) uploadMaxSizeKB.value = limit
  } catch { /* use default */ } finally {
    limitLoading.value = false
  }
}

function refreshLimitOnVisible() {
  if (document.visibilityState === 'visible') loadUploadLimit()
}

function handleDrop(e: DragEvent) {
  dragOver.value = false
  const files = e.dataTransfer?.files
  if (files) uploadFiles(files)
}

async function handleFileInput(e: Event) {
  const files = (e.target as HTMLInputElement).files
  if (files) await uploadFiles(files)
}

async function uploadFiles(files: FileList | File[]) {
  await loadUploadLimit()
  for (const file of Array.from(files)) {
    uploadError.value = ''
    const sizeKB = file.size / 1024
    if (uploadMaxSizeKB.value > 0 && sizeKB > uploadMaxSizeKB.value) {
      uploadError.value = `${file.name} 超过上传大小限制，${uploadLimitLabel.value}`
      toast.error(uploadError.value)
      continue
    }
    const formData = new FormData()
    formData.append('file', file)
    uploading.value = true
    progress.value  = 0
    try {
      const res = await imagesApi.upload(formData, (p) => { progress.value = p })
      if (res) uploadedImages.value.unshift(res)
      toast.success('上传成功')
    } catch (e: any) {
      uploadError.value = e.message || '上传失败，请检查图片格式、大小或存储策略'
      toast.error(e.message || '上传失败')
    }
    uploading.value = false
  }
  if (fileInput.value) fileInput.value.value = ''
}

function copyLink(url: string) {
  navigator.clipboard.writeText(url)
  toast.success('链接已复制')
}

onMounted(() => {
  loadUploadLimit()
  window.addEventListener('focus', loadUploadLimit)
  document.addEventListener('visibilitychange', refreshLimitOnVisible)
})

onUnmounted(() => {
  window.removeEventListener('focus', loadUploadLimit)
  document.removeEventListener('visibilitychange', refreshLimitOnVisible)
})
</script>

<template>
  <div>
    <!-- ── Page header ─────────────────────────────────────────────────── -->
    <div class="animate-fade-in-up mb-8 flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
      <div>
        <p class="text-xs font-semibold uppercase tracking-widest text-violet-400">Upload</p>
        <h1 class="mt-1.5 text-3xl font-semibold tracking-tight">上传图片</h1>
        <p class="mt-2 text-sm leading-6 text-muted-foreground">
          拖拽、点击选择或从剪贴板粘贴图片，上传后可立即复制链接。
          <span v-if="limitLoading" class="text-muted-foreground/60">正在同步限制...</span>
        </p>
      </div>
    </div>

    <!-- ── Drop zone ────────────────────────────────────────────────────── -->
    <div
      class="animate-fade-in-up delay-75 relative cursor-pointer overflow-hidden rounded-2xl border transition-all duration-300"
      :class="dragOver
        ? 'border-violet-500/60 bg-violet-500/8 shadow-2xl shadow-violet-500/20'
        : 'border-dashed border-white/12 bg-[oklch(9%_0.008_270)] hover:border-violet-500/40 hover:bg-white/[0.04]'"
      @dragover.prevent="dragOver = true"
      @dragleave="dragOver = false"
      @drop.prevent="handleDrop"
      @click="fileInput?.click()"
    >
      <!-- Animated corner accents when dragging -->
      <div
        v-if="dragOver"
        aria-hidden="true"
        class="pointer-events-none absolute inset-0"
      >
        <div class="absolute top-0 left-0 h-px w-24 bg-gradient-to-r from-violet-500 to-transparent" />
        <div class="absolute top-0 left-0 w-px h-24 bg-gradient-to-b from-violet-500 to-transparent" />
        <div class="absolute top-0 right-0 h-px w-24 bg-gradient-to-l from-violet-500 to-transparent" />
        <div class="absolute top-0 right-0 w-px h-24 bg-gradient-to-b from-violet-500 to-transparent" />
        <div class="absolute bottom-0 left-0 h-px w-24 bg-gradient-to-r from-violet-500 to-transparent" />
        <div class="absolute bottom-0 left-0 w-px h-24 bg-gradient-to-t from-violet-500 to-transparent" />
        <div class="absolute bottom-0 right-0 h-px w-24 bg-gradient-to-l from-violet-500 to-transparent" />
        <div class="absolute bottom-0 right-0 w-px h-24 bg-gradient-to-t from-violet-500 to-transparent" />
        <!-- Center glow -->
        <div
          class="absolute inset-0 rounded-2xl"
          style="background: radial-gradient(ellipse 60% 40% at 50% 50%, oklch(60% 0.22 293 / 0.08), transparent)"
        />
      </div>

      <div class="flex min-h-72 flex-col items-center justify-center p-8 text-center">
        <!-- Floating icon -->
        <div
          class="mb-5 flex h-16 w-16 items-center justify-center rounded-3xl bg-violet-500/12 text-violet-400 animate-float shadow-lg shadow-violet-500/15 ring-1 ring-violet-500/20"
          :class="dragOver ? 'scale-110 transition-transform duration-200' : ''"
        >
          <CloudUpload class="h-8 w-8" />
        </div>

        <p class="text-xl font-semibold">
          {{ dragOver ? '松手即可上传' : '把图片拖到这里' }}
        </p>
        <p class="mt-2 max-w-md text-sm leading-6 text-muted-foreground">
          支持 JPG、PNG、GIF、WebP 等格式。多个文件会按顺序上传，{{ uploadLimitLabel }}。
        </p>
        <Button class="mt-6 pointer-events-none" type="button">
          <ImagePlus class="mr-2 h-4 w-4" />
          选择图片
        </Button>
        <input ref="fileInput" type="file" accept="image/*" multiple class="hidden" @change="handleFileInput" />
      </div>
    </div>

    <!-- ── Error ────────────────────────────────────────────────────────── -->
    <div
      v-if="uploadError"
      class="animate-fade-in-up mt-4 flex items-start gap-3 rounded-xl border border-red-500/25 bg-red-500/8 p-4 text-red-400"
    >
      <AlertTriangle class="mt-0.5 h-5 w-5 shrink-0" />
      <div class="min-w-0 flex-1">
        <p class="font-semibold text-red-300">上传失败</p>
        <p class="mt-1 text-sm leading-6">{{ uploadError }}</p>
      </div>
      <Button
        variant="ghost"
        size="icon"
        class="h-8 w-8 text-red-400 hover:bg-red-500/10"
        @click="uploadError = ''"
      >
        <X class="h-4 w-4" />
      </Button>
    </div>

    <!-- ── Progress ─────────────────────────────────────────────────────── -->
    <div v-if="uploading" class="mt-4 space-y-2">
      <div class="flex items-center justify-between text-sm">
        <span class="text-muted-foreground">上传中...</span>
        <span class="tabular-nums font-medium">{{ progress }}%</span>
      </div>
      <Progress :model-value="progress" class="h-1.5" />
    </div>

    <!-- ── Uploaded results ─────────────────────────────────────────────── -->
    <div v-if="uploadedImages.length > 0" class="mt-10">
      <div class="mb-5 flex items-center gap-3">
        <CheckCircle2 class="h-5 w-5 text-emerald-400" />
        <h2 class="text-lg font-semibold">已上传的图片</h2>
        <span class="rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-medium text-emerald-400">
          {{ uploadedImages.length }}
        </span>
      </div>
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <Card
          v-for="img in uploadedImages"
          :key="img.key"
          class="group animate-scale-in overflow-hidden"
        >
          <div class="relative overflow-hidden">
            <img
              :src="img.links?.url"
              :alt="img.alias_name || img.origin_name"
              class="h-44 w-full object-cover transition-transform duration-500 group-hover:scale-105"
            />
            <!-- hover overlay -->
            <div class="absolute inset-0 bg-black/0 transition-all duration-300 group-hover:bg-black/25" />
          </div>
          <CardContent class="p-3">
            <p class="truncate text-sm font-medium">{{ img.alias_name || img.origin_name }}</p>
            <div class="mt-2 flex gap-2">
              <Button size="sm" variant="outline" class="flex-1 text-xs" @click.stop="copyLink(img.links?.url)">
                <Clipboard class="mr-1.5 h-3 w-3" />
                复制链接
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
