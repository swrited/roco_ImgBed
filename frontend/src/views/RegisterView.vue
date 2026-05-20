<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'vue-sonner'
import { Image } from 'lucide-vue-next'

const auth = useAuthStore()
const router = useRouter()

const name = ref('')
const email = ref('')
const password = ref('')
const passwordConfirmation = ref('')
const loading = ref(false)

async function handleRegister() {
  if (password.value !== passwordConfirmation.value) {
    toast.error('两次密码输入不一致')
    return
  }
  loading.value = true
  try {
    await auth.register({
      name: name.value,
      email: email.value,
      password: password.value,
      password_confirmation: passwordConfirmation.value,
    })
    toast.success('注册成功，请登录')
    router.push('/login')
  } catch (e: any) {
    toast.error(e.message || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center px-4 py-12">
    <Card class="w-full max-w-md">
      <CardHeader class="text-center">
        <div class="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-purple-500/20">
          <Image class="h-6 w-6 text-purple-400" />
        </div>
        <CardTitle class="text-2xl">创建账户</CardTitle>
        <CardDescription>注册后即可使用全部功能</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="space-y-2">
          <Label for="name">用户名</Label>
          <Input id="name" v-model="name" placeholder="你的用户名" />
        </div>
        <div class="space-y-2">
          <Label for="email">邮箱</Label>
          <Input id="email" v-model="email" type="email" placeholder="your@email.com" />
        </div>
        <div class="space-y-2">
          <Label for="password">密码</Label>
          <Input id="password" v-model="password" type="password" placeholder="••••••••" />
        </div>
        <div class="space-y-2">
          <Label for="password_confirmation">确认密码</Label>
          <Input id="password_confirmation" v-model="passwordConfirmation" type="password" placeholder="••••••••" />
        </div>
        <Button class="w-full" :disabled="loading" @click="handleRegister">
          {{ loading ? '注册中...' : '注册' }}
        </Button>
      </CardContent>
      <CardFooter class="justify-center">
        <p class="text-sm text-muted-foreground">
          已有账户？
          <router-link to="/login" class="text-purple-400 hover:underline">立即登录</router-link>
        </p>
      </CardFooter>
    </Card>
  </div>
</template>
