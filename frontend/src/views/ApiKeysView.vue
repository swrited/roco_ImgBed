<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiKeysApi, type ApiKeyItem } from '@/api/apikeys'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'vue-sonner'
import { Key, Copy, Plus, Trash2, Clock } from 'lucide-vue-next'

const keys = ref<ApiKeyItem[]>([])
const newKeyName = ref('')
const creating = ref(false)

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
  navigator.clipboard.writeText(key)
  toast.success('已复制到剪贴板')
}

function formatDate(date: string | null): string {
  if (!date) return '从未使用'
  return new Date(date).toLocaleString()
}

onMounted(loadKeys)
</script>

<template>
  <div class="max-w-2xl mx-auto space-y-6">
    <h1 class="text-2xl font-bold">API Key 管理</h1>

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
        <CardDescription>在 API 请求中使用以下任一方式传递 Key</CardDescription>
      </CardHeader>
      <CardContent class="space-y-3">
        <div class="space-y-2">
          <p class="text-sm font-medium">请求头方式：</p>
          <code class="block rounded-lg bg-[#0a0a0f] px-3 py-2 text-sm font-mono break-all">
            X-Api-Key: lsky-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
          </code>
        </div>
        <div class="space-y-2">
          <p class="text-sm font-medium">Query 参数方式：</p>
          <code class="block rounded-lg bg-[#0a0a0f] px-3 py-2 text-sm font-mono break-all">
            GET /api/v1/images?api_key=lsky-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
          </code>
        </div>
      </CardContent>
    </Card>
  </div>
</template>
