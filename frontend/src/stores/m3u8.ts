import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { M3U8Task, MediaInfo } from '../types'
import { DownloadM3U8, CancelDownload, SelectOutputDir, ProbeM3U8URL } from '../../wailsjs/go/controller/M3U8App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

function generateId(): string {
  return Math.random().toString(36).substring(2, 9) + Date.now().toString(36)
}

function clearListeners(id: string) {
  EventsOff('m3u8-progress-' + id)
  EventsOff('m3u8-done-' + id)
  EventsOff('m3u8-error-' + id)
}

export const useM3U8Store = defineStore('m3u8', () => {
  const m3u8Url = ref('')
  const outputDir = ref('')
  const tasks = ref<M3U8Task[]>([])
  const probeResult = ref<MediaInfo | null>(null)
  const probeError = ref('')
  const probing = ref(false)

  const probe = async () => {
    const url = m3u8Url.value.trim()
    if (!url) return
    probing.value = true
    probeError.value = ''
    probeResult.value = null
    try {
      probeResult.value = await ProbeM3U8URL(url)
    } catch (err: any) {
      probeError.value = String(err)
    } finally {
      probing.value = false
    }
  }

  const startDownload = async () => {
    if (!m3u8Url.value.trim()) return
    const id = generateId()
    tasks.value.push({
      id, url: m3u8Url.value,
      fileName: m3u8Url.value.split('/').pop() || 'stream',
      status: 'processing', progressTime: '正在连接...'
    })

    EventsOn('m3u8-progress-' + id, (t: string) => {
      const tsk = tasks.value.find(x => x.id === id)
      if (tsk) tsk.progressTime = t
    })
    EventsOn('m3u8-done-' + id, (path: string) => {
      const tsk = tasks.value.find(x => x.id === id)
      if (tsk) { tsk.status = 'success'; tsk.progressTime = '✅ 完成'; tsk.outputPath = path }
      clearListeners(id)
    })
    EventsOn('m3u8-error-' + id, (err: string) => {
      const tsk = tasks.value.find(x => x.id === id)
      if (tsk) { tsk.status = 'error'; tsk.progressTime = err === '任务已取消' ? '⏹️ 已取消' : '❌ ' + err }
      clearListeners(id)
    })

    try {
      await DownloadM3U8(id, m3u8Url.value, outputDir.value)
    } catch (err: any) {
      clearListeners(id)
      const tsk = tasks.value.find(x => x.id === id)
      if (tsk) { tsk.status = 'error'; tsk.progressTime = '❌ ' + err }
    }
  }

  const cancelDownload = (task: M3U8Task) => {
    if (task.status === 'processing') {
      task.progressTime = '正在停止...'
      CancelDownload(task.id)
    }
  }

  const removeTask = (i: number) => {
    const t = tasks.value[i]; clearListeners(t.id); tasks.value.splice(i, 1)
  }

  const clearTasks = () => {
    tasks.value.forEach(t => clearListeners(t.id)); tasks.value = []
  }

  const selectDir = async () => {
    const dir = await SelectOutputDir()
    if (dir) outputDir.value = dir
  }

  return {
    m3u8Url, outputDir, tasks,
    probeResult, probeError, probing,
    startDownload, cancelDownload, removeTask, clearTasks, selectDir, probe
  }
})
