<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { Upload, Image, FolderOpen, Zap, ShieldCheck, ServerCog, ArrowRight } from 'lucide-vue-next'

const auth = useAuthStore()
const router = useRouter()

const previewImages = [
  'https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=640&q=80',
  'https://images.unsplash.com/photo-1493246507139-91e8fad9978e?auto=format&fit=crop&w=640&q=80',
  'https://images.unsplash.com/photo-1519681393784-d120267933ba?auto=format&fit=crop&w=640&q=80',
  'https://images.unsplash.com/photo-1500534314209-a25ddb2bd429?auto=format&fit=crop&w=640&q=80',
]

const features = [
  { icon: Upload, title: '快速上传', text: '拖拽、粘贴、批量上传，上传后立即复制链接。' },
  { icon: FolderOpen, title: '相册归档', text: '按项目、客户、场景组织图片，管理路径更清晰。' },
  { icon: ServerCog, title: '多存储策略', text: '本地、S3、COS 等策略统一配置，迁移更平滑。' },
  { icon: Zap, title: 'API 集成', text: '保留原版的接口习惯，适合接入脚本和第三方系统。' },
]
</script>

<template>
  <div class="min-h-screen overflow-hidden">
    <header class="mx-auto flex w-full max-w-7xl items-center justify-between px-5 py-5 lg:px-8">
      <router-link to="/" class="flex items-center gap-3 font-semibold">
        <span class="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-purple-600 to-indigo-500 text-white shadow-lg shadow-purple-500/20">
          <Image class="h-5 w-5" />
        </span>
        <span>
          <span class="block text-base text-white">洛克图床</span>
          <span class="block text-xs font-medium text-slate-500">Image Hosting</span>
        </span>
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
      <section class="mx-auto grid min-h-[calc(100vh-92px)] w-full max-w-7xl items-center gap-10 px-5 pb-10 pt-6 lg:grid-cols-[0.95fr_1.05fr] lg:px-8">
        <div class="max-w-2xl">
          <Badge class="mb-5 border-primary/20 bg-primary/10 text-primary hover:bg-primary/10">
            Go 重构版控制台
          </Badge>
          <h1 class="text-4xl font-semibold leading-tight sm:text-5xl lg:text-6xl">
            图片托管、相册管理和多云存储，放在一个清爽的工作台里。
          </h1>
          <p class="mt-5 max-w-xl text-base leading-7 text-muted-foreground sm:text-lg">
            面向个人图床和团队素材库的管理界面。上传、归档、复制链接、切换存储策略都保持低摩擦。
          </p>
          <div class="mt-8 flex flex-wrap gap-3">
            <Button size="lg" class="h-11 px-5" @click="router.push(auth.isAuthenticated ? '/upload' : '/login')">
              <Upload class="mr-2 h-5 w-5" />
              开始上传
            </Button>
            <Button size="lg" variant="outline" class="h-11 px-5 bg-white/5 hover:bg-white/10" @click="router.push('/gallery')">
              <Image class="mr-2 h-5 w-5" />
              浏览画廊
            </Button>
          </div>
          <div class="mt-8 grid max-w-lg grid-cols-3 gap-4 text-sm">
            <div>
              <p class="text-2xl font-semibold">10+</p>
              <p class="text-muted-foreground">存储策略</p>
            </div>
            <div>
              <p class="text-2xl font-semibold">API</p>
              <p class="text-muted-foreground">开放接口</p>
            </div>
            <div>
              <p class="text-2xl font-semibold">Vue</p>
              <p class="text-muted-foreground">新控制台</p>
            </div>
          </div>
        </div>

        <div class="relative">
          <div class="grid grid-cols-2 gap-3 sm:gap-4">
            <div class="space-y-3 sm:space-y-4">
              <img
                v-for="(src, index) in previewImages.slice(0, 2)"
                :key="src"
                :src="src"
                :class="index === 0 ? 'aspect-[4/5]' : 'aspect-[4/3]'"
                class="w-full rounded-3xl object-cover shadow-2xl shadow-black/30 ring-1 ring-purple-500/10"
                alt=""
              />
            </div>
            <div class="mt-10 space-y-3 sm:space-y-4">
              <img
                v-for="(src, index) in previewImages.slice(2)"
                :key="src"
                :src="src"
                :class="index === 0 ? 'aspect-[4/3]' : 'aspect-[4/5]'"
                class="w-full rounded-3xl object-cover shadow-2xl shadow-black/30 ring-1 ring-purple-500/10"
                alt=""
              />
            </div>
          </div>
          <Card class="absolute bottom-6 left-6 max-w-xs border-purple-500/15 bg-[#0f0f15] shadow-md">
            <CardContent class="p-4">
              <div class="mb-3 flex items-center justify-between">
                <div class="flex items-center gap-2 text-sm font-semibold text-slate-200">
                  <ShieldCheck class="h-4 w-4 text-purple-400" />
                  存储策略在线
                </div>
                <span class="h-2 w-2 rounded-full bg-emerald-500" />
              </div>
              <div class="space-y-2 text-xs text-muted-foreground">
                <div class="flex justify-between">
                  <span>默认本地策略</span>
                  <span>可用</span>
                </div>
                <div class="h-2 rounded-full bg-white/10">
                  <div class="h-2 w-2/3 rounded-full bg-purple-500" />
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </section>

      <section class="border-t border-purple-500/10 bg-[#0a0a0f] px-5 py-14 lg:px-8">
        <div class="mx-auto grid max-w-7xl gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card v-for="feature in features" :key="feature.title">
            <CardContent class="p-5">
              <div class="mb-4 flex h-10 w-10 items-center justify-center rounded-2xl bg-purple-500/15 text-purple-400">
                <component :is="feature.icon" class="h-5 w-5" />
              </div>
              <h2 class="text-base font-semibold">{{ feature.title }}</h2>
              <p class="mt-2 text-sm leading-6 text-muted-foreground">{{ feature.text }}</p>
            </CardContent>
          </Card>
        </div>
      </section>
    </main>
  </div>
</template>
