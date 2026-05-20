<script setup lang="ts">
import { ref } from 'vue'
import { imagesApi } from '@/api/images'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { toast } from 'vue-sonner'
import { Upload, Clipboard, ImagePlus, CloudUpload } from 'lucide-vue-next'

const uploading = ref(false)
const progress = ref(0)
const uploadedImages = ref<any[]>([])
const dragOver = ref(false)
const fileInput = ref<HTMLInputElement>()

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
  for (const file of Array.from(files)) {
    const formData = new FormData()
    formData.append('file', file)
    uploading.value = true
    progress.value = 0
    try {
      const res = await imagesApi.upload(formData, (p) => {
        progress.value = p
      })
      if (res) {
        uploadedImages.value.unshift(res)
      }
      toast.success('上传成功')
    } catch (e: any) {
      toast.error(e.message || '上传失败')
    }
    uploading.value = false
  }
}

function copyLink(url: string) {
  navigator.clipboard.writeText(url)
  toast.success('链接已复制')
}
</script>

<template>
  <div>
    <div class="mb-6 flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
      <div>
        <p class="text-sm font-medium text-primary">Upload</p>
        <h1 class="mt-1 text-3xl font-semibold">上传图片</h1>
        <p class="mt-2 text-sm text-muted-foreground">拖拽、点击选择或从剪贴板粘贴图片，上传后可立即复制链接。</p>
      </div>
    </div>

    <!-- Drop zone -->
    <Card
      class="relative cursor-pointer border-dashed bg-[#0f0f15] transition-all hover:border-primary/60 hover:bg-white/5"
      :class="dragOver ? 'border-primary bg-primary/5 shadow-xl shadow-primary/10' : ''"
      @dragover.prevent="dragOver = true"
      @dragleave="dragOver = false"
      @drop.prevent="handleDrop"
      @click="fileInput?.click()"
    >
      <CardContent class="flex min-h-72 flex-col items-center justify-center p-8 text-center">
        <div class="mb-5 flex h-16 w-16 items-center justify-center rounded-3xl bg-primary/10 text-primary">
          <CloudUpload class="h-8 w-8" />
        </div>
        <p class="text-xl font-semibold">把图片拖到这里</p>
        <p class="mt-2 max-w-md text-sm leading-6 text-muted-foreground">支持 JPG、PNG、GIF、WebP 等格式。多个文件会按顺序上传。</p>
        <Button class="mt-6" type="button">
          <ImagePlus class="mr-2 h-4 w-4" />
          选择图片
        </Button>
        <input ref="fileInput" type="file" accept="image/*" multiple class="hidden" @change="handleFileInput" />
      </CardContent>
    </Card>

    <!-- Progress -->
    <div v-if="uploading" class="mt-4 space-y-2">
      <p class="text-sm">上传中...</p>
      <Progress :model-value="progress" class="h-2" />
    </div>

    <!-- Uploaded images -->
    <div v-if="uploadedImages.length > 0" class="mt-8">
      <h2 class="mb-4 text-lg font-semibold">已上传的图片</h2>
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <Card v-for="img in uploadedImages" :key="img.key">
          <CardContent class="p-3">
            <img :src="img.links?.url" :alt="img.name" class="h-44 w-full rounded-xl object-cover" />
            <p class="text-sm truncate">{{ img.name || img.origin_name }}</p>
            <div class="flex gap-2 mt-2">
              <Button size="sm" variant="outline" @click.stop="copyLink(img.links?.url)">
                <Clipboard class="mr-1 h-3 w-3" /> 复制链接
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
