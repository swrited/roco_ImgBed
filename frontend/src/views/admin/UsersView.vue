<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { adminApi } from '@/api/admin'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { toast } from 'vue-sonner'
import ConfirmDialog from '@/components/ConfirmDialog.vue'
import { Plus, Pencil, Trash2, Search, Users, HardDrive } from 'lucide-vue-next'
import type { User } from '@/types'

// 用户列表数据
const users = ref<User[]>([])
const total = ref(0)
const currentPage = ref(1)
const lastPage = ref(1)
const loading = ref(false)
const searchQuery = ref('')

// 删除确认对话框
const deleteDialogOpen = ref(false)
const deleteTargetId = ref<number | null>(null)
const deleteTargetName = ref('')

// 编辑对话框
const showEditDialog = ref(false)
const editingUser = ref<User | null>(null)
const editSaving = ref(false)
const editName = ref('')
const editEmail = ref('')
const editCapacity = ref(0)
const editAdminer = ref(false)

// 导入 ConfirmDialog 组件
// 格式化容量，输入单位为 KB
function formatSize(kb: number): string {
  if (!kb || kb <= 0) return '无限制'
  if (kb < 1024) return kb + ' KB'
  if (kb < 1024 * 1024) return (kb / 1024).toFixed(1) + ' MB'
  return (kb / (1024 * 1024)).toFixed(2) + ' GB'
}

function formatUsage(u: User): string {
  const used = formatSize(u.used_capacity || 0)
  if (!u.capacity || u.capacity <= 0) return `${used} / 无限制`
  return `${used} / ${formatSize(u.capacity)}`
}

function remainingText(u: User): string {
  if (!u.capacity || u.capacity <= 0) return '剩余无限制'
  return `剩余 ${formatSize(u.remaining_capacity || 0)}`
}

// 加载用户列表
async function loadUsers(page = 1) {
  loading.value = true
  try {
    const params: Record<string, any> = { page }
    if (searchQuery.value.trim()) {
      params.q = searchQuery.value.trim()
    }
    const res = await adminApi.listUsers(params)
    users.value = res.data
    total.value = res.total
    currentPage.value = res.current_page
    lastPage.value = res.last_page
  } catch (e: any) {
    toast.error('加载用户失败')
  } finally {
    loading.value = false
  }
}

// 搜索
function handleSearch() {
  loadUsers(1)
}

// 编辑用户
function editUser(u: User) {
  editingUser.value = u
  editName.value = u.name
  editEmail.value = u.email
  editCapacity.value = u.capacity
  editAdminer.value = u.is_adminer
  showEditDialog.value = true
}

// 保存用户
async function saveUser() {
  if (!editingUser.value) return

  // 表单校验
  if (!editName.value.trim()) {
    toast.error('请输入用户名')
    return
  }
  if (!editEmail.value.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(editEmail.value)) {
    toast.error('请输入有效的邮箱地址')
    return
  }

  editSaving.value = true
  try {
    await adminApi.updateUser(editingUser.value.id, {
      name: editName.value.trim(),
      email: editEmail.value.trim(),
      capacity: editCapacity.value,
      is_adminer: editAdminer.value,
    })
    toast.success('更新成功')
    showEditDialog.value = false
    loadUsers(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '更新失败')
  } finally {
    editSaving.value = false
  }
}

// 打开删除确认
function confirmDeleteUser(id: number, name: string) {
  deleteTargetId.value = id
  deleteTargetName.value = name
  deleteDialogOpen.value = true
}

// 执行删除
async function handleDeleteConfirm() {
  if (deleteTargetId.value === null) return
  try {
    await adminApi.deleteUser(deleteTargetId.value)
    toast.success('删除成功')
    loadUsers(currentPage.value)
  } catch (e: any) {
    toast.error(e.message || '删除失败')
  } finally {
    deleteDialogOpen.value = false
    deleteTargetId.value = null
    deleteTargetName.value = ''
  }
}

// 分页页码计算
const totalPages = computed(() => lastPage.value)

const visiblePages = computed(() => {
  const pages: (number | 'ellipsis')[] = []
  const current = currentPage.value
  const last = lastPage.value

  if (last <= 7) {
    for (let i = 1; i <= last; i++) pages.push(i)
    return pages
  }

  pages.push(1)
  if (current > 3) pages.push('ellipsis')

  const start = Math.max(2, current - 1)
  const end = Math.min(last - 1, current + 1)

  for (let i = start; i <= end; i++) pages.push(i)

  if (current < last - 2) pages.push('ellipsis')
  pages.push(last)

  return pages
})

function goToPage(page: number) {
  if (page >= 1 && page <= lastPage.value) {
    loadUsers(page)
  }
}

onMounted(() => loadUsers())
</script>

<template>
  <div>
    <!-- 页面标题 -->
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold">用户管理</h1>
      <Button>
        <Plus class="h-4 w-4 mr-2" />
        新建用户
      </Button>
    </div>

    <!-- 搜索栏 -->
    <div class="relative mb-4">
      <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
      <Input
        v-model="searchQuery"
        placeholder="搜索用户名或邮箱..."
        class="pl-9"
        @keyup.enter="handleSearch"
      />
    </div>

    <!-- 表格 -->
    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead class="w-[60px]">ID</TableHead>
            <TableHead>用户名</TableHead>
            <TableHead>邮箱</TableHead>
            <TableHead>存储空间</TableHead>
            <TableHead>图片数</TableHead>
            <TableHead>角色</TableHead>
            <TableHead>注册时间</TableHead>
            <TableHead class="w-[140px]">操作</TableHead>
          </TableRow>
        </TableHeader>

        <!-- 加载骨架屏 -->
        <TableBody v-if="loading">
          <TableRow v-for="i in 6" :key="'skeleton-' + i">
            <TableCell colspan="8">
              <div class="h-10 bg-muted animate-pulse rounded" />
            </TableCell>
          </TableRow>
        </TableBody>

        <!-- 空状态 -->
        <TableBody v-else-if="users.length === 0">
          <TableRow>
            <TableCell colspan="8">
              <div class="text-center py-16 text-muted-foreground">
                <Users class="h-12 w-12 mx-auto mb-3 opacity-40" />
                <p class="text-base font-medium">暂无用户</p>
                <p class="text-sm mt-1">系统中还没有任何用户</p>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>

        <!-- 数据行 -->
        <TableBody v-else>
          <TableRow v-for="u in users" :key="u.id">
            <TableCell>{{ u.id }}</TableCell>
            <TableCell class="font-medium">{{ u.name }}</TableCell>
            <TableCell>{{ u.email }}</TableCell>
            <TableCell class="min-w-[190px]">
              <div class="space-y-1.5">
                <div class="flex items-center justify-between gap-3">
                  <span class="font-medium">{{ formatUsage(u) }}</span>
                  <Badge v-if="u.capacity && u.capacity > 0" variant="secondary">
                    {{ u.capacity_percent || 0 }}%
                  </Badge>
                  <Badge v-else variant="outline">无限制</Badge>
                </div>
                <div v-if="u.capacity && u.capacity > 0" class="h-1.5 rounded-full bg-muted overflow-hidden">
                  <div
                    class="h-full rounded-full bg-primary transition-all"
                    :style="{ width: `${Math.min(100, u.capacity_percent || 0)}%` }"
                  />
                </div>
                <div class="flex items-center gap-1 text-xs text-muted-foreground">
                  <HardDrive class="h-3.5 w-3.5" />
                  {{ remainingText(u) }}
                </div>
              </div>
            </TableCell>
            <TableCell>{{ u.image_num }}</TableCell>
            <TableCell>
              <Badge :variant="u.is_adminer ? 'default' : 'secondary'">
                {{ u.is_adminer ? '管理员' : '用户' }}
              </Badge>
            </TableCell>
            <TableCell>{{ u.created_at }}</TableCell>
            <TableCell>
              <div class="flex gap-1">
                <Button variant="ghost" size="icon" title="编辑" @click="editUser(u)">
                  <Pencil class="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  title="删除"
                  @click="confirmDeleteUser(u.id, u.name)"
                >
                  <Trash2 class="h-4 w-4 text-destructive" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>

    <!-- 分页 -->
    <Pagination v-if="lastPage > 1 && !loading" class="mt-4">
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious
            :disabled="currentPage <= 1"
            @click="goToPage(currentPage - 1)"
          />
        </PaginationItem>

        <template v-for="(item, idx) in visiblePages" :key="idx">
          <PaginationItem v-if="item === 'ellipsis'">
            <PaginationEllipsis />
          </PaginationItem>
          <PaginationItem v-else>
            <PaginationLink
              :is-active="item === currentPage"
              @click="goToPage(item)"
            >
              {{ item }}
            </PaginationLink>
          </PaginationItem>
        </template>

        <PaginationItem>
          <PaginationNext
            :disabled="currentPage >= lastPage"
            @click="goToPage(currentPage + 1)"
          />
        </PaginationItem>
      </PaginationContent>
    </Pagination>

    <!-- 编辑对话框 -->
    <Dialog v-model:open="showEditDialog">
      <DialogContent>
        <DialogHeader>
          <DialogTitle>编辑用户</DialogTitle>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-2">
            <Label for="edit-name">用户名</Label>
            <Input id="edit-name" v-model="editName" placeholder="请输入用户名" />
          </div>
          <div class="space-y-2">
            <Label for="edit-email">邮箱</Label>
            <Input id="edit-email" v-model="editEmail" placeholder="请输入邮箱" />
          </div>
          <div class="space-y-2">
            <Label for="edit-capacity">容量 (KB，0 表示无限制)</Label>
            <Input id="edit-capacity" v-model.number="editCapacity" type="number" min="0" />
            <p v-if="editingUser" class="text-xs text-muted-foreground">
              当前已用 {{ formatSize(editingUser.used_capacity || 0) }}，{{ remainingText(editingUser) }}。
            </p>
          </div>
          <div class="flex items-center gap-2">
            <Label for="edit-adminer">管理员</Label>
            <Switch id="edit-adminer" v-model="editAdminer" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showEditDialog = false">取消</Button>
          <Button :disabled="editSaving" @click="saveUser">
            {{ editSaving ? '保存中...' : '保存' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 删除确认对话框 -->
    <ConfirmDialog
      v-model:open="deleteDialogOpen"
      title="删除用户"
      :description="`确定要删除用户「${deleteTargetName}」吗？此操作不可撤销，该用户的所有数据将被永久删除。`"
      confirm-text="删除"
      cancel-text="取消"
      variant="destructive"
      @confirm="handleDeleteConfirm"
    />
  </div>
</template>
