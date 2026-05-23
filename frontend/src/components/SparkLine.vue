<script setup lang="ts">
import { computed, ref } from 'vue'

interface Props {
  data: { date: string; count: number }[]
  height?: number
}

const props = withDefaults(defineProps<Props>(), { height: 160 })

const chartHeight = computed(() => props.height)
const padding = { top: 18, right: 8, bottom: 24, left: 8 }
const innerWidth = computed(() => 100 - padding.left - padding.right)
const innerHeight = computed(() => chartHeight.value - padding.top - padding.bottom)
const gradientId = `spark-area-${Math.random().toString(36).slice(2)}`

const points = computed(() => {
  const counts = props.data.map((d) => d.count)
  const max = Math.max(...counts, 1)
  const stepX = innerWidth.value / Math.max(props.data.length - 1, 1)
  return counts.map((count, i) => ({
    x: padding.left + i * stepX,
    y: padding.top + innerHeight.value - (count / max) * innerHeight.value,
    count,
  }))
})

const polylinePoints = computed(() => points.value.map((p) => `${p.x},${p.y}`).join(' '))
const linePath = computed(() => {
  if (points.value.length === 0) return ''
  if (points.value.length === 1) {
    const p = points.value[0]!
    return `M ${p.x} ${p.y}`
  }

  const [first, ...rest] = points.value
  const commands = [`M ${first!.x} ${first!.y}`]
  rest.forEach((point, index) => {
    const prev = points.value[index]!
    const midX = (prev.x + point.x) / 2
    commands.push(`Q ${prev.x} ${prev.y} ${midX} ${(prev.y + point.y) / 2}`)
    commands.push(`T ${point.x} ${point.y}`)
  })
  return commands.join(' ')
})

const areaPoints = computed(() => {
  if (points.value.length === 0) return ''
  const first = points.value[0]!
  const last = points.value[points.value.length - 1]!
  return `${first.x},${chartHeight.value - padding.bottom} ${polylinePoints.value} ${last.x},${chartHeight.value - padding.bottom}`
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
    const parts = props.data[i]!.date.split('-')
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
  const stepX = innerWidth.value / Math.max(props.data.length - 1, 1)
  const idx = Math.round((relX - padding.left) / stepX)
  hoverIndex.value = Math.max(0, Math.min(idx, props.data.length - 1))
}

function onMouseLeave() {
  hoverIndex.value = -1
}

const hoverPoint = computed(() => {
  if (hoverIndex.value < 0 || hoverIndex.value >= points.value.length) return null
  const p = points.value[hoverIndex.value]!
  const d = props.data[hoverIndex.value]!
  return { x: p.x, y: p.y, date: formatLabel(d.date), count: d.count }
})

const tooltipStyle = computed(() => {
  if (!hoverPoint.value) return {}
  const x = hoverPoint.value.x
  const horizontal = x > 78 ? 'translateX(-100%)' : x < 22 ? 'translateX(0)' : 'translateX(-50%)'
  const left = x > 78 ? `calc(${x}% - 10px)` : x < 22 ? `calc(${x}% + 10px)` : `${x}%`
  return {
    left,
    top: `${Math.max(8, hoverPoint.value.y - 42)}px`,
    transform: horizontal,
  }
})

function labelForIndex(index: number): string {
  const item = props.data[index]
  return item ? formatLabel(item.date) : ''
}
</script>

<template>
  <div class="w-full">
    <div class="mb-4 flex flex-wrap items-center gap-2">
      <div class="rounded-full border border-white/10 bg-white/[0.03] px-3 py-1 text-xs text-muted-foreground">
        总计 <span class="text-foreground font-semibold">{{ total }}</span>
      </div>
      <div class="rounded-full border border-white/10 bg-white/[0.03] px-3 py-1 text-xs text-muted-foreground">
        日均 <span class="text-foreground font-semibold">{{ avg }}</span>
      </div>
      <div v-if="maxCount > 0" class="rounded-full border border-emerald-400/20 bg-emerald-400/10 px-3 py-1 text-xs text-muted-foreground">
        峰值 <span class="text-emerald-400 font-semibold">{{ maxCount }}</span>
        <span class="ml-0.5">({{ maxDate }})</span>
      </div>
    </div>

    <!-- Chart SVG -->
    <div
      ref="svgContainer"
      class="relative overflow-visible rounded-2xl border border-white/10 bg-[#09090d] px-2"
      :style="{ height: chartHeight + 'px' }"
      @mousemove="onMouseMove"
      @mouseleave="onMouseLeave"
    >
      <svg
        :viewBox="`0 0 100 ${chartHeight}`"
        preserveAspectRatio="none"
        class="absolute inset-0 h-full w-full"
      >
        <!-- Grid lines -->
        <line
          v-for="y in [0, 0.25, 0.5, 0.75]"
          :key="y"
          :x1="padding.left"
          :x2="100 - padding.right"
          :y1="padding.top + innerHeight * (1 - y)"
          :y2="padding.top + innerHeight * (1 - y)"
          stroke="rgba(255,255,255,0.06)"
          stroke-width="0.5"
        />

        <!-- Area fill -->
        <polygon
          v-if="points.length > 1"
          :points="areaPoints"
          :fill="`url(#${gradientId})`"
        />

        <!-- Line -->
        <path
          v-if="points.length > 1"
          :d="linePath"
          fill="none"
          stroke="#a78bfa"
          stroke-width="2.4"
          stroke-linecap="round"
          stroke-linejoin="round"
          vector-effect="non-scaling-stroke"
        />

        <!-- Hover vertical line -->
        <line
          v-if="hoverPoint"
          :x1="hoverPoint.x" :x2="hoverPoint.x"
          :y1="padding.top"
          :y2="chartHeight - padding.bottom"
          stroke="rgba(167,139,250,0.42)"
          stroke-width="0.5"
          stroke-dasharray="3 3"
        />

        <defs>
          <linearGradient :id="gradientId" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#a78bfa" stop-opacity="0.32" />
            <stop offset="72%" stop-color="#8b5cf6" stop-opacity="0.08" />
            <stop offset="100%" stop-color="#8b5cf6" stop-opacity="0" />
          </linearGradient>
        </defs>
      </svg>

      <!-- HTML dots (not stretched by preserveAspectRatio) -->
      <div
        v-for="(p, i) in points"
        :key="'d' + i"
        class="pointer-events-none absolute rounded-full bg-purple-300/80 ring-2 ring-[#09090d]"
        :style="{
          width: i === hoverIndex ? '7px' : '4px',
          height: i === hoverIndex ? '7px' : '4px',
          left: p.x + '%',
          top: p.y + 'px',
          transform: 'translate(-50%, -50%)',
          opacity: i === hoverIndex || p.count > 0 ? 1 : 0.45,
        }"
      />

      <!-- Hover dot -->
      <div
        v-if="hoverPoint"
        class="pointer-events-none absolute rounded-full"
        :style="{
          width: '12px',
          height: '12px',
          left: hoverPoint.x + '%',
          top: hoverPoint.y + 'px',
          transform: 'translate(-50%, -50%)',
          backgroundColor: '#a78bfa',
          border: '2px solid #fff',
          zIndex: 10,
        }"
      />

      <!-- Hover tooltip -->
      <div
        v-if="hoverPoint"
        class="pointer-events-none absolute z-20 whitespace-nowrap rounded-xl border border-purple-400/30 bg-[#151520] px-3 py-2 text-xs text-slate-100 shadow-xl shadow-black/30"
        :style="tooltipStyle"
      >
        <span class="text-slate-400">{{ hoverPoint.date }}</span>
        <span class="mx-1.5 text-slate-600">/</span>
        <span class="font-semibold text-purple-200">{{ hoverPoint.count }} 张</span>
      </div>
    </div>

    <!-- Date labels below chart -->
    <div v-if="data.length > 0" class="relative mt-1 h-5 w-full">
      <span
        v-for="i in labelPositions"
        :key="i"
        class="absolute -translate-x-1/2 text-[10px] text-muted-foreground select-none whitespace-nowrap"
        :style="{ left: padding.left + (i / Math.max(data.length - 1, 1)) * innerWidth + '%' }"
      >
        {{ labelForIndex(i) }}
      </span>
    </div>
  </div>
</template>
