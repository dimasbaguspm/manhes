/// <reference lib="webworker" />

export type WorkerInMessage = { urls: string[] }

export type WorkerOutMessage =
  | { type: 'page'; index: number; bitmap: ImageBitmap }
  | { type: 'progress'; loaded: number; total: number }
  | { type: 'error'; index: number; message: string }
  | { type: 'done' }

self.onmessage = async (e: MessageEvent<WorkerInMessage>) => {
  const { urls } = e.data
  const total = urls.length
  let loaded = 0

  await Promise.all(
    urls.map(async (url, index) => {
      try {
        const res = await fetch(url)
        if (!res.ok) throw new Error(`HTTP ${res.status}`)
        const blob = await res.blob()
        const bitmap = await createImageBitmap(blob)
        loaded++
        // Transfer ownership (zero-copy) — bitmap is no longer usable in the worker after this.
        self.postMessage({ type: 'page', index, bitmap } satisfies WorkerOutMessage, [bitmap])
        self.postMessage({ type: 'progress', loaded, total } satisfies WorkerOutMessage)
      } catch (err) {
        loaded++
        self.postMessage(
          { type: 'error', index, message: String(err instanceof Error ? err.message : err) } satisfies WorkerOutMessage,
        )
        self.postMessage({ type: 'progress', loaded, total } satisfies WorkerOutMessage)
      }
    }),
  )

  self.postMessage({ type: 'done' } satisfies WorkerOutMessage)
}
