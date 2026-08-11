<script setup lang="ts">
defineProps<{
  open: boolean
  title: string
  stage?: string
  percent?: number
}>()

defineEmits<{ close: [] }>()
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/40 p-4"
    @click.self="$emit('close')"
  >
    <div class="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
      <h3 class="text-lg font-semibold text-slate-900">{{ title }}</h3>
      <p class="mt-1 text-sm text-slate-500">{{ stage || 'Menyiapkan…' }}</p>

      <div class="mt-5 h-2 overflow-hidden rounded-full bg-slate-100">
        <div
          class="h-full rounded-full bg-teal-500 transition-all duration-300"
          :style="{ width: `${Math.min(Math.max(percent ?? 0, 0), 100)}%` }"
        />
      </div>
      <p class="mt-2 text-right text-xs font-medium text-slate-500">{{ percent ?? 0 }}%</p>
    </div>
  </div>
</template>
