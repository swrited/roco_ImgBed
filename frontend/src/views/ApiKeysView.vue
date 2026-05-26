<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiKeysApi, type ApiKeyItem } from '@/api/apikeys'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'vue-sonner'
import { Key, Copy, Plus, Trash2, Clock, RefreshCw, Images, Terminal } from 'lucide-vue-next'
import { copyToClipboard } from '@/utils/clipboard'
import { usersApi } from '@/api/users'

const keys = ref<ApiKeyItem[]>([])
const newKeyName = ref('')
const creating = ref(false)
const imageReadToken = ref('')
const tokenLoading = ref(false)
const tokenResetting = ref(false)

async function loadKeys() {
  try {
    keys.value = await apiKeysApi.list()
  } catch {
    toast.error('加载 API Keys 失败')
  }
}

async function createKey() {
  if (!newKeyName.value.trim()) {
    toast.error('请输入 Key 名称')
    return
  }
  creating.value = true
  try {
    await apiKeysApi.create({ name: newKeyName.value.trim() })
    newKeyName.value = ''
    toast.success('API Key 已创建')
    await loadKeys()
  } catch (e: any) {
    toast.error(e.message || '创建失败')
  } finally {
    creating.value = false
  }
}

async function loadImageReadToken() {
  tokenLoading.value = true
  try {
    const res = await usersApi.imageReadToken()
    imageReadToken.value = res.token
  } catch (e: any) {
    toast.error(e.message || '加载图库只读令牌失败')
  } finally {
    tokenLoading.value = false
  }
}

async function resetImageReadToken() {
  if (!confirm('重置后，使用旧图库只读令牌的请求将立即失效。确定继续吗？')) return
  tokenResetting.value = true
  try {
    const res = await usersApi.resetImageReadToken()
    imageReadToken.value = res.token
    toast.success('图库只读令牌已重置')
  } catch (e: any) {
    toast.error(e.message || '重置图库只读令牌失败')
  } finally {
    tokenResetting.value = false
  }
}

async function revokeKey(id: number) {
  if (!confirm('确定要撤销这个 API Key 吗？此操作不可恢复。')) return
  try {
    await apiKeysApi.revoke(id)
    toast.success('API Key 已撤销')
    await loadKeys()
  } catch (e: any) {
    toast.error(e.message || '撤销失败')
  }
}

function copyKey(key: string) {
  copyToClipboard(key)
  toast.success('已复制到剪贴板')
}

function formatDate(date: string | null): string {
  if (!date) return '从未使用'
  return new Date(date).toLocaleString()
}

onMounted(() => {
  loadKeys()
  loadImageReadToken()
})
</script>

<template>
  <div class="max-w-2xl mx-auto space-y-6">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-2xl font-bold">API Key 管理</h1>
      <Button variant="outline" @click="$router.push('/api-test')">
        <Terminal class="mr-2 h-4 w-4" /> 打开 API 测试台
      </Button>
    </div>

    <!-- Create new key -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2 text-lg">
          <Plus class="h-5 w-5" /> 创建新 Key
        </CardTitle>
        <CardDescription>API Key 可用于通过 API 接口访问您的图片数据</CardDescription>
      </CardHeader>
      <CardContent>
        <div class="flex gap-3">
          <div class="flex-1 space-y-2">
            <Label for="key-name">Key 名称</Label>
            <Input id="key-name" v-model="newKeyName" placeholder="例如：PicGo 图床" @keyup.enter="createKey" />
          </div>
          <div class="flex items-end">
            <Button @click="createKey" :disabled="creating">
              <Plus class="mr-2 h-4 w-4" />
              {{ creating ? '创建中...' : '创建' }}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card class="border-emerald-500/20 bg-emerald-500/5">
      <CardHeader>
        <CardTitle class="flex items-center justify-between gap-3 text-lg">
          <span class="flex min-w-0 items-center gap-2 text-emerald-300">
            <Images class="h-5 w-5 shrink-0" /> 图库只读令牌
          </span>
          <Button variant="outline" size="sm" :disabled="tokenResetting || tokenLoading" @click="resetImageReadToken">
            <RefreshCw class="mr-1 h-3.5 w-3.5" :class="{ 'animate-spin': tokenResetting }" />
            重置
          </Button>
        </CardTitle>
        <CardDescription class="text-emerald-100/70">
          供桌面组件、壁纸脚本或只读展示程序获取你的图片数据。
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <p class="rounded-xl border border-emerald-500/15 bg-black/20 p-3 text-sm leading-6 text-muted-foreground">
          该令牌只能读取你的图片列表、随机图和设备适配图，不能上传、修改或删除内容。可以在 API 测试台选择“图库只读令牌”验证调用；分享单张私密图片时，请使用图片详情页生成的单图链接。
        </p>
        <div class="space-y-2">
          <div class="flex items-center justify-between gap-3">
            <Label class="text-xs text-emerald-200">令牌</Label>
            <Button variant="ghost" size="sm" :disabled="!imageReadToken" @click="copyKey(imageReadToken)">
              <Copy class="mr-1 h-3.5 w-3.5" /> 复制令牌
            </Button>
          </div>
          <code class="block rounded-lg border border-emerald-500/20 bg-[#0a0a0f] px-3 py-2 font-mono text-xs break-all text-emerald-300">
            {{ tokenLoading ? '加载中...' : imageReadToken || '暂无令牌' }}
          </code>
        </div>
        <div class="space-y-2">
          <Label class="text-xs text-emerald-200">只读调用示例</Label>
          <code class="block rounded-lg border border-white/10 bg-[#0a0a0f] px-3 py-2 font-mono text-xs break-all">
            GET /api/v1/t/{{ imageReadToken || '{READ_TOKEN}' }}/images/random
          </code>
          <p class="text-xs text-muted-foreground">
            也支持 <span class="font-mono">/images</span> 与 <span class="font-mono">/images/adaptive</span>。
          </p>
        </div>
      </CardContent>
    </Card>

    <!-- Existing keys -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2 text-lg">
          <Key class="h-5 w-5" /> 我的 Keys
        </CardTitle>
        <CardDescription>{{ keys.length }} / 10 个已创建</CardDescription>
      </CardHeader>
      <CardContent class="space-y-3">
        <div v-if="keys.length === 0" class="py-8 text-center text-muted-foreground text-sm">
          还没有任何 API Key，请创建一个
        </div>
        <div
          v-for="k in keys" :key="k.id"
          class="flex items-center gap-3 rounded-xl border border-white/5 bg-white/5 p-3"
        >
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <p class="text-sm font-medium truncate">{{ k.name }}</p>
              <span class="shrink-0 text-xs text-muted-foreground font-mono bg-[#0f0f15] px-1.5 py-0.5 rounded">
                lsky-...{{ k.key.slice(-8) }}
              </span>
            </div>
            <div class="flex items-center gap-1 mt-1 text-xs text-muted-foreground">
              <Clock class="h-3 w-3" />
              {{ formatDate(k.last_used) }}
            </div>
          </div>
          <div class="flex items-center gap-1 shrink-0">
            <Button variant="ghost" size="icon" class="h-8 w-8" title="复制" @click="copyKey(k.key)">
              <Copy class="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" size="icon" class="h-8 w-8 text-destructive" title="撤销" @click="revokeKey(k.id)">
              <Trash2 class="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <!-- Usage tips -->
    <Card>
      <CardHeader>
        <CardTitle class="text-lg">使用方式</CardTitle>
        <CardDescription>在业务 API 请求头中传递 Key</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <p class="text-xs text-muted-foreground bg-white/5 p-3 rounded-lg border border-white/10">
          你可以前往左侧菜单的 <strong class="text-primary">API 测试台</strong>，填写 API Key 后直接生成和发送请求。私密图片分享链接由图片管理页单独生成，不需要暴露 API Key。
        </p>

        <div class="space-y-2">
          <p class="text-sm font-medium">请求头方式：</p>
          <code class="block rounded-lg bg-[#0a0a0f] px-3 py-2 text-sm font-mono break-all">
            X-Api-Key: lsky-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
          </code>
        </div>
      </CardContent>
    </Card>

  </div>
</template>
