<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useToast } from '../composables/useToast'
import { LibraryService } from '../api/services'
import type { LibrarySummary, PHPExtension } from '../types'

const toast = useToast()
const loading = ref(true)
const busy = ref<string | null>(null)
const query = ref('')
const summary = ref<LibrarySummary>({
  installed: false,
  php_version: '',
  php_path: '',
  apache_running: false,
  message: '',
  extensions: [],
})

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  const list = summary.value.extensions || []
  if (!q) return list
  return list.filter(
    (e) =>
      e.display_name.toLowerCase().includes(q) ||
      e.name.toLowerCase().includes(q) ||
      (e.dll || '').toLowerCase().includes(q),
  )
})

const counts = computed(() => {
  const list = summary.value.extensions || []
  return {
    enabled: list.filter((e) => e.status === 'enabled').length,
    disabled: list.filter((e) => e.status === 'disabled').length,
    missing: list.filter((e) => e.status === 'not_installed').length,
  }
})

async function load() {
  loading.value = true
  try {
    summary.value = (await LibraryService.GetSummary()) as LibrarySummary
  } catch (err) {
    toast.error(String(err))
  } finally {
    loading.value = false
  }
}

async function toggle(ext: PHPExtension) {
  if (!ext.toggleable || busy.value) return
  if (ext.status === 'not_installed') {
    toast.error(`${ext.display_name} belum terinstal. Tidak dapat diaktifkan sebelum DLL tersedia.`)
    return
  }
  busy.value = ext.name
  try {
    const enable = ext.status !== 'enabled'
    const result = enable
      ? await LibraryService.Enable(ext.name)
      : await LibraryService.Disable(ext.name)
    if (result.ok) toast.success(result.message)
    else toast.error(result.message)
    await load()
  } catch (err) {
    toast.error(String(err))
  } finally {
    busy.value = null
  }
}

function statusLabel(status: string): string {
  switch (status) {
    case 'enabled':
      return 'Enabled'
    case 'disabled':
      return 'Disabled'
    default:
      return 'Not Installed'
  }
}

onMounted(load)
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-wrap items-end justify-between gap-4">
      <div>
        <h2 class="font-display text-2xl font-semibold text-slate-900">Library</h2>
        <p class="mt-1 text-sm text-slate-500">
          Kelola extension PHP aktif tanpa mengedit php.ini secara manual.
        </p>
      </div>
      <button class="btn-secondary" type="button" :disabled="loading" @click="load">
        Refresh
      </button>
    </header>

    <div class="grid gap-3 sm:grid-cols-3">
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3">
        <p class="text-xs font-medium tracking-wide text-slate-500 uppercase">PHP Version</p>
        <p class="mt-1 text-lg font-semibold text-slate-900">{{ summary.php_version || '—' }}</p>
        <p class="mt-1 truncate font-mono text-xs text-slate-500" :title="summary.php_path">
          {{ summary.php_path || 'PHP belum aktif' }}
        </p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3">
        <p class="text-xs font-medium tracking-wide text-slate-500 uppercase">Extensions</p>
        <p class="mt-1 text-lg font-semibold text-slate-900">
          {{ counts.enabled }} enabled · {{ counts.disabled }} disabled
        </p>
        <p class="mt-1 text-xs text-slate-500">{{ counts.missing }} not installed</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white px-4 py-3">
        <p class="text-xs font-medium tracking-wide text-slate-500 uppercase">Web Server</p>
        <p class="mt-1 text-sm font-medium text-slate-800">
          {{ summary.apache_running ? 'Apache running — perubahan akan me-restart Apache' : 'Apache stopped — Start Apache agar berlaku di web' }}
        </p>
      </div>
    </div>

    <p
      v-if="summary.message"
      class="rounded-lg border border-amber-100 bg-amber-50 px-3 py-2 text-sm text-amber-800"
    >
      {{ summary.message }}
    </p>

    <div class="flex items-center gap-3">
      <input
        v-model="query"
        class="input max-w-sm"
        type="search"
        placeholder="Cari library (cURL, OpenSSL, GD…)"
      />
    </div>

    <p v-if="loading" class="text-sm text-slate-500">Memuat daftar library…</p>

    <div v-else class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <table class="min-w-full text-left text-sm">
        <thead class="bg-slate-50 text-xs tracking-wide text-slate-500 uppercase">
          <tr>
            <th class="px-4 py-3 font-medium">Library</th>
            <th class="px-4 py-3 font-medium">Status</th>
            <th class="px-4 py-3 font-medium">Version</th>
            <th class="px-4 py-3 font-medium">DLL</th>
            <th class="px-4 py-3 font-medium">Enable</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="ext in filtered" :key="ext.name" class="border-t border-slate-100">
            <td class="px-4 py-3">
              <p class="font-medium text-slate-800">{{ ext.display_name }}</p>
              <p class="text-xs text-slate-500">{{ ext.name }}{{ ext.zend ? ' · zend' : '' }}{{ ext.builtin ? ' · builtin' : '' }}</p>
            </td>
            <td class="px-4 py-3">
              <span
                class="rounded-md px-2 py-1 text-xs font-medium"
                :class="{
                  'bg-emerald-50 text-emerald-700': ext.status === 'enabled',
                  'bg-amber-50 text-amber-700': ext.status === 'disabled',
                  'bg-slate-100 text-slate-600': ext.status === 'not_installed',
                }"
              >
                {{ statusLabel(ext.status) }}
              </span>
            </td>
            <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ ext.version || '—' }}</td>
            <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ ext.dll || '—' }}</td>
            <td class="px-4 py-3">
              <button
                type="button"
                class="relative h-6 w-11 rounded-full transition"
                :class="ext.status === 'enabled' ? 'bg-teal-600' : 'bg-slate-300'"
                :disabled="!ext.toggleable || busy === ext.name"
                :title="
                  ext.status === 'not_installed'
                    ? 'Belum terinstal'
                    : ext.builtin
                      ? 'Extension bawaan'
                      : ext.status === 'enabled'
                        ? 'Disable'
                        : 'Enable'
                "
                @click="toggle(ext)"
              >
                <span
                  class="absolute top-0.5 left-0.5 h-5 w-5 rounded-full bg-white shadow transition"
                  :class="ext.status === 'enabled' ? 'translate-x-5' : ''"
                />
              </button>
            </td>
          </tr>
          <tr v-if="filtered.length === 0">
            <td colspan="5" class="px-4 py-6 text-center text-sm text-slate-500">
              Tidak ada library yang cocok.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
