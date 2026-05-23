<script setup lang="ts">
interface Field {
  name: string
  type: string
  required?: string
  desc: string
}

withDefaults(defineProps<{
  fields: Field[]
  location?: string
  nameLabel?: string
  typeLabel?: string
}>(), {
  location: '-',
  nameLabel: '字段',
  typeLabel: '类型',
})
</script>

<template>
  <div class="overflow-x-auto rounded-xl border border-white/10 bg-[#09090d]">
    <table class="w-full min-w-[720px] text-sm">
      <thead class="bg-white/5">
        <tr>
          <th class="w-28 px-3 py-2 text-left font-medium">位置</th>
          <th class="w-44 px-3 py-2 text-left font-medium">{{ nameLabel }}</th>
          <th class="w-32 px-3 py-2 text-left font-medium">{{ typeLabel }}</th>
          <th class="w-20 px-3 py-2 text-left font-medium">必填</th>
          <th class="px-3 py-2 text-left font-medium">说明</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="field in fields" :key="field.name" class="border-t border-white/10">
          <td class="px-3 py-2 text-xs text-slate-400">{{ location }}</td>
          <td class="px-3 py-2 font-mono text-xs text-slate-100">{{ field.name }}</td>
          <td class="px-3 py-2 text-xs text-slate-200">{{ field.type }}</td>
          <td class="px-3 py-2 text-xs">{{ field.required || '-' }}</td>
          <td class="px-3 py-2 text-xs leading-5 text-slate-400">{{ field.desc }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
