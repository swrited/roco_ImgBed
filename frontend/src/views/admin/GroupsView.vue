<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminApi } from '@/api/admin'
import { Button } from '@/components/ui/button'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import {
  Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle, DialogDescription,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { Checkbox } from '@/components/ui/checkbox'
import { toast } from 'vue-sonner'
import { Plus, Pencil, Trash2, RotateCcw, Settings, CheckCircle2, XCircle } from 'lucide-vue-next'
import type { Group, Strategy } from '@/types'

// --------------- State ---------------
const groups = ref<Group[]>([])
const allStrategies = ref<Strategy[]>([])
const loading = ref(false)

// Group create/edit dialog
const showDialog = ref(false)
const editingGroup = ref<Group | null>(null)
const groupName = ref('')
const groupIsDefault = ref(false)
const groupIsGuest = ref(false)

// Rules/strategy association dialog
const showRulesDialog = ref(false)
const rulesGroup = ref<Group | null>(null)
const selectedStrategyIds = ref<number[]>([])
const rulesCapacity = ref('')
const rulesImageScan = ref(false)
const rulesOriginalProtection = ref(false)
const rulesWatermark = ref(false)
const savingRules = ref(false)

// Confirm dialog state
const showDeleteConfirm = ref(false)
const deleteTargetId = ref<number | null>(null)

// --------------- Actions ---------------
async function loadGroups() {
  loading.value = true
  try {
    const res = await adminApi.listGroups()
    groups.value = res || []
  } catch {
    toast.error('加载角色组失败')
  } finally {
    loading.value = false
  }
}

async function loadStrategies() {
  try {
    const res = await adminApi.listStrategies()
    allStrategies.value = res || []
  } catch { /* ignore */ }
}

function openCreate() {
  editingGroup.value = null
  groupName.value = ''
  groupIsDefault.value = false
  groupIsGuest.value = false
  showDialog.value = true
}

function openEdit(group: Group) {
  editingGroup.value = group
  groupName.value = group.name
  groupIsDefault.value = group.is_default || false
  groupIsGuest.value = group.is_guest || false
  showDialog.value = true
}

function boolConfig(group: Group, key: string): boolean {
  const value = group.configs?.[key]
  return value === true || value === 1 || value === '1'
}

function statusText(enabled: boolean): string {
  return enabled ? '开启' : '关闭'
}

function countStrategies(group: Group): number {
  return group.strategies_count ?? group.strategies?.length ?? 0
}

async function handleSubmit() {
  const name = groupName.value.trim()
  if (!name) {
    toast.error('请输入角色组名称')
    return
  }
  try {
    if (editingGroup.value) {
      await adminApi.updateGroup(editingGroup.value.id, {
        name,
        is_default: groupIsDefault.value,
        is_guest: groupIsGuest.value,
      })
      toast.success('更新成功')
    } else {
      await adminApi.createGroup({
        name,
        is_default: groupIsDefault.value,
        is_guest: groupIsGuest.value,
      })
      toast.success('创建成功')
    }
    showDialog.value = false
    loadGroups()
  } catch (e: any) {
    toast.error(e.message || '操作失败')
  }
}

function confirmDeleteGroup(id: number) {
  const group = groups.value.find(g => g.id === id)
  if (group?.is_default || group?.is_guest) {
    toast.error('默认组和游客组无法删除')
    return
  }
  deleteTargetId.value = id
  showDeleteConfirm.value = true
}

async function deleteGroup() {
  if (deleteTargetId.value === null) return
  try {
    await adminApi.deleteGroup(deleteTargetId.value)
    toast.success('删除成功')
    loadGroups()
  } catch (e: any) {
    toast.error(e.message || '删除失败')
  }
}

async function clearCache() {
  try {
    await adminApi.clearGroupCache()
    toast.success('缓存已清除')
  } catch {
    toast.error('清除失败')
  }
}

// --------------- Rules dialog ---------------
function openRules(g: Group) {
  rulesGroup.value = g
  // Parse existing strategy IDs
  const existing = g.strategies || []
  selectedStrategyIds.value = existing.map((s: any) => s.id)
  rulesCapacity.value = g.configs?.capacity ? String(g.configs.capacity) : ''
  rulesImageScan.value = boolConfig(g, 'is_enable_scan')
  rulesOriginalProtection.value = boolConfig(g, 'is_enable_original_protection')
  rulesWatermark.value = boolConfig(g, 'is_enable_watermark')
  showRulesDialog.value = true
}

function toggleStrategy(id: number) {
  const idx = selectedStrategyIds.value.indexOf(id)
  if (idx >= 0) {
    selectedStrategyIds.value.splice(idx, 1)
  } else {
    selectedStrategyIds.value.push(id)
  }
}

async function saveRules() {
  if (!rulesGroup.value) return
  savingRules.value = true
  try {
    const configs: Record<string, any> = { ...(rulesGroup.value.configs || {}) }
    if (rulesCapacity.value) {
      configs.capacity = Number(rulesCapacity.value)
    } else {
      delete configs.capacity
    }
    configs.is_enable_scan = rulesImageScan.value
    configs.is_enable_original_protection = rulesOriginalProtection.value
    configs.is_enable_watermark = rulesWatermark.value
    await adminApi.updateGroup(rulesGroup.value.id, {
      configs,
      strategy_ids: selectedStrategyIds.value,
    })
    toast.success('规则已保存')
    showRulesDialog.value = false
    loadGroups()
  } catch (e: any) {
    toast.error(e.message || '保存失败')
  } finally {
    savingRules.value = false
  }
}

onMounted(() => {
  loadGroups()
  loadStrategies()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">角色组管理</h1>
      <div class="flex gap-2">
        <Button variant="outline" @click="clearCache">
          <RotateCcw class="mr-2 h-4 w-4" /> 清除缓存
        </Button>
        <Button @click="openCreate">
          <Plus class="mr-2 h-4 w-4" /> 新建角色组
        </Button>
      </div>
    </div>

    <!-- Loading skeleton -->
    <Table v-if="loading">
      <TableHeader>
        <TableRow>
          <TableHead>ID</TableHead>
          <TableHead>名称</TableHead>
          <TableHead>是否默认</TableHead>
          <TableHead>是否为游客组</TableHead>
          <TableHead>图片审核</TableHead>
          <TableHead>原图保护</TableHead>
          <TableHead>水印</TableHead>
          <TableHead>用户数量</TableHead>
          <TableHead>策略数量</TableHead>
          <TableHead>操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="i in 6" :key="'skel-' + i">
          <TableCell colspan="10">
            <div class="h-10 bg-muted animate-pulse rounded" />
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <!-- Empty state -->
    <div v-else-if="groups.length === 0" class="flex flex-col items-center justify-center py-20 text-center">
      <p class="text-lg font-semibold text-muted-foreground">暂无角色组</p>
      <p class="text-sm text-muted-foreground mt-1">还没有创建任何角色组</p>
    </div>

    <!-- Data table -->
    <Table v-else>
      <TableHeader>
        <TableRow>
          <TableHead>ID</TableHead>
          <TableHead>名称</TableHead>
          <TableHead>是否默认</TableHead>
          <TableHead>是否为游客组</TableHead>
          <TableHead>图片审核</TableHead>
          <TableHead>原图保护</TableHead>
          <TableHead>水印</TableHead>
          <TableHead>用户数量</TableHead>
          <TableHead>策略数量</TableHead>
          <TableHead>操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-for="g in groups" :key="g.id">
          <TableCell>{{ g.id }}</TableCell>
          <TableCell class="font-medium">{{ g.name }}</TableCell>
          <TableCell>
            <Badge :variant="g.is_default ? 'default' : 'secondary'" class="gap-1">
              <component :is="g.is_default ? CheckCircle2 : XCircle" class="h-3.5 w-3.5" />
              {{ g.is_default ? '是' : '否' }}
            </Badge>
          </TableCell>
          <TableCell>
            <Badge :variant="g.is_guest ? 'default' : 'secondary'" class="gap-1">
              <component :is="g.is_guest ? CheckCircle2 : XCircle" class="h-3.5 w-3.5" />
              {{ g.is_guest ? '是' : '否' }}
            </Badge>
          </TableCell>
          <TableCell>
            <Badge :variant="boolConfig(g, 'is_enable_scan') ? 'default' : 'secondary'" class="gap-1">
              <component :is="boolConfig(g, 'is_enable_scan') ? CheckCircle2 : XCircle" class="h-3.5 w-3.5" />
              {{ statusText(boolConfig(g, 'is_enable_scan')) }}
            </Badge>
          </TableCell>
          <TableCell>
            <Badge :variant="boolConfig(g, 'is_enable_original_protection') ? 'default' : 'secondary'" class="gap-1">
              <component :is="boolConfig(g, 'is_enable_original_protection') ? CheckCircle2 : XCircle" class="h-3.5 w-3.5" />
              {{ statusText(boolConfig(g, 'is_enable_original_protection')) }}
            </Badge>
          </TableCell>
          <TableCell>
            <Badge :variant="boolConfig(g, 'is_enable_watermark') ? 'default' : 'secondary'" class="gap-1">
              <component :is="boolConfig(g, 'is_enable_watermark') ? CheckCircle2 : XCircle" class="h-3.5 w-3.5" />
              {{ statusText(boolConfig(g, 'is_enable_watermark')) }}
            </Badge>
          </TableCell>
          <TableCell>
            {{ g.users_count ?? 0 }}
          </TableCell>
          <TableCell>
            {{ countStrategies(g) }}
          </TableCell>
          <TableCell>
            <div class="flex gap-2">
              <Button variant="outline" size="sm" @click="openEdit(g)">
                <Pencil class="mr-1 h-3 w-3" /> 编辑
              </Button>
              <Button variant="secondary" size="sm" @click="openRules(g)">
                <Settings class="mr-1 h-3 w-3" /> 规则
              </Button>
              <Button v-if="!g.is_default && !g.is_guest" variant="destructive" size="sm" @click="confirmDeleteGroup(g.id)">
                <Trash2 class="mr-1 h-3 w-3" /> 删除
              </Button>
            </div>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <!-- Create/Edit Dialog -->
    <Dialog v-model:open="showDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ editingGroup ? '编辑角色组' : '新建角色组' }}</DialogTitle>
          <DialogDescription>
            {{ editingGroup ? '修改角色组名称、默认组和游客组状态' : '创建一个新的角色组' }}
          </DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <Label>名称</Label>
            <Input v-model="groupName" placeholder="输入角色组名称" />
          </div>
          <div class="flex items-center justify-between">
            <Label>设为默认</Label>
            <Switch v-model="groupIsDefault" />
          </div>
          <div class="flex items-center justify-between">
            <Label>设为游客组</Label>
            <Switch v-model="groupIsGuest" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showDialog = false">取消</Button>
          <Button @click="handleSubmit">确认</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Rules/Strategy Association Dialog -->
    <Dialog v-model:open="showRulesDialog">
      <DialogContent class="max-w-md max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>角色组规则 · {{ rulesGroup?.name }}</DialogTitle>
          <DialogDescription>配置该角色组的权限开关、可用策略和容量限制</DialogDescription>
        </DialogHeader>
        <div class="space-y-6">
          <!-- Capacity -->
          <div class="space-y-2">
            <Label>存储容量限制 (KB，0 = 不限制)</Label>
            <Input v-model="rulesCapacity" type="number" placeholder="0" />
          </div>

          <div class="space-y-3">
            <Label>权限开关</Label>
            <div class="space-y-3 rounded-lg border p-3">
              <div class="flex items-center justify-between gap-4">
                <div>
                  <p class="text-sm font-medium">图片审核</p>
                  <p class="text-xs text-muted-foreground">开启后上传图片需要经过审核流程</p>
                </div>
                <Switch v-model="rulesImageScan" />
              </div>
              <div class="flex items-center justify-between gap-4">
                <div>
                  <p class="text-sm font-medium">原图保护</p>
                  <p class="text-xs text-muted-foreground">开启后图片不直接返回原图直链</p>
                </div>
                <Switch v-model="rulesOriginalProtection" />
              </div>
              <div class="flex items-center justify-between gap-4">
                <div>
                  <p class="text-sm font-medium">水印</p>
                  <p class="text-xs text-muted-foreground">开启后该组用户上传图片应用水印配置</p>
                </div>
                <Switch v-model="rulesWatermark" />
              </div>
            </div>
          </div>

          <!-- Strategy association -->
          <div class="space-y-2">
            <Label>可用存储策略</Label>
            <p class="text-xs text-muted-foreground">未选中任何策略表示不限制</p>
            <div class="space-y-2 mt-2">
              <div
                v-for="s in allStrategies"
                :key="s.id"
                class="flex items-center gap-3 rounded-lg border p-3 cursor-pointer hover:bg-muted/50 transition-colors"
                @click="toggleStrategy(s.id)"
              >
                <Checkbox :checked="selectedStrategyIds.includes(s.id)" />
                <div class="flex-1 min-w-0">
                  <p class="text-sm font-medium">{{ s.name }}</p>
                  <p class="text-xs text-muted-foreground truncate">{{ s.intro || '-' }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showRulesDialog = false">取消</Button>
          <Button @click="saveRules" :disabled="savingRules">
            {{ savingRules ? '保存中...' : '保存规则' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Delete Confirm Dialog -->
    <ConfirmDialog
      v-model:open="showDeleteConfirm"
      title="删除角色组"
      :description="`确定要删除此角色组吗？此操作不可撤销。`"
      confirm-text="确认删除"
      cancel-text="取消"
      variant="destructive"
      @confirm="deleteGroup"
    />
  </div>
</template>
