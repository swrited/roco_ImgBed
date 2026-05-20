<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useRouter } from 'vue-router'
import { ArrowLeft } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const router = useRouter()
</script>

<template>
  <div class="flex flex-col h-full">
    <header class="flex h-14 items-center gap-4 border-b bg-card px-4">
      <Button variant="ghost" size="icon" @click="router.push('/')">
        <ArrowLeft class="h-5 w-5" />
      </Button>
      <h1 class="font-semibold">管理后台</h1>
    </header>
    <main class="flex-1 overflow-auto p-6">
      <RouterView />
    </main>
  </div>
</template>
