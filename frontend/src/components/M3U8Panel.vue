<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import { useM3U8Store } from '../stores/m3u8'

declare const Hls: any

const store = useM3U8Store()
const videoRef = ref<HTMLVideoElement | null>(null)
const playing = ref(false)
const playerError = ref('')
let hls: any = null

onUnmounted(() => {
  if (hls) { hls.destroy(); hls = null }
})

function stopPlayback() {
  if (hls) { hls.destroy(); hls = null }
  if (videoRef.value) { videoRef.value.src = '' }
  playing.value = false
  playerError.value = ''
}

function handlePlay() {
  const url = store.m3u8Url.trim()
  if (!url) return
  stopPlayback()

  if (Hls.isSupported()) {
    hls = new Hls()
    hls.loadSource(url)
    hls.attachMedia(videoRef.value!)
    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      playing.value = true
      videoRef.value?.play()
    })
    hls.on(Hls.Events.ERROR, (_event: any, data: any) => {
      if (data.fatal) {
        playerError.value = '播放失败: ' + (data.type || '未知错误')
      }
    })
  } else if (videoRef.value?.canPlayType('application/vnd.apple.mpegurl')) {
    videoRef.value.src = url
    playing.value = true
  } else {
    playerError.value = '当前浏览器不支持 HLS 播放'
  }
}

function formatUrlDisplay(url: string) {
  try {
    const u = new URL(url)
    return u.hostname + u.pathname.substring(0, 40) + (u.pathname.length > 40 ? '...' : '')
  } catch {
    return url.substring(0, 50)
  }
}
</script>

<template>
  <main class="app-layout">
    <div class="workspace">
      <div class="m3u8-content">
        <div class="m3u8-card">
          <h2 class="section-title">M3U8 流媒体工具</h2>

          <div class="input-field" style="margin-bottom: 12px;">
            <label>视频地址</label>
            <input v-model="store.m3u8Url" placeholder="https://example.com/playlist.m3u8" class="m3u8-url-input" />
          </div>

          <div class="m3u8-actions">
            <button class="btn btn-primary" @click="handlePlay" :disabled="!store.m3u8Url.trim()">
              ▶ 在线播放
            </button>
            <button class="btn btn-secondary" @click="store.startDownload" :disabled="!store.m3u8Url.trim()">
              ⬇ 下载
            </button>
            <button class="btn btn-outline" @click="store.probe" :disabled="!store.m3u8Url.trim() || store.probing" style="display:flex;align-items:center;gap:4px;">
              {{ store.probing ? '分析中...' : '验证有效性' }}
              <span class="btn-help" title="检查m3u8链接有效性，并解析出基础信息。如果视频较大可能需要几十秒，请耐心等待">!</span>
            </button>
          </div>

          <div class="divider"></div>

          <div class="input-field" style="margin-bottom: 12px;">
            <label>下载目录（留空则保存到系统下载文件夹）</label>
            <div class="dir-input-row">
              <input v-model="store.outputDir" placeholder="如: C:\Videos" />
              <button class="btn btn-secondary" @click="store.selectDir">浏览</button>
            </div>
          </div>
        </div>

        <div v-if="playerError" class="m3u8-card m3u8-error-card">
          <span style="color: #ef4444;">{{ playerError }}</span>
          <button class="btn btn-action danger" @click="playerError = ''" style="margin-left: 12px;">关闭</button>
        </div>

        <div v-show="playing" class="m3u8-player-card">
          <div class="m3u8-player-header">
            <span class="status-badge success" style="display: inline-flex;">▶ 播放中</span>
            <button class="btn btn-action danger" @click="stopPlayback">停止播放</button>
          </div>
          <div class="m3u8-player-wrapper">
            <video ref="videoRef" controls class="m3u8-video"></video>
          </div>
        </div>

        <div v-if="store.probeResult" class="m3u8-card probe-card">
          <div class="probe-header">
            <h3 class="section-title" style="margin:0;">流信息</h3>
            <button class="btn btn-action" @click="store.probeResult = null">关闭</button>
          </div>
          <div class="probe-body">
            <div class="probe-row" v-if="store.probeResult.format"><span class="probe-label">封装格式</span><span>{{ store.probeResult.format }}</span></div>
            <div class="probe-row" v-if="store.probeResult.duration && store.probeResult.duration !== '0'"><span class="probe-label">总时长</span><span>{{ store.probeResult.duration }}</span></div>
            <div class="probe-row" v-if="store.probeResult.bitRate"><span class="probe-label">总码率</span><span>{{ store.probeResult.bitRate }}</span></div>
            <template v-for="(vs, idx) in store.probeResult.video" :key="'v'+idx">
              <div class="probe-row"><span class="probe-label">视频流 {{ idx+1 }}</span><span>{{ vs.codec }} · {{ vs.width }}x{{ vs.height }} · {{ vs.frameRate }} · {{ vs.bitRate || '未知码率' }}</span></div>
            </template>
            <template v-for="(as, idx) in store.probeResult.audio" :key="'a'+idx">
              <div class="probe-row"><span class="probe-label">音频流 {{ idx+1 }}</span><span>{{ as.codec }} · {{ as.channels }}ch · {{ as.sampleRate }}Hz · {{ as.bitRate || '未知码率' }} · {{ as.language }}</span></div>
            </template>
          </div>
        </div>
        <div v-else-if="store.probeError" class="m3u8-card m3u8-error-card">
          <span style="color: #ef4444;">{{ store.probeError }}</span>
          <button class="btn btn-action danger" @click="store.probeError = ''" style="margin-left: 12px;">关闭</button>
        </div>

        <div v-if="store.tasks.length > 0" class="m3u8-card">
          <div class="m3u8-task-header">
            <h3 class="section-title" style="margin:0;">下载任务</h3>
            <button class="btn btn-danger-soft" @click="store.clearTasks">清空列表</button>
          </div>
          <div v-for="(task, i) in store.tasks" :key="task.id" class="m3u8-task-row"
               :class="{ 'row-error': task.status === 'error' }">
            <div class="m3u8-task-info">
              <span class="m3u8-task-name" :title="task.url">{{ formatUrlDisplay(task.url) }}</span>
              <span class="m3u8-task-progress" :class="task.status === 'processing' ? 'processing' : task.status">
                {{ task.progressTime }}
              </span>
              <span v-if="task.status === 'processing' && task.progressTime.includes('总时长')" class="m3u8-task-legend">总时长 | 速率 | 当前已处理时长</span>
            </div>
            <div class="m3u8-task-actions">
              <button v-if="task.status === 'processing'" class="btn-action danger" @click="store.cancelDownload(task)">取消</button>
              <template v-else>
                <button class="btn-action danger" @click="store.removeTask(i)">移除</button>
              </template>
            </div>
          </div>
        </div>

        <div v-if="!store.m3u8Url.trim() && !playing && store.tasks.length === 0" class="empty-state" style="height:200px; margin-top: 40px;">
          <div class="empty-icon">📺</div>
          <h3>输入 M3U8 地址</h3>
          <p>粘贴 M3U8 视频链接进行在线播放或下载</p>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.m3u8-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px 32px;
  background: #f1f5f9;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.m3u8-card {
  background: #fff;
  border-radius: 10px;
  padding: 20px 24px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}
.m3u8-error-card {
  display: flex;
  align-items: center;
  background: #fef2f2;
  border: 1px solid #fecaca;
}
.m3u8-url-input {
  font-family: "JetBrains Mono", Consolas, monospace;
  font-size: 13px;
}
.m3u8-actions {
  display: flex;
  gap: 10px;
}
.dir-input-row {
  display: flex;
  gap: 8px;
}
.dir-input-row input { flex: 1; }
.m3u8-player-card {
  background: #000;
  border-radius: 10px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.2);
}
.m3u8-player-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  background: #1e293b;
  border-radius: 10px 10px 0 0;
}
.m3u8-player-wrapper {
  position: relative;
  width: 100%;
  max-height: 480px;
  background: #000;
}
.m3u8-video {
  max-height: 480px;
  width: 100%;
  display: block;
  outline: none;
  border-radius: 0 0 10px 10px;
}
.m3u8-task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.m3u8-task-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f1f5f9;
}
.m3u8-task-row:last-child { border-bottom: none; }
.m3u8-task-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}
.m3u8-task-name {
  font-size: 13px;
  font-weight: 500;
  color: #1e293b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.m3u8-task-progress {
  font-size: 12px;
  font-family: "JetBrains Mono", Consolas, monospace;
}
.m3u8-task-progress.success { color: #10b981; }
.m3u8-task-progress.error { color: #ef4444; }
.m3u8-task-progress.processing { color: #2563eb; }
.m3u8-task-legend {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 2px;
}
.m3u8-task-actions {
  flex-shrink: 0;
  margin-left: 16px;
}
.probe-card { background: #fff; }
.probe-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.probe-body { display: flex; flex-direction: column; gap: 6px; }
.probe-row {
  display: flex;
  font-size: 13px;
  padding: 4px 0;
  border-bottom: 1px solid #f1f5f9;
}
.probe-row:last-child { border-bottom: none; }
.probe-label {
  flex: 0 0 90px;
  color: #64748b;
  font-weight: 500;
}
.btn-help {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  font-size: 11px;
  font-weight: 700;
  color: #94a3b8;
  border: 1px solid #cbd5e1;
  border-radius: 50%;
  cursor: help;
  flex-shrink: 0;
}
</style>
