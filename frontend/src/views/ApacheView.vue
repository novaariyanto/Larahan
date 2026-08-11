<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import StatusCard from '../components/StatusCard.vue'
import { useToast } from '../composables/useToast'
import { onInstallDone } from '../composables/useInstallProgress'
import { usePolling } from '../composables/usePolling'
import { ApacheService } from '../api/services'
import type { ComponentInfo } from '../types'

const toast = useToast()
const busy = ref(false)
const info = ref<ComponentInfo>({
  type: 'apache',
  version: '',
  status: 'not_installed',
  path: '',
  installed: false,
})
let offDone: (() => void) | undefined

async function load() {
  try {
    info.value = (await ApacheService.GetInfo()) as ComponentInfo
  } catch (err) {
    console.warn(err)
  }
}

async function call(method: 'Install' | 'Uninstall' | 'Start' | 'Stop' | 'Restart') {
  if (busy.value) return
  busy.value = true
  try {
    const result = await ApacheService[method]()
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
    if (type === 'apache') load()
  })
})

onUnmounted(() => {
  offDone?.()
})
</script>

<template>
  <section class="space-y-6">
    <header>
      <h2 class="font-display text-2xl font-semibold text-slate-900">Apache</h2>
      <p class="mt-1 text-sm text-slate-500">Install, kontrol proses, dan lihat status Apache.</p>
    </header>

    <StatusCard
      title="Apache"
      :status="info.status"
      :version="info.version"
      :path="info.path"
      :installed="info.installed"
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
          :disabled="busy || !info.installed"
          @click="call('Start')"
        >
          Start
        </button>
        <button
          class="btn-secondary"
          type="button"
          :disabled="busy || !info.installed"
          @click="call('Stop')"
        >
          Stop
        </button>
        <button
          class="btn-secondary"
          type="button"
          :disabled="busy || !info.installed"
          @click="call('Restart')"
        >
          Restart
        </button>
      </template>
    </StatusCard>
  </section>
</template>
