import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { MediaTask, MediaSettings, Quality, HwAccel } from '../types'
import { CheckInputFile, SubmitMediaTask, CancelMediaTask } from '../../wailsjs/go/controller/MediaApp'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useModal } from '../composables/useModal'

function generateId(): string {
  return Math.random().toString(36).substring(2, 9) + Date.now().toString(36)
}

function clearMediaListeners(id: string) {
  EventsOff('ffmpeg-progress-' + id)
  EventsOff('ffmpeg-done-' + id)
  EventsOff('ffmpeg-error-' + id)
}

export const useMediaStore = defineStore('media', () => {
  const mediaTasks = ref<MediaTask[]>([])
  const mediaSettings = ref<MediaSettings>({
    mediaType: 'video',
    globalTargetFormat: 'mp4',
    globalQuality: 'Medium',
    videoCRF: '23',
    audioVBR: '2',
    outputDir: '',
    hwAccel: 'cpu'
  })

  const videoFormats = ['mp4', 'mkv', 'avi', 'gif']
  const audioFormats = ['mp3', 'wav', 'aac', 'flac']
  const currentFormatOptions = computed(() =>
    mediaSettings.value.mediaType === 'video' ? videoFormats : audioFormats
  )

  const hasErrorTasks = computed(() => mediaTasks.value.some(t => t.status === 'error'))

  const addMediaFiles = async (paths: string[]) => {
    for (const p of paths) {
      if (mediaTasks.value.some(t => t.path === p)) continue
      const name = p.split('\\').pop() || p.split('/').pop() || p
      const task: MediaTask = {
        id: generateId(), path: p, name,
        targetFormat: mediaSettings.value.globalTargetFormat,
        status: 'pending', progressTime: '等待中...'
      }
      mediaTasks.value.push(task)
      try {
        const r = await CheckInputFile(p)
        if (r !== 'success') {
          task.status = 'error'
          task.errorMessage = r
          task.progressTime = '❌ ' + r
        }
      } catch { /* skip */ }
    }
  }

  const startMediaTask = async (task: MediaTask, forceDropSubtitle = false) => {
    if (task.status === 'processing') return
    task.status = 'processing'; task.progressTime = '启动引擎...'; task.errorMessage = ''

    EventsOn('ffmpeg-progress-' + task.id, (t: string) => { task.progressTime = '处理中: ' + t })
    EventsOn('ffmpeg-done-' + task.id, () => { task.status = 'success'; task.progressTime = '✅ 完成'; clearMediaListeners(task.id) })
    EventsOn('ffmpeg-error-' + task.id, (err: string) => {
      task.status = 'error'; task.errorMessage = err
      task.progressTime = err === '任务已取消' ? '⏹️ 已取消' : '❌ 引擎崩溃 (查看详情)'
      if (err !== '任务已取消') {
        useModal().showModal('转码失败', `文件: ${task.name}\n\n原因: ${err}`, 'error')
      }
      clearMediaListeners(task.id)
    })

    const crf = mediaSettings.value.mediaType === 'video' ? String(mediaSettings.value.videoCRF) : ''
    const vbr = mediaSettings.value.mediaType === 'audio' ? String(mediaSettings.value.audioVBR) : ''

    try {
      await SubmitMediaTask({
        ID: task.id, InputPath: task.path,
        OutputPath: mediaSettings.value.outputDir,
        Format: task.targetFormat,
        Quality: mediaSettings.value.globalQuality as Quality,
        VideoCRF: crf, AudioVBR: vbr,
        HwAccel: mediaSettings.value.hwAccel as HwAccel,
        ForceDropSubtitle: forceDropSubtitle
      })
    } catch (err: any) {
      clearMediaListeners(task.id)
      if (err === 'WARN_SUBTITLE') {
        const ok = await useModal().showModal(
          '⚠️ 预检拦截',
          `文件 [${task.name}]\n包含 MP4 不兼容的内封字幕流！\n强行转换可能导致程序崩溃或丢弃字幕。`,
          'warning', true
        )
        if (ok) { task.status = 'pending'; startMediaTask(task, true) }
        else { task.status = 'pending'; task.progressTime = '⏸️ 等待处理字幕' }
      } else {
        task.status = 'error'; task.progressTime = '❌ ' + err
        useModal().showModal('系统异常', err, 'error')
      }
    }
  }

  const startAllMediaTasks = () => {
    mediaTasks.value.forEach(t => {
      if (t.status === 'pending' || t.status === 'error' || t.progressTime === '⏸️ 等待处理字幕') startMediaTask(t)
    })
  }

  const stopMediaTask = (task: MediaTask) => {
    if (task.status === 'processing') { task.progressTime = '正在停止...'; CancelMediaTask(task.id) }
  }

  const removeMediaTask = (i: number) => {
    const t = mediaTasks.value[i]; clearMediaListeners(t.id); mediaTasks.value.splice(i, 1)
  }

  const clearMediaTasks = () => {
    mediaTasks.value.forEach(t => clearMediaListeners(t.id)); mediaTasks.value = []
  }

  const clearErrorTasks = () => {
    mediaTasks.value.filter(t => t.status === 'error').forEach(t => clearMediaListeners(t.id))
    mediaTasks.value = mediaTasks.value.filter(t => t.status !== 'error')
  }

  return {
    mediaTasks, mediaSettings, videoFormats, audioFormats,
    currentFormatOptions, hasErrorTasks,
    addMediaFiles, startMediaTask, startAllMediaTasks,
    stopMediaTask, removeMediaTask, clearMediaTasks, clearErrorTasks
  }
})
