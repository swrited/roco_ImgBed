<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  data: { date: string; count: number }[]
  height?: number
  showLabels?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  height: 160,
  showLabels: false,
})

const chartHeight = computed(() => props.height)
const chartViewBox = computed(() => `0 0 100 ${chartHeight.value}`)
const padding = { top: 10, right: 4, bottom: 22, left: 4 }
const innerHeight = computed(() => chartHeight.value - padding.top - padding.bottom)

const points = computed(() => {
  const counts = props.data.map((d) => d.count)
  const max = Math.max(...counts, 1)
  const stepX = 100 / Math.max(props.data.length - 1, 1)

  return counts.map((count, i) => ({
    x: i * stepX,
    y: innerHeight.value - (count / max) * innerHeight.value,
    count,
  }))
})

const polylinePoints = computed(() =>
  points.value.map((p) => `${p.x},${p.y}`).join(' ')
)

const areaPoints = computed(() => {
  if (points.value.length === 0) return ''
  const first = points.value[0]
  const last = points.value[points.value.length - 1]
  return `${first.x},${innerHeight.value} ${polylinePoints.value} ${last.x},${innerHeight.value}`
})

const maxCount = computed(() => Math.max(...props.data.map((d) => d.count), 0))
const total = computed(() => props.data.reduce((sum, d) => sum + d.count, 0))
const avg = computed(() => (props.data.length > 0 ? Math.round(total.value / props.data.length) : 0))

const maxDate = computed(() => {
  if (props.data.length === 0) return ''
  let maxVal = 0
  let maxDate = ''
  for (const d of props.data) {
    if (d.count >= maxVal) {
      maxVal = d.count
      maxDate = d.date
    }
  }
  return maxDate.split('-').length === 3 ? maxDate.slice(5) : maxDate
})

// Labels: show first date, first day of each month, last date
const labelIndices = computed(() => {
  const indices = new Set<number>()
  if (props.data.length === 0) return indices
  indices.add(0)
  indices.add(props.data.length - 1)
  for (let i = 1; i < props.data.length - 1; i++) {
    const day = props.data[i].date.split('-')[2]
    if (day === '01') indices.add(i)
  }
  return indices
})

function formatLabel(date: string): string {
  const parts = date.split('-')
  if (parts.length >= 3) return `${parts[1]}-${parts[2]}`
  return date
}
</script>

<template>
  <div class="w-full">
    <div class="flex items-center gap-4 mb-1">
      <div class="text-xs text-muted-foreground">
        总上传 <span class="text-foreground font-semibold">{{ total }}</span>
      </div>
      <div class="text-xs text-muted-foreground">
        日均 <span class="text-foreground font-semibold">{{ avg }}</span>
      </div>
      <div v-if="maxCount > 0" class="text-xs text-muted-foreground">
        峰值 <span class="text-emerald-400 font-semibold">{{ maxCount }}</span>
        <span class="ml-0.5">({{ maxDate }})</span>
      </div>
    </div>
    <svg
      :viewBox="chartViewBox"
      preserveAspectRatio="none"
      class="w-full overflow-visible"
      :style="{ height: chartHeight + 'px' }"
    >
      <!-- Grid lines -->
      <line
        v-for="y in [0, 0.25, 0.5, 0.75]"
        :key="y"
        x1="0" :x2="100"
        :y1="padding.top + innerHeight * (1 - y)"
        :y2="padding.top + innerHeight * (1 - y)"
        stroke="rgba(255,255,255,0.04)"
        stroke-width="0.5"
      />

      <!-- Area fill -->
      <polygon
        v-if="points.length > 1"
        :points="areaPoints"
        fill="url(#areaGradient)"
        class="transition-all duration-300"
      />

      <!-- Line -->
      <polyline
        v-if="points.length > 1"
        :points="polylinePoints"
        fill="none"
        stroke="#8b5cf6"
        stroke-width="1.8"
        class="transition-all duration-300"
        vector-effect="non-scaling-stroke"
      />

      <!-- Dots -->
      <circle
        v-for="(p, i) in points"
        :key="i"
        :cx="p.x"
        :cy="p.y"
        r="1.2"
        fill="#8b5cf6"
        class="transition-all duration-300"
      />

      <!-- Date labels -->
      <text
        v-for="i in labelIndices"
        :key="'l' + i"
        :x="i === 0 ? padding.left : i === data.length - 1 ? 100 - padding.right : (i / Math.max(data.length - 1, 1)) * 100"
        :y="chartHeight - 2"
        text-anchor="middle"
        class="fill-muted-foreground select-none"
        style="font-size: 2.5px"
      >
        {{ formatLabel(data[i].date) }}
      </text>

      <defs>
        <linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stop-color="#8b5cf6" stop-opacity="0.25" />
          <stop offset="100%" stop-color="#8b5cf6" stop-opacity="0.0" />
        </linearGradient>
      </defs>
    </svg>
  </div>
</template>
