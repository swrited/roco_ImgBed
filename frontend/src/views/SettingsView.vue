<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { usersApi } from '@/api/users'
import { strategiesApi } from '@/api/strategies'
import { albumsApi } from '@/api/albums'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import { toast } from 'vue-sonner'
import { UserCog, Key, HardDrive, FolderOpen, LogOut } from 'lucide-vue-next'
import type { Album } from '@/types'

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

const albums = ref<Album[]>([])
const currentAlbumId = ref<number | null>(null)
const selectedAlbumId = ref('__none__')
const savingAlbum = ref(false)

const selectedPermission = ref('0')
const currentPermission = ref<number>(0)
const savingPermission = ref(false)

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
    const albumID = configs?.default_album_id ? Number(configs.default_album_id) : null
    currentAlbumId.value = albumID
    selectedAlbumId.value = albumID ? String(albumID) : '__none__'
    const perm = Number(configs?.default_permission ?? 0)
    currentPermission.value = perm
    selectedPermission.value = String(perm)
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

async function loadAlbums() {
  try {
    albums.value = await albumsApi.list()
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

async function saveUploadPreferences() {
  savingAlbum.value = true
  savingPermission.value = true
  try {
    const albumId = selectedAlbumId.value === '__none__' ? null : Number(selectedAlbumId.value)
    const perm = Number(selectedPermission.value)
    await Promise.all([
      usersApi.setAlbum(albumId),
      usersApi.setPermission(perm),
    ])
    currentAlbumId.value = albumId
    currentPermission.value = perm
    toast.success('默认上传偏好已保存')
  } catch (e: any) {
    toast.error(e.message || '设置失败')
  } finally {
    savingAlbum.value = false
    savingPermission.value = false
  }
}

onMounted(() => {
  loadProfile()
  loadStrategies()
  loadAlbums()
})
</script>

<template>
  <div class="mx-auto max-w-5xl space-y-6">
    <div class="flex flex-col gap-2">
      <p class="text-sm font-medium text-primary">Account</p>
      <h1 class="text-3xl font-semibold">用户设置</h1>
      <p class="text-sm text-muted-foreground">管理账户资料、默认上传行为和安全设置。</p>
    </div>

    <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
      <div class="space-y-6">
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

        <!-- Default Upload -->
        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2 text-lg">
              <FolderOpen class="h-5 w-5" /> 默认上传
            </CardTitle>
          </CardHeader>
          <CardContent class="space-y-5">
            <div class="grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <Label>默认相册</Label>
                <Select v-model="selectedAlbumId">
                  <SelectTrigger>
                    <SelectValue placeholder="不指定相册" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__none__">不指定相册</SelectItem>
                    <SelectItem v-for="album in albums" :key="album.id" :value="String(album.id)">
                      {{ album.name }}
                    </SelectItem>
                  </SelectContent>
                </Select>
                <p class="text-xs text-muted-foreground">不指定时，新图片会进入未分类图片。</p>
              </div>

              <div class="space-y-2">
                <Label>默认权限</Label>
                <Select v-model="selectedPermission">
                  <SelectTrigger>
                    <SelectValue placeholder="选择默认权限" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="1">公开</SelectItem>
                    <SelectItem value="0">私密</SelectItem>
                  </SelectContent>
                </Select>
                <p class="text-xs text-muted-foreground">私密图片不会出现在公开画廊中。</p>
              </div>
            </div>

            <Button variant="outline" @click="saveUploadPreferences" :disabled="savingAlbum || savingPermission">
              {{ savingAlbum || savingPermission ? '保存中...' : '保存上传偏好' }}
            </Button>
          </CardContent>
        </Card>
      </div>

      <div class="space-y-6">
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

        <Card>
          <CardHeader>
            <CardTitle class="flex items-center gap-2 text-lg">
              <LogOut class="h-5 w-5" /> 退出登录
            </CardTitle>
          </CardHeader>
          <CardContent class="space-y-4">
            <p class="text-sm text-muted-foreground">退出当前账户并返回登录页。</p>
            <Button variant="destructive" @click="auth.logout()">
              <LogOut class="mr-2 h-4 w-4" /> 退出登录
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  </div>
</template>
