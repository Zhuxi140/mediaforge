<script setup lang="ts">
import { ref } from 'vue'
import { SelectMediaFile, GetMediaInfo } from '../../wailsjs/go/controller/MediaApp'
import { media } from '../../wailsjs/go/models'

const mediaInfo = ref<media.MediaInfo | null>(null)
const currentPath = ref('')
const loading = ref(false)
const errorMsg = ref('')
const collapsed = ref<string[]>([])

function toggleCollapse(key: string) {
  const idx = collapsed.value.indexOf(key)
  if (idx >= 0) collapsed.value.splice(idx, 1)
  else collapsed.value.push(key)
}

function isCollapsed(key: string) {
  return collapsed.value.includes(key)
}

async function handleSelectFile() {
  const path = await SelectMediaFile()
  if (!path) return
  await loadFile(path)
}

async function loadFile(path: string) {
  currentPath.value = path
  loading.value = true
  errorMsg.value = ''
  try {
    const info = await GetMediaInfo(path)
    if (!info) {
      errorMsg.value = '无法获取元数据，文件可能不存在或格式不支持'
    } else {
      mediaInfo.value = info
    }
  } catch (err: any) {
    errorMsg.value = String(err)
  } finally {
    loading.value = false
  }
}

function getChannelLabel(n: number) {
  if (n === 1) return '单声道'
  if (n === 2) return '立体声'
  if (n === 6) return '5.1 环绕'
  if (n === 8) return '7.1 环绕'
  return n + ' 声道'
}
</script>

<template>
  <main class="app-layout">
    <div class="workspace">
      <aside class="control-sidebar">
        <div class="panel-content">
          <h2 class="section-title">文件选择</h2>
          <div class="input-stack">
            <div class="input-field">
              <label>当前文件</label>
              <div class="file-path-display" :title="currentPath || '未选择'">
                <span class="text-truncate">{{ currentPath || '未选择' }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="global-actions">
          <button class="btn btn-primary w-full shadow-sm" @click="handleSelectFile">
            {{ currentPath ? '更换文件' : '选择媒体文件' }}
          </button>
          <div class="action-row mt-2">
            <p class="tiny-hint">支持视频、音频、GIF 文件</p>
          </div>
        </div>
      </aside>
      <main class="data-view mediainfo-view" style="--wails-drop-target: drop">
        <div v-if="loading" class="empty-state">
          <div class="empty-icon">🔍</div>
          <h3>正在分析...</h3>
          <p>调用 FFprobe 解析媒体元数据</p>
        </div>
        <div v-else-if="errorMsg" class="empty-state">
          <div class="empty-icon">⚠️</div>
          <h3>分析失败</h3>
          <p style="color: #ef4444;">{{ errorMsg }}</p>
        </div>
        <div v-else-if="!mediaInfo" class="empty-state">
          <div class="empty-icon">📊</div>
          <h3>等待导入文件</h3>
          <p>选择一个媒体文件查看完整元数据</p>
        </div>
        <div v-else class="info-cards">
          <div class="card">
            <div class="card-header" @click="toggleCollapse('general')">
              <span class="card-title">📋 通用信息</span>
              <span class="collapse-icon">{{ isCollapsed('general') ? '▶' : '▼' }}</span>
            </div>
            <div v-show="!isCollapsed('general')" class="card-body">
              <div class="kv-row"><span class="kv-label">文件路径</span><span class="kv-value path">{{ mediaInfo.filePath }}</span></div>
              <div class="kv-row"><span class="kv-label">容器格式</span><span class="kv-value">{{ mediaInfo.format }}</span></div>
              <div class="kv-row"><span class="kv-label">文件大小</span><span class="kv-value">{{ mediaInfo.fileSize }}</span></div>
              <div class="kv-row"><span class="kv-label">时长</span><span class="kv-value">{{ mediaInfo.duration }}</span></div>
              <div class="kv-row"><span class="kv-label">总比特率</span><span class="kv-value">{{ mediaInfo.bitRate || '未知' }}</span></div>
            </div>
          </div>

          <div class="card" v-if="mediaInfo.video.length > 0">
            <div class="card-header" @click="toggleCollapse('video')">
              <span class="card-title">🎬 视频流 ({{ mediaInfo.video.length }})</span>
              <span class="collapse-icon">{{ isCollapsed('video') ? '▶' : '▼' }}</span>
            </div>
            <div v-show="!isCollapsed('video')" class="card-body">
              <div v-for="(v, i) in mediaInfo.video" :key="i" :class="{ 'stream-divider': i > 0 }">
                <div class="stream-badge">#{{ v.index }}</div>
                <div class="kv-row"><span class="kv-label">编码</span><span class="kv-value">{{ v.codec }}</span></div>
                <div class="kv-row"><span class="kv-label">分辨率</span><span class="kv-value">{{ v.width }}×{{ v.height }}</span></div>
                <div class="kv-row"><span class="kv-label">帧率</span><span class="kv-value">{{ v.frameRate }}</span></div>
                <div class="kv-row"><span class="kv-label">比特率</span><span class="kv-value">{{ v.bitRate || '未知' }}</span></div>
                <div class="kv-row"><span class="kv-label">像素格式</span><span class="kv-value">{{ v.pixelFormat }}</span></div>
                <div class="kv-row"><span class="kv-label">编码等级</span><span class="kv-value">{{ v.profile }}</span></div>
              </div>
            </div>
          </div>

          <div class="card" v-if="mediaInfo.audio.length > 0">
            <div class="card-header" @click="toggleCollapse('audio')">
              <span class="card-title">🎵 音频流 ({{ mediaInfo.audio.length }})</span>
              <span class="collapse-icon">{{ isCollapsed('audio') ? '▶' : '▼' }}</span>
            </div>
            <div v-show="!isCollapsed('audio')" class="card-body">
              <div v-for="(a, i) in mediaInfo.audio" :key="i" :class="{ 'stream-divider': i > 0 }">
                <div class="stream-badge">#{{ a.index }}</div>
                <div class="kv-row"><span class="kv-label">编码</span><span class="kv-value">{{ a.codec }}</span></div>
                <div class="kv-row"><span class="kv-label">采样率</span><span class="kv-value">{{ a.sampleRate }} Hz</span></div>
                <div class="kv-row"><span class="kv-label">声道</span><span class="kv-value">{{ getChannelLabel(a.channels) }}</span></div>
                <div class="kv-row"><span class="kv-label">比特率</span><span class="kv-value">{{ a.bitRate || '未知' }}</span></div>
                <div class="kv-row"><span class="kv-label">语言</span><span class="kv-value">{{ a.language }}</span></div>
              </div>
            </div>
          </div>

          <div class="card" :class="{ 'card-dimmed': mediaInfo.subtitle.length === 0 }">
            <div class="card-header" @click="toggleCollapse('subtitle')">
              <span class="card-title">📝 字幕流 ({{ mediaInfo.subtitle.length }})</span>
              <span class="collapse-icon">{{ isCollapsed('subtitle') ? '▶' : '▼' }}</span>
            </div>
            <div v-show="!isCollapsed('subtitle')" class="card-body">
              <div v-if="mediaInfo.subtitle.length > 0">
                <div v-for="(s, i) in mediaInfo.subtitle" :key="i" :class="{ 'stream-divider': i > 0 }">
                  <div class="stream-badge">#{{ s.Index }}</div>
                  <div class="kv-row"><span class="kv-label">编码</span><span class="kv-value">{{ s.Codec }}</span></div>
                  <div class="kv-row"><span class="kv-label">语言</span><span class="kv-value">{{ s.Language }}</span></div>
                </div>
              </div>
              <div v-else class="no-stream-msg">
                ⚠️ 未检测到内封字幕流<br><span class="tiny-hint">硬烧录（hardsub）字幕已渲染到视频帧中，无法被探测</span>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </main>
</template>

<style scoped>
.mediainfo-view {
  overflow-y: auto;
  background: #f1f5f9;
}
.info-cards {
  padding: 16px 24px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.card {
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
  overflow: hidden;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  user-select: none;
  border-bottom: 1px solid #f1f5f9;
}
.card-header:hover {
  background: #f8fafc;
}
.card-title {
  font-weight: 600;
  font-size: 14px;
  color: #1e293b;
}
.collapse-icon {
  font-size: 10px;
  color: #94a3b8;
}
.card-body {
  padding: 8px 16px 16px;
}
.kv-row {
  display: flex;
  padding: 5px 0;
  font-size: 13px;
  line-height: 1.5;
}
.kv-label {
  flex: 0 0 100px;
  color: #64748b;
  font-weight: 500;
}
.kv-value {
  flex: 1;
  color: #1e293b;
  font-weight: 400;
}
.kv-value.path {
  word-break: break-all;
  font-size: 12px;
}
.stream-badge {
  display: inline-block;
  background: #e2e8f0;
  color: #475569;
  font-size: 11px;
  font-weight: 600;
  padding: 1px 8px;
  border-radius: 4px;
  margin-bottom: 4px;
}
.stream-divider {
  padding-top: 10px;
  border-top: 1px solid #f1f5f9;
  margin-top: 6px;
}
.file-path-display {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 8px 10px;
  font-size: 11px;
  color: #475569;
  line-height: 1.4;
}
.card-dimmed {
  opacity: 0.65;
}
.no-stream-msg {
  padding: 12px 0;
  text-align: center;
  color: #94a3b8;
  font-size: 13px;
  line-height: 1.8;
}
</style>
