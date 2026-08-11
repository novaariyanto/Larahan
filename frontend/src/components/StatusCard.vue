<script setup lang="ts">
import type { ComponentStatus } from '../types'
import { statusLabel } from '../types'

defineProps<{
  title: string
  status: ComponentStatus
  version?: string
  path?: string
  detail?: string
  installed?: boolean
}>()
</script>

<template>
  <article class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm transition hover:border-slate-300">
    <div class="flex items-start justify-between gap-3">
      <div>
        <h3 class="text-sm font-semibold text-slate-800">{{ title }}</h3>
        <p v-if="detail" class="mt-1 text-xs text-slate-500">{{ detail }}</p>
      </div>
      <span
        class="inline-flex items-center gap-1.5 rounded-md px-2 py-1 text-xs font-medium"
        :class="{
          'bg-emerald-50 text-emerald-700': status === 'running',
          'bg-teal-50 text-teal-700': status === 'ready',
          'bg-amber-50 text-amber-700': status === 'stopped',
          'bg-slate-100 text-slate-600': status === 'not_installed',
          'bg-rose-50 text-rose-700': status === 'error',
          'bg-sky-50 text-sky-700': status === 'installing',
        }"
      >
        <span
          class="h-1.5 w-1.5 rounded-full"
          :class="{
            'bg-emerald-500': status === 'running',
            'bg-teal-500': status === 'ready',
            'bg-amber-500': status === 'stopped',
            'bg-slate-400': status === 'not_installed',
            'bg-rose-500': status === 'error',
            'animate-pulse bg-sky-500': status === 'installing',
          }"
        />
        {{ statusLabel(status) }}
      </span>
    </div>

    <dl class="mt-4 space-y-2 text-sm">
      <div class="flex justify-between gap-4">
        <dt class="text-slate-500">Installed</dt>
        <dd class="font-medium" :class="installed ? 'text-emerald-700' : 'text-slate-500'">
          {{ installed ? 'Yes' : 'No' }}
        </dd>
      </div>
      <div class="flex justify-between gap-4">
        <dt class="text-slate-500">Version</dt>
        <dd class="font-medium text-slate-800">{{ version || '—' }}</dd>
      </div>
      <div v-if="path !== undefined" class="flex justify-between gap-4">
        <dt class="text-slate-500">Path</dt>
        <dd class="truncate text-right font-mono text-xs text-slate-700" :title="path">
          {{ path || '—' }}
        </dd>
      </div>
    </dl>

    <div v-if="$slots.actions" class="mt-4 flex flex-wrap gap-2 border-t border-slate-100 pt-4">
      <slot name="actions" />
    </div>
  </article>
</template>
