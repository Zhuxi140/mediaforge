import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import type { SubtitleTask, SubConvertTask, SubtitleStream } from '../types'
import { ScanSubtitles, ExtractSubtitle, ConvertSubtitle } from '../../wailsjs/go/controller/MediaApp'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useModal } from '../composables/useModal'

function generateId(): string {
  return Math.random().toString(36).substring(2, 9) + Date.now().toString(36)
}

export const useSubtitleStore = defineStore('subtitle', () => {
  const currentSubMode = ref<'Extract' | 'Convert'>('Extract')

  // Extract state
  const subTasks = ref<SubtitleTask[]>([])
  const subSettings = ref({ outputDir: '', format: 'srt' })

  // Convert state
  const subConvertTasks = ref<SubConvertTask[]>([])
  const subConvertSettings = ref({ outputDir: '', globalTargetFormat: 'ass' })

  const supportedSubFormats = ['srt', 'ass', 'vtt', 'ssa', 'ttml', 'smi', 'sub']

  watch(() => subConvertSettings.value.globalTargetFormat, (f) => {
    subConvertTasks.value.forEach(t => { if (t.status === 'pending') t.targetFormat = f })
  })

  const hasErrorSubTasks = computed(() => subTasks.value.some(t => t.status === 'error' || t.status === 'no_sub'))

  // --- Extract ---
  const addSubtitleFiles = async (paths: string[]) => {
    for (const p of paths) {
      if (subTasks.value.some(t => t.path === p)) continue
      subTasks.value.push({
        id: generateId(), path: p,
        name: p.split('\\').pop() || p.split('/').pop() || p,
        status: 'scanning', streams: [], selectedStreams: [], progressText: '🔍 扫描中...'
      })
      const proxy = subTasks.value[subTasks.value.length - 1]
      try {
        const streams = await ScanSubtitles(p)
        if (streams && streams.length > 0) {
          proxy.streams = streams as SubtitleStream[]
          proxy.status = 'ready'
          proxy.progressText = `发现 ${streams.length} 条字幕`
          proxy.selectedStreams = streams.map((s: any) => s.Index)
        } else {
          proxy.status = 'no_sub'
          proxy.progressText = '未发现字幕流'
        }
      } catch (err) {
        proxy.status = 'error'
        proxy.progressText = '❌ ' + err
      }
    }
  }

  const extractSubtitles = async (task: SubtitleTask) => {
    if (task.selectedStreams.length === 0) {
      useModal().showModal('提示', '请至少选择一条字幕流进行提取！', 'warning')
      return
    }
    const hasUnsupported = task.selectedStreams.some(idx => {
      const s = task.streams.find(st => st.Index === idx)
      return s && (s.Codec.toLowerCase() === 'microdvd' || s.Codec.toLowerCase() === 'dvd_subtitle')
    })
    if (hasUnsupported) {
      const agree = await useModal().showModal(
        '⚠️ 格式硬核预警',
        '您勾选了 MicroDVD / VobSub 格式的字幕。强行提取可能失败。确定继续吗？',
        'warning', true
      )
      if (!agree) return
    }

    task.status = 'processing'; task.progressText = '提取中...'
    const expected = task.selectedStreams.length
    let finished = 0; let success = 0
    const checkDone = () => { if (finished === expected) { task.status = 'success'; task.progressText = `✅ 已提取 ${success} 条` } }

    for (const idx of task.selectedStreams) {
      const subId = task.id + '_' + idx
      EventsOff('ffmpeg-done-' + subId); EventsOff('ffmpeg-error-' + subId)
      EventsOn('ffmpeg-done-' + subId, () => { success++; finished++; checkDone() })
      EventsOn('ffmpeg-error-' + subId, (err) => { useModal().showModal('提取出错', String(err), 'error'); finished++; checkDone() })
      ExtractSubtitle(subId, task.path, idx, subSettings.value.outputDir, subSettings.value.format)
        .catch(() => { finished++; checkDone() })
    }
  }

  const extractAllTasks = () => { subTasks.value.forEach(t => { if (t.status === 'ready') extractSubtitles(t) }) }
  const removeSubTask = (i: number) => { subTasks.value.splice(i, 1) }
  const clearSubTasks = () => { subTasks.value = [] }
  const clearErrorSubTasks = () => { subTasks.value = subTasks.value.filter(t => t.status !== 'error' && t.status !== 'no_sub') }

  // --- Convert ---
  const addConvertSubtitleFiles = (paths: string[]) => {
    for (const p of paths) {
      if (subConvertTasks.value.some(t => t.path === p)) continue
      subConvertTasks.value.push({
        id: generateId(), path: p,
        name: p.split('\\').pop() || p.split('/').pop() || p,
        outputName: '',
        targetFormat: subConvertSettings.value.globalTargetFormat,
        status: 'pending', progressText: '等待转换'
      })
    }
  }

  const runSubConvert = async (task: SubConvertTask) => {
    const outDir = subConvertSettings.value.outputDir.trim()
    const outName = task.outputName.trim()
    if (!outDir && !outName) {
      task.status = 'error'; task.progressText = '缺少输出信息'
      useModal().showModal('缺失', `文件 [${task.name}]\n输出目录和新文件名至少填一个`, 'warning')
      return
    }
    task.status = 'processing'; task.progressText = '转换中...'
    try {
      await ConvertSubtitle(task.path, outDir, outName, task.targetFormat)
      task.status = 'success'; task.progressText = '转换成功'
    } catch (err) {
      task.status = 'error'; task.progressText = '失败'
      useModal().showModal('转换失败', String(err), 'error')
    }
  }

  const runAllSubConvert = () => { subConvertTasks.value.forEach(t => { if (t.status === 'pending' || t.status === 'error') runSubConvert(t) }) }
  const clearSubConvertTasks = () => { subConvertTasks.value = [] }
  const removeSubConvertTask = (i: number) => { subConvertTasks.value.splice(i, 1) }

  return {
    currentSubMode, subTasks, subSettings, subConvertTasks, subConvertSettings,
    supportedSubFormats, hasErrorSubTasks,
    addSubtitleFiles, extractSubtitles, extractAllTasks,
    removeSubTask, clearSubTasks, clearErrorSubTasks,
    addConvertSubtitleFiles, runSubConvert, runAllSubConvert,
    clearSubConvertTasks, removeSubConvertTask
  }
})
