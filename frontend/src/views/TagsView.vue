<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { tagsApi } from '@/api/tags'
import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
import { toast } from 'vue-sonner'
import { Hash, Trash2, ChevronRight } from 'lucide-vue-next'
import type { Tag } from '@/types'

const router = useRouter()
const tags = ref<Tag[]>([])
const loading = ref(false)

async function loadTags() {
  loading.value = true
  try {
    const res = await tagsApi.list()
    tags.value = res || []
  } catch (e: any) {
    toast.error('加载标签失败')
  } finally {
    loading.value = false
  }
}

async function deleteTag(id: number) {
  if (!confirm('确定删除此标签？该标签将从所有图片上移除。')) return
  try {
    await tagsApi.delete(id)
    toast.success('删除成功')
    loadTags()
  } catch (e: any) {
    toast.error(e.message || '删除失败')
  }
}

function goToImages(tag: Tag) {
  router.push({ name: 'library.images', query: { tag_id: tag.id } })
}

onMounted(() => {
  loadTags()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-end mb-4">
      <p class="text-sm text-muted-foreground">在上传或编辑图片时可添加新标签</p>
    </div>

    <div v-if="loading" class="grid gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
      <div v-for="i in 10" :key="i" class="h-24 bg-white/5 animate-pulse rounded-xl" />
    </div>

    <div v-else-if="tags.length > 0" class="grid gap-4 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
      <Card
        v-for="tag in tags"
        :key="tag.id"
        class="cursor-pointer hover:border-violet-500/50 transition-colors group bg-white/[0.02]"
        @click="goToImages(tag)"
      >
        <CardHeader class="p-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2 min-w-0">
              <Hash class="h-5 w-5 text-violet-400 shrink-0" />
              <div class="min-w-0">
                <CardTitle class="text-base truncate">{{ tag.name }}</CardTitle>
              </div>
            </div>
            <div class="flex items-center shrink-0">
              <Button variant="ghost" size="icon" class="h-8 w-8 text-muted-foreground hover:text-destructive" @click.stop="deleteTag(tag.id)">
                <Trash2 class="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent class="px-4 pb-4 pt-0">
          <div class="flex items-center justify-between">
            <p class="text-sm text-muted-foreground">{{ tag.image_num }} 张图片</p>
            <ChevronRight class="h-4 w-4 text-muted-foreground group-hover:text-violet-400 transition-colors" />
          </div>
        </CardContent>
      </Card>
    </div>

    <div v-else class="text-center py-24 text-muted-foreground">
      <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-3xl bg-white/5 mx-auto">
        <Hash class="h-8 w-8 text-muted-foreground/50" />
      </div>
      <p class="text-lg font-medium">暂无标签</p>
      <p class="text-sm mt-1">上传图片时可以为图片添加标签</p>
    </div>
  </div>
</template>
