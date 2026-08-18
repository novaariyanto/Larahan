<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const items = [
  { name: 'Dashboard', path: '/', icon: 'dashboard' },
  { name: 'Apache', path: '/apache', icon: 'apache' },
  { name: 'PHP', path: '/php', icon: 'php' },
  { name: 'Library', path: '/library', icon: 'library' },
  { name: 'MySQL', path: '/mysql', icon: 'mysql' },
  { name: 'phpMyAdmin', path: '/phpmyadmin', icon: 'phpmyadmin' },
  { name: 'Settings', path: '/settings', icon: 'settings' },
] as const

const activePath = computed(() => route.path)

function go(path: string) {
  router.push(path)
}
</script>

<template>
  <aside class="flex h-full w-56 shrink-0 flex-col border-r border-slate-200/80 bg-slate-950 text-slate-100">
    <div class="border-b border-white/10 px-5 py-5">
      <p class="text-xs font-medium tracking-[0.18em] text-teal-300/90 uppercase">Local Stack</p>
      <h1 class="mt-1 font-display text-2xl font-semibold tracking-tight text-white">Larahan</h1>
    </div>

    <nav class="flex flex-1 flex-col gap-1 overflow-y-auto p-3">
      <button
        v-for="item in items"
        :key="item.path"
        type="button"
        class="rounded-lg px-3 py-2.5 text-left text-sm font-medium transition"
        :class="
          activePath === item.path
            ? 'bg-teal-500/20 text-teal-200 ring-1 ring-teal-400/30'
            : 'text-slate-300 hover:bg-white/5 hover:text-white'
        "
        @click="go(item.path)"
      >
        {{ item.name }}
      </button>
    </nav>

    <div class="border-t border-white/10 px-4 py-3 text-xs text-slate-500">
      v0.1.0 MVP
    </div>
  </aside>
</template>
