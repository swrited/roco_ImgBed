<script setup lang="ts">
import { RouterView } from 'vue-router'
import { useRouter } from 'vue-router'
import { ArrowLeft } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

const router = useRouter()
</script>

<template>
  <div class="flex min-w-0 flex-col gap-6">
    <header class="flex h-12 items-center gap-3 border-b border-white/10">
      <Button variant="ghost" size="icon" @click="router.push('/')">
        <ArrowLeft class="h-5 w-5" />
      </Button>
      <h1 class="font-semibold">管理后台</h1>
    </header>
    <main class="min-w-0">
      <RouterView />
    </main>
  </div>
</template>
