<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { aiImagesApi } from '@/api/aiImages'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { toast } from 'vue-sonner'
import { Copy, FolderOpen, ImagePlus, Loader2, Sparkles } from 'lucide-vue-next'

const router = useRouter()

const prompt = ref('')
const aspectRatio = ref('1:1')
const count = ref(1)
const promptOptimizer = ref(true)
const loading = ref(false)
const cooldownRemaining = ref(0)
const generated = ref<any[]>([])
const album = ref<{ id: number; name: string } | null>(null)
const quota = ref<{ limit: number; used: number; remaining: number } | null>(null)
let cooldownTimer: number | undefined

const canGenerate = computed(() => prompt.value.trim().length > 0 && !loading.value && cooldownRemaining.value === 0)

const ratioOptions = [
  { value: '1:1', label: '1:1 正方形', size: '1024 x 1024' },
  { value: '16:9', label: '16:9 横版', size: '1280 x 720' },
  { value: '4:3', label: '4:3 横版', size: '1152 x 864' },
  { value: '3:2', label: '3:2 横版', size: '1248 x 832' },
  { value: '2:3', label: '2:3 竖版', size: '832 x 1248' },
  { value: '3:4', label: '3:4 竖版', size: '864 x 1152' },
  { value: '9:16', label: '9:16 竖版', size: '720 x 1280' },
  { value: '21:9', label: '21:9 超宽', size: '1344 x 576' },
]

async function generateImages() {
  if (!prompt.value.trim()) {
    toast.error('请输入提示词')
    return
  }
  if (cooldownRemaining.value > 0) {
    toast.error(`请求过于频繁，请 ${cooldownRemaining.value} 秒后再试`)
    return
  }
  loading.value = true
  try {
    const res = await aiImagesApi.generate({
      prompt: prompt.value.trim(),
      aspect_ratio: aspectRatio.value,
      count: Number(count.value) || 1,
      prompt_optimizer: promptOptimizer.value,
    })
    generated.value = res.images || []
    album.value = res.album || null
    quota.value = res.quota || null
    toast.success(`已生成 ${generated.value.length} 张图片`)
    startCooldown(5)
  } catch (e: any) {
    toast.error(e.message || '生成失败')
    if (String(e.message || '').includes('频繁')) startCooldown(5)
  } finally {
    loading.value = false
  }
}

function startCooldown(seconds: number) {
  cooldownRemaining.value = seconds
  if (cooldownTimer) window.clearInterval(cooldownTimer)
  cooldownTimer = window.setInterval(() => {
    cooldownRemaining.value -= 1
    if (cooldownRemaining.value <= 0 && cooldownTimer) {
      window.clearInterval(cooldownTimer)
      cooldownTimer = undefined
      cooldownRemaining.value = 0
    }
  }, 1000)
}

function copyText(value: string, message = '已复制') {
  navigator.clipboard.writeText(value)
  toast.success(message)
}

function openAlbum() {
  if (!album.value?.id) return
  router.push({ name: 'images', query: { album_id: album.value.id } })
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <div class="mb-2 flex items-center gap-2">
          <Badge class="border-purple-500/20 bg-purple-500/10 text-purple-300 hover:bg-purple-500/10">
            MiniMax image-01
          </Badge>
        </div>
        <h1 class="text-2xl font-bold">AI 生图</h1>
        <p class="mt-1 text-sm text-muted-foreground">输入提示词生成图片，结果会自动保存到“AI 生成”相册。每个用户每天默认 10 次免费额度。</p>
      </div>
      <Button v-if="album" variant="outline" @click="openAlbum">
        <FolderOpen class="mr-2 h-4 w-4" />
        打开 AI 生成相册
      </Button>
    </div>

    <div class="grid gap-6 lg:grid-cols-[minmax(0,440px)_1fr]">
      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Sparkles class="h-5 w-5 text-purple-400" />
            生成参数
          </CardTitle>
          <CardDescription>
            生成会消耗 MiniMax 额度，保存时使用你的默认存储策略。
            <span v-if="quota" class="mt-1 block text-purple-300">
              今日已用 {{ quota.used }} / {{ quota.limit || '不限' }} 次<span v-if="quota.remaining >= 0">，剩余 {{ quota.remaining }} 次</span>
            </span>
          </CardDescription>
        </CardHeader>
        <CardContent class="space-y-5">
          <div class="space-y-2">
            <Label for="prompt">提示词</Label>
            <Textarea
              id="prompt"
              v-model="prompt"
              class="min-h-40"
              maxlength="1500"
              placeholder="例如：一张极简科技风的产品海报，黑色背景，紫色霓虹边缘光，中心是一朵玻璃质感的花..."
            />
            <div class="text-right text-xs text-muted-foreground">{{ prompt.length }}/1500</div>
          </div>

          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label>图片比例</Label>
              <Select v-model="aspectRatio">
                <SelectTrigger class="w-full">
                  <SelectValue placeholder="选择比例" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="option in ratioOptions" :key="option.value" :value="option.value">
                    {{ option.label }} · {{ option.size }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div class="space-y-2">
              <Label for="count">生成数量</Label>
              <Input id="count" v-model.number="count" type="number" min="1" max="9" />
            </div>
          </div>

          <div class="flex items-center justify-between rounded-xl border border-white/5 px-4 py-3">
            <div class="space-y-0.5">
              <Label class="text-sm font-medium">提示词优化</Label>
              <p class="text-xs text-muted-foreground">让 MiniMax 自动增强提示词细节</p>
            </div>
            <Switch v-model="promptOptimizer" />
          </div>

          <Button class="h-10 w-full" :disabled="!canGenerate" @click="generateImages">
            <Loader2 v-if="loading" class="mr-2 h-4 w-4 animate-spin" />
            <ImagePlus v-else class="mr-2 h-4 w-4" />
            {{ loading ? '生成并保存中...' : cooldownRemaining > 0 ? `${cooldownRemaining} 秒后可再次生成` : '生成图片' }}
          </Button>
        </CardContent>
      </Card>

      <div class="space-y-4">
        <Card v-if="generated.length === 0" class="min-h-[420px]">
          <CardContent class="flex min-h-[420px] flex-col items-center justify-center text-center">
            <div class="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-purple-500/15 text-purple-300">
              <ImagePlus class="h-7 w-7" />
            </div>
            <h2 class="text-lg font-semibold">等待生成</h2>
            <p class="mt-2 max-w-sm text-sm leading-6 text-muted-foreground">
              生成完成后，图片会出现在这里，并已经保存到你的图床里。
            </p>
          </CardContent>
        </Card>

        <div v-else class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <Card v-for="img in generated" :key="img.key" class="overflow-hidden">
            <div class="aspect-square overflow-hidden bg-muted">
              <img :src="img.links.url" :alt="img.origin_name" class="h-full w-full object-cover" />
            </div>
            <CardContent class="space-y-3 p-3">
              <div>
                <p class="truncate text-sm font-medium">{{ img.origin_name }}</p>
                <p class="text-xs text-muted-foreground">{{ img.width }}x{{ img.height }} · {{ img.size }} KB</p>
              </div>
              <div class="grid grid-cols-2 gap-2">
                <Button variant="outline" size="sm" @click="copyText(img.links.url, '直链已复制')">
                  <Copy class="mr-1 h-3.5 w-3.5" />
                  直链
                </Button>
                <Button variant="outline" size="sm" @click="copyText(img.links.markdown, 'Markdown 已复制')">
                  <Copy class="mr-1 h-3.5 w-3.5" />
                  Markdown
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  </div>
</template>
