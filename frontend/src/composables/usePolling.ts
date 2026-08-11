import { onMounted, onUnmounted, ref } from 'vue'

/**
 * Periodically invokes `fn` while the component is mounted.
 * Skips overlapping runs if the previous tick is still in flight.
 */
export function usePolling(fn: () => void | Promise<void>, intervalMs = 3000) {
  const ticking = ref(false)
  let timer: number | undefined

  async function tick() {
    if (ticking.value) return
    ticking.value = true
    try {
      await fn()
    } finally {
      ticking.value = false
    }
  }

  function start() {
    stop()
    void tick()
    timer = window.setInterval(() => {
      void tick()
    }, intervalMs)
  }

  function stop() {
    if (timer !== undefined) {
      window.clearInterval(timer)
      timer = undefined
    }
  }

  onMounted(start)
  onUnmounted(stop)

  return { start, stop, tick, ticking }
}
