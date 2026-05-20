<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { usersApi } from '@/api/users'
import { strategiesApi } from '@/api/strategies'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import { toast } from 'vue-sonner'
import { UserCog, Key, HardDrive } from 'lucide-vue-next'

const auth = useAuthStore()

const name = ref('')
const email = ref('')
const url = ref('')
const capacity = ref(0)
const isAdmin = ref(false)
const savingProfile = ref(false)

const oldPassword = ref('')
const newPassword = ref('')
const passwordConfirmation = ref('')
const savingPassword = ref(false)

const strategies = ref<{ id: number; name: string; key: number }[]>([])
const currentStrategyId = ref<number | null>(null)
const selectedStrategyId = ref('')
const savingStrategy = ref(false)

function formatSize(kb: number): string {
  if (kb <= 0) return '无限制'
  if (kb < 1024) return `${kb} KB`
  if (kb < 1024 * 1024) return `${(kb / 1024).toFixed(1)} MB`
  return `${(kb / 1024 / 1024).toFixed(2)} GB`
}

async function loadProfile() {
  try {
    const data = await usersApi.profile()
    name.value = data.name || ''
    email.value = data.email || ''
    url.value = data.url || ''
    capacity.value = data.capacity || 0
    isAdmin.value = data.is_adminer || false
    const configs = (data as any).configs
    currentStrategyId.value = configs?.default_strategy ?? null
    selectedStrategyId.value = currentStrategyId.value ? String(currentStrategyId.value) : ''
  } catch {
    toast.error('加载用户信息失败')
  }
}

async function loadStrategies() {
  try {
    const res = await strategiesApi.list()
    strategies.value = res
  } catch { /**/ }
}

async function saveProfile() {
  if (!name.value.trim()) {
    toast.error('用户名不能为空')
    return
  }
  savingProfile.value = true
  try {
    await usersApi.updateProfile({ name: name.value.trim(), url: url.value.trim() })
    auth.fetchUser()
    toast.success('个人信息已更新')
  } catch (e: any) {
    toast.error(e.message || '更新失败')
  } finally {
    savingProfile.value = false
  }
}

async function changePassword() {
  if (!oldPassword.value) {
    toast.error('请输入当前密码')
    return
  }
  if (!newPassword.value) {
    toast.error('请输入新密码')
    return
  }
  if (newPassword.value.length < 6) {
    toast.error('密码长度至少6位')
    return
  }
  if (newPassword.value !== passwordConfirmation.value) {
    toast.error('两次输入的密码不一致')
    return
  }
  savingPassword.value = true
  try {
    await usersApi.updateProfile({ old_password: oldPassword.value, password: newPassword.value })
    oldPassword.value = ''
    newPassword.value = ''
    passwordConfirmation.value = ''
    toast.success('密码已更新')
  } catch (e: any) {
    toast.error(e.message || '密码更新失败')
  } finally {
    savingPassword.value = false
  }
}

async function saveStrategy() {
  if (!selectedStrategyId.value) {
    toast.error('请选择存储策略')
    return
  }
  savingStrategy.value = true
  try {
    await usersApi.setStrategy(Number(selectedStrategyId.value))
    currentStrategyId.value = Number(selectedStrategyId.value)
    toast.success('默认存储策略已更新')
  } catch (e: any) {
    toast.error(e.message || '设置失败')
  } finally {
    savingStrategy.value = false
  }
}

onMounted(() => {
  loadProfile()
  loadStrategies()
})
</script>

<template>
  <div class="max-w-2xl mx-auto space-y-6">
    <h1 class="text-2xl font-bold">用户设置</h1>

    <!-- Profile -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2 text-lg">
          <UserCog class="h-5 w-5" /> 个人信息
        </CardTitle>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="space-y-2">
          <Label for="name">用户名</Label>
          <Input id="name" v-model="name" placeholder="请输入用户名" />
        </div>
        <div class="space-y-2">
          <Label for="email">邮箱</Label>
          <Input id="email" :model-value="email" disabled />
          <p class="text-xs text-muted-foreground">邮箱暂不支持修改</p>
        </div>
        <div class="space-y-2">
          <Label for="url">个人主页</Label>
          <Input id="url" v-model="url" placeholder="https://example.com" />
        </div>
        <div class="flex items-center justify-between rounded-lg bg-white/[0.02] p-3">
          <span class="text-sm text-muted-foreground">存储容量</span>
          <span class="text-sm font-medium">{{ formatSize(capacity) }}</span>
        </div>
        <div class="flex items-center justify-between rounded-lg bg-white/[0.02] p-3">
          <span class="text-sm text-muted-foreground">账户角色</span>
          <span class="text-sm font-medium">{{ isAdmin ? '管理员' : '普通用户' }}</span>
        </div>
        <Button @click="saveProfile" :disabled="savingProfile">
          {{ savingProfile ? '保存中...' : '保存修改' }}
        </Button>
      </CardContent>
    </Card>

    <!-- Password -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2 text-lg">
          <Key class="h-5 w-5" /> 修改密码
        </CardTitle>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="space-y-2">
          <Label for="old-password">当前密码</Label>
          <Input id="old-password" v-model="oldPassword" type="password" placeholder="请输入当前密码" />
        </div>
        <div class="space-y-2">
          <Label for="new-password">新密码</Label>
          <Input id="new-password" v-model="newPassword" type="password" placeholder="至少6位字符" />
        </div>
        <div class="space-y-2">
          <Label for="password-confirmation">确认密码</Label>
          <Input id="password-confirmation" v-model="passwordConfirmation" type="password" placeholder="再次输入新密码" />
        </div>
        <Button variant="outline" @click="changePassword" :disabled="savingPassword">
          {{ savingPassword ? '更新中...' : '更新密码' }}
        </Button>
      </CardContent>
    </Card>

    <!-- Default Strategy -->
    <Card>
      <CardHeader>
        <CardTitle class="flex items-center gap-2 text-lg">
          <HardDrive class="h-5 w-5" /> 默认存储策略
        </CardTitle>
      </CardHeader>
      <CardContent class="space-y-4">
        <p class="text-sm text-muted-foreground">上传图片时优先使用的存储策略</p>
        <div class="space-y-2">
          <Label>存储策略</Label>
          <Select v-model="selectedStrategyId">
            <SelectTrigger>
              <SelectValue placeholder="选择存储策略" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem v-for="s in strategies" :key="s.id" :value="String(s.id)">
                {{ s.name }}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <Button variant="outline" @click="saveStrategy" :disabled="savingStrategy">
          {{ savingStrategy ? '保存中...' : '保存策略' }}
        </Button>
      </CardContent>
    </Card>
  </div>
</template>
