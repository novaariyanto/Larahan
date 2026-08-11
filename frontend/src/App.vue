<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppSidebar from './components/AppSidebar.vue'
import ProgressModal from './components/ProgressModal.vue'
import Toast from './components/Toast.vue'
import { useInstallProgress } from './composables/useInstallProgress'

const { open, title, stage, percent } = useInstallProgress()
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-slate-50 text-slate-900">
    <AppSidebar />
    <main class="relative flex-1 overflow-y-auto">
      <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,_rgba(13,148,136,0.08),_transparent_45%),linear-gradient(180deg,#f8fafc_0%,#f1f5f9_100%)]" />
      <div class="relative mx-auto max-w-6xl px-6 py-8">
        <RouterView />
      </div>
    </main>
    <Toast />
    <ProgressModal :open="open" :title="title" :stage="stage" :percent="percent" />
  </div>
</template>
