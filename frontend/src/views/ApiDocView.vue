<script setup lang="ts">
import { computed, ref } from 'vue'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { toast } from 'vue-sonner'
import { Check, Copy, Terminal } from 'lucide-vue-next'
import FieldTable from '@/components/api/FieldTable.vue'
import { copyToClipboard } from '@/utils/clipboard'

type HttpMethod = 'GET' | 'POST'
type AuthMode = 'none' | 'api-key' | 'read-token'

interface Field {
  name: string
  type: string
  required?: string
  desc: string
}

interface Endpoint {
  label: string
  method: HttpMethod
  path: string
  desc: string
  auth: AuthMode
  query?: Field[]
  body?: Field[]
  bodyExample?: string
  response: Field[]
  responseExample: string
  notes?: Field[]
}

const defaultApiBaseUrl = window.location.port === '5173'
  ? `${window.location.protocol}//${window.location.hostname}:8000`
  : window.location.origin
const baseUrl = ref(defaultApiBaseUrl)
const apiKey = ref('')
const readToken = ref('')
const copied = ref<string | null>(null)

const endpoints: Endpoint[] = [
  {
    label: '健康检查',
    method: 'GET',
    path: '/api/ping',
    desc: '检查后端服务是否在线。该接口不需要 API Key。',
    auth: 'none',
    response: [
      { name: 'message', type: 'String', desc: '固定返回 pong' },
    ],
    responseExample: `{
  "message": "pong"
}`,
  },
  {
    label: '短令牌图片列表',
    method: 'GET',
    path: '/api/v1/t/{READ_TOKEN}/images',
    desc: '使用图库只读令牌获取令牌所属用户的图片列表。该令牌不能上传、修改或删除图片。',
    auth: 'read-token',
    query: [
      { name: 'page', type: 'Integer', required: '否', desc: '页码，默认 1' },
      { name: 'per_page', type: 'Integer', required: '否', desc: '每页数量，最大 100' },
      { name: 'album_id', type: 'Integer', required: '否', desc: '相册 ID，传 0 表示未分类' },
      { name: 'permission', type: 'Integer', required: '否', desc: '1 公开，0 私密' },
      { name: 'q', type: 'String', required: '否', desc: '按名称搜索' },
    ],
    response: [
      { name: 'data', type: 'Array<Image>', desc: '当前用户图片列表' },
      { name: 'current_page', type: 'Integer', desc: '当前页码' },
      { name: 'last_page', type: 'Integer', desc: '最后一页' },
      { name: 'total', type: 'Integer', desc: '总数量' },
    ],
    responseExample: `{
  "status": true,
  "message": "success",
  "data": {
    "data": [{ "key": "abc123", "links": { "url": "https://..." } }],
    "current_page": 1,
    "last_page": 1,
    "total": 1
  }
}`,
  },
  {
    label: '短令牌随机图',
    method: 'GET',
    path: '/api/v1/t/{READ_TOKEN}/images/random',
    desc: '使用图库只读令牌随机返回当前用户的一张图片，可用于壁纸脚本或桌面组件。',
    auth: 'read-token',
    query: [
      { name: 'album_id', type: 'Integer', required: '否', desc: '相册 ID，传 0 表示未分类' },
      { name: 'orientation', type: 'String', required: '否', desc: 'landscape / portrait / square' },
      { name: 'ratio', type: 'String', required: '否', desc: '16:9、4:3、3:2 或小数 1.778，容差 ±5%' },
      { name: 'min_width', type: 'Integer', required: '否', desc: '最小宽度 px' },
      { name: 'max_width', type: 'Integer', required: '否', desc: '最大宽度 px' },
      { name: 'min_height', type: 'Integer', required: '否', desc: '最小高度 px' },
      { name: 'max_height', type: 'Integer', required: '否', desc: '最大高度 px' },
    ],
    response: imageResponseFields(),
    responseExample: imageResponseExample(),
  },
  {
    label: '短令牌设备自适应图',
    method: 'GET',
    path: '/api/v1/t/{READ_TOKEN}/images/adaptive',
    desc: '使用图库只读令牌按请求 User-Agent 返回当前用户适配画幅的图片。',
    auth: 'read-token',
    notes: [
      { name: 'iPhone / iPod', type: 'portrait', desc: '竖版图片，优先匹配竖向比例' },
      { name: 'Android 手机', type: 'portrait', desc: '竖版图片，优先匹配竖向比例' },
      { name: 'iPad', type: '4:3', desc: '横版 4:3 图片' },
      { name: 'Windows / Mac', type: '16:9', desc: '横版 16:9 图片' },
    ],
    response: imageResponseFields(),
    responseExample: imageResponseExample(),
  },
  {
    label: '随机图',
    method: 'GET',
    path: '/api/v1/images/random',
    desc: '随机返回当前 API Key 所属用户的一张图片。支持按相册、方向、比例和尺寸筛选。',
    auth: 'api-key',
    query: [
      { name: 'album_id', type: 'Integer', required: '否', desc: '相册 ID，传 0 表示未分类' },
      { name: 'orientation', type: 'String', required: '否', desc: 'landscape / portrait / square' },
      { name: 'ratio', type: 'String', required: '否', desc: '16:9、4:3、3:2 或小数 1.778，容差 ±5%' },
      { name: 'min_width', type: 'Integer', required: '否', desc: '最小宽度 px' },
      { name: 'max_width', type: 'Integer', required: '否', desc: '最大宽度 px' },
      { name: 'min_height', type: 'Integer', required: '否', desc: '最小高度 px' },
      { name: 'max_height', type: 'Integer', required: '否', desc: '最大高度 px' },
    ],
    response: imageResponseFields(),
    responseExample: imageResponseExample(),
  },
  {
    label: '设备自适应图',
    method: 'GET',
    path: '/api/v1/images/adaptive',
    desc: '根据请求 User-Agent 自动返回适配设备画幅的图片。回退规则：比例不匹配 -> 仅按方向 -> 完全随机。',
    auth: 'api-key',
    notes: [
      { name: 'iPhone / iPod', type: 'portrait', desc: '竖版图片，优先匹配竖向比例' },
      { name: 'Android 手机', type: 'portrait', desc: '竖版图片，优先匹配竖向比例' },
      { name: 'iPad', type: '4:3', desc: '横版 4:3 图片' },
      { name: 'Android 平板', type: '16:9', desc: '横版 16:9 图片' },
      { name: 'Windows / Mac', type: '16:9', desc: '横版 16:9 图片' },
      { name: '其他设备', type: 'landscape', desc: '横版 fallback；比例不匹配时会继续回退到方向或完全随机' },
    ],
    response: imageResponseFields(),
    responseExample: imageResponseExample(),
  },
  {
    label: 'AI 生图',
    method: 'POST',
    path: '/api/v1/ai/images',
    desc: '调用管理员配置的 AI 生图渠道生成图片，并自动保存到当前用户的“AI 生成”相册。',
    auth: 'api-key',
    body: [
      { name: 'prompt', type: 'String', required: '是', desc: '提示词，最多 1500 字符' },
      { name: 'aspect_ratio', type: 'String', required: '否', desc: '1:1 / 16:9 / 4:3 / 3:2 / 2:3 / 3:4 / 9:16 / 21:9' },
      { name: 'count', type: 'Integer', required: '否', desc: '生成数量，受后台单次最大数量限制' },
      { name: 'prompt_optimizer', type: 'Boolean', required: '否', desc: '是否启用提示词优化，仅 MiniMax 渠道生效' },
    ],
    bodyExample: `{
  "prompt": "一张极简科技风的紫色星空图片",
  "aspect_ratio": "1:1",
  "count": 1,
  "prompt_optimizer": true
}`,
    response: [
      { name: 'provider', type: 'String', desc: '实际使用的生图渠道：minimax / openai / siliconflow / compatible' },
      { name: 'album', type: 'Object', desc: 'AI 生成相册信息' },
      { name: 'images', type: 'Array<Image>', desc: '生成并保存后的图片列表' },
    ],
    responseExample: `{
  "status": true,
  "message": "生成成功",
  "data": {
    "provider": "minimax",
    "album": { "id": 4, "name": "AI 生成" },
    "images": [{ "key": "abc123", "links": { "url": "https://..." } }]
  }
}`,
  },
  {
    label: '上传图片',
    method: 'POST',
    path: '/api/v1/upload',
    desc: '上传图片。请在 X-Api-Key 请求头中携带 API Key。',
    auth: 'api-key',
    body: [
      { name: 'file', type: 'File', required: '是', desc: '图片文件（必须使用 multipart/form-data 格式）' },
      { name: 'strategy_id', type: 'Integer', required: '否', desc: '存储策略 ID，不传则使用默认策略' },
      { name: 'album_id', type: 'Integer', required: '否', desc: '相册 ID，不传则默认未分类' },
      { name: 'permission', type: 'Integer', required: '否', desc: '1 为公开，0 为私密' },
    ],
    response: imageResponseFields(),
    responseExample: imageResponseExample(),
  },
  {
    label: '图片列表',
    method: 'GET',
    path: '/api/v1/images',
    desc: '获取当前用户图片列表。',
    auth: 'api-key',
    query: [
      { name: 'page', type: 'Integer', required: '否', desc: '页码，默认 1' },
      { name: 'per_page', type: 'Integer', required: '否', desc: '每页数量，最大 100' },
      { name: 'album_id', type: 'Integer', required: '否', desc: '相册 ID，传 0 表示未分类' },
      { name: 'permission', type: 'Integer', required: '否', desc: '1 公开，0 私密' },
      { name: 'q', type: 'String', required: '否', desc: '按名称搜索' },
      { name: 'sort', type: 'String', required: '否', desc: 'newest / earliest / utmost / least' },
    ],
    response: [
      { name: 'data', type: 'Array<Image>', desc: '图片列表' },
      { name: 'current_page', type: 'Integer', desc: '当前页码' },
      { name: 'last_page', type: 'Integer', desc: '最后一页' },
      { name: 'total', type: 'Integer', desc: '总数量' },
    ],
    responseExample: `{
  "status": true,
  "message": "success",
  "data": {
    "data": [{ "key": "abc123", "url": "https://..." }],
    "current_page": 1,
    "last_page": 1,
    "total": 1
  }
}`,
  },
  {
    label: '相册列表',
    method: 'GET',
    path: '/api/v1/albums',
    desc: '获取当前用户的所有相册。',
    auth: 'api-key',
    response: [
      { name: 'id', type: 'Integer', desc: '相册 ID' },
      { name: 'name', type: 'String', desc: '相册名称' },
      { name: 'intro', type: 'String', desc: '相册简介' },
      { name: 'image_num', type: 'Integer', desc: '图片数量' },
      { name: 'created_at', type: 'String', desc: '创建时间' },
    ],
    responseExample: `{
  "status": true,
  "message": "success",
  "data": [
    { "id": 1, "name": "测试相册", "intro": "API 测试创建", "image_num": 0 }
  ]
}`,
  },
  {
    label: '创建相册',
    method: 'POST',
    path: '/api/v1/albums',
    desc: '创建一个新相册。',
    auth: 'api-key',
    body: [
      { name: 'name', type: 'String', required: '是', desc: '相册名称' },
      { name: 'intro', type: 'String', required: '否', desc: '相册简介' },
    ],
    bodyExample: `{
  "name": "测试相册",
  "intro": "API 测试创建"
}`,
    response: [
      { name: 'id', type: 'Integer', desc: '相册 ID' },
      { name: 'name', type: 'String', desc: '相册名称' },
      { name: 'intro', type: 'String', desc: '相册简介' },
      { name: 'image_num', type: 'Integer', desc: '图片数量' },
    ],
    responseExample: `{
  "status": true,
  "message": "创建成功",
  "data": { "id": 1, "name": "测试相册", "intro": "API 测试创建" }
}`,
  },
]

const endpointCount = computed(() => endpoints.length)

const authFields: Field[] = [
  { name: 'X-Api-Key', type: 'String', required: '是', desc: '在请求 Header 中传入用户 API Key。仅标记“需 API Key”的接口需要携带。' },
]

const readTokenFields: Field[] = [
  { name: 'READ_TOKEN', type: 'String', required: '是', desc: '在 API Key 页面创建或重置的图库只读令牌，放在请求路径中。仅可读取图片。' },
]

const linkFields: Field[] = [
  { name: 'links.url', type: 'String', desc: '公开图返回永久直链；私密图返回 /api/v1/images/{短令牌}/{key}，持有链接可读取该单张图' },
  { name: 'links.html', type: 'String', desc: 'HTML img 标签代码' },
  { name: 'links.markdown', type: 'String', desc: 'Markdown 图片代码' },
  { name: 'links.bbcode', type: 'String', desc: 'BBCode 图片代码' },
  { name: 'links.thumbnail_url', type: 'String', desc: '缩略图地址；私密图使用同样的短令牌链接' },
]

function imageResponseFields(): Field[] {
  return [
    { name: 'key', type: 'String', desc: '图片唯一密钥' },
    { name: 'name', type: 'String', desc: '图片名称' },
    { name: 'origin_name', type: 'String', desc: '原始文件名' },
    { name: 'pathname', type: 'String', desc: '图片路径名' },
    { name: 'size', type: 'Float', desc: '图片大小，单位 KB' },
    { name: 'width', type: 'Integer', desc: '图片宽度 px' },
    { name: 'height', type: 'Integer', desc: '图片高度 px' },
    { name: 'md5', type: 'String', desc: '图片 md5 值' },
    { name: 'sha1', type: 'String', desc: '图片 sha1 值' },
    { name: 'links', type: 'Object', desc: '链接对象，含 url / html / markdown / bbcode 等' },
  ]
}

function imageResponseExample(): string {
  return `{
  "status": true,
  "message": "success",
  "data": {
    "key": "abc123",
    "name": "a1b2c3.jpg",
    "origin_name": "photo.jpg",
    "pathname": "user_1/2026/05/23/a1b2c3.jpg",
    "size": 256.5,
    "width": 1920,
    "height": 1080,
    "md5": "...",
    "sha1": "...",
    "links": {
      "url": "http://localhost:8000/api/v1/images/u9Fd2aK1mZqP/abc123",
      "html": "<img src=\\"...\\" />",
      "markdown": "![](...)",
      "bbcode": "[img]...[/img]"
    }
  }
}`
}

function fullUrl(ep: Endpoint): string {
  const path = ep.path.replace('{READ_TOKEN}', readToken.value || 'READ_TOKEN')
  return `${baseUrl.value.replace(/\/+$/, '')}${path}`
}

function curlExample(ep: Endpoint): string {
  const headers = ep.auth === 'api-key' ? ` \\\n  -H 'X-Api-Key: ${apiKey.value || 'YOUR_API_KEY'}'` : ''
  const body = ep.bodyExample ? ` \\\n  -H 'Content-Type: application/json' \\\n  -d '${ep.bodyExample.replace(/\n/g, ' ')}'` : ''
  return `curl -X ${ep.method} '${fullUrl(ep)}'${headers}${body}`
}

function copyText(text: string, key: string) {
  copyToClipboard(text)
  copied.value = key
  toast.success('已复制')
  setTimeout(() => {
    if (copied.value === key) copied.value = null
  }, 1500)
}
</script>

<template>
  <div class="space-y-6 text-slate-100">
    <div class="flex flex-col justify-between gap-4 lg:flex-row lg:items-end">
      <div>
        <p class="text-sm font-medium text-purple-400">Developer API</p>
        <h1 class="mt-1 text-3xl font-semibold">API 文档</h1>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-slate-400">
          图库只读令牌（短令牌）只读取图片列表、随机图和设备适配图；上传、AI 生图和管理操作必须使用 API Key。私密图返回的单图链接可直接用于展示。
        </p>
      </div>
      <Button variant="outline" @click="$router.push('/api-test')">
        <Terminal class="mr-2 h-4 w-4" /> 打开测试台
      </Button>
    </div>

    <Card class="border-white/10 bg-[#0f0f15] text-slate-100">
      <CardContent class="grid gap-4 p-4 lg:grid-cols-[1fr_1fr_1fr_120px]">
        <div class="space-y-1.5">
          <Label class="text-xs text-slate-400">Base URL</Label>
          <Input v-model="baseUrl" class="h-10 border-white/10 bg-[#09090d] text-slate-100" />
        </div>
        <div class="space-y-1.5">
          <Label class="text-xs text-slate-400">API Key</Label>
          <Input v-model="apiKey" placeholder="lsky-..." class="h-10 border-white/10 bg-[#09090d] text-slate-100 placeholder:text-slate-500" />
        </div>
        <div class="space-y-1.5">
          <Label class="text-xs text-slate-400">图库只读令牌</Label>
          <Input v-model="readToken" placeholder="READ_TOKEN" class="h-10 border-white/10 bg-[#09090d] text-slate-100 placeholder:text-slate-500" />
        </div>
        <div class="rounded-xl border border-white/10 bg-white/5 p-3 text-sm">
          <p class="font-semibold">{{ endpointCount }}</p>
          <p class="text-slate-400">个接口</p>
        </div>
      </CardContent>
    </Card>

    <Card class="border-white/10 bg-[#0f0f15] text-slate-100">
      <CardContent class="grid gap-3 p-4 lg:grid-cols-3">
        <div class="rounded-lg border border-emerald-500/20 bg-emerald-500/5 p-3">
          <p class="text-sm font-semibold text-emerald-200">健康检查</p>
          <p class="mt-1 text-xs leading-5 text-slate-400">下方测试接口中，服务存活检查无需认证，可直接请求。</p>
        </div>
        <div class="rounded-lg border border-blue-500/20 bg-blue-500/5 p-3">
          <p class="text-sm font-semibold text-blue-200">API Key</p>
          <p class="mt-1 text-xs leading-5 text-slate-400">访问账户图库、随机图、设备适配图、AI 生图及管理操作。</p>
        </div>
        <div class="rounded-lg border border-amber-500/20 bg-amber-500/5 p-3">
          <p class="text-sm font-semibold text-amber-200">图库只读令牌（短令牌）</p>
          <p class="mt-1 text-xs leading-5 text-slate-400">供脚本读取图片列表、随机图和适配图，不能写入或管理内容。</p>
        </div>
      </CardContent>
    </Card>

    <section class="space-y-4">
      <Card v-for="ep in endpoints" :key="ep.method + ep.path" class="overflow-hidden border-white/10 bg-[#0f0f15] text-slate-100">
        <CardHeader class="border-b border-white/10">
          <div class="flex flex-wrap items-center gap-3">
            <Badge :class="ep.method === 'GET' ? 'bg-emerald-500 text-white hover:bg-emerald-500' : 'bg-blue-500 text-white hover:bg-blue-500'">
              {{ ep.method }}
            </Badge>
            <code class="rounded-lg border border-white/10 bg-[#09090d] px-2.5 py-1 text-sm font-semibold">{{ ep.path }}</code>
            <Badge
              variant="outline"
              :class="ep.auth === 'api-key'
                ? 'border-blue-400/30 bg-blue-400/10 text-blue-200'
                : ep.auth === 'read-token'
                  ? 'border-amber-400/30 bg-amber-400/10 text-amber-200'
                  : 'border-emerald-400/30 bg-emerald-400/10 text-emerald-200'"
            >
              {{ ep.auth === 'api-key' ? '需 API Key' : ep.auth === 'read-token' ? '需短令牌' : '无需认证' }}
            </Badge>
          </div>
          <CardTitle class="pt-2 text-lg">{{ ep.label }}</CardTitle>
          <CardDescription class="leading-6 text-slate-400">{{ ep.desc }}</CardDescription>
        </CardHeader>
        <CardContent class="space-y-5 p-5">
          <div v-if="ep.auth === 'api-key'" class="space-y-2">
            <h3 class="text-xs font-semibold uppercase text-slate-400">认证参数</h3>
            <FieldTable :fields="authFields" location="Header" />
          </div>
          <div v-if="ep.auth === 'read-token'" class="space-y-2">
            <h3 class="text-xs font-semibold uppercase text-slate-400">认证参数</h3>
            <FieldTable :fields="readTokenFields" location="Path" />
          </div>

          <div v-if="ep.query?.length" class="space-y-2">
            <h3 class="text-xs font-semibold uppercase text-slate-400">Query 参数</h3>
            <FieldTable :fields="ep.query" location="Query" />
          </div>

          <div v-if="ep.body?.length" class="space-y-2">
            <h3 class="text-xs font-semibold uppercase text-slate-400">JSON Body</h3>
            <FieldTable :fields="ep.body" location="Body" />
          </div>

          <div v-if="ep.notes?.length" class="space-y-2">
            <h3 class="text-xs font-semibold uppercase text-slate-400">匹配规则</h3>
            <FieldTable :fields="ep.notes" location="User-Agent" name-label="设备" type-label="目标画幅" />
          </div>

          <div v-if="ep.bodyExample" class="space-y-2">
            <div class="flex items-center justify-between">
              <h3 class="text-xs font-semibold uppercase text-slate-400">请求示例</h3>
              <Button variant="ghost" size="sm" @click="copyText(ep.bodyExample, ep.path + 'body')">
                <component :is="copied === ep.path + 'body' ? Check : Copy" class="mr-1 h-3 w-3" /> 复制
              </Button>
            </div>
            <pre class="overflow-x-auto rounded-xl border border-white/10 bg-[#09090d] p-4 text-xs leading-relaxed"><code>{{ ep.bodyExample }}</code></pre>
          </div>

          <div class="space-y-2">
            <h3 class="text-xs font-semibold uppercase text-slate-400">返回参数</h3>
            <FieldTable :fields="ep.response" location="Response" />
          </div>

          <div v-if="ep.response.some((field) => field.name === 'links')" class="space-y-2">
            <h3 class="text-xs font-semibold uppercase text-slate-400">links 字段说明</h3>
            <FieldTable :fields="linkFields" location="Response" />
          </div>

          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <h3 class="text-xs font-semibold uppercase text-slate-400">返回示例</h3>
              <Button variant="ghost" size="sm" @click="copyText(ep.responseExample, ep.path + 'response')">
                <component :is="copied === ep.path + 'response' ? Check : Copy" class="mr-1 h-3 w-3" /> 复制
              </Button>
            </div>
            <pre class="overflow-x-auto rounded-xl border border-white/10 bg-[#09090d] p-4 text-xs leading-relaxed"><code>{{ ep.responseExample }}</code></pre>
          </div>

          <div class="rounded-xl border border-white/10 bg-[#09090d] p-3">
            <div class="mb-2 flex items-center justify-between">
              <span class="flex items-center gap-1.5 text-xs font-semibold text-slate-400">
                <Terminal class="h-3.5 w-3.5" /> cURL
              </span>
              <Button variant="ghost" size="sm" @click="copyText(curlExample(ep), ep.path + 'curl')">
                <component :is="copied === ep.path + 'curl' ? Check : Copy" class="mr-1 h-3 w-3" /> 复制
              </Button>
            </div>
            <pre class="overflow-x-auto text-xs leading-relaxed text-slate-400">{{ curlExample(ep) }}</pre>
          </div>
        </CardContent>
      </Card>
    </section>
  </div>
</template>
