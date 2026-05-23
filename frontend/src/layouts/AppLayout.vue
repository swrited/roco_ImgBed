<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { RouterView } from 'vue-router'
import {
  Sheet, SheetContent, SheetTrigger,
} from '@/components/ui/sheet'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  Menu, LayoutDashboard, Upload, Image, FolderOpen, Images,
  BookOpen, Settings, Shield, HardDrive, BarChart3, Users,
  Key, Terminal, Activity, WandSparkles,
} from 'lucide-vue-next'
import { computed, onMounted } from 'vue'

const auth = useAuthStore()
const router = useRouter()

const userInitials = computed(() => auth.user?.name?.charAt(0).toUpperCase() || 'U')
const usedCapacity  = computed(() => auth.user?.used_capacity || 0)
const totalCapacity = computed(() => auth.user?.capacity     || 0)
const capacityLabel = computed(() => {
  if (!totalCapacity.value || totalCapacity.value <= 0) return `${formatSize(usedCapacity.value)} / 无限制`
  return `${formatSize(usedCapacity.value)} / ${formatSize(totalCapacity.value)}`
})
const capacityPercent = computed(() => {
  if (!totalCapacity.value || totalCapacity.value <= 0) return 0
  return Math.min(100, Math.round((usedCapacity.value / totalCapacity.value) * 100))
})

const navItems = [
  { to: '/dashboard',   label: '控制面板', icon: LayoutDashboard, requiresAuth: true  },
  { to: '/upload',      label: '上传图片', icon: Upload,          requiresAuth: true  },
  { to: '/ai-generate', label: 'AI 生图',  icon: WandSparkles,   requiresAuth: true  },
  { to: '/images',      label: '图片管理', icon: Image,           requiresAuth: true  },
  { to: '/albums',      label: '相册管理', icon: FolderOpen,      requiresAuth: true  },
  { to: '/api-keys',    label: 'API Keys', icon: Key,             requiresAuth: true  },
  { to: '/api-usage',   label: 'API 统计', icon: Activity,        requiresAuth: true  },
  { to: '/gallery',     label: '画廊',     icon: Images                               },
  { to: '/api-doc',     label: 'API 文档', icon: BookOpen                             },
  { to: '/api-test',    label: 'API 测试', icon: Terminal                             },
]

const adminNavItems = [
  { to: '/admin',              label: '控制台',   icon: BarChart3, exact: true },
  { to: '/admin/users',        label: '用户管理', icon: Users                  },
  { to: '/admin/images',       label: '图片管理', icon: Image                  },
  { to: '/admin/groups',       label: '角色组',   icon: Shield                 },
  { to: '/admin/strategies',   label: '存储策略', icon: HardDrive              },
  { to: '/admin/api-usage',    label: 'API 统计', icon: Activity               },
  { to: '/admin/settings',     label: '系统设置', icon: Settings               },
]

function isActive(path: string, exact = false) {
  const currentPath = router.currentRoute.value.path
  return exact ? currentPath === path : currentPath === path || currentPath.startsWith(path + '/')
}

function formatSize(kb: number): string {
  if (!kb || kb <= 0) return '0 KB'
  if (kb < 1024)        return `${kb.toFixed(0)} KB`
  if (kb < 1024 * 1024) return `${(kb / 1024).toFixed(1)} MB`
  return `${(kb / 1024 / 1024).toFixed(2)} GB`
}

onMounted(() => {
  if (auth.isAuthenticated) auth.fetchProfile().catch(() => {})
})
</script>

<template>
  <div class="flex h-screen bg-background text-foreground">

    <!-- ── Desktop Sidebar ───────────────────────────────────────────────── -->
    <aside class="relative hidden lg:flex w-72 flex-col border-r border-violet-500/10 bg-[oklch(7%_0.007_270)]">

      <!-- Top ambient glow blob -->
      <div
        aria-hidden="true"
        class="pointer-events-none absolute -top-20 left-1/2 h-56 w-56 -translate-x-1/2 rounded-full blur-3xl"
        style="background: radial-gradient(circle, oklch(60% 0.22 293 / 0.2) 0%, transparent 70%)"
      />
      <!-- Bottom ambient glow -->
      <div
        aria-hidden="true"
        class="pointer-events-none absolute -bottom-16 left-8 h-48 w-48 rounded-full blur-3xl"
        style="background: radial-gradient(circle, oklch(55% 0.2 265 / 0.15) 0%, transparent 70%)"
      />

      <!-- Logo -->
      <div class="relative z-10 flex h-16 items-center border-b border-violet-500/10 px-5">
        <router-link to="/" class="flex items-center gap-3 font-semibold">
          <img
            src="/roco-logo.svg"
            alt="洛克图床"
            class="h-10 w-10 rounded-xl object-contain shadow-lg shadow-purple-500/25"
          />
          <img src="/roco-wordmark.svg" alt="洛克图床" class="h-11 w-32 object-contain object-left" />
        </router-link>
      </div>

      <!-- Nav -->
      <nav class="relative z-10 flex-1 overflow-auto px-3 py-5 space-y-0.5">
        <div v-for="item in navItems" :key="item.to">
          <router-link
            v-if="!item.requiresAuth || auth.isAuthenticated"
            :to="item.to"
            class="relative flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-sm transition-all duration-150 border border-transparent"
            :class="isActive(item.to) ? 'nav-item-active' : 'text-slate-400 hover:bg-white/5 hover:text-slate-100'"
          >
            <component :is="item.icon" class="h-4 w-4 shrink-0" />
            {{ item.label }}
          </router-link>
        </div>

        <template v-if="auth.isAdmin">
          <Separator class="my-4" />
          <p class="px-3.5 pb-1 text-xs font-semibold uppercase tracking-wider text-slate-500">
            管理后台
          </p>
          <div v-for="item in adminNavItems" :key="item.to">
            <router-link
              :to="item.to"
              class="relative flex items-center gap-3 rounded-xl px-3.5 py-2.5 text-sm transition-all duration-150 border border-transparent"
              :class="isActive(item.to, item.exact) ? 'nav-item-active' : 'text-slate-400 hover:bg-white/5 hover:text-slate-100'"
            >
              <component :is="item.icon" class="h-4 w-4 shrink-0" />
              {{ item.label }}
            </router-link>
          </div>
        </template>
      </nav>

      <!-- User footer -->
      <div class="relative z-10 border-t border-violet-500/10 p-4 space-y-3">
        <template v-if="auth.isAuthenticated">
          <!-- Capacity widget -->
          <div class="rounded-xl border border-white/8 bg-white/[0.03] p-3">
            <div class="mb-2 flex items-center justify-between text-xs">
              <span class="font-medium text-slate-300">存储空间</span>
              <span class="tabular-nums text-slate-500">{{ capacityLabel }}</span>
            </div>
            <div class="h-1.5 overflow-hidden rounded-full bg-white/8">
              <div
                class="h-full rounded-full transition-all duration-700"
                :style="{
                  width: totalCapacity > 0 ? `${capacityPercent}%` : '100%',
                  background: 'linear-gradient(90deg, oklch(55% 0.22 265), oklch(62% 0.24 295) 50%, oklch(68% 0.18 310))',
                  boxShadow: '0 0 8px oklch(60% 0.22 293 / 0.5)',
                }"
              />
            </div>
          </div>

          <!-- Profile button -->
          <button
            type="button"
            class="flex w-full min-w-0 items-center gap-3 rounded-xl border border-white/5 bg-white/[0.04] p-2.5 text-left transition-all duration-150 hover:border-violet-500/30 hover:bg-violet-500/10"
            @click="router.push('/settings')"
          >
            <Avatar class="h-9 w-9 ring-2 ring-violet-500/25 ring-offset-1 ring-offset-[oklch(7%_0.007_270)]">
              <AvatarImage :src="auth.user?.avatar || ''" />
              <AvatarFallback class="bg-violet-500/20 text-violet-300 font-semibold">
                {{ userInitials }}
              </AvatarFallback>
            </Avatar>
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-medium text-slate-100">{{ auth.user?.name }}</p>
              <p class="truncate text-xs text-slate-500">{{ auth.user?.email }}</p>
            </div>
            <Settings class="h-3.5 w-3.5 shrink-0 text-slate-600" />
          </button>
        </template>

        <!-- Not logged in -->
        <div v-else class="space-y-3 rounded-xl border border-violet-500/20 bg-violet-500/10 p-3.5">
          <div>
            <p class="text-sm font-semibold text-slate-100">登录后管理图片</p>
            <p class="mt-1 text-xs leading-5 text-slate-400">
              上传图片、创建相册和生成 API Key。
            </p>
          </div>
          <div class="grid grid-cols-2 gap-2">
            <Button class="h-9" @click="router.push('/login')">登录</Button>
            <Button variant="outline" class="h-9 border-white/10 bg-white/5" @click="router.push('/register')">注册</Button>
          </div>
        </div>
      </div>
    </aside>

    <!-- ── Mobile header + content ──────────────────────────────────────── -->
    <div class="flex flex-1 flex-col overflow-hidden">

      <!-- Mobile header -->
      <header class="flex h-14 items-center gap-4 border-b border-violet-500/10 bg-[oklch(6%_0.005_270)]/90 px-4 backdrop-blur-xl lg:hidden">
        <Sheet>
          <SheetTrigger as-child>
            <Button variant="ghost" size="icon">
              <Menu class="h-5 w-5" />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" class="w-64 p-0">
            <div class="flex h-14 items-center border-b border-violet-500/10 px-4">
              <router-link to="/" class="flex items-center gap-2 font-semibold text-white">
                <img src="/roco-logo.svg" alt="洛克图床" class="h-8 w-8 rounded-lg object-contain" />
                <img src="/roco-wordmark.svg" alt="洛克图床" class="h-12 w-32 object-contain object-left" />
              </router-link>
            </div>
            <nav class="flex-1 overflow-auto py-4 px-2 space-y-0.5">
              <div v-for="item in navItems" :key="item.to">
                <router-link
                  v-if="!item.requiresAuth || auth.isAuthenticated"
                  :to="item.to"
                  class="relative flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition-all border border-transparent"
                  :class="isActive(item.to) ? 'nav-item-active' : 'text-slate-400 hover:bg-white/5 hover:text-slate-100'"
                >
                  <component :is="item.icon" class="h-4 w-4 shrink-0" />
                  {{ item.label }}
                </router-link>
              </div>
            </nav>
          </SheetContent>
        </Sheet>

        <router-link to="/" class="flex items-center gap-2 font-semibold text-white">
          <img src="/roco-logo.svg" alt="洛克图床" class="h-8 w-8 rounded-lg object-contain" />
          <img src="/roco-wordmark.svg" alt="洛克图床" class="h-12 w-32 object-contain object-left" />
        </router-link>
        <div class="flex-1" />

        <Button
          v-if="auth.isAuthenticated"
          variant="ghost"
          size="icon"
          class="rounded-full"
          @click="router.push('/settings')"
        >
          <Avatar class="h-8 w-8 ring-2 ring-violet-500/25">
            <AvatarFallback class="bg-violet-500/20 text-violet-300">{{ userInitials }}</AvatarFallback>
          </Avatar>
        </Button>
        <div v-else class="flex gap-2">
          <Button variant="ghost" size="sm" @click="router.push('/login')">登录</Button>
          <Button size="sm" @click="router.push('/register')">注册</Button>
        </div>
      </header>

      <!-- Page content -->
      <main class="flex-1 overflow-auto" style="background: radial-gradient(ellipse 80% 50% at 50% -10%, oklch(60% 0.22 293 / 0.06), transparent);">
        <div class="page-shell">
          <RouterView v-slot="{ Component, route }">
            <Transition name="page-slide" mode="out-in">
              <component :is="Component" :key="route.fullPath" />
            </Transition>
          </RouterView>
        </div>
      </main>
    </div>
  </div>
</template>
