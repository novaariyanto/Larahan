<script setup lang="ts">
import { useToast } from '../composables/useToast'

const { toasts } = useToast()
</script>

<template>
  <div class="pointer-events-none fixed right-4 bottom-4 z-50 flex w-80 flex-col gap-2">
    <div
      v-for="toast in toasts"
      :key="toast.id"
      class="pointer-events-auto rounded-lg border px-4 py-3 text-sm shadow-lg"
      :class="{
        'border-slate-200 bg-white text-slate-800': toast.kind === 'info',
        'border-emerald-200 bg-emerald-50 text-emerald-800': toast.kind === 'success',
        'border-rose-200 bg-rose-50 text-rose-800': toast.kind === 'error',
      }"
    >
      {{ toast.message }}
    </div>
  </div>
</template>
