import { onMounted, onUnmounted, ref } from 'vue'
import { Events } from '@wailsio/runtime'
import { useToast } from './useToast'

export interface DownloadProgress {
  type: string
  version: string
  percent: number
  bytes: number
  total: number
  cached: boolean
  filename: string
}

export interface InstallStageEvent {
  type: string
  version: string
  stage: string
  message: string
}

export interface InstallErrorEvent {
  type: string
  version: string
  message: string
}

const open = ref(false)
const title = ref('Instalasi')
const stage = ref('')
const percent = ref(0)

type DoneHandler = (componentType: string, version: string) => void
const doneHandlers = new Set<DoneHandler>()

export function onInstallDone(handler: DoneHandler) {
  doneHandlers.add(handler)
  return () => doneHandlers.delete(handler)
}

export function useInstallProgress() {
  const toast = useToast()
  const unsubs: Array<() => void> = []

  function stageLabel(s: string): string {
    switch (s) {
      case 'download':
        return 'Mengunduh…'
      case 'verify':
        return 'Memverifikasi…'
      case 'extract':
        return 'Mengekstrak…'
      case 'configure':
        return 'Mengonfigurasi…'
      case 'done':
        return 'Selesai'
      default:
        return s
    }
  }

  onMounted(() => {
    unsubs.push(
      Events.On('download:progress', (event: { data?: DownloadProgress }) => {
        const data = event.data
        if (!data) return
        open.value = true
        title.value = `Install ${data.type} ${data.version}`
        percent.value = Math.round(data.percent || 0)
        if (data.cached) {
          stage.value = 'Menggunakan cache lokal'
          percent.value = 100
        } else {
          stage.value = `Mengunduh ${data.filename || ''}`.trim()
        }
      }),
    )

    unsubs.push(
      Events.On('install:stage', (event: { data?: InstallStageEvent }) => {
        const data = event.data
        if (!data) return
        open.value = true
        title.value = `Install ${data.type} ${data.version}`
        stage.value = data.message || stageLabel(data.stage)
        if (data.stage === 'verify' || data.stage === 'extract' || data.stage === 'configure') {
          percent.value = Math.max(percent.value, 100)
        }
        if (data.stage === 'done') {
          percent.value = 100
          stage.value = data.message || 'Selesai'
          toast.success(`${data.type} ${data.version} terinstal`)
          window.setTimeout(() => {
            open.value = false
          }, 800)
          doneHandlers.forEach((h) => h(data.type, data.version))
        }
      }),
    )

    unsubs.push(
      Events.On('install:error', (event: { data?: InstallErrorEvent }) => {
        const data = event.data
        if (!data) return
        open.value = false
        toast.error(data.message || 'Instalasi gagal')
      }),
    )
  })

  onUnmounted(() => {
    unsubs.forEach((u) => u())
  })

  return { open, title, stage, percent }
}
