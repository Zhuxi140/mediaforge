<script setup lang="ts">
import { ref } from 'vue'
import { useMediaStore } from '../stores/media'
import { useModal } from '../composables/useModal'
import { GetMediaInfo } from '../../wailsjs/go/controller/MediaApp'
import { media } from '../../wailsjs/go/models'

const store = useMediaStore()
const { showModal } = useModal()

const selectedInfo = ref<media.MediaInfo | null>(null)
const infoLoading = ref(false)
const infoError = ref('')
const infoFile = ref('')

async function showMetadata(path: string) {
  infoFile.value = path
  infoLoading.value = true
  infoError.value = ''
  selectedInfo.value = null
  try {
    const info = await GetMediaInfo(path)
    if (!info) {
      infoError.value = '无法获取元数据'
    } else {
      selectedInfo.value = info
    }
  } catch (err: any) {
    infoError.value = String(err)
  } finally {
    infoLoading.value = false
  }
}

function getChannelLabel(n: number) {
  if (n === 1) return '单声道'
  if (n === 2) return '立体声'
  if (n === 6) return '5.1 环绕'
  if (n === 8) return '7.1 环绕'
  return n + ' 声道'
}

const handleStartAll = async () => {
  const pending = store.mediaTasks.filter(t => t.status === 'pending' || t.status === 'error' || t.progressTime === '⏸️ 等待处理字幕').length
  if (pending === 0) return
  const confirmed = await showModal('确认执行', `即将启动 ${pending} 个转换任务，是否继续？`, 'warning', true)
  if (confirmed) store.startAllMediaTasks()
}

const qualityOptions = [
  { label: '极速/普清 (Low)', value: 'Low' },
  { label: '平衡 (Medium)', value: 'Medium' },
  { label: '高质量 (High)', value: 'High' },
  { label: '自定义 (Custom)', value: 'Custom' }
]
const hwAccelOptions = [
  { label: 'CPU 软解 (兼容性最强)', value: 'cpu' },
  { label: 'NVIDIA 显卡加速 (NVENC)', value: 'nvidia' },
  { label: 'Intel 核显加速 (QSV)', value: 'intel' },
  { label: 'AMD 显卡加速 (AMF)', value: 'amd' },
  { label: 'Apple Mac 加速', value: 'mac' }
]
</script>

<template>
  <main class="app-layout">
    <div class="workspace">
      <aside class="control-sidebar">
        <div class="panel-content">
          <h2 class="section-title">全局转换设定</h2>
          <div class="input-stack">
            <div class="input-field"><label>媒体类型</label>
              <div class="segmented-control mini-segmented">
                <button :class="{ active: store.mediaSettings.mediaType === 'video' }" @click="store.mediaSettings.mediaType = 'video'" style="flex: 1;">视频处理</button>
                <button :class="{ active: store.mediaSettings.mediaType === 'audio' }" @click="store.mediaSettings.mediaType = 'audio'" style="flex: 1;">音频提取</button>
              </div>
            </div>
            <div class="input-field"><label>目标格式</label>
              <select v-model="store.mediaSettings.globalTargetFormat" class="native-select">
                <option v-for="fmt in store.currentFormatOptions" :key="fmt" :value="fmt">{{ fmt.toUpperCase() }}</option>
              </select>
            </div>
            <div v-if="store.mediaSettings.mediaType === 'video'" class="input-field mt-1"><label>编码加速引擎</label>
              <select v-model="store.mediaSettings.hwAccel" class="native-select" style="font-weight: 500; color: #2563eb;">
                <option v-for="opt in hwAccelOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </div>
            <div class="input-field mt-1"><label>输出质量参数</label>
              <select v-model="store.mediaSettings.globalQuality" class="native-select">
                <option v-for="q in qualityOptions" :key="q.value" :value="q.value">{{ q.label }}</option>
              </select>
            </div>
            <div v-if="store.mediaSettings.globalQuality === 'Custom'" class="geek-panel">
              <div v-if="store.mediaSettings.mediaType === 'video'" class="input-field"><label>视频 CRF</label>
                <p class="tiny-hint">画质参数, 越小越清，推荐(18-28)</p>
                <input v-model="store.mediaSettings.videoCRF" type="number" min="0" max="51" class="code-input" />
              </div>
              <div v-if="store.mediaSettings.mediaType === 'audio'" class="input-field mt-2"><label>音频 VBR</label>
                <p class="tiny-hint">音质参数, 越小越好，推荐(0-5)</p>
                <input v-model="store.mediaSettings.audioVBR" type="number" min="0" max="9" class="code-input" />
              </div>
            </div>
            <div class="divider" style="margin: 16px 0;"></div>
            <div class="input-field"><label>输出目录 (留空为原文件同级)</label>
              <input v-model="store.mediaSettings.outputDir" placeholder="如: C:\Videos\Output" style="font-size: 12px; color: #64748b;" />
            </div>
          </div>
        </div>
        <div class="global-actions">
          <button class="btn btn-primary w-full shadow-sm" :disabled="store.mediaTasks.length === 0" @click="handleStartAll">一键全部转换</button>
          <div class="action-row mt-2">
            <button class="btn btn-secondary w-full" @click="$emit('select-files')">选择媒体文件</button>
          </div>
        </div>
      </aside>
      <main class="data-view" style="--wails-drop-target: drop">
        <div v-if="store.mediaTasks.length === 0" class="empty-state">
          <div class="empty-icon">🎞</div><h3>等待导入媒体</h3><p>选择视频或音频文件至此区域开始转换</p>
        </div>
        <div v-else class="table-container">
          <div class="table-toolbar">
            <button class="btn btn-warning-soft" @click="store.clearErrorTasks" v-if="store.hasErrorTasks">清除失败任务</button>
            <button class="btn btn-danger-soft" @click="store.clearMediaTasks">清空列表</button>
          </div>
          <table class="native-table">
            <thead><tr><th style="width: 35%;">文件名称</th><th style="width: 15%;">目标格式</th><th style="width: 35%;">转换进度</th><th style="width: 15%; text-align: right; padding-right: 20px;">操作</th></tr></thead>
            <tbody>
              <tr v-for="(task, index) in store.mediaTasks" :key="task.id" :class="{'row-error': task.status === 'error'}">
                <td class="col-new-name"><span class="text-truncate font-medium" :title="task.name">{{ task.name }}</span></td>
                <td>
                  <select v-model="task.targetFormat" class="native-select mini" :disabled="task.status === 'processing'">
                    <option v-for="fmt in store.currentFormatOptions" :key="fmt" :value="fmt">{{ fmt.toUpperCase() }}</option>
                  </select>
                </td>
                <td>
                  <div class="progress-display" :class="task.status">
                    <span v-if="task.status === 'processing'" class="pulsing-dot"></span>
                    <span class="progress-text" :title="task.errorMessage || task.progressTime">{{ task.progressTime }}</span>
                  </div>
                </td>
                <td class="col-actions">
                  <button v-if="task.status !== 'processing' && task.status !== 'success'" class="btn-action" @click="showMetadata(task.path)" title="查看元数据">元数据</button>
                  <button v-if="task.status !== 'processing' && task.status !== 'success'" class="btn-action primary" @click="store.startMediaTask(task)">▶️ 转换</button>
                  <button v-if="task.status === 'processing'" class="btn-action danger" @click="store.stopMediaTask(task)">⏹️ 停止</button>
                  <button v-if="task.status !== 'processing'" class="btn-action danger" @click="store.removeMediaTask(index)">移除</button>
                </td>
              </tr>
            </tbody>
          </table>

          <div v-if="infoLoading" class="empty-state" style="padding: 20px 0;">
            <div class="empty-icon">🔍</div><h3>正在分析...</h3>
          </div>
          <div v-else-if="infoError" class="empty-state" style="padding: 20px 0;">
            <div class="empty-icon">⚠️</div><p style="color:#ef4444;">{{ infoError }}</p>
          </div>
          <div v-else-if="selectedInfo" class="meta-card">
            <div class="meta-card-header">
              <h4 style="margin:0;font-size:14px;font-weight:600;color:#1e293b;">📊 元数据 — {{ infoFile.split('\\').pop()?.split('/').pop() }}</h4>
              <button class="btn btn-action" @click="selectedInfo = null">关闭</button>
            </div>
            <div class="meta-card-body">
              <div class="kv-row"><span class="kv-label">容器格式</span><span class="kv-value">{{ selectedInfo.format }}</span></div>
              <div class="kv-row"><span class="kv-label">文件大小</span><span class="kv-value">{{ selectedInfo.fileSize }}</span></div>
              <div class="kv-row"><span class="kv-label">时长</span><span class="kv-value">{{ selectedInfo.duration }}</span></div>
              <div class="kv-row"><span class="kv-label">总比特率</span><span class="kv-value">{{ selectedInfo.bitRate || '未知' }}</span></div>
              <template v-for="(vs, i) in selectedInfo.video" :key="'v'+i">
                <div class="kv-divider"></div>
                <div class="stream-badge">#{{ vs.index }}</div>
                <div class="kv-row"><span class="kv-label">视频编码</span><span class="kv-value">{{ vs.codec }} · {{ vs.width }}×{{ vs.height }} · {{ vs.frameRate }}</span></div>
                <div class="kv-row"><span class="kv-label">比特率</span><span class="kv-value">{{ vs.bitRate || '未知' }}</span></div>
                <div class="kv-row"><span class="kv-label">像素格式</span><span class="kv-value">{{ vs.pixelFormat }}</span></div>
              </template>
              <template v-for="(as, i) in selectedInfo.audio" :key="'a'+i">
                <div class="kv-divider"></div>
                <div class="stream-badge">#{{ as.index }}</div>
                <div class="kv-row"><span class="kv-label">音频编码</span><span class="kv-value">{{ as.codec }} · {{ getChannelLabel(as.channels) }}</span></div>
                <div class="kv-row"><span class="kv-label">采样率</span><span class="kv-value">{{ as.sampleRate }} Hz</span></div>
                <div class="kv-row"><span class="kv-label">比特率</span><span class="kv-value">{{ as.bitRate || '未知' }}</span></div>
                <div class="kv-row"><span class="kv-label">语言</span><span class="kv-value">{{ as.language }}</span></div>
              </template>
              <template v-if="selectedInfo.subtitle && selectedInfo.subtitle.length > 0">
                <div class="kv-divider"></div>
                <div v-for="(ss, i) in selectedInfo.subtitle" :key="'s'+i" style="margin-bottom:4px;">
                  <div class="stream-badge">#{{ ss.Index }}</div>
                  <div class="kv-row"><span class="kv-label">字幕编码</span><span class="kv-value">{{ ss.Codec }} · {{ ss.Language }}</span></div>
                </div>
              </template>
            </div>
          </div>
        </div>
      </main>
    </div>
  </main>
</template>

<style scoped>
.meta-card {
  background: #fff;
  border-radius: 10px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
  margin-top: 16px;
  overflow: hidden;
}
.meta-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f1f5f9;
  background: #f8fafc;
}
.meta-card-body {
  padding: 8px 16px 16px;
}
.kv-row {
  display: flex;
  padding: 5px 0;
  font-size: 13px;
  line-height: 1.5;
}
.kv-label {
  flex: 0 0 90px;
  color: #64748b;
  font-weight: 500;
}
.kv-value {
  flex: 1;
  color: #1e293b;
}
.kv-divider {
  border-top: 1px solid #f1f5f9;
  margin: 8px 0 4px;
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
</style>
