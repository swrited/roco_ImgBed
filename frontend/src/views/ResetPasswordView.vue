<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { authApi } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'vue-sonner'
import { Image } from 'lucide-vue-next'
import { useRouter } from 'vue-router'

const router = useRouter()
const route = useRoute()
const token = route.params.token as string

const email = ref('')
const password = ref('')
const passwordConfirmation = ref('')
const loading = ref(false)

async function handleReset() {
  if (password.value !== passwordConfirmation.value) {
    toast.error('两次密码输入不一致')
    return
  }
  loading.value = true
  try {
    await authApi.resetPassword({
      token,
      email: email.value,
      password: password.value,
      password_confirmation: passwordConfirmation.value,
    })
    toast.success('密码重置成功，请登录')
    router.push('/login')
  } catch (e: any) {
    toast.error(e.message || '重置失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center px-4 py-12">
    <Card class="w-full max-w-md">
      <CardHeader class="text-center">
        <div class="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
          <Image class="h-6 w-6 text-primary" />
        </div>
        <CardTitle class="text-2xl">重置密码</CardTitle>
        <CardDescription>输入你的新密码</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="space-y-2">
          <Label for="email">邮箱</Label>
          <Input id="email" v-model="email" type="email" placeholder="your@email.com" />
        </div>
        <div class="space-y-2">
          <Label for="password">新密码</Label>
          <Input id="password" v-model="password" type="password" />
        </div>
        <div class="space-y-2">
          <Label for="password_confirmation">确认密码</Label>
          <Input id="password_confirmation" v-model="passwordConfirmation" type="password" />
        </div>
        <Button class="w-full" :disabled="loading" @click="handleReset">
          {{ loading ? '重置中...' : '重置密码' }}
        </Button>
      </CardContent>
      <CardFooter class="justify-center">
        <router-link to="/login" class="text-sm text-primary hover:underline">返回登录</router-link>
      </CardFooter>
    </Card>
  </div>
</template>
