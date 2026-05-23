<script setup lang="ts">
import { ref } from 'vue'
import { authApi } from '@/api/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { toast } from 'vue-sonner'

const email = ref('')
const loading = ref(false)
const sent = ref(false)

async function handleSubmit() {
  loading.value = true
  try {
    await authApi.forgotPassword(email.value)
    sent.value = true
    toast.success('重置链接已发送到你的邮箱')
  } catch (e: any) {
    toast.error(e.message || '发送失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center px-4 py-12">
    <Card class="w-full max-w-md">
      <CardHeader class="text-center">
        <img src="/roco-logo.svg" alt="洛克图床" class="mx-auto mb-2 h-12 w-12 rounded-full object-contain" />
        <CardTitle class="text-2xl">忘记密码</CardTitle>
        <CardDescription>{{ sent ? '请检查你的邮箱' : '输入邮箱接收重置链接' }}</CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="space-y-2">
          <Label for="email">邮箱</Label>
          <Input id="email" v-model="email" type="email" placeholder="your@email.com" :disabled="sent" />
        </div>
        <Button v-if="!sent" class="w-full" :disabled="loading" @click="handleSubmit">
          {{ loading ? '发送中...' : '发送重置链接' }}
        </Button>
      </CardContent>
      <CardFooter class="justify-center">
        <router-link to="/login" class="text-sm text-primary hover:underline">返回登录</router-link>
      </CardFooter>
    </Card>
  </div>
</template>
