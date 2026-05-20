<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { Button } from '@/components/ui/button'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'vue-sonner'
import { ArrowLeft, Plus, Pencil, Save, ServerCog, Trash2 } from 'lucide-vue-next'
import type { Strategy } from '@/types'

// --------------- Storage type definitions ---------------
interface ConfigField {
  key: string
  label: string
  placeholder?: string
  type?: 'text' | 'password' | 'number'
  required?: boolean
  defaultVal?: string
}

const keyNames: Record<string, string> = {
  '1': '本地', '2': 'AWS S3', '3': '阿里云 OSS', '4': '腾讯云 COS',
  '5': '七牛云 Kodo', '6': '又拍云 USS', '7': 'SFTP', '8': 'FTP',
  '9': 'WebDav', '10': 'Minio',
}

// Config fields per storage type
const configSchemas: Record<string, ConfigField[]> = {
  '1': [ // 本地
    { key: 'root', label: '存储根目录', placeholder: 'uploads', defaultVal: 'uploads' },
  ],
  '2': [ // S3
    { key: 'access_key_id', label: 'AccessKeyId', placeholder: 'AKIAXXX...', required: true, type: 'password' },
    { key: 'secret_access_key', label: 'SecretAccessKey', placeholder: 'xxx...', required: true, type: 'password' },
    { key: 'region', label: '区域', placeholder: 'us-east-1', defaultVal: 'us-east-1' },
    { key: 'bucket', label: 'Bucket', placeholder: 'my-bucket', required: true },
    { key: 'endpoint', label: 'Endpoint', placeholder: 'https://s3.amazonaws.com（兼容 Minio 等需填写）' },
  ],
  '3': [ // OSS
    { key: 'access_key_id', label: 'AccessKeyId', placeholder: 'LTAI...', required: true, type: 'password' },
    { key: 'access_key_secret', label: 'AccessKeySecret', placeholder: 'xxx...', required: true, type: 'password' },
    { key: 'endpoint', label: 'Endpoint', placeholder: 'oss-cn-hangzhou.aliyuncs.com', required: true },
    { key: 'bucket', label: 'Bucket', placeholder: 'my-bucket', required: true },
  ],
  '4': [ // COS
    { key: 'secret_id', label: 'SecretId', placeholder: 'AKIDxxx...', required: true, type: 'password' },
    { key: 'secret_key', label: 'SecretKey', placeholder: 'xxx...', required: true, type: 'password' },
    { key: 'region', label: '地域', placeholder: 'ap-guangzhou', defaultVal: 'ap-guangzhou' },
    { key: 'bucket', label: '储存桶名称', placeholder: 'mybucket（不含 AppId 后缀）', required: true },
    { key: 'app_id', label: 'AppId', placeholder: '1250000000' },
  ],
  '5': [ // Kodo
    { key: 'access_key', label: 'AccessKey', placeholder: 'xxx...', required: true, type: 'password' },
    { key: 'secret_key', label: 'SecretKey', placeholder: 'xxx...', required: true, type: 'password' },
    { key: 'bucket', label: 'Bucket', placeholder: 'my-bucket', required: true },
    { key: 'domain', label: '访问域名', placeholder: 'xxx.bkt.clouddn.com' },
  ],
  '6': [ // USS
    { key: 'access_key', label: 'AccessKey', placeholder: 'xxx...', required: true, type: 'password' },
    { key: 'secret_key', label: 'SecretKey', placeholder: 'xxx...', required: true, type: 'password' },
    { key: 'bucket', label: 'Bucket', placeholder: 'my-bucket', required: true },
    { key: 'domain', label: '访问域名', placeholder: 'xxx.upyun.com' },
  ],
  '7': [ // SFTP
    { key: 'host', label: '主机', placeholder: '192.168.1.1', required: true },
    { key: 'port', label: '端口', placeholder: '22', type: 'number', defaultVal: '22' },
    { key: 'username', label: '用户名', placeholder: 'root', required: true },
    { key: 'password', label: '密码', placeholder: 'xxx...', type: 'password', required: true },
    { key: 'root', label: '根目录', placeholder: '/var/www/uploads', defaultVal: '/var/www/uploads' },
  ],
  '8': [ // FTP
    { key: 'host', label: '主机', placeholder: '192.168.1.1', required: true },
    { key: 'port', label: '端口', placeholder: '21', type: 'number', defaultVal: '21' },
    { key: 'username', label: '用户名', placeholder: 'ftpuser', required: true },
    { key: 'password', label: '密码', placeholder: 'xxx...', type: 'password', required: true },
    { key: 'ssl', label: 'SSL', placeholder: 'false', defaultVal: 'false' },
    { key: 'passive', label: '被动模式', placeholder: 'true', defaultVal: 'true' },
    { key: 'root', label: '根目录', placeholder: '/uploads', defaultVal: '/uploads' },
  ],
  '9': [ // WebDav
    { key: 'base_uri', label: '基础地址', placeholder: 'https://dav.example.com', required: true },
    { key: 'username', label: '用户名', placeholder: 'admin', required: true },
    { key: 'password', label: '密码', placeholder: 'xxx...', type: 'password', required: true },
  ],
  '10': [ // Minio
    { key: 'access_key_id', label: 'AccessKeyId', placeholder: 'minioadmin', required: true, type: 'password' },
    { key: 'secret_access_key', label: 'SecretAccessKey', placeholder: 'minioadmin', required: true, type: 'password' },
    { key: 'region', label: '区域', placeholder: 'us-east-1', defaultVal: 'us-east-1' },
    { key: 'bucket', label: 'Bucket', placeholder: 'my-bucket', required: true },
    { key: 'endpoint', label: 'Endpoint', placeholder: 'http://127.0.0.1:9000', required: true },
  ],
}

// Common fields appended to all types
const commonFields: ConfigField[] = [
  { key: 'url', label: '自定义访问域名', placeholder: '留空使用默认' },
  { key: 'queries', label: 'URL 查询参数', placeholder: '例如: token=xxx' },
]

// --------------- State ---------------
const strategies = ref<Strategy[]>([])
const loading = ref(false)
const showForm = ref(false)
const editingStrategy = ref<Strategy | null>(null)
const sName = ref('')
const sIntro = ref('')
const sKey = ref<string>('1')
const configValues = ref<Record<string, string>>({})
const submitting = ref(false)

// Confirm dialog state
const showDeleteConfirm = ref(false)
const deleteTargetId = ref<number | null>(null)

const currentFields = computed(() => {
  const fields = configSchemas[sKey.value] || []
  return [...fields, ...commonFields]
})

function resetConfigs(key: string, existingConfigs?: Record<string, any>) {
  sKey.value = key
  const vals: Record<string, string> = {}
  const fields = [...(configSchemas[key] || []), ...commonFields]
  for (const f of fields) {
    const existing = existingConfigs?.[f.key]
    vals[f.key] = existing !== undefined && existing !== null
      ? String(existing)
      : (f.defaultVal || '')
  }
  configValues.value = vals
}

// --------------- Actions ---------------
async function loadStrategies() {
  loading.value = true
  try {
    const res = await adminApi.listStrategies()
    strategies.value = res
  } catch (e: any) {
    toast.error('加载策略失败')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingStrategy.value = null
  sName.value = ''
  sIntro.value = ''
  resetConfigs('1')
  showForm.value = true
}

function openEdit(s: Strategy) {
  editingStrategy.value = s
  sName.value = s.name
  sIntro.value = s.intro || ''
  resetConfigs(String(s.key), s.configs)
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingStrategy.value = null
}

async function handleSubmit() {
  const name = sName.value.trim()
  if (!name) {
    toast.error('请输入策略名称')
    return
  }
  if (submitting.value) return
  submitting.value = true
  try {
    // Build configs object (skip empty optional fields)
    const configs: Record<string, any> = {}
    const requiredKeys = new Set(
      (configSchemas[sKey.value] || [])
        .filter(f => f.required)
        .map(f => f.key)
    )
    for (const f of currentFields.value) {
      const val = configValues.value[f.key]?.trim() || ''
      if (val || requiredKeys.has(f.key)) {
        // Convert number fields
        if (f.type === 'number' && val) {
          configs[f.key] = Number(val)
        } else {
          configs[f.key] = val
        }
      }
    }

    if (editingStrategy.value) {
      await adminApi.updateStrategy(editingStrategy.value.id, {
        name,
        intro: sIntro.value,
        configs,
      })
      toast.success('更新成功')
    } else {
      await adminApi.createStrategy({
        key: Number(sKey.value),
        name,
        intro: sIntro.value,
        configs,
      })
      toast.success('创建成功')
    }
    showForm.value = false
    await loadStrategies()
  } catch (e: any) {
    toast.error(e.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

function confirmDeleteStrategy(id: number) {
  deleteTargetId.value = id
  showDeleteConfirm.value = true
}

async function deleteStrategy() {
  if (deleteTargetId.value === null) return
  try {
    await adminApi.deleteStrategy(deleteTargetId.value)
    toast.success('删除成功')
    loadStrategies()
  } catch (e: any) {
    toast.error(e.message || '删除失败')
  }
}

// Supported indicator - only Local, S3, COS are fully implemented
const implStatus: Record<string, string> = { '1': '', '2': '', '4': '' }

onMounted(loadStrategies)
</script>

<template>
  <div>
    <template v-if="showForm">
      <div class="mb-6 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <Button variant="ghost" class="-ml-2 mb-3" @click="closeForm">
            <ArrowLeft class="mr-2 h-4 w-4" />
            返回策略列表
          </Button>
          <p class="text-sm font-medium text-primary">Storage Strategy</p>
          <h1 class="mt-1 text-3xl font-semibold">
            {{ editingStrategy ? '编辑存储策略' : '新建存储策略' }}
          </h1>
          <p class="mt-2 max-w-2xl text-sm leading-6 text-muted-foreground">
            配置上传文件的落点、访问域名和连接参数。策略保存后可分配给用户组或作为用户默认策略使用。
          </p>
        </div>
        <div class="flex gap-2">
          <Button variant="outline" :disabled="submitting" @click="closeForm">取消</Button>
          <Button :disabled="submitting" @click="handleSubmit">
            <Save class="mr-2 h-4 w-4" />
            {{ submitting ? '保存中...' : '保存策略' }}
          </Button>
        </div>
      </div>

      <div class="grid gap-6 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div class="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle class="flex items-center gap-2 text-base">
                <ServerCog class="h-4 w-4 text-primary" />
                基本信息
              </CardTitle>
            </CardHeader>
            <CardContent class="grid gap-5 md:grid-cols-2">
              <div v-if="!editingStrategy" class="space-y-2">
                <Label>存储类型 <span class="text-red-500">*</span></Label>
                <Select v-model="sKey" @update:model-value="(v) => resetConfigs(String(v || '1'))">
                  <SelectTrigger class="h-10">
                    <SelectValue placeholder="选择存储类型" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="(name, key) in keyNames" :key="key" :value="String(key)">
                      {{ name }}
                      <span v-if="implStatus[key] !== undefined && !implStatus[key]" class="ml-2 text-xs text-muted-foreground">(完全支持)</span>
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div v-else class="space-y-2">
                <Label>存储类型</Label>
                <Input class="h-10" :model-value="keyNames[sKey] || `类型 ${sKey}`" disabled />
              </div>

              <div class="space-y-2">
                <Label>名称 <span class="text-red-500">*</span></Label>
                <Input v-model="sName" class="h-10" placeholder="例如：默认本地策略" />
              </div>

              <div class="space-y-2 md:col-span-2">
                <Label>简介</Label>
                <Textarea v-model="sIntro" placeholder="简要说明此策略的用途，例如：主站图片默认上传策略" />
              </div>
            </CardContent>
          </Card>

          <Card v-if="currentFields.length > 0">
            <CardHeader>
              <CardTitle class="text-base">连接与访问配置</CardTitle>
            </CardHeader>
            <CardContent class="grid gap-5 md:grid-cols-2">
              <div v-for="field in currentFields" :key="field.key" class="space-y-2">
                <Label>
                  {{ field.label }}
                  <span v-if="field.required" class="text-red-500">*</span>
                </Label>
                <Input
                  :model-value="configValues[field.key]"
                  class="h-10"
                  @update:model-value="(v) => configValues[field.key] = String(v || '')"
                  :type="field.type === 'password' ? 'password' : 'text'"
                  :placeholder="field.placeholder || ''"
                />
              </div>
            </CardContent>
          </Card>
        </div>

        <aside class="space-y-4">
          <Card>
            <CardContent class="p-5">
              <p class="text-sm font-semibold">当前类型</p>
              <p class="mt-2 text-2xl font-semibold">{{ keyNames[sKey] || `类型 ${sKey}` }}</p>
              <p class="mt-3 text-sm leading-6 text-muted-foreground">
                必填项为空时仍会提交为空字符串，后端需要继续做存储驱动级校验。
              </p>
            </CardContent>
          </Card>
          <Card class="bg-purple-500/10">
            <CardContent class="p-5 text-sm leading-6 text-muted-foreground">
              自定义访问域名用于生成图片外链；URL 查询参数适合需要签名或防盗链参数的存储服务。
            </CardContent>
          </Card>
        </aside>
      </div>
    </template>

    <template v-else>
      <div class="mb-6 flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
        <div>
          <p class="text-sm font-medium text-primary">Storage</p>
          <h1 class="mt-1 text-3xl font-semibold">存储策略</h1>
          <p class="mt-2 text-sm text-muted-foreground">管理本地、对象存储和远程文件系统的上传策略。</p>
        </div>
        <Button @click="openCreate">
          <Plus class="mr-2 h-4 w-4" /> 新建策略
        </Button>
      </div>

      <Table v-if="loading">
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">序号</TableHead>
            <TableHead>名称</TableHead>
            <TableHead>类型</TableHead>
            <TableHead>简介</TableHead>
            <TableHead class="text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="i in 6" :key="'skel-' + i">
            <TableCell colspan="5">
              <div class="h-10 bg-muted animate-pulse rounded" />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <Card v-else-if="strategies.length === 0">
        <CardContent class="flex flex-col items-center justify-center py-20 text-center">
          <ServerCog class="mb-4 h-12 w-12 text-muted-foreground/60" />
          <p class="text-lg font-semibold">暂无存储策略</p>
          <p class="mt-1 text-sm text-muted-foreground">创建一个策略后即可开始配置上传落点</p>
          <Button class="mt-5" @click="openCreate">
            <Plus class="mr-2 h-4 w-4" /> 新建策略
          </Button>
        </CardContent>
      </Card>

      <Table v-else>
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">序号</TableHead>
            <TableHead>名称</TableHead>
            <TableHead>类型</TableHead>
            <TableHead>简介</TableHead>
            <TableHead class="text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="(s, index) in strategies" :key="s.id">
            <TableCell>{{ index + 1 }}</TableCell>
            <TableCell class="font-medium">{{ s.name }}</TableCell>
            <TableCell>
              <Badge variant="outline">{{ keyNames[String(s.key)] || `类型 ${s.key}` }}</Badge>
            </TableCell>
            <TableCell class="max-w-xs truncate">{{ s.intro || '-' }}</TableCell>
            <TableCell>
              <div class="flex gap-2">
                <Button variant="outline" size="sm" @click="openEdit(s)">
                  <Pencil class="mr-1 h-3 w-3" /> 编辑
                </Button>
                <Button variant="destructive" size="sm" @click="confirmDeleteStrategy(s.id)">
                  <Trash2 class="mr-1 h-3 w-3" /> 删除
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </template>

    <!-- Delete Confirm Dialog -->
    <ConfirmDialog
      v-model:open="showDeleteConfirm"
      title="删除存储策略"
      :description="`确定要删除此存储策略吗？此操作不可撤销。`"
      confirm-text="确认删除"
      cancel-text="取消"
      variant="destructive"
      @confirm="deleteStrategy"
    />
  </div>
</template>
