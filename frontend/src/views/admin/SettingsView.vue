<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { toast } from 'vue-sonner'
import { Save, Mail, RefreshCw, ArrowUp, HardDrive, WandSparkles } from 'lucide-vue-next'
import type { Strategy } from '@/types'

const settings = ref<Record<string, any>>({})
const strategies = ref<Strategy[]>([])
const loading = ref(true)
const testingMail = ref(false)
const checkingUpgrade = ref(false)
const upgrading = ref(false)
const upgradeInfo = ref<string | null>(null)
const upgradeDialogOpen = ref(false)
const saveDialogOpen = ref(false)

const BOOL_KEYS = ['is_enable_registration', 'is_enable_guest_upload', 'is_enable_gallery', 'is_enable_api', 'is_enable_ai_image']

function strToBool(v: any): boolean {
  if (typeof v === 'boolean') return v
  return v === '1' || v === 1 || v === 'true'
}

function boolToStr(v: any): string {
  return v ? '1' : '0'
}

function normalizeSettings(raw: Record<string, any>): Record<string, any> {
  const s = { ...raw }
  for (const k of BOOL_KEYS) {
    if (k in s) s[k] = strToBool(s[k])
  }
  s.user_initial_capacity = s.user_initial_capacity ?? s.user_capacity ?? '512000'
  delete s.user_capacity
  s.default_strategy_id = s.default_strategy_id ? String(s.default_strategy_id) : '__auto__'
  s.is_enable_ai_image = strToBool(s.is_enable_ai_image ?? false)
  s.minimax_api_endpoint = s.minimax_api_endpoint || 'https://api.minimaxi.com/v1/image_generation'
  s.minimax_model = s.minimax_model || 'image-01'
  s.ai_image_max_count = s.ai_image_max_count || '4'
  s.ai_image_rate_limit_seconds = s.ai_image_rate_limit_seconds || '30'
  s.ai_image_daily_limit = s.ai_image_daily_limit || '10'
  s.api_key_minute_limit = s.api_key_minute_limit || '60'
  s.api_key_daily_limit = s.api_key_daily_limit || '1000'
  s.upload_max_size = s.upload_max_size || '10240'
  return s
}

function denormalizeSettings(s: Record<string, any>): Record<string, any> {
  const data = { ...s }
  for (const k of BOOL_KEYS) {
    if (k in data && typeof data[k] === 'boolean') data[k] = boolToStr(data[k])
  }
  if (data.default_strategy_id === '__auto__') data.default_strategy_id = ''
  return data
}

async function loadSettings() {
  loading.value = true
  try {
    const [settingsRes, strategiesRes] = await Promise.all([
      adminApi.getSettings(),
      adminApi.listStrategies(),
    ])
    settings.value = normalizeSettings(settingsRes)
    strategies.value = strategiesRes
  } catch { /* ignore */ }
  finally { loading.value = false }
}

function confirmSave() {
  saveDialogOpen.value = true
}

async function handleSaveConfirm() {
  try {
    await adminApi.updateSettings(denormalizeSettings(settings.value))
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
          <Button variant="outline" class="mt-2" @click="confirmSave">
            <Save class="mr-2 h-4 w-4" /> 保存设置
          </Button>
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
            <Switch v-model="settings.is_enable_registration" />
          </div>
          <div class="flex items-center justify-between rounded-xl border border-white/5 px-4 py-3">
            <div class="space-y-0.5">
              <Label class="text-sm font-medium">游客上传</Label>
              <p class="text-xs text-muted-foreground">未登录用户可以直接上传图片</p>
            </div>
            <Switch v-model="settings.is_enable_guest_upload" />
          </div>
          <div class="space-y-2">
            <Label>默认用户容量 (KB，0 = 无限制)</Label>
            <Input v-model="settings.user_initial_capacity" type="number" placeholder="512000" />
          </div>
          <div class="space-y-2">
            <Label>单张图片大小限制 (KB，0 = 不限制)</Label>
            <Input v-model="settings.upload_max_size" type="number" min="0" placeholder="10240" />
            <p class="text-xs text-muted-foreground">普通用户上传页会展示该限制，后端会强制校验。</p>
          </div>
          <div class="space-y-2">
            <Label class="flex items-center gap-2">
              <HardDrive class="h-4 w-4" />
              默认存储策略
            </Label>
            <Select v-model="settings.default_strategy_id">
              <SelectTrigger class="w-full">
                <SelectValue placeholder="未指定，按用户组策略自动选择" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__auto__">未指定，按用户组策略自动选择</SelectItem>
                <SelectItem v-for="strategy in strategies" :key="strategy.id" :value="String(strategy.id)">
                  {{ strategy.name }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p class="text-xs text-muted-foreground">设置后，未单独选择策略的用户会默认使用该存储策略上传。</p>
          </div>
          <Button variant="outline" class="mt-2" @click="confirmSave">
            <Save class="mr-2 h-4 w-4" /> 保存设置
          </Button>
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
            <Switch v-model="settings.is_enable_gallery" />
          </div>
          <div class="flex items-center justify-between rounded-xl border border-white/5 px-4 py-3">
            <div class="space-y-0.5">
              <Label class="text-sm font-medium">API 接口</Label>
              <p class="text-xs text-muted-foreground">开放外部接口供第三方工具和脚本调用</p>
            </div>
            <Switch v-model="settings.is_enable_api" />
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label>API Key 每分钟限制 (0 = 不限制)</Label>
              <Input v-model="settings.api_key_minute_limit" type="number" min="0" placeholder="60" />
              <p class="text-xs text-muted-foreground">用于限制 API 测试台和外部脚本的高频请求。</p>
            </div>
            <div class="space-y-2">
              <Label>API Key 每日限制 (0 = 不限制)</Label>
              <Input v-model="settings.api_key_daily_limit" type="number" min="0" placeholder="1000" />
              <p class="text-xs text-muted-foreground">按 API Key 统计，超限后返回 429。</p>
            </div>
          </div>
          <Button variant="outline" class="mt-2" @click="confirmSave">
            <Save class="mr-2 h-4 w-4" /> 保存设置
          </Button>
        </CardContent>
      </Card>

      <!-- AI 生图 -->
      <Card>
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <WandSparkles class="h-5 w-5 text-purple-400" />
            AI 生图
          </CardTitle>
          <CardDescription>配置 MiniMax 图片生成服务，普通用户不会看到 API Key。</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4">
          <div class="flex items-center justify-between rounded-xl border border-white/5 px-4 py-3">
            <div class="space-y-0.5">
              <Label class="text-sm font-medium">启用 AI 生图</Label>
              <p class="text-xs text-muted-foreground">启用后，用户可以调用 MiniMax 生成并保存图片</p>
            </div>
            <Switch v-model="settings.is_enable_ai_image" />
          </div>
          <div class="space-y-2">
            <Label>MiniMax API Key</Label>
            <Input v-model="settings.minimax_api_key" type="password" placeholder="填入 MiniMax 开放平台 API Key" />
          </div>
          <div class="space-y-2">
            <Label>接口地址</Label>
            <Input v-model="settings.minimax_api_endpoint" placeholder="https://api.minimaxi.com/v1/image_generation" />
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-2">
              <Label>模型</Label>
              <Input v-model="settings.minimax_model" placeholder="image-01" />
            </div>
            <div class="space-y-2">
              <Label>单次最大生成数量</Label>
              <Input v-model="settings.ai_image_max_count" type="number" min="1" max="9" placeholder="4" />
            </div>
            <div class="space-y-2">
              <Label>生成冷却时间 (秒，0 = 不限制)</Label>
              <Input v-model="settings.ai_image_rate_limit_seconds" type="number" min="0" placeholder="30" />
            </div>
            <div class="space-y-2">
              <Label>每用户每日免费次数 (0 = 不限制)</Label>
              <Input v-model="settings.ai_image_daily_limit" type="number" min="0" placeholder="10" />
              <p class="text-xs text-muted-foreground">默认每个用户每天 10 次，按成功发起生成请求计数。</p>
            </div>
          </div>
          <Button variant="outline" class="mt-2" @click="confirmSave">
            <Save class="mr-2 h-4 w-4" /> 保存设置
          </Button>
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
          <Button variant="outline" class="mt-2 ml-2" @click="confirmSave">
            <Save class="mr-2 h-4 w-4" /> 保存设置
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

    <!-- 粘性底部保存栏 -->
    <div class="sticky bottom-0 z-10 -mx-4 px-4 py-3 bg-background/95 backdrop-blur border-t border-border mt-8">
      <div class="max-w-2xl mx-auto flex justify-end">
        <Button @click="confirmSave">
          <Save class="mr-2 h-4 w-4" /> 保存设置
        </Button>
      </div>
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

    <ConfirmDialog
      :open="saveDialogOpen"
      title="保存系统设置"
      description="确定要保存当前的系统设置吗？部分设置可能需要在保存后重启服务才能生效。"
      confirm-text="确认保存"
      cancel-text="取消"
      @update:open="saveDialogOpen = $event"
      @confirm="handleSaveConfirm"
    />
  </div>
</template>
