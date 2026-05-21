<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  data: { date: string; count: number }[]
  height?: number
}

const props = withDefaults(defineProps<Props>(), { height: 160 })

const chartHeight = computed(() => props.height)
const padding = { top: 10, right: 4, bottom: 12, left: 4 }
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

const polylinePoints = computed(() => points.value.map((p) => `${p.x},${p.y}`).join(' '))

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
  let maxVal = 0, maxDate = ''
  for (const d of props.data) {
    if (d.count >= maxVal) { maxVal = d.count; maxDate = d.date }
  }
  return maxDate.split('-').length === 3 ? maxDate.slice(5) : maxDate
})

// Label positions
const labelPositions = computed(() => {
  const indices: number[] = []
  if (props.data.length === 0) return indices
  indices.push(0)
  if (props.data.length > 1) indices.push(props.data.length - 1)
  for (let i = 1; i < props.data.length - 1; i++) {
    const parts = props.data[i].date.split('-')
    const day = parts.length >= 3 ? parts[2] : parts[1]
    if (day === '01') indices.push(i)
  }
  return indices.sort((a, b) => a - b)
})

function formatLabel(date: string): string {
  const parts = date.split('-')
  if (parts.length >= 3) return `${parts[1]}-${parts[2]}`
  return date
}

// Hover interaction
const hoverIndex = ref(-1)
const svgContainer = ref<HTMLElement>()

function onMouseMove(e: MouseEvent) {
  if (!svgContainer.value || props.data.length === 0) return
  const rect = svgContainer.value.getBoundingClientRect()
  const relX = ((e.clientX - rect.left) / rect.width) * 100
  const stepX = 100 / Math.max(props.data.length - 1, 1)
  const idx = Math.round(relX / stepX)
  hoverIndex.value = Math.max(0, Math.min(idx, props.data.length - 1))
}

function onMouseLeave() {
  hoverIndex.value = -1
}

const hoverPoint = computed(() => {
  if (hoverIndex.value < 0 || hoverIndex.value >= points.value.length) return null
  const p = points.value[hoverIndex.value]
  const d = props.data[hoverIndex.value]
  return { x: p.x, y: p.y, date: formatLabel(d.date), count: d.count }
})
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

    <!-- Chart SVG -->
    <div
      ref="svgContainer"
      class="relative"
      :style="{ height: chartHeight + 'px' }"
      @mousemove="onMouseMove"
      @mouseleave="onMouseLeave"
    >
      <svg
        :viewBox="`0 0 100 ${chartHeight}`"
        preserveAspectRatio="none"
        class="absolute inset-0 w-full h-full"
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
        />

        <!-- Line -->
        <polyline
          v-if="points.length > 1"
          :points="polylinePoints"
          fill="none"
          stroke="#8b5cf6"
          stroke-width="1.8"
          vector-effect="non-scaling-stroke"
        />

        <!-- Dots -->
        <circle
          v-for="(p, i) in points"
          :key="i"
          :cx="p.x" :cy="p.y"
          r="1.2"
          fill="#8b5cf6"
        />

        <!-- Hover vertical line -->
        <line
          v-if="hoverPoint"
          :x1="hoverPoint.x" :x2="hoverPoint.x"
          :y1="padding.top"
          :y2="chartHeight - padding.bottom"
          stroke="rgba(255,255,255,0.15)"
          stroke-width="0.5"
          stroke-dasharray="2 2"
        />

        <!-- Hover dot -->
        <circle
          v-if="hoverPoint"
          :cx="hoverPoint.x" :cy="hoverPoint.y"
          r="1.8"
          fill="#8b5cf6"
          stroke="#fff"
          stroke-width="0.5"
        />

        <defs>
          <linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#8b5cf6" stop-opacity="0.25" />
            <stop offset="100%" stop-color="#8b5cf6" stop-opacity="0.0" />
          </linearGradient>
        </defs>
      </svg>

      <!-- Hover tooltip -->
      <div
        v-if="hoverPoint"
        class="absolute pointer-events-none bg-popover text-popover-foreground text-xs rounded px-1.5 py-0.5 -translate-x-1/2 shadow border border-border whitespace-nowrap"
        :style="{
          left: hoverPoint.x + '%',
          top: Math.max(0, hoverPoint.y - 20) + 'px',
        }"
      >
        {{ hoverPoint.date }} — {{ hoverPoint.count }} 张
      </div>
    </div>

    <!-- Date labels below chart -->
    <div v-if="data.length > 0" class="relative h-5 w-full mt-0.5">
      <span
        v-for="i in labelPositions"
        :key="i"
        class="absolute -translate-x-1/2 text-[10px] text-muted-foreground select-none whitespace-nowrap"
        :style="{ left: (i / Math.max(data.length - 1, 1)) * 100 + '%' }"
      >
        {{ formatLabel(data[i].date) }}
      </span>
    </div>
  </div>
</template>
