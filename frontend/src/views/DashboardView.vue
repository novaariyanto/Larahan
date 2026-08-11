<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import StatusCard from '../components/StatusCard.vue'
import { useToast } from '../composables/useToast'
import { onInstallDone } from '../composables/useInstallProgress'
import { usePolling } from '../composables/usePolling'
import { DashboardService, SettingsService } from '../api/services'
import type { DashboardSummary, Settings } from '../types'

const router = useRouter()
const toast = useToast()
const busy = ref(false)
const loading = ref(true)
const hasLoaded = ref(false)
const lastUpdated = ref<Date | null>(null)
const settings = ref<Settings | null>(null)

const summary = ref<DashboardSummary>({
  apache: { type: 'apache', version: '', status: 'not_installed', path: '', installed: false },
  mysql: { type: 'mysql', version: '', status: 'not_installed', path: '', installed: false },
  php: { type: 'php', version: '', status: 'not_installed', path: '', installed: false },
  phpmyadmin: { type: 'phpmyadmin', version: '', status: 'not_installed', path: '', installed: false },
  active_php: '',
})

const installedCount = computed(() => {
  let n = 0
  if (summary.value.apache.installed) n++
  if (summary.value.php.installed || !!summary.value.active_php) n++
  if (summary.value.mysql.installed) n++
  if (summary.value.phpmyadmin.installed) n++
  return n
})

const runningCount = computed(() => {
  let n = 0
  if (summary.value.apache.status === 'running') n++
  if (summary.value.mysql.status === 'running') n++
  return n
})

const canControl = computed(() => summary.value.apache.installed || summary.value.mysql.installed)

const updatedLabel = computed(() => {
  if (!lastUpdated.value) return '—'
  return lastUpdated.value.toLocaleTimeString()
})

async function loadSummary(silent = false) {
  if (!silent) loading.value = true
  try {
    summary.value = (await DashboardService.GetSummary()) as DashboardSummary
    lastUpdated.value = new Date()
    hasLoaded.value = true
  } catch (err) {
    console.warn('DashboardService unavailable:', err)
  } finally {
    loading.value = false
  }
}

async function loadSettings() {
  try {
    settings.value = (await SettingsService.Get()) as Settings
  } catch (err) {
    console.warn(err)
  }
}

async function runAction(action: 'StartAll' | 'StopAll' | 'RestartAll') {
  if (busy.value) return
  busy.value = true
  try {
    const result = await DashboardService[action]()
    if (result.ok) toast.success(result.message)
    else toast.error(result.message)
    await loadSummary(true)
  } catch (err) {
    toast.error(String(err))
  } finally {
    busy.value = false
  }
}

usePolling(() => loadSummary(hasLoaded.value), 3000)

let offDone: (() => void) | undefined
onMounted(() => {
  void loadSettings()
  offDone = onInstallDone(() => {
    void loadSummary(true)
  })
})
onUnmounted(() => {
  offDone?.()
})
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h2 class="font-display text-2xl font-semibold text-slate-900">Dashboard</h2>
        <p class="mt-1 text-sm text-slate-500">
          Ringkasan environment lokal · diperbarui {{ updatedLabel }}
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        <button class="btn-secondary" type="button" :disabled="busy" @click="loadSummary()">
          Refresh
        </button>
        <button
          class="btn-primary"
          type="button"
          :disabled="busy || !canControl"
          @click="runAction('StartAll')"
        >
          {{ busy ? 'Working…' : 'Start' }}
        </button>
        <button
          class="btn-secondary"
          type="button"
          :disabled="busy || !canControl"
          @click="runAction('StopAll')"
        >
          Stop
        </button>
        <button
          class="btn-secondary"
          type="button"
          :disabled="busy || !canControl"
          @click="runAction('RestartAll')"
        >
          Restart
        </button>
      </div>
    </header>

    <div class="grid gap-3 sm:grid-cols-3">
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3">
        <p class="text-xs font-medium tracking-wide text-slate-500 uppercase">Installed</p>
        <p class="mt-1 text-2xl font-semibold text-slate-900">{{ installedCount }}/4</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3">
        <p class="text-xs font-medium tracking-wide text-slate-500 uppercase">Running</p>
        <p class="mt-1 text-2xl font-semibold text-slate-900">{{ runningCount }}</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3">
        <p class="text-xs font-medium tracking-wide text-slate-500 uppercase">Ports</p>
        <p class="mt-1 text-sm font-medium text-slate-800">
          Apache {{ settings?.apache_port ?? 80 }} · MySQL {{ settings?.mysql_port ?? 3306 }}
        </p>
        <p class="mt-1 text-xs text-slate-500">
          PHP {{ summary.active_php || summary.php.version || '—' }}
        </p>
      </div>
    </div>

    <p v-if="loading && !hasLoaded" class="text-sm text-slate-500">Memuat status…</p>

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <StatusCard
        title="Apache"
        :status="summary.apache.status"
        :version="summary.apache.version"
        :path="summary.apache.path"
        :installed="summary.apache.installed"
      >
        <template #actions>
          <button class="btn-secondary" type="button" @click="router.push('/apache')">
            Kelola
          </button>
        </template>
      </StatusCard>

      <StatusCard
        title="PHP"
        :status="summary.php.status"
        :version="summary.active_php || summary.php.version"
        :path="summary.php.path"
        :installed="summary.php.installed || !!summary.active_php"
        detail="Versi aktif"
      >
        <template #actions>
          <button class="btn-secondary" type="button" @click="router.push('/php')">
            Kelola
          </button>
        </template>
      </StatusCard>

      <StatusCard
        title="MySQL"
        :status="summary.mysql.status"
        :version="summary.mysql.version"
        :path="summary.mysql.path"
        :installed="summary.mysql.installed"
      >
        <template #actions>
          <button class="btn-secondary" type="button" @click="router.push('/mysql')">
            Kelola
          </button>
        </template>
      </StatusCard>

      <StatusCard
        title="phpMyAdmin"
        :status="summary.phpmyadmin.status"
        :version="summary.phpmyadmin.version"
        :path="summary.phpmyadmin.path"
        :installed="summary.phpmyadmin.installed"
        detail="UI MySQL"
      >
        <template #actions>
          <button class="btn-secondary" type="button" @click="router.push('/phpmyadmin')">
            Kelola
          </button>
        </template>
      </StatusCard>
    </div>
  </section>
</template>
