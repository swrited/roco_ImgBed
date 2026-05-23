<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Upload, Image, FolderOpen, Zap, ServerCog, ArrowRight } from 'lucide-vue-next'

const auth = useAuthStore()
const router = useRouter()

const previewImages = [
  'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=640&q=80',
  'https://images.unsplash.com/photo-1493246507139-91e8fad9978e?auto=format&fit=crop&w=640&q=80',
  'https://images.unsplash.com/photo-1519681393784-d120267933ba?auto=format&fit=crop&w=640&q=80',
  'https://images.unsplash.com/photo-1500534314209-a25ddb2bd429?auto=format&fit=crop&w=640&q=80',
]

const features = [
  { icon: Upload,    title: '快速上传',   text: '拖拽、粘贴、批量上传，上传后立即复制链接。',      color: 'violet' },
  { icon: FolderOpen, title: '相册归档', text: '按项目、客户、场景组织图片，管理路径更清晰。',    color: 'indigo' },
  { icon: ServerCog, title: '多存储策略', text: '本地、S3、COS 等策略统一配置，迁移更平滑。',    color: 'sky'   },
  { icon: Zap,       title: 'API 集成',  text: '保留原版的接口习惯，适合接入脚本和第三方系统。', color: 'amber' },
]

const featureColors: Record<string, string> = {
  violet: 'bg-violet-500/10 text-violet-400 group-hover:bg-violet-500/18 group-hover:shadow-violet-500/20',
  indigo: 'bg-indigo-500/10 text-indigo-400 group-hover:bg-indigo-500/18 group-hover:shadow-indigo-500/20',
  sky:    'bg-sky-500/10    text-sky-400    group-hover:bg-sky-500/18    group-hover:shadow-sky-500/20',
  amber:  'bg-amber-500/10  text-amber-400  group-hover:bg-amber-500/18  group-hover:shadow-amber-500/20',
}
</script>

<template>
  <div class="min-h-screen overflow-hidden">

    <!-- ── Header ─────────────────────────────────────────────────────── -->
    <header class="relative z-10 mx-auto flex w-full max-w-7xl items-center justify-between px-5 py-5 lg:px-8">
      <router-link to="/" class="flex items-center gap-3 font-semibold">
        <img
          src="/roco-logo.svg"
          alt="洛克图床"
          class="h-10 w-10 rounded-xl object-contain shadow-lg shadow-purple-500/20"
        />
        <img src="/roco-wordmark.svg" alt="洛克图床" class="h-11 w-32 object-contain object-left" />
      </router-link>
      <div class="flex items-center gap-2">
        <Button variant="ghost" @click="router.push('/gallery')">画廊</Button>
        <Button v-if="auth.isAuthenticated" @click="router.push('/dashboard')">
          控制台
          <ArrowRight class="ml-1 h-4 w-4" />
        </Button>
        <template v-else>
          <Button variant="outline" @click="router.push('/login')">登录</Button>
          <Button @click="router.push('/register')">注册</Button>
        </template>
      </div>
    </header>

    <main>
      <!-- ── Hero ─────────────────────────────────────────────────────── -->
      <section class="relative mx-auto grid min-h-[calc(100vh-92px)] w-full max-w-7xl items-center gap-10 px-5 pb-10 pt-6 lg:grid-cols-[0.95fr_1.05fr] lg:px-8">

        <!-- Ambient glow balls -->
        <div aria-hidden="true" class="pointer-events-none absolute inset-0 overflow-hidden">
          <!-- Main violet glow (top-left area) -->
          <div class="ambient-glow ambient-glow-violet absolute -top-32 -left-32 h-[480px] w-[480px]" />
          <!-- Secondary indigo glow (center-right) -->
          <div class="ambient-glow ambient-glow-indigo absolute top-1/3 right-0 h-[360px] w-[360px]" />
          <!-- Sky glow (bottom-left) -->
          <div class="ambient-glow ambient-glow-sky absolute bottom-0 left-1/4 h-[280px] w-[280px]" />
          <!-- Subtle grid overlay -->
          <div
            class="absolute inset-0 opacity-[0.025]"
            style="background-image: linear-gradient(oklch(80% 0.15 293) 1px, transparent 1px),
                                      linear-gradient(90deg, oklch(80% 0.15 293) 1px, transparent 1px);
                   background-size: 48px 48px;"
          />
        </div>

        <!-- Left: copy -->
        <div class="relative max-w-2xl">
          <Badge
            class="animate-fade-in-up mb-5 border-violet-500/25 bg-violet-500/10 text-violet-400 hover:bg-violet-500/10"
          >
            <span class="mr-1.5 inline-block h-1.5 w-1.5 rounded-full bg-violet-400 animate-glow-pulse" />
            Go 重构版控制台
          </Badge>

          <h1 class="animate-fade-in-up delay-75 text-4xl font-semibold leading-tight sm:text-5xl lg:text-6xl">
            图片托管、相册管理和多云存储，放在一个清爽的工作台里。
          </h1>

          <p class="animate-fade-in-up delay-150 mt-5 max-w-xl text-base leading-7 text-muted-foreground sm:text-lg">
            面向个人图床和团队素材库的管理界面。上传、归档、复制链接、切换存储策略都保持低摩擦。
          </p>

          <div class="animate-fade-in-up delay-225 mt-8 flex flex-wrap gap-3">
            <Button size="lg" class="h-11 px-5" @click="router.push(auth.isAuthenticated ? '/upload' : '/login')">
              <Upload class="mr-2 h-5 w-5" />
              开始上传
            </Button>
            <Button
              size="lg"
              variant="outline"
              class="h-11 bg-white/5 px-5 hover:bg-white/10"
              @click="router.push('/gallery')"
            >
              <Image class="mr-2 h-5 w-5" />
              浏览画廊
            </Button>
          </div>

          <!-- Stats strip -->
          <div class="animate-fade-in-up delay-300 mt-8 grid max-w-lg grid-cols-3 gap-4 text-sm">
            <div
              v-for="stat in [{ val: '10+', label: '存储策略' }, { val: 'API', label: '开放接口' }, { val: 'Vue', label: '新控制台' }]"
              :key="stat.label"
              class="rounded-xl border border-white/5 bg-white/[0.03] p-3"
            >
              <p class="text-2xl font-semibold text-white">{{ stat.val }}</p>
              <p class="mt-0.5 text-muted-foreground">{{ stat.label }}</p>
            </div>
          </div>
        </div>

        <!-- Right: image mosaic -->
        <div class="animate-fade-in-up delay-150 relative">
          <!-- Glow behind images -->
          <div
            aria-hidden="true"
            class="absolute inset-0 -z-10 scale-95 rounded-3xl blur-2xl"
            style="background: radial-gradient(ellipse at 50% 50%, oklch(60% 0.22 293 / 0.14) 0%, transparent 70%)"
          />
          <div class="grid grid-cols-2 gap-3 sm:gap-4">
            <div class="space-y-3 sm:space-y-4">
              <img
                v-for="(src, index) in previewImages.slice(0, 2)"
                :key="src"
                :src="src"
                :class="index === 0 ? 'aspect-[4/5]' : 'aspect-[4/3]'"
                class="w-full rounded-3xl object-cover shadow-2xl shadow-black/40 ring-1 ring-purple-500/15 transition-transform duration-500 hover:scale-[1.02]"
                alt=""
                loading="lazy"
              />
            </div>
            <div class="mt-10 space-y-3 sm:space-y-4">
              <img
                v-for="(src, index) in previewImages.slice(2)"
                :key="src"
                :src="src"
                :class="index === 0 ? 'aspect-[4/3]' : 'aspect-[4/5]'"
                class="w-full rounded-3xl object-cover shadow-2xl shadow-black/40 ring-1 ring-purple-500/15 transition-transform duration-500 hover:scale-[1.02]"
                alt=""
                loading="lazy"
              />
            </div>
          </div>
        </div>
      </section>

      <!-- ── Features ─────────────────────────────────────────────────── -->
      <section class="relative border-t border-purple-500/10 bg-[oklch(7.5%_0.007_270)] px-5 py-16 lg:px-8">
        <!-- Subtle top-center glow -->
        <div
          aria-hidden="true"
          class="pointer-events-none absolute inset-x-0 top-0 h-px"
          style="background: linear-gradient(90deg, transparent, oklch(60% 0.22 293 / 0.4), transparent)"
        />
        <div class="mx-auto grid max-w-7xl gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card
            v-for="feature in features"
            :key="feature.title"
            class="group cursor-default transition-all duration-300 hover:-translate-y-1 hover:border-white/10"
          >
            <CardContent class="p-5">
              <div
                :class="featureColors[feature.color]"
                class="mb-4 flex h-10 w-10 items-center justify-center rounded-2xl shadow-lg transition-all duration-300"
              >
                <component :is="feature.icon" class="h-5 w-5 transition-transform duration-300 group-hover:scale-110" />
              </div>
              <h2 class="text-base font-semibold text-foreground">{{ feature.title }}</h2>
              <p class="mt-2 text-sm leading-6 text-muted-foreground">{{ feature.text }}</p>
            </CardContent>
          </Card>
        </div>
      </section>
    </main>
  </div>
</template>
