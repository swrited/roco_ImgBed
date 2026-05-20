<script setup lang="ts">
import { ref, computed } from 'vue'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { toast } from 'vue-sonner'
import { Copy, Check, Terminal, Image, FolderOpen, KeyRound, CloudUpload, FileImage, Shield, Trash2, Pencil, FolderInput, Eye, EyeOff, BarChart3 } from 'lucide-vue-next'

const baseUrl = ref(`${window.location.protocol}//${window.location.host}`)
const token = ref('')

type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

interface Endpoint {
  method: HttpMethod
  path: string
  title: string
  desc: string
  auth: boolean
  body?: string
  params?: { name: string; type: string; required: string; desc: string }[]
  query?: { name: string; default?: string; desc: string }[]
  response?: string
}

const methodColors: Record<HttpMethod, string> = {
  GET: 'bg-emerald-500 hover:bg-emerald-500 text-white',
  POST: 'bg-blue-500 hover:bg-blue-500 text-white',
  PUT: 'bg-amber-500 hover:bg-amber-500 text-white',
  DELETE: 'bg-red-500 hover:bg-red-500 text-white',
}

const sections = [
  {
    key: 'auth',
    label: '认证',
    icon: KeyRound,
    endpoints: [
      {
        method: 'POST' as HttpMethod,
        path: '/api/v1/tokens',
        title: '获取 Token',
        desc: '使用邮箱和密码换取访问令牌，后续请求需在 Header 中携带 Authorization: Bearer {token}',
        auth: false,
        body: `{
  "email": "admin@admin.com",
  "password": "123456"
}`,
        response: `{
  "status": true,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer"
  }
}`,
      },
      {
        method: 'POST' as HttpMethod,
        path: '/api/v1/register',
        title: '用户注册',
        desc: '注册新用户账户',
        auth: false,
        body: `{
  "name": "用户名",
  "email": "user@example.com",
  "password": "your-password"
}`,
      },
      {
        method: 'DELETE' as HttpMethod,
        path: '/api/v1/tokens',
        title: '退出登录',
        desc: '使当前 Token 失效',
        auth: true,
      },
    ] as Endpoint[],
  },
  {
    key: 'upload',
    label: '上传',
    icon: CloudUpload,
    endpoints: [
      {
        method: 'POST' as HttpMethod,
        path: '/api/v1/upload',
        title: '上传图片',
        desc: '支持 multipart/form-data 方式上传图片文件。游客上传需在后台开启对应开关。',
        auth: false,
        params: [
          { name: 'file', type: 'File', required: '是', desc: '图片文件（jpg/png/gif/webp 等）' },
          { name: 'strategy_id', type: 'Int', required: '否', desc: '指定存储策略 ID，留空使用默认策略' },
          { name: 'album_id', type: 'Int', required: '否', desc: '指定归属相册 ID' },
        ],
        response: `{
  "status": true,
  "message": "上传成功",
  "data": {
    "key": "abc123",
    "name": "2025/05/20/a1b2c3d4e5f6.jpg",
    "pathname": "2025/05/20/a1b2c3d4e5f6.jpg",
    "origin_name": "photo.jpg",
    "size": 256.5,
    "mimetype": "image/jpeg",
    "extension": "jpg",
    "md5": "...",
    "sha1": "...",
    "links": {
      "url": "https://example.com/i/2025/05/20/a1b2c3d4e5f6.jpg",
      "html": "<img src=\"...\" />",
      "bbcode": "[img]...[/img]",
      "markdown": "![](...)",
      "markdown_with_link": "[![](...)](...)",
      "thumbnail_url": "..."
    }
  }
}`,
      },
    ] as Endpoint[],
  },
  {
    key: 'images',
    label: '图片',
    icon: FileImage,
    endpoints: [
      {
        method: 'GET' as HttpMethod,
        path: '/api/v1/images',
        title: '图片列表',
        desc: '获取当前登录用户的图片列表，支持分页、相册筛选和排序。',
        auth: true,
        query: [
          { name: 'page', default: '1', desc: '页码' },
          { name: 'per_page', default: '20', desc: '每页数量（最大 100）' },
          { name: 'album_id', desc: '按相册筛选' },
          { name: 'permission', desc: '按权限筛选（1=公开，0=私密）' },
          { name: 'q', desc: '按名称搜索' },
          { name: 'sort', default: 'newest', desc: '排序：newest/earliest/utmost/least' },
        ],
      },
      {
        method: 'PUT' as HttpMethod,
        path: '/api/v1/images/rename',
        title: '重命名',
        desc: '修改图片别名',
        auth: true,
        body: `{
  "key": "abc123",
  "alias_name": "新名称.jpg"
}`,
      },
      {
        method: 'PUT' as HttpMethod,
        path: '/api/v1/images/movement',
        title: '移动到相册',
        desc: '批量移动图片到指定相册，album_id 传 null 则移出相册',
        auth: true,
        body: `{
  "keys": ["abc123", "def456"],
  "album_id": 1
}`,
      },
      {
        method: 'PUT' as HttpMethod,
        path: '/api/v1/images/permission',
        title: '设置权限',
        desc: '批量设置图片公开或私密',
        auth: true,
        body: `{
  "keys": ["abc123", "def456"],
  "permission": 1
}`,
      },
      {
        method: 'DELETE' as HttpMethod,
        path: '/api/v1/images/{key}',
        title: '删除单张',
        desc: '根据 key 删除单张图片',
        auth: true,
      },
      {
        method: 'DELETE' as HttpMethod,
        path: '/api/v1/images',
        title: '批量删除',
        desc: '根据 keys 数组批量删除',
        auth: true,
        body: `{
  "keys": ["abc123", "def456"]
}`,
      },
    ] as Endpoint[],
  },
  {
    key: 'albums',
    label: '相册',
    icon: FolderOpen,
    endpoints: [
      {
        method: 'GET' as HttpMethod,
        path: '/api/v1/albums',
        title: '相册列表',
        desc: '获取当前用户的所有相册',
        auth: true,
      },
      {
        method: 'POST' as HttpMethod,
        path: '/api/v1/albums',
        title: '创建相册',
        desc: '创建新相册',
        auth: true,
        body: `{
  "name": "旅行照片",
  "intro": "2024 年旅行记录"
}`,
      },
      {
        method: 'PUT' as HttpMethod,
        path: '/api/v1/albums/{id}',
        title: '更新相册',
        desc: '修改相册名称或简介',
        auth: true,
        body: `{
  "name": "新名称",
  "intro": "新简介"
}`,
      },
      {
        method: 'DELETE' as HttpMethod,
        path: '/api/v1/albums/{id}',
        title: '删除相册',
        desc: '删除相册（相册内图片不会被删除）',
        auth: true,
      },
    ] as Endpoint[],
  },
  {
    key: 'public',
    label: '公开',
    icon: Image,
    endpoints: [
      {
        method: 'GET' as HttpMethod,
        path: '/api/v1/gallery',
        title: '公开画廊',
        desc: '获取所有用户设置为公开的图片列表，无需登录',
        auth: false,
        query: [
          { name: 'page', default: '1', desc: '页码' },
          { name: 'per_page', default: '20', desc: '每页数量' },
          { name: 'q', desc: '按名称搜索' },
          { name: 'user_id', desc: '按用户筛选' },
        ],
      },
      {
        method: 'GET' as HttpMethod,
        path: '/api/v1/strategies',
        title: '存储策略列表',
        desc: '获取系统启用的存储策略列表，无需登录',
        auth: false,
      },
    ] as Endpoint[],
  },
  {
    key: 'user',
    label: '用户',
    icon: Shield,
    endpoints: [
      {
        method: 'GET' as HttpMethod,
        path: '/api/v1/profile',
        title: '个人信息',
        desc: '获取当前登录用户的基本信息',
        auth: true,
      },
      {
        method: 'PUT' as HttpMethod,
        path: '/api/v1/profile',
        title: '更新信息',
        desc: '修改用户名称、头像等信息',
        auth: true,
        body: `{
  "name": "新昵称",
  "avatar": "https://..."
}`,
      },
      {
        method: 'GET' as HttpMethod,
        path: '/api/v1/dashboard',
        title: '仪表盘数据',
        desc: '获取用户统计概览（图片数、占用空间等）',
        auth: true,
      },
      {
        method: 'GET' as HttpMethod,
        path: '/api/v1/user/settings',
        title: '用户设置',
        desc: '获取当前用户的偏好设置',
        auth: true,
      },
      {
        method: 'PUT' as HttpMethod,
        path: '/api/v1/user/settings',
        title: '更新设置',
        desc: '修改用户偏好设置',
        auth: true,
      },
      {
        method: 'PUT' as HttpMethod,
        path: '/api/v1/user/settings/strategy',
        title: '设置默认策略',
        desc: '设置用户默认使用的存储策略',
        auth: true,
        body: `{
  "strategy_id": 2
}`,
      },
    ] as Endpoint[],
  },
]

const activeTab = ref('auth')

function copyCode(text: string) {
  navigator.clipboard.writeText(text)
  toast.success('已复制到剪贴板')
}

function curlExample(ep: Endpoint): string {
  const headers = ep.auth ? ` -H "Authorization: Bearer ${token.value || 'YOUR_TOKEN'}"` : ''
  const url = `${baseUrl.value}${ep.path}`
  if (ep.method === 'GET') {
    return `curl -X GET "${url}"${headers}`
  }
  if (ep.body) {
    return `curl -X ${ep.method} "${url}"${headers} -H "Content-Type: application/json" -d '${ep.body.replace(/\n/g, ' ')}'`
  }
  return `curl -X ${ep.method} "${url}"${headers}`
}

const copied = ref<string | null>(null)
function copyCurl(ep: Endpoint) {
  const text = curlExample(ep)
  navigator.clipboard.writeText(text)
  copied.value = ep.path + ep.method
  setTimeout(() => copied.value = null, 1500)
}
</script>

<template>
  <div class="max-w-4xl">
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold">API 文档</h1>
        <p class="text-sm text-muted-foreground mt-1">洛克图床 REST API 接口参考</p>
      </div>
      <Badge variant="secondary" class="text-xs">v1</Badge>
    </div>

    <!-- Base URL & Token config -->
    <Card class="mb-6 border-white/5">
      <CardContent class="p-4">
        <div class="grid gap-4 sm:grid-cols-2">
          <div class="space-y-1.5">
            <Label class="text-xs font-medium text-muted-foreground">Base URL</Label>
            <Input v-model="baseUrl" class="h-9" />
          </div>
          <div class="space-y-1.5">
            <Label class="text-xs font-medium text-muted-foreground">测试 Token（用于生成 curl 示例）</Label>
            <Input v-model="token" placeholder="Bearer token..." class="h-9" />
          </div>
        </div>
      </CardContent>
    </Card>

    <Tabs v-model="activeTab">
      <TabsList class="mb-5 flex-wrap h-auto gap-y-1 bg-[#0f0f15] border border-white/5 p-1 rounded-xl">
        <TabsTrigger
          v-for="sec in sections" :key="sec.key" :value="sec.key"
          class="rounded-lg text-xs px-3 py-1.5 data-[state=active]:shadow-sm"
        >
          <component :is="sec.icon" class="mr-1.5 h-3.5 w-3.5 inline" />
          {{ sec.label }}
        </TabsTrigger>
      </TabsList>

      <TabsContent v-for="sec in sections" :key="sec.key" :value="sec.key" class="space-y-4">
        <Card v-for="ep in sec.endpoints" :key="ep.path + ep.method" class="border-white/5 overflow-hidden">
          <CardHeader class="pb-3">
            <div class="flex flex-wrap items-center gap-3">
              <Badge :class="['font-mono text-xs px-2 py-0.5', methodColors[ep.method]]">
                {{ ep.method }}
              </Badge>
              <code class="text-sm font-semibold text-foreground">{{ ep.path }}</code>
              <Badge v-if="ep.auth" variant="outline" class="text-xs text-amber-600 border-amber-200 bg-amber-50">
                需登录
              </Badge>
              <Badge v-else variant="outline" class="text-xs text-emerald-600 border-emerald-200 bg-emerald-50">
                公开
              </Badge>
            </div>
            <CardTitle class="text-base mt-2">{{ ep.title }}</CardTitle>
            <CardDescription>{{ ep.desc }}</CardDescription>
          </CardHeader>

          <CardContent class="space-y-4 pt-0">
            <!-- Request Params -->
            <div v-if="ep.params?.length">
              <h4 class="text-xs font-semibold uppercase text-muted-foreground mb-2">请求参数 (form-data / body)</h4>
              <div class="overflow-x-auto rounded-lg border border-white/5">
                <table class="w-full text-sm">
                  <thead class="bg-white/5">
                    <tr>
                      <th class="text-left px-3 py-2 font-medium">参数</th>
                      <th class="text-left px-3 py-2 font-medium">类型</th>
                      <th class="text-left px-3 py-2 font-medium">必填</th>
                      <th class="text-left px-3 py-2 font-medium">说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="p in ep.params" :key="p.name" class="border-t border-white/5">
                      <td class="px-3 py-2 font-mono text-xs">{{ p.name }}</td>
                      <td class="px-3 py-2 text-xs">{{ p.type }}</td>
                      <td class="px-3 py-2 text-xs">{{ p.required }}</td>
                      <td class="px-3 py-2 text-xs text-muted-foreground">{{ p.desc }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Query Params -->
            <div v-if="ep.query?.length">
              <h4 class="text-xs font-semibold uppercase text-muted-foreground mb-2">查询参数</h4>
              <div class="overflow-x-auto rounded-lg border border-white/5">
                <table class="w-full text-sm">
                  <thead class="bg-white/5">
                    <tr>
                      <th class="text-left px-3 py-2 font-medium">参数</th>
                      <th class="text-left px-3 py-2 font-medium">默认</th>
                      <th class="text-left px-3 py-2 font-medium">说明</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="q in ep.query" :key="q.name" class="border-t border-white/5">
                      <td class="px-3 py-2 font-mono text-xs">{{ q.name }}</td>
                      <td class="px-3 py-2 text-xs">{{ q.default || '-' }}</td>
                      <td class="px-3 py-2 text-xs text-muted-foreground">{{ q.desc }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Request Body -->
            <div v-if="ep.body">
              <div class="flex items-center justify-between mb-2">
                <h4 class="text-xs font-semibold uppercase text-muted-foreground">请求体</h4>
                <Button variant="ghost" size="sm" class="h-7 text-xs" @click="copyCode(ep.body)">
                  <Copy class="mr-1 h-3 w-3" /> 复制
                </Button>
              </div>
              <pre class="bg-[#0a0a0f] text-slate-200 p-3 rounded-lg text-xs overflow-x-auto leading-relaxed"><code>{{ ep.body }}</code></pre>
            </div>

            <!-- Response -->
            <div v-if="ep.response">
              <div class="flex items-center justify-between mb-2">
                <h4 class="text-xs font-semibold uppercase text-muted-foreground">响应示例</h4>
                <Button variant="ghost" size="sm" class="h-7 text-xs" @click="copyCode(ep.response)">
                  <Copy class="mr-1 h-3 w-3" /> 复制
                </Button>
              </div>
              <pre class="bg-[#0a0a0f] text-slate-200 p-3 rounded-lg text-xs overflow-x-auto leading-relaxed"><code>{{ ep.response }}</code></pre>
            </div>

            <!-- Curl -->
            <div class="rounded-lg border border-white/5 bg-white/5 p-3">
              <div class="flex items-center justify-between mb-1.5">
                <div class="flex items-center gap-1.5 text-xs font-semibold text-muted-foreground">
                  <Terminal class="h-3.5 w-3.5" />
                  cURL 示例
                </div>
                <Button variant="ghost" size="sm" class="h-7 text-xs" @click="copyCurl(ep)">
                  <component :is="copied === (ep.path + ep.method) ? Check : Copy" class="mr-1 h-3 w-3" />
                  {{ copied === (ep.path + ep.method) ? '已复制' : '复制' }}
                </Button>
              </div>
              <pre class="text-xs text-muted-foreground overflow-x-auto leading-relaxed font-mono">{{ curlExample(ep) }}</pre>
            </div>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <!-- Common Response Format -->
    <Card class="mt-6 border-white/5">
      <CardHeader>
        <CardTitle class="text-base">通用响应格式</CardTitle>
        <CardDescription>所有接口均遵循以下响应结构</CardDescription>
      </CardHeader>
      <CardContent class="space-y-3">
        <div class="rounded-lg border border-white/5 bg-white/5 p-3">
          <p class="text-xs font-mono text-muted-foreground mb-1">成功响应（HTTP 200）</p>
          <pre class="bg-[#0a0a0f] text-slate-200 p-3 rounded-lg text-xs overflow-x-auto leading-relaxed"><code>{
  "status": true,
  "message": "success",
  "data": { ... }
}</code></pre>
        </div>
        <div class="rounded-lg border border-white/5 bg-white/5 p-3">
          <p class="text-xs font-mono text-muted-foreground mb-1">错误响应（HTTP 4xx/5xx）</p>
          <pre class="bg-[#0a0a0f] text-slate-200 p-3 rounded-lg text-xs overflow-x-auto leading-relaxed"><code>{
  "status": false,
  "message": "错误描述"
}</code></pre>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
