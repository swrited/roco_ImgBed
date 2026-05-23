<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { toast } from 'vue-sonner'
import { AlertCircle, ArrowLeft } from 'lucide-vue-next'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const loading = ref(false)
const errorDialogOpen = ref(false)
const errorMessage = ref('')

function showError(message: string) {
  errorMessage.value = message
  errorDialogOpen.value = true
}

async function handleLogin() {
  loading.value = true
  try {
    await auth.login({ email: email.value, password: password.value })
    toast.success('登录成功')
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.push(redirect)
  } catch (e: any) {
    showError(e.message || '账号或密码错误')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="grid min-h-screen lg:grid-cols-[1.05fr_0.95fr]">
    <section class="relative hidden overflow-hidden lg:block">
      <img
        src="https://images.unsplash.com/photo-1497366754035-f200968a6e72?auto=format&fit=crop&w=1400&q=80"
        alt=""
        class="h-full w-full object-cover"
      />
      <div class="absolute inset-0 bg-black/50" />
      <div class="absolute bottom-10 left-10 max-w-md text-white">
        <img src="/roco-wordmark.svg" alt="洛克图床" class="mb-4 h-16 w-44 object-contain object-left" />
        <h1 class="text-4xl font-semibold leading-tight">把图片资产管理变成一个安静、高效的工作流。</h1>
      </div>
    </section>

    <section class="flex items-center justify-center px-5 py-10">
      <div class="w-full max-w-md">
        <Button variant="ghost" class="mb-6 -ml-2" @click="router.push('/')">
          <ArrowLeft class="mr-2 h-4 w-4" />
          返回首页
        </Button>
        <Card>
          <CardHeader>
            <img src="/roco-logo.svg" alt="洛克图床" class="mb-2 h-12 w-12 rounded-2xl object-contain" />
            <CardTitle class="text-2xl">登录账户</CardTitle>
            <CardDescription>进入控制台管理图片、相册和存储策略</CardDescription>
          </CardHeader>
          <CardContent class="space-y-4">
            <div class="space-y-2">
              <Label for="email">邮箱</Label>
              <Input id="email" v-model="email" type="email" placeholder="admin@admin.com" />
            </div>
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <Label for="password">密码</Label>
                <router-link to="/forgot-password" class="text-sm text-purple-400 hover:underline">
                  忘记密码?
                </router-link>
              </div>
              <Input id="password" v-model="password" type="password" placeholder="请输入密码" @keyup.enter="handleLogin" />
            </div>
            <Button class="h-10 w-full" :disabled="loading" @click="handleLogin">
              {{ loading ? '登录中...' : '登录' }}
            </Button>
          </CardContent>
          <CardFooter class="justify-center">
            <p class="text-sm text-muted-foreground">
              没有账户？
              <router-link to="/register" class="text-purple-400 hover:underline">立即注册</router-link>
            </p>
          </CardFooter>
        </Card>
      </div>
    </section>

    <Dialog v-model:open="errorDialogOpen">
      <DialogContent class="sm:max-w-md">
        <DialogHeader>
          <div class="mb-2 flex h-11 w-11 items-center justify-center rounded-full bg-destructive/10 text-destructive">
            <AlertCircle class="h-5 w-5" />
          </div>
          <DialogTitle>登录失败</DialogTitle>
          <DialogDescription class="text-base leading-6">
            {{ errorMessage }}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button @click="errorDialogOpen = false">知道了</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
