import { ref } from 'vue'

export type ToastKind = 'info' | 'success' | 'error'

export interface ToastItem {
  id: number
  kind: ToastKind
  message: string
}

const toasts = ref<ToastItem[]>([])
let seq = 0

export function useToast() {
  function push(message: string, kind: ToastKind = 'info') {
    const id = ++seq
    toasts.value.push({ id, kind, message })
    window.setTimeout(() => {
      toasts.value = toasts.value.filter((t) => t.id !== id)
    }, 3200)
  }

  return {
    toasts,
    info: (message: string) => push(message, 'info'),
    success: (message: string) => push(message, 'success'),
    error: (message: string) => push(message, 'error'),
  }
}
