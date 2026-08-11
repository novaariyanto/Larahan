<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import StatusCard from '../components/StatusCard.vue'
import { useToast } from '../composables/useToast'
import { onInstallDone } from '../composables/useInstallProgress'
import { usePolling } from '../composables/usePolling'
import { PhpMyAdminService } from '../api/services'
import type { ComponentInfo } from '../types'

const toast = useToast()
const busy = ref(false)
const url = ref('')
const info = ref<ComponentInfo>({
  type: 'phpmyadmin',
  version: '',
  status: 'not_installed',
  path: '',
  installed: false,
})
let offDone: (() => void) | undefined

const canOpen = computed(() => info.value.installed)

async function load() {
  try {
    info.value = (await PhpMyAdminService.GetInfo()) as ComponentInfo
    url.value = await PhpMyAdminService.GetURL()
  } catch (err) {
    console.warn(err)
  }
}

async function call(method: 'Install' | 'Uninstall' | 'Open') {
  if (busy.value) return
  busy.value = true
  try {
    const result = await PhpMyAdminService[method]()
    if (result.ok) {
      if (method === 'Install') toast.info(result.message)
      else toast.success(result.message)
    } else {
      toast.error(result.message)
    }
    if (method !== 'Install') await load()
  } catch (err) {
    toast.error(String(err))
  } finally {
    busy.value = false
  }
}

usePolling(load, 3000)

onMounted(() => {
  offDone = onInstallDone((type) => {
    if (type === 'phpmyadmin') load()
  })
})

onUnmounted(() => {
  offDone?.()
})
</script>

<template>
  <section class="space-y-6">
    <header>
      <h2 class="font-display text-2xl font-semibold text-slate-900">phpMyAdmin</h2>
      <p class="mt-1 text-sm text-slate-500">
        UI web untuk MySQL lokal — terintegrasi via Alias Apache
        <code class="rounded bg-slate-100 px-1 py-0.5 text-xs">/phpmyadmin</code>.
      </p>
    </header>

    <StatusCard
      title="phpMyAdmin"
      :status="info.status"
      :version="info.version"
      :path="info.path"
      :installed="info.installed"
      :detail="url || undefined"
    >
      <template #actions>
        <button class="btn-primary" type="button" :disabled="busy" @click="call('Install')">
          Install
        </button>
        <button
          class="btn-secondary"
          type="button"
          :disabled="busy || !info.installed"
          @click="call('Uninstall')"
        >
          Uninstall
        </button>
        <button
          class="btn-secondary"
          type="button"
          :disabled="busy || !canOpen"
          @click="call('Open')"
        >
          Buka di Browser
        </button>
      </template>
    </StatusCard>

    <div class="rounded-xl border border-slate-200 bg-white p-5 text-sm text-slate-600 shadow-sm">
      <h3 class="font-semibold text-slate-800">Prasyarat & integrasi</h3>
      <ul class="mt-2 list-disc space-y-1 pl-5">
        <li>Apache harus terinstal (modul PHP aktif).</li>
        <li>MySQL sebaiknya running — login otomatis sebagai <code>root</code> tanpa password.</li>
        <li>
          URL:
          <span class="font-mono text-xs text-slate-800">{{ url || 'http://127.0.0.1/phpmyadmin/' }}</span>
        </li>
      </ul>
    </div>
  </section>
</template>
