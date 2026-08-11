<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useToast } from '../composables/useToast'
import { SettingsService } from '../api/services'
import type { AppInfo, Paths, Settings } from '../types'

const toast = useToast()
const saving = ref(false)
const paths = ref<Paths | null>(null)
const appInfo = ref<AppInfo>({
  name: 'Larahan',
  version: '0.1.0',
  description: '',
})

const form = reactive<Settings>({
  install_path: 'C:\\Larahan',
  active_php: '',
  apache_port: 80,
  mysql_port: 3306,
  first_run: true,
})

async function load() {
  try {
    const settings = (await SettingsService.Get()) as Settings
    Object.assign(form, settings)
    paths.value = (await SettingsService.GetPaths()) as Paths
    appInfo.value = (await SettingsService.GetAppInfo()) as AppInfo
  } catch (err) {
    console.warn(err)
  }
}

async function save() {
  if (saving.value) return
  saving.value = true
  try {
    const result = await SettingsService.Save({ ...form })
    if (result.ok) toast.success(result.message)
    else toast.error(result.message)
    await load()
  } catch (err) {
    toast.error(String(err))
  } finally {
    saving.value = false
  }
}

async function openDir(kind: string) {
  try {
    const result = await SettingsService.OpenDirectory(kind)
    if (!result.ok) toast.error(result.message)
  } catch (err) {
    toast.error(String(err))
  }
}

onMounted(load)
</script>

<template>
  <section class="space-y-6">
    <header>
      <h2 class="font-display text-2xl font-semibold text-slate-900">Settings</h2>
      <p class="mt-1 text-sm text-slate-500">Konfigurasi runtime dan informasi aplikasi.</p>
    </header>

    <div class="max-w-xl space-y-4 rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <h3 class="text-sm font-semibold text-slate-800">Runtime</h3>

      <label class="block text-sm">
        <span class="mb-1 block font-medium text-slate-700">Install Path</span>
        <input v-model="form.install_path" class="input bg-slate-50" type="text" readonly />
        <span class="mt-1 block text-xs text-slate-500">Tetap di C:\Larahan pada MVP ini.</span>
      </label>

      <div class="grid grid-cols-2 gap-4">
        <label class="block text-sm">
          <span class="mb-1 block font-medium text-slate-700">Apache Port</span>
          <input v-model.number="form.apache_port" class="input" type="number" min="1" max="65535" />
        </label>
        <label class="block text-sm">
          <span class="mb-1 block font-medium text-slate-700">MySQL Port</span>
          <input v-model.number="form.mysql_port" class="input" type="number" min="1" max="65535" />
        </label>
      </div>

      <label class="block text-sm">
        <span class="mb-1 block font-medium text-slate-700">Active PHP</span>
        <input v-model="form.active_php" class="input bg-slate-50" type="text" readonly placeholder="—" />
        <span class="mt-1 block text-xs text-slate-500">Ubah lewat menu PHP → Switch.</span>
      </label>

      <p class="text-xs text-slate-500">
        Jika service sedang running, perubahan port akan diterapkan dan service di-restart otomatis.
      </p>

      <button class="btn-primary" type="button" :disabled="saving" @click="save">
        {{ saving ? 'Menyimpan…' : 'Simpan' }}
      </button>
    </div>

    <div v-if="paths" class="max-w-xl rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <h3 class="text-sm font-semibold text-slate-800">Directory Layout</h3>
      <dl class="mt-3 space-y-2 text-sm">
        <div
          v-for="(value, key) in paths"
          :key="key"
          class="flex items-center justify-between gap-3 border-b border-slate-50 py-2 last:border-0"
        >
          <div class="min-w-0">
            <dt class="text-xs font-medium tracking-wide text-slate-500 uppercase">{{ key }}</dt>
            <dd class="truncate font-mono text-xs text-slate-700" :title="value">{{ value }}</dd>
          </div>
          <button class="btn-secondary shrink-0" type="button" @click="openDir(String(key))">
            Open
          </button>
        </div>
      </dl>
    </div>

    <div class="max-w-xl rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
      <h3 class="text-sm font-semibold text-slate-800">About</h3>
      <dl class="mt-3 space-y-2 text-sm">
        <div class="flex justify-between gap-4">
          <dt class="text-slate-500">Name</dt>
          <dd class="font-medium text-slate-800">{{ appInfo.name }}</dd>
        </div>
        <div class="flex justify-between gap-4">
          <dt class="text-slate-500">Version</dt>
          <dd class="font-medium text-slate-800">{{ appInfo.version }}</dd>
        </div>
        <div class="flex justify-between gap-4">
          <dt class="text-slate-500">Description</dt>
          <dd class="text-right text-slate-700">{{ appInfo.description }}</dd>
        </div>
      </dl>
    </div>
  </section>
</template>
