<script setup lang="ts">
import AlbumsView from './AlbumsView.vue'
</script>

<template>
  <div class="space-y-6">
    <div>
      <p class="text-xs font-semibold uppercase tracking-widest text-violet-400">Library</p>
      <h1 class="mt-1.5 text-3xl font-semibold tracking-tight">我的图库</h1>
      <p class="mt-2 text-sm text-muted-foreground">选择相册文件夹查看图片，未指定相册的上传会进入未分类图片。</p>
    </div>

    <div class="rounded-2xl border border-white/5 bg-[oklch(15%_0.01_270)] p-6 shadow-2xl shadow-black/20">
      <AlbumsView />
    </div>
  </div>
</template>
