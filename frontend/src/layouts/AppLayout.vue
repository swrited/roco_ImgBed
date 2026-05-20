<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { RouterView } from 'vue-router'
import {
  Sheet, SheetContent, SheetTrigger,
} from '@/components/ui/sheet'
import {
  DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel,
  DropdownMenuSeparator, DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Menu, LayoutDashboard, Upload, Image, FolderOpen, Settings, LogOut, Images, BookOpen, UserCog, Shield, HardDrive, BarChart3, Users, Sparkles, Key } from 'lucide-vue-next'
import { computed } from 'vue'

const auth = useAuthStore()
const router = useRouter()

const userInitials = computed(() => {
  return auth.user?.name?.charAt(0).toUpperCase() || 'U'
})

const navItems = [
  { to: '/dashboard', label: '控制面板', icon: LayoutDashboard, requiresAuth: true },
  { to: '/upload', label: '上传图片', icon: Upload, requiresAuth: true },
  { to: '/images', label: '图片管理', icon: Image, requiresAuth: true },
  { to: '/albums', label: '相册管理', icon: FolderOpen, requiresAuth: true },
  { to: '/api-keys', label: 'API Keys', icon: Key, requiresAuth: true },
  { to: '/gallery', label: '画廊', icon: Images },
  { to: '/api-doc', label: 'API 文档', icon: BookOpen },
]

const adminNavItems = [
  { to: '/admin', label: '控制台', icon: BarChart3 },
  { to: '/admin/users', label: '用户管理', icon: Users },
  { to: '/admin/images', label: '图片管理', icon: Image },
  { to: '/admin/groups', label: '角色组', icon: Shield },
  { to: '/admin/strategies', label: '存储策略', icon: HardDrive },
  { to: '/admin/settings', label: '系统设置', icon: Settings },
]

function isActive(path: string) {
  return router.currentRoute.value.path === path || router.currentRoute.value.path.startsWith(path + '/')
}
</script>

<template>
  <div class="flex h-screen bg-background text-foreground">
    <!-- Desktop Sidebar -->
    <aside class="hidden lg:flex w-72 flex-col border-r border-purple-500/10 bg-[#0a0a0f]">
      <div class="flex h-16 items-center border-b border-purple-500/10 px-5">
        <router-link to="/" class="flex items-center gap-3 font-semibold">
          <span class="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-purple-600 to-indigo-500 text-white shadow-lg shadow-purple-500/20">
            <Image class="h-5 w-5" />
          </span>
          <span class="leading-tight">
            <span class="block text-base text-white">洛克图床</span>
            <span class="block text-xs font-medium text-slate-500">Image Console</span>
          </span>
        </router-link>
      </div>
      <nav class="flex-1 overflow-auto px-3 py-5 space-y-1.5">
        <div v-for="item in navItems" :key="item.to">
          <router-link
            v-if="!item.requiresAuth || auth.isAuthenticated"
            :to="item.to"
            class="flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-sm transition-all border border-transparent"
            :class="isActive(item.to) ? 'bg-purple-500/10 text-purple-400 border-purple-500/20 shadow-[0_0_10px_rgba(139,92,246,0.05)]' : 'text-slate-400 hover:bg-white/5 hover:text-slate-200'"
          >
            <component :is="item.icon" class="h-4 w-4" />
            {{ item.label }}
          </router-link>
        </div>

        <template v-if="auth.isAdmin">
          <Separator class="my-4" />
          <p class="px-3 text-xs font-semibold text-slate-500 uppercase tracking-wider">管理后台</p>
          <div v-for="item in adminNavItems" :key="item.to">
            <router-link
              :to="item.to"
              class="flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-sm transition-all border border-transparent"
              :class="isActive(item.to) ? 'bg-purple-500/10 text-purple-400 border-purple-500/20 shadow-[0_0_10px_rgba(139,92,246,0.05)]' : 'text-slate-400 hover:bg-white/5 hover:text-slate-200'"
            >
              <component :is="item.icon" class="h-4 w-4" />
              {{ item.label }}
            </router-link>
          </div>
        </template>
      </nav>
      <div class="border-t border-purple-500/10 p-4">
        <div class="mb-3 rounded-2xl border border-purple-500/15 bg-purple-500/5 p-3 text-xs">
          <div class="mb-1 flex items-center gap-2 font-semibold text-purple-400">
            <Sparkles class="h-3.5 w-3.5" />
            重构预览版
          </div>
          <p class="text-slate-500">Go API + Vue 控制台正在迁移中</p>
        </div>
        <div class="flex items-center gap-3 rounded-2xl bg-white/5 border border-white/5 p-2">
          <Avatar class="h-9 w-9 ring-2 ring-purple-500/20">
            <AvatarImage :src="auth.user?.avatar || ''" />
            <AvatarFallback class="bg-purple-500/20 text-purple-400">{{ userInitials }}</AvatarFallback>
          </Avatar>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-slate-200 truncate">{{ auth.user?.name }}</p>
            <p class="text-xs text-slate-500 truncate">{{ auth.user?.email }}</p>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" size="icon" class="h-8 w-8 rounded-xl text-slate-400 hover:text-white">
                <Settings class="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>账户</DropdownMenuLabel>
              <DropdownMenuItem @click="router.push('/settings')">
                <UserCog class="mr-2 h-4 w-4" /> 用户设置
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem @click="auth.logout()">
                <LogOut class="mr-2 h-4 w-4" /> 退出登录
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </aside>

    <!-- Mobile header + content -->
    <div class="flex flex-1 flex-col overflow-hidden">
      <!-- Mobile header -->
      <header class="flex h-14 items-center gap-4 border-b border-purple-500/10 bg-[#050505]/90 backdrop-blur-xl px-4 lg:hidden">
        <Sheet>
          <SheetTrigger as-child>
            <Button variant="ghost" size="icon">
              <Menu class="h-5 w-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" class="w-64 p-0">
            <div class="flex h-14 items-center border-b border-purple-500/10 px-4">
              <router-link to="/" class="flex items-center gap-2 font-semibold text-white">
                <span class="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-purple-600 to-indigo-500">
                  <Image class="h-4 w-4 text-white" />
                </span>
                <span>洛克图床</span>
              </router-link>
            </div>
            <nav class="flex-1 overflow-auto py-4 px-2 space-y-1">
              <div v-for="item in navItems" :key="item.to">
                <router-link
                  v-if="!item.requiresAuth || auth.isAuthenticated"
                  :to="item.to"
                  class="flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors"
                  :class="isActive(item.to) ? 'bg-purple-500/10 text-purple-400' : 'text-slate-400 hover:bg-white/5 hover:text-slate-200'"
                >
                  <component :is="item.icon" class="h-4 w-4" />
                  {{ item.label }}
                </router-link>
              </div>
            </nav>
          </SheetContent>
        </Sheet>
        <router-link to="/" class="flex items-center gap-2 font-semibold text-white">
          <span class="flex h-7 w-7 items-center justify-center rounded-lg bg-gradient-to-br from-purple-600 to-indigo-500">
            <Image class="h-4 w-4 text-white" />
          </span>
          <span>洛克图床</span>
        </router-link>
        <div class="flex-1" />
        <DropdownMenu v-if="auth.isAuthenticated">
          <DropdownMenuTrigger as-child>
            <Button variant="ghost" size="icon" class="rounded-full">
              <Avatar class="h-8 w-8 ring-2 ring-purple-500/20">
                <AvatarFallback class="bg-purple-500/20 text-purple-400">{{ userInitials }}</AvatarFallback>
              </Avatar>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>{{ auth.user?.name }}</DropdownMenuLabel>
            <DropdownMenuItem @click="router.push('/settings')">设置</DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem @click="auth.logout()">退出</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <div v-else class="flex gap-2">
          <Button variant="ghost" size="sm" @click="router.push('/login')">登录</Button>
          <Button size="sm" @click="router.push('/register')">注册</Button>
        </div>
      </header>

      <!-- Page content -->
      <main class="flex-1 overflow-auto bg-gradient-to-br from-transparent via-purple-900/5 to-transparent">
        <div class="page-shell">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>
