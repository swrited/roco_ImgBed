<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { toast } from 'vue-sonner'
import { Save, Mail, RefreshCw, ArrowUp } from 'lucide-vue-next'

const settings = ref<Record<string, any>>({})
const loading = ref(true)
const testingMail = ref(false)
const checkingUpgrade = ref(false)
const upgrading = ref(false)
const upgradeInfo = ref<string | null>(null)
const upgradeDialogOpen = ref(false)

async function loadSettings() {
  loading.value = true
  try {
    const res = await adminApi.getSettings()
    settings.value = { ...res }
  } catch { /* ignore */ }
  finally { loading.value = false }
}

async function saveSettings() {
  try {
    await adminApi.updateSettings(settings.value)
    toast.success('保存成功')
  } catch (e: any) {
    toast.error(e.message || '保存失败')
  }
}

async function testMail() {
  testingMail.value = true
  try {
    await adminApi.testMail(settings.value)
    toast.success('测试邮件已发送，请检查邮箱')
  } catch (e: any) {
    toast.error(e.message || '邮件发送失败')
  } finally {
    testingMail.value = false
  }
}

async function checkUpgrade() {
  checkingUpgrade.value = true
  upgradeInfo.value = null
  try {
    const res = await adminApi.checkUpgrade()
    if (res?.version) {
      upgradeInfo.value = `发现新版本: ${res.version}`
      toast.info(upgradeInfo.value!)
    } else {
      upgradeInfo.value = '当前已是最新版本'
      toast.success(upgradeInfo.value)
    }
  } catch (e: any) {
    toast.error(e.message || '检查更新失败')
  } finally {
    checkingUpgrade.value = false
  }
}

function doUpgrade() {
  upgradeDialogOpen.value = true
}

async function handleUpgradeConfirm() {
  upgrading.value = true
  try {
    await adminApi.upgrade()
    toast.success('升级完成')
    upgradeInfo.value = null
  } catch (e: any) {
    toast.error(e.message || '升级失败')
  } finally {
    upgrading.value = false
  }
}

onMounted(loadSettings)
</script>

<template>
  <div class="max-w-2xl">
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">系统设置</h1>
      <Button @click="saveSettings">
        <Save class="mr-2 h-4 w-4" /> 保存设置
      </Button>
    </div>

    <div class="space-y-6">
      <!-- 基本设置 -->
      <Card>
        <CardHeader>
          <CardTitle>基本设置</CardTitle>
          <CardDescription>配置站点名称、描述等基本信息</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>站点名称</Label>
            <Input v-model="settings.name" placeholder="洛克图床" />
          </div>
          <div class="space-y-2">
            <Label>站点描述</Label>
            <Input v-model="settings.description" placeholder="站点描述" />
          </div>
          <div class="space-y-2">
            <Label>站点关键词</Label>
            <Input v-model="settings.keywords" placeholder="关键词" />
          </div>
        </CardContent>
      </Card>

      <!-- 用户设置 -->
      <Card>
        <CardHeader>
          <CardTitle>用户设置</CardTitle>
          <CardDescription>配置注册和用户相关选项</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="flex items-center justify-between rounded-xl border border-white/5 px-4 py-3">
            <div class="space-y-0.5">
              <Label class="text-sm font-medium">开放注册</Label>
              <p class="text-xs text-muted-foreground">允许新用户在首页注册账户</p>
            </div>
            <Switch v-model:checked="settings.is_enable_registration" />
          </div>
          <div class="flex items-center justify-between rounded-xl border border-white/5 px-4 py-3">
            <div class="space-y-0.5">
              <Label class="text-sm font-medium">游客上传</Label>
              <p class="text-xs text-muted-foreground">未登录用户可以直接上传图片</p>
            </div>
            <Switch v-model:checked="settings.is_enable_guest_upload" />
          </div>
          <div class="space-y-2">
            <Label>默认用户容量 (KB，0 = 无限制)</Label>
            <Input v-model="settings.user_capacity" type="number" placeholder="0" />
          </div>
        </CardContent>
      </Card>

      <!-- 功能开关 -->
      <Card>
        <CardHeader>
          <CardTitle>功能开关</CardTitle>
          <CardDescription>控制系统核心功能的启用状态</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="flex items-center justify-between rounded-xl border border-white/5 px-4 py-3">
            <div class="space-y-0.5">
              <Label class="text-sm font-medium">画廊功能</Label>
              <p class="text-xs text-muted-foreground">允许游客和注册用户浏览公开图库</p>
            </div>
            <Switch v-model:checked="settings.is_enable_gallery" />
          </div>
          <div class="flex items-center justify-between rounded-xl border border-white/5 px-4 py-3">
            <div class="space-y-0.5">
              <Label class="text-sm font-medium">API 接口</Label>
              <p class="text-xs text-muted-foreground">开放外部接口供第三方工具和脚本调用</p>
            </div>
            <Switch v-model:checked="settings.is_enable_api" />
          </div>
        </CardContent>
      </Card>

      <!-- 邮件设置 -->
      <Card>
        <CardHeader>
          <CardTitle>邮件设置</CardTitle>
          <CardDescription>配置 SMTP 邮件服务</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="space-y-2">
            <Label>SMTP 主机</Label>
            <Input v-model="settings.mail_host" placeholder="smtp.example.com" />
          </div>
          <div class="space-y-2">
            <Label>SMTP 端口</Label>
            <Input v-model="settings.mail_port" placeholder="587" />
          </div>
          <div class="space-y-2">
            <Label>发件邮箱</Label>
            <Input v-model="settings.mail_username" placeholder="noreply@example.com" />
          </div>
          <div class="space-y-2">
            <Label>邮箱密码</Label>
            <Input v-model="settings.mail_password" type="password" placeholder="••••••••" />
          </div>
          <Button class="rounded-xl border border-white/5 hover:shadow-md transition-all" @click="testMail" :disabled="testingMail">
            <Mail class="mr-2 h-4 w-4" /> {{ testingMail ? '发送中...' : '测试邮件' }}
          </Button>
        </CardContent>
      </Card>

      <!-- 系统升级 -->
      <Card>
        <CardHeader>
          <CardTitle>系统升级</CardTitle>
          <CardDescription>检查更新并执行在线升级</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div v-if="upgradeInfo" class="rounded-lg bg-muted/50 p-3">
            <p class="text-sm font-medium">{{ upgradeInfo }}</p>
          </div>
          <div class="flex gap-2">
            <Button class="rounded-xl border border-white/5 hover:shadow-md transition-all" @click="checkUpgrade" :disabled="checkingUpgrade">
              <RefreshCw class="mr-2 h-4 w-4" /> {{ checkingUpgrade ? '检查中...' : '检查更新' }}
            </Button>
            <Button class="rounded-xl shadow-sm hover:shadow-md transition-all" @click="doUpgrade" :disabled="upgrading" variant="secondary">
              <ArrowUp class="mr-2 h-4 w-4" /> {{ upgrading ? '升级中...' : '执行升级' }}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>

    <ConfirmDialog
      :open="upgradeDialogOpen"
      title="执行在线升级"
      description="升级过程中服务可能会暂时不可用，确定要继续吗？"
      confirm-text="确认升级"
      cancel-text="取消"
      variant="destructive"
      @update:open="upgradeDialogOpen = $event"
      @confirm="handleUpgradeConfirm"
    />
  </div>
</template>
