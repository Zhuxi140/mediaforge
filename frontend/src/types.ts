export type AppMode = 'renamer' | 'ffmpeg' | 'subtitle' | 'mediainfo' | 'm3u8'
export type SubMode = 'Extract' | 'Convert'
export type RenameMode = 'BasicMode' | 'SmartMode'
export type TaskStatus = 'pending' | 'processing' | 'success' | 'error' | 'scanning' | 'ready' | 'no_sub'
export type MediaType = 'video' | 'audio'
export type Quality = 'Low' | 'Medium' | 'High' | 'Custom'
export type ModalType = 'success' | 'warning' | 'error' | 'info'
export type HwAccel = 'cpu' | 'nvidia' | 'intel' | 'amd' | 'mac'

export interface ExtractRule {
  Name: string
  Pattern: string
}

export interface RenameRule {
  Mode: RenameMode
  Prefix: string
  Suffix: string
  ReplaceOld: string
  ReplaceNew: string
  SmartRules: ExtractRule[]
  SmartTemplate: string
  CleanChars: string
}

export interface RenamePreview {
  originalPath?: string
  originalName: string
  newName: string
  newPath: string
  hasConflict: boolean
  formatError: string
}

export interface FFmpegTask {
  ID: string
  InputPath: string
  OutputPath: string
  Format: string
  Quality: Quality
  VideoCRF: string
  AudioVBR: string
  HwAccel: HwAccel
  ForceDropSubtitle: boolean
}

export interface MediaTask {
  id: string
  path: string
  name: string
  targetFormat: string
  status: TaskStatus
  progressTime: string
  errorMessage?: string
}

export interface MediaSettings {
  mediaType: MediaType
  globalTargetFormat: string
  globalQuality: Quality
  videoCRF: string
  audioVBR: string
  outputDir: string
  hwAccel: HwAccel
}

export interface SubtitleStream {
  Index: string
  Language: string
  Codec: string
}

export interface SubtitleTask {
  id: string
  path: string
  name: string
  status: TaskStatus
  streams: SubtitleStream[]
  selectedStreams: string[]
  progressText: string
}

export interface SubConvertTask {
  id: string
  path: string
  name: string
  outputName: string
  targetFormat: string
  status: TaskStatus
  progressText: string
}

export interface ModalState {
  visible: boolean
  title: string
  message: string
  type: ModalType
  isConfirm: boolean
}

export interface VideoStream {
  index: number
  codec: string
  width: number
  height: number
  frameRate: string
  bitRate: string
  pixelFormat: string
  profile: string
}

export interface AudioStream {
  index: number
  codec: string
  sampleRate: number
  channels: number
  bitRate: string
  language: string
}

export interface MediaInfo {
  filePath: string
  fileSize: string
  format: string
  duration: string
  bitRate: string
  video: VideoStream[]
  audio: AudioStream[]
  subtitle: SubtitleStream[]
}

export interface CleanCharOption {
  label: string
  value: string
}

export interface M3U8Task {
  id: string
  url: string
  fileName: string
  status: TaskStatus
  progressTime: string
  outputPath?: string
}
