<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { toast } from 'vue-sonner'
import { Copy, ExternalLink, Play, RotateCcw, Terminal } from 'lucide-vue-next'
import { copyToClipboard } from '@/utils/clipboard'

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'
type AuthMode = 'none' | 'api-key' | 'read-token'

interface Preset {
  label: string
  method: HttpMethod
  path: string
  auth: AuthMode
  query?: string
  body?: string
}

interface ImagePreview {
  url: string
  name: string
  originName: string
  width?: number
  height?: number
  size?: number
}

const categories: Array<{ value: AuthMode; label: string; desc: string }> = [
  { value: 'none', label: '公开 API', desc: '无需认证即可读取的公开数据' },
  { value: 'api-key', label: 'API Key 调用', desc: '读取账户数据与执行写入操作' },
  { value: 'read-token', label: '短令牌调用', desc: '仅用于读取个人图库图片' },
]

const presets: Preset[] = [
  { label: '健康检查', method: 'GET', path: '/api/ping', auth: 'none' },
  { label: '公开设置', method: 'GET', path: '/api/v1/settings/public', auth: 'none' },
  { label: '公开画廊', method: 'GET', path: '/api/v1/gallery', auth: 'none', query: 'page=1' },
  { label: '公开存储策略', method: 'GET', path: '/api/v1/strategies', auth: 'none' },
  { label: '短令牌图片列表', method: 'GET', path: '/api/v1/t/{READ_TOKEN}/images', auth: 'read-token', query: 'page=1&per_page=20' },
  { label: '短令牌随机图', method: 'GET', path: '/api/v1/t/{READ_TOKEN}/images/random', auth: 'read-token' },
  { label: '短令牌适配图', method: 'GET', path: '/api/v1/t/{READ_TOKEN}/images/adaptive', auth: 'read-token' },
  { label: 'API Key 随机图', method: 'GET', path: '/api/v1/images/random', auth: 'api-key' },
  { label: 'API Key 适配图', method: 'GET', path: '/api/v1/images/adaptive', auth: 'api-key' },
  { label: 'AI 生图', method: 'POST', path: '/api/v1/ai/images', auth: 'api-key', body: '{\n  "prompt": "一张极简科技风的紫色星空图片",\n  "aspect_ratio": "1:1",\n  "count": 1,\n  "prompt_optimizer": true\n}' },
  { label: '图片列表', method: 'GET', path: '/api/v1/images', auth: 'api-key', query: 'page=1&per_page=20' },
  { label: '相册列表', method: 'GET', path: '/api/v1/albums', auth: 'api-key' },
  { label: '创建相册', method: 'POST', path: '/api/v1/albums', auth: 'api-key', body: '{\n  "name": "测试相册",\n  "intro": "API 测试创建"\n}' },
]

const baseUrl = ref(window.location.origin)
const method = ref<HttpMethod>('GET')
const path = ref('/api/ping')
const authMode = ref<AuthMode>('none')
const query = ref('')
const apiKey = ref('')
const readToken = ref('')
const body = ref('')
const sending = ref(false)
const cooldownRemaining = ref(0)
const status = ref<number | null>(null)
const elapsedMs = ref<number | null>(null)
const responseText = ref('')
const errorText = ref('')
const imagePreview = ref<ImagePreview | null>(null)
const bodyPlaceholder = '{\n  "name": "测试"\n}'
const lastAutoBody = ref('')
const randomAlbumId = ref('')
const randomOrientation = ref('any')
const randomRatio = ref('')
const randomMinWidth = ref('')
const randomMaxWidth = ref('')
const randomMinHeight = ref('')
const randomMaxHeight = ref('')
let cooldownTimer: number | undefined

const fullUrl = computed(() => {
  const trimmedBase = baseUrl.value.replace(/\/+$/, '')
  const rawPath = path.value.startsWith('/') ? path.value : `/${path.value}`
  const normalizedPath = rawPath.replace('{READ_TOKEN}', readToken.value.trim() || 'READ_TOKEN')
  const params = new URLSearchParams(query.value)
  const qString = params.toString()
  return `${trimmedBase}${normalizedPath}${qString ? `?${qString}` : ''}`
})

const requestCommand = computed(() => {
  const lines = [`curl -X ${method.value} ${shellQuote(fullUrl.value)}`]
  if (authMode.value === 'api-key') {
    lines.push(`  -H ${shellQuote(`X-Api-Key: ${apiKey.value.trim() || 'YOUR_API_KEY'}`)}`)
  }
  if (method.value !== 'GET' && body.value.trim()) {
    lines.push(`  -H ${shellQuote('Content-Type: application/json')}`)
    lines.push(`  -d ${shellQuote(body.value.trim())}`)
  }
  return lines.join(' \\\n')
})

const isRandomImageRequest = computed(() => path.value.replace(/\/+$/, '').endsWith('/images/random'))
const visiblePresets = computed(() => presets.filter((preset) => preset.auth === authMode.value))
const activeCategoryDescription = computed(() => categories.find((category) => category.value === authMode.value)?.desc || '')
const credentialLabel = computed(() => authMode.value === 'read-token' ? '图库只读令牌' : 'API Key')
const activeCredential = computed({
  get: () => authMode.value === 'read-token' ? readToken.value : apiKey.value,
  set: (value: string) => {
    if (authMode.value === 'read-token') {
      readToken.value = value
    } else {
      apiKey.value = value
    }
  },
})
const sendsBody = computed(() => method.value !== 'GET')

const bodyTemplates: Record<string, string> = {
  'POST /api/v1/register': '{\n  "name": "用户名",\n  "email": "user@example.com",\n  "password": "your-password"\n}',
  'POST /api/v1/albums': '{\n  "name": "测试相册",\n  "intro": "API 测试创建"\n}',
  'POST /api/v1/ai/images': '{\n  "prompt": "一张极简科技风的紫色星空图片",\n  "aspect_ratio": "1:1",\n  "count": 1,\n  "prompt_optimizer": true\n}',
  'PUT /api/v1/albums/{id}': '{\n  "name": "新名称",\n  "intro": "新简介"\n}',
  'PUT /api/v1/images/rename': '{\n  "key": "abc123",\n  "alias_name": "新名称.jpg"\n}',
  'PUT /api/v1/images/movement': '{\n  "keys": ["abc123", "def456"],\n  "album_id": 1\n}',
  'PUT /api/v1/images/permission': '{\n  "keys": ["abc123", "def456"],\n  "permission": 1\n}',
  'DELETE /api/v1/images': '{\n  "keys": ["abc123", "def456"]\n}',
}

function applyPreset(preset: Preset) {
  method.value = preset.method
  path.value = preset.path
  authMode.value = preset.auth
  query.value = preset.query || ''
  const template = preset.body || templateForRequest(preset.method, preset.path)
  body.value = template
  lastAutoBody.value = template
  status.value = null
  elapsedMs.value = null
  responseText.value = ''
  errorText.value = ''
  imagePreview.value = null
}

function selectCategory(category: AuthMode) {
  authMode.value = category
  const preset = presets.find((item) => item.auth === category)
  if (preset) applyPreset(preset)
}

function resetForm() {
  selectCategory(authMode.value)
}

function applyRandomFilters() {
  const params = new URLSearchParams()
  if (randomAlbumId.value.trim()) params.set('album_id', randomAlbumId.value.trim())
  if (randomOrientation.value !== 'any') params.set('orientation', randomOrientation.value)
  if (randomRatio.value.trim()) params.set('ratio', randomRatio.value.trim())
  if (randomMinWidth.value.trim()) params.set('min_width', randomMinWidth.value.trim())
  if (randomMaxWidth.value.trim()) params.set('max_width', randomMaxWidth.value.trim())
  if (randomMinHeight.value.trim()) params.set('min_height', randomMinHeight.value.trim())
  if (randomMaxHeight.value.trim()) params.set('max_height', randomMaxHeight.value.trim())
  query.value = params.toString()
}

function clearRandomFilters() {
  randomAlbumId.value = ''
  randomOrientation.value = 'any'
  randomRatio.value = ''
  randomMinWidth.value = ''
  randomMaxWidth.value = ''
  randomMinHeight.value = ''
  randomMaxHeight.value = ''
  query.value = ''
}

function normalizePath(value: string): string {
  return value.trim().replace(/\/+$/, '')
}

function templateForRequest(requestMethod = method.value, requestPath = path.value): string {
  return bodyTemplates[`${requestMethod} ${normalizePath(requestPath)}`] || ''
}

function shellQuote(value: string): string {
  return `'${value.replace(/'/g, `'\\''`)}'`
}

function syncBodyTemplate() {
  const next = sendsBody.value ? templateForRequest() : ''
  if (!body.value || body.value === lastAutoBody.value) {
    body.value = next
    lastAutoBody.value = next
  }
}

function formatResponse(text: string) {
  try {
    return JSON.stringify(JSON.parse(text), null, 2)
  } catch {
    return text
  }
}

function extractImagePreview(text: string): ImagePreview | null {
  try {
    const payload = JSON.parse(text)
    const image = Array.isArray(payload?.data?.images) ? payload.data.images[0] : payload?.data
    const url = image?.links?.url
    if (!url || typeof url !== 'string') return null
    return {
      url,
      name: image?.name || '',
      originName: image?.origin_name || '',
      width: image?.width,
      height: image?.height,
      size: image?.size,
    }
  } catch {
    return null
  }
}

function formatSize(kb?: number): string {
  if (!kb || kb <= 0) return ''
  if (kb >= 1048576) return `${(kb / 1048576).toFixed(2)} GB`
  if (kb >= 1024) return `${(kb / 1024).toFixed(2)} MB`
  return `${Math.round(kb)} KB`
}

async function sendRequest() {
  if (cooldownRemaining.value > 0) {
    toast.error(`请求过于频繁，请 ${cooldownRemaining.value} 秒后再试`)
    return
  }
  if (authMode.value === 'api-key' && !apiKey.value.trim()) {
    toast.error('请填写 API Key')
    return
  }
  if (authMode.value === 'read-token' && !readToken.value.trim()) {
    toast.error('请填写图库只读令牌')
    return
  }

  sending.value = true
  status.value = null
  elapsedMs.value = null
  responseText.value = ''
  errorText.value = ''
  imagePreview.value = null

  const headers: Record<string, string> = {}
  if (authMode.value === 'api-key') {
    headers['X-Api-Key'] = apiKey.value.trim()
  }

  const options: RequestInit = { method: method.value, headers }
  if (method.value !== 'GET' && body.value.trim()) {
    headers['Content-Type'] = 'application/json'
    options.body = body.value
  }

  const started = performance.now()
  try {
    const res = await fetch(fullUrl.value, options)
    status.value = res.status
    const text = await res.text()
    imagePreview.value = extractImagePreview(text)
    responseText.value = formatResponse(text)
  } catch (error: any) {
    errorText.value = error?.message || '请求失败'
    imagePreview.value = null
  } finally {
    elapsedMs.value = Math.round(performance.now() - started)
    sending.value = false
    startCooldown(2)
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

function copyResponse() {
  copyToClipboard(responseText.value || errorText.value)
  toast.success('响应内容已复制')
}

function copyRequestCommand() {
  copyToClipboard(requestCommand.value)
  toast.success('请求命令已复制')
}

watch([method, path], syncBodyTemplate)
</script>

<template>
  <div class="space-y-6 text-slate-100">
    <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-end">
      <div>
        <p class="text-sm font-medium text-purple-400">API Playground</p>
        <h1 class="mt-1 text-3xl font-semibold">API 测试台</h1>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-slate-400">
          图库只读令牌用于获取个人图片列表、随机图和设备适配图；写入与管理操作使用 API Key。
        </p>
      </div>
      <Button variant="outline" @click="$router.push('/api-doc')">
        <Terminal class="mr-2 h-4 w-4" /> 查看文档
      </Button>
    </div>

    <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_420px]">
      <div class="space-y-6">
        <Card class="border-white/10 bg-[#0f0f15] text-slate-100">
          <CardHeader>
            <CardTitle class="text-base">请求配置</CardTitle>
            <CardDescription class="text-slate-400">先选择调用类型，再选择同一认证范围内的接口</CardDescription>
          </CardHeader>
          <CardContent class="space-y-5">
            <div class="space-y-2">
              <div class="inline-flex w-fit max-w-full flex-wrap rounded-xl border border-white/10 bg-white/[0.03] p-1">
                <Button
                  v-for="category in categories"
                  :key="category.value"
                  type="button"
                  variant="ghost"
                  size="sm"
                  class="h-8 px-3"
                  :class="authMode === category.value ? 'bg-violet-500/15 text-violet-200' : 'text-muted-foreground'"
                  @click="selectCategory(category.value)"
                >
                  {{ category.label }}
                </Button>
              </div>
              <p class="text-xs text-slate-400">{{ activeCategoryDescription }}</p>
            </div>

            <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              <Button
                v-for="preset in visiblePresets"
                :key="preset.label"
                type="button"
                variant="outline"
                class="h-auto justify-start border-white/10 bg-[#09090d] p-3 text-left text-slate-100 hover:bg-white/5"
                @click="applyPreset(preset)"
              >
                <span>
                  <span class="block text-sm font-semibold">{{ preset.label }}</span>
                  <span class="mt-1 block font-mono text-xs text-slate-400">{{ preset.method }} {{ preset.path }}</span>
                </span>
              </Button>
            </div>

            <div class="grid gap-4 lg:grid-cols-[160px_1fr]">
              <div class="space-y-2">
                <Label>方法</Label>
                <Select v-model="method">
                  <SelectTrigger class="h-10 w-full border-white/10 bg-[#09090d] text-slate-100">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent class="border-white/10 bg-[#09090d] text-slate-100">
                    <SelectItem value="GET">GET</SelectItem>
                    <SelectItem value="POST">POST</SelectItem>
                    <SelectItem value="PUT">PUT</SelectItem>
                    <SelectItem value="DELETE">DELETE</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div class="space-y-2">
                <Label>Base URL</Label>
                <Input v-model="baseUrl" class="h-10 border-white/10 bg-[#09090d] text-slate-100" />
              </div>
            </div>

            <div class="space-y-2">
              <Label>路径</Label>
              <Input v-model="path" class="h-10 border-white/10 bg-[#09090d] font-mono text-slate-100 placeholder:text-slate-500" placeholder="/api/v1/gallery" />
            </div>

            <div class="space-y-2">
              <Label>Query 参数</Label>
              <Input v-model="query" class="h-10 border-white/10 bg-[#09090d] font-mono text-slate-100 placeholder:text-slate-500" placeholder="page=1&per_page=20" />
            </div>

            <div v-if="isRandomImageRequest" class="rounded-xl border border-white/10 bg-[#09090d] p-4">
              <div class="mb-4 flex flex-col justify-between gap-2 sm:flex-row sm:items-center">
                <div>
                  <p class="text-sm font-semibold">随机图筛选</p>
                  <p class="mt-1 text-xs text-slate-400">不填则完全随机；填写后会自动组合到 Query 参数。</p>
                </div>
                <div class="flex gap-2">
                  <Button variant="outline" size="sm" @click="clearRandomFilters">清空</Button>
                  <Button size="sm" @click="applyRandomFilters">应用筛选</Button>
                </div>
              </div>

              <div class="grid gap-4 lg:grid-cols-3">
                <div class="space-y-2">
                  <Label>画幅方向</Label>
                  <Select v-model="randomOrientation">
                    <SelectTrigger class="h-10 w-full border-white/10 bg-[#0f0f15] text-slate-100">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent class="border-white/10 bg-[#09090d] text-slate-100">
                      <SelectItem value="any">不限制</SelectItem>
                      <SelectItem value="landscape">横版 landscape</SelectItem>
                      <SelectItem value="portrait">竖版 portrait</SelectItem>
                      <SelectItem value="square">正方形 square</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div class="space-y-2">
                  <Label>画幅比例</Label>
                  <Input v-model="randomRatio" class="h-10 border-white/10 bg-[#0f0f15] font-mono text-slate-100 placeholder:text-slate-500" placeholder="16:9 / 4:3 / 1.778" />
                </div>
                <div class="space-y-2">
                  <Label>相册 ID</Label>
                  <Input v-model="randomAlbumId" class="h-10 border-white/10 bg-[#0f0f15] font-mono text-slate-100 placeholder:text-slate-500" placeholder="0 表示未分类" />
                </div>
              </div>

              <div class="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <div class="space-y-2">
                  <Label>最小宽度</Label>
                  <Input v-model="randomMinWidth" type="number" min="0" class="h-10 border-white/10 bg-[#0f0f15] font-mono text-slate-100 placeholder:text-slate-500" placeholder="min_width" />
                </div>
                <div class="space-y-2">
                  <Label>最大宽度</Label>
                  <Input v-model="randomMaxWidth" type="number" min="0" class="h-10 border-white/10 bg-[#0f0f15] font-mono text-slate-100 placeholder:text-slate-500" placeholder="max_width" />
                </div>
                <div class="space-y-2">
                  <Label>最小高度</Label>
                  <Input v-model="randomMinHeight" type="number" min="0" class="h-10 border-white/10 bg-[#0f0f15] font-mono text-slate-100 placeholder:text-slate-500" placeholder="min_height" />
                </div>
                <div class="space-y-2">
                  <Label>最大高度</Label>
                  <Input v-model="randomMaxHeight" type="number" min="0" class="h-10 border-white/10 bg-[#0f0f15] font-mono text-slate-100 placeholder:text-slate-500" placeholder="max_height" />
                </div>
              </div>
            </div>

            <div class="grid gap-4 lg:grid-cols-[180px_1fr]">
              <div class="space-y-2">
                <Label>认证方式</Label>
                <div class="flex h-10 items-center rounded-md border border-white/10 bg-[#09090d] px-3 text-sm text-slate-300">
                  {{ authMode === 'api-key' ? 'API Key (Header)' : authMode === 'read-token' ? '图库只读令牌 (URL)' : '无需认证' }}
                </div>
              </div>
              <div v-if="authMode !== 'none'" class="space-y-2">
                <Label>{{ credentialLabel }}</Label>
                <Input v-model="activeCredential" class="h-10 border-white/10 bg-[#09090d] font-mono text-slate-100 placeholder:text-slate-500" :placeholder="authMode === 'read-token' ? '图库只读令牌' : 'lsky-...'" />
              </div>
              <div v-else class="space-y-2">
                <Label>凭据</Label>
                <div class="flex h-10 items-center rounded-md border border-emerald-500/20 bg-emerald-500/5 px-3 text-sm text-emerald-200">
                  此类接口无需填写凭据
                </div>
              </div>
            </div>

            <div class="space-y-2">
              <div class="flex items-center justify-between gap-3">
                <Label>JSON Body</Label>
                <span class="text-xs text-slate-500">
                  {{ sendsBody ? '会随请求发送' : 'GET 请求不会发送 body' }}
                </span>
              </div>
              <Textarea
                v-model="body"
                class="min-h-48 border-white/10 bg-[#09090d] font-mono text-sm text-slate-100 placeholder:text-slate-500"
                :placeholder="bodyPlaceholder"
              />
            </div>
          </CardContent>
        </Card>
      </div>

      <aside class="space-y-4 xl:sticky xl:top-6 xl:self-start">
        <Card class="border-white/10 bg-[#0f0f15] text-slate-100">
          <CardHeader>
            <CardTitle class="text-base">即将发送</CardTitle>
            <CardDescription class="break-all font-mono text-xs text-slate-400">{{ fullUrl }}</CardDescription>
          </CardHeader>
          <CardContent class="flex gap-2">
            <Button class="flex-1" :disabled="sending || cooldownRemaining > 0" @click="sendRequest">
              <Play class="mr-2 h-4 w-4" />
              {{ sending ? '请求中...' : cooldownRemaining > 0 ? `${cooldownRemaining} 秒后重试` : '发送请求' }}
            </Button>
            <Button variant="outline" size="icon" @click="resetForm">
              <RotateCcw class="h-4 w-4" />
            </Button>
          </CardContent>
        </Card>

        <Card class="border-white/10 bg-[#0f0f15] text-slate-100">
          <CardHeader class="border-b border-white/10">
            <div class="flex items-center justify-between gap-3">
              <div>
                <CardTitle class="text-base">完整请求命令</CardTitle>
                <CardDescription class="text-slate-400">可直接复制到终端或接口文档中使用</CardDescription>
              </div>
              <Button variant="ghost" size="sm" @click="copyRequestCommand">
                <Copy class="mr-1 h-3 w-3" /> 复制
              </Button>
            </div>
          </CardHeader>
          <CardContent class="p-0">
            <pre class="max-h-56 overflow-auto bg-[#09090d] p-4 text-xs leading-relaxed text-slate-100"><code>{{ requestCommand }}</code></pre>
          </CardContent>
        </Card>

        <Card class="border-white/10 bg-[#0f0f15] text-slate-100">
          <CardHeader class="border-b border-white/10">
            <div class="flex items-center justify-between gap-3">
              <div>
                <CardTitle class="text-base">响应结果</CardTitle>
                <CardDescription class="text-slate-400">
                  <span v-if="status">HTTP {{ status }}</span>
                  <span v-if="elapsedMs !== null"> · {{ elapsedMs }}ms</span>
                  <span v-if="!status && !errorText">等待请求</span>
                </CardDescription>
              </div>
              <Badge v-if="status" :variant="status >= 200 && status < 300 ? 'default' : 'destructive'">
                {{ status }}
              </Badge>
            </div>
          </CardHeader>
          <CardContent class="p-0">
            <div v-if="imagePreview" class="border-b border-white/10 bg-[#09090d] p-3">
              <div class="mb-3 flex items-center justify-between gap-3">
                <div class="min-w-0">
                  <p class="truncate text-sm font-semibold">{{ imagePreview.originName || imagePreview.name || '图片预览' }}</p>
                  <p class="mt-1 text-xs text-slate-400">
                    <span v-if="imagePreview.width && imagePreview.height">{{ imagePreview.width }} × {{ imagePreview.height }}</span>
                    <span v-if="imagePreview.size"> · {{ formatSize(imagePreview.size) }}</span>
                  </p>
                </div>
                <Button variant="outline" size="sm" as-child>
                  <a :href="imagePreview.url" target="_blank" rel="noreferrer">
                    <ExternalLink class="mr-1 h-3.5 w-3.5" /> 打开
                  </a>
                </Button>
              </div>
              <div class="overflow-hidden rounded-xl border border-white/10 bg-black">
                <img
                  :src="imagePreview.url"
                  :alt="imagePreview.originName || imagePreview.name || 'API image preview'"
                  class="max-h-[360px] w-full object-contain"
                />
              </div>
            </div>
            <div class="flex justify-end border-b border-white/10 px-3 py-2">
              <Button variant="ghost" size="sm" :disabled="!responseText && !errorText" @click="copyResponse">
                <Copy class="mr-1 h-3 w-3" /> 复制
              </Button>
            </div>
            <pre class="max-h-[520px] min-h-64 overflow-auto bg-[#09090d] p-4 text-xs leading-relaxed text-slate-100"><code>{{ responseText || errorText || '点击“发送请求”后在这里查看响应。' }}</code></pre>
          </CardContent>
        </Card>
      </aside>
    </div>
  </div>
</template>
