<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useToast } from '../composables/useToast'
import { onInstallDone } from '../composables/useInstallProgress'
import { PHPService } from '../api/services'
import type { PHPVersionInfo } from '../types'

const toast = useToast()
const versions = ref<PHPVersionInfo[]>([])
const active = ref('')
let offDone: (() => void) | undefined

async function load() {
  try {
    versions.value = (await PHPService.ListVersions()) as PHPVersionInfo[]
    active.value = await PHPService.GetActive()
  } catch (err) {
    console.warn(err)
  }
}

async function run(method: 'Install' | 'Delete' | 'Switch', version: string) {
  try {
    const result = await PHPService[method](version)
    if (result.ok) {
      if (method === 'Install') toast.info(result.message)
      else toast.success(result.message)
    } else {
      toast.info(result.message)
    }
    if (method !== 'Install') await load()
  } catch (err) {
    toast.error(String(err))
  }
}

onMounted(() => {
  load()
  offDone = onInstallDone((type) => {
    if (type === 'php') load()
  })
})

onUnmounted(() => {
  offDone?.()
})
</script>

<template>
  <section class="space-y-6">
    <header>
      <h2 class="font-display text-2xl font-semibold text-slate-900">PHP</h2>
      <p class="mt-1 text-sm text-slate-500">
        Kelola multi-versi PHP. Active:
        <span class="font-medium text-slate-800">{{ active || '—' }}</span>
      </p>
      <p class="mt-2 rounded-lg border border-teal-100 bg-teal-50 px-3 py-2 text-xs text-teal-800">
        PHP tidak punya tombol Start sendiri — berjalan sebagai modul Apache.
        Install versi → Switch → lalu Start Apache. Status PHP jadi
        <strong>Running</strong> saat Apache aktif.
      </p>
    </header>

    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <table class="min-w-full text-left text-sm">
        <thead class="bg-slate-50 text-xs tracking-wide text-slate-500 uppercase">
          <tr>
            <th class="px-4 py-3 font-medium">Version</th>
            <th class="px-4 py-3 font-medium">Status</th>
            <th class="px-4 py-3 font-medium">Path</th>
            <th class="px-4 py-3 font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in versions" :key="row.version" class="border-t border-slate-100">
            <td class="px-4 py-3 font-medium text-slate-800">
              PHP {{ row.version }}
              <span
                v-if="row.active"
                class="ml-2 rounded bg-teal-50 px-1.5 py-0.5 text-xs font-medium text-teal-700"
              >
                Active
              </span>
            </td>
            <td class="px-4 py-3 text-slate-600">
              {{ row.installed ? 'Installed' : 'Not installed' }}
            </td>
            <td class="px-4 py-3 font-mono text-xs text-slate-600">{{ row.path || '—' }}</td>
            <td class="px-4 py-3">
              <div class="flex flex-wrap gap-2">
                <button
                  v-if="!row.installed"
                  class="btn-primary"
                  type="button"
                  @click="run('Install', row.version)"
                >
                  Install
                </button>
                <template v-else>
                  <button
                    v-if="!row.active"
                    class="btn-secondary"
                    type="button"
                    @click="run('Switch', row.version)"
                  >
                    Switch
                  </button>
                  <button class="btn-danger" type="button" @click="run('Delete', row.version)">
                    Delete
                  </button>
                </template>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
