<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import RenamerPanel from './components/RenamerPanel.vue'
import FFmpegPanel from './components/FFmpegPanel.vue'
import SubtitlePanel from './components/SubtitlePanel.vue'
import M3U8Panel from './components/M3U8Panel.vue'
import ModalOverlay from './components/ModalOverlay.vue'
import { useRenamerStore } from './stores/renamer'
import { useMediaStore } from './stores/media'
import { useSubtitleStore } from './stores/subtitle'
import { SelectFiles } from '../wailsjs/go/controller/RenamerApp'
import { LoadSettings, SaveSettings } from '../wailsjs/go/controller/MediaApp'
import { OnFileDrop, OnFileDropOff } from '../wailsjs/runtime/runtime'
import * as models from '../wailsjs/go/models'

const activeApp = ref<'renamer' | 'ffmpeg' | 'subtitle' | 'm3u8'>('subtitle')

const renamer = useRenamerStore()
const media = useMediaStore()
const subtitle = useSubtitleStore()

// 配置持久化
onMounted(async () => {
  try {
    const s = await LoadSettings() as models.config.Settings
    if (s.mediaOutputDir) media.mediaSettings.outputDir = s.mediaOutputDir
    if (s.mediaHwAccel) media.mediaSettings.hwAccel = s.mediaHwAccel as any
    if (s.mediaQuality) media.mediaSettings.globalQuality = s.mediaQuality as any
    if (s.mediaVideoCRF) media.mediaSettings.videoCRF = s.mediaVideoCRF
    if (s.mediaAudioVBR) media.mediaSettings.audioVBR = s.mediaAudioVBR
    if (s.mediaTargetFmt) media.mediaSettings.globalTargetFormat = s.mediaTargetFmt
    if (s.subOutputDir) subtitle.subSettings.outputDir = s.subOutputDir
    if (s.subFormat) subtitle.subSettings.format = s.subFormat
    if (s.subConvertDir) subtitle.subConvertSettings.outputDir = s.subConvertDir
    if (s.subConvertFmt) subtitle.subConvertSettings.globalTargetFormat = s.subConvertFmt
  } catch { /* silent */ }
})

// 自动保存配置（debounced）
let saveTimer: number | undefined
const autoSave = () => {
  if (saveTimer) clearTimeout(saveTimer)
  saveTimer = window.setTimeout(async () => {
    try {
      await SaveSettings({
        mediaOutputDir: media.mediaSettings.outputDir,
        mediaHwAccel: media.mediaSettings.hwAccel,
        mediaQuality: media.mediaSettings.globalQuality,
        mediaVideoCRF: media.mediaSettings.videoCRF,
        mediaAudioVBR: media.mediaSettings.audioVBR,
        mediaTargetFmt: media.mediaSettings.globalTargetFormat,
        subOutputDir: subtitle.subSettings.outputDir,
        subFormat: subtitle.subSettings.format,
        subConvertDir: subtitle.subConvertSettings.outputDir,
        subConvertFmt: subtitle.subConvertSettings.globalTargetFormat,
      } as any)
    } catch { /* silent */ }
  }, 1000)
}

watch(() => media.mediaSettings, autoSave, { deep: true })
watch(() => subtitle.subSettings, autoSave, { deep: true })
watch(() => subtitle.subConvertSettings, autoSave, { deep: true })

const handleSelectFiles = async () => {
  const files = await SelectFiles()
  if (!files || files.length === 0) return
  if (activeApp.value === 'renamer') {
    const newFiles = files.filter(f => !renamer.filePaths.includes(f))
    renamer.filePaths = [...renamer.filePaths, ...newFiles]
  } else if (activeApp.value === 'ffmpeg') {
    media.addMediaFiles(files)
  } else if (activeApp.value === 'subtitle') {
    if (subtitle.currentSubMode === 'Extract') {
      subtitle.addSubtitleFiles(files)
    } else {
      subtitle.addConvertSubtitleFiles(files)
    }
  }
}

onMounted(() => {
  OnFileDrop((x, y, paths) => {
    if (activeApp.value === 'renamer') {
      const newFiles = paths.filter((f: string) => !renamer.filePaths.includes(f))
      renamer.filePaths = [...renamer.filePaths, ...newFiles]
    } else if (activeApp.value === 'ffmpeg') {
      media.addMediaFiles(paths)
    } else if (activeApp.value === 'subtitle') {
      if (subtitle.currentSubMode === 'Extract') {
        subtitle.addSubtitleFiles(paths)
      } else {
        subtitle.addConvertSubtitleFiles(paths)
      }
    }
  }, false)
})

onUnmounted(() => OnFileDropOff())
</script>

<template>
  <div class="app-root">
    <nav class="tab-bar">
      <div class="tab-logo">
        <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="url(#logo-grad)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <defs><linearGradient id="logo-grad" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" stop-color="#3b82f6" /><stop offset="100%" stop-color="#8b5cf6" /></linearGradient></defs>
          <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
        </svg>
        <span class="tab-brand">MediaForge</span>
      </div>
      <div class="tab-items">
        <button class="tab-item" :class="{ active: activeApp === 'renamer' }" @click="activeApp = 'renamer'">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 13.5V4a2 2 0 0 1 2-2h8.5L20 7.5V20a2 2 0 0 1-2 2h-5.5"/><polyline points="14 2 14 8 20 8"/><path d="M10.42 12.61a2.1 2.1 0 1 1 2.97 2.97L7.95 21 4 22l.99-3.95 5.43-5.44Z"/></svg>
          <span>重命名</span>
        </button>
        <button class="tab-item" :class="{ active: activeApp === 'ffmpeg' }" @click="activeApp = 'ffmpeg'">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.2 6 3 11l-.9-2.4c-.3-1.1.4-2.2 1.5-2.5l13.5-4c1.1-.3 2.2.4 2.5 1.5z"/><path d="m6.2 5.3 3.1 3.9"/><path d="m12.4 3.4 3.1 4"/><path d="M3 11h18v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z"/></svg>
          <span>影音工厂</span>
        </button>
        <button class="tab-item" :class="{ active: activeApp === 'subtitle' }" @click="activeApp = 'subtitle'">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/><line x1="9" y1="9" x2="15" y2="9"/><line x1="9" y1="13" x2="15" y2="13"/></svg>
          <span>字幕处理</span>
        </button>
        <button class="tab-item" :class="{ active: activeApp === 'm3u8' }" @click="activeApp = 'm3u8'">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="23 7 16 12 23 17 23 7"/><rect x="1" y="5" width="15" height="14" rx="2" ry="2"/></svg>
          <span>M3U8流</span>
        </button>
      </div>
    </nav>

    <div class="panel-container">
      <RenamerPanel v-show="activeApp === 'renamer'" @select-files="handleSelectFiles" />
      <FFmpegPanel v-show="activeApp === 'ffmpeg'" @select-files="handleSelectFiles" />
      <SubtitlePanel v-show="activeApp === 'subtitle'" @select-files="handleSelectFiles" />
      <M3U8Panel v-show="activeApp === 'm3u8'" />
    </div>

    <ModalOverlay />
  </div>
</template>
