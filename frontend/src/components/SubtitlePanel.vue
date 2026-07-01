<script setup lang="ts">
import { useSubtitleStore } from '../stores/subtitle'

const store = useSubtitleStore()
</script>

<template>
  <main class="app-layout">
    <div class="workspace">
      <aside class="control-sidebar">
        <div class="panel-content">
          <div class="input-field" style="margin-bottom: 20px;">
            <label>处理模式</label>
            <div class="segmented-control mini-segmented">
              <button :class="{ active: store.currentSubMode === 'Extract' }" @click="store.currentSubMode = 'Extract'" style="flex: 1;">视频提取</button>
              <button :class="{ active: store.currentSubMode === 'Convert' }" @click="store.currentSubMode = 'Convert'" style="flex: 1;">格式互转</button>
            </div>
          </div>

          <template v-if="store.currentSubMode === 'Extract'">
            <h2 class="section-title">剥离设定</h2>
            <div class="input-stack">
              <div class="input-field"><label>输出格式</label>
                <select v-model="store.subSettings.format" class="native-select">
                  <option value="srt">通用 SubRip 格式 (.srt)</option>
                  <option value="ass">高级特效字幕 (.ass)</option>
                  <option value="ssa">标准特效字幕 (.ssa)</option>
                  <option value="vtt">网页 HTML5 格式 (.vtt)</option>
                  <option value="ttml">广电/流媒体标准 (.ttml / .xml)</option>
                  <option value="smi">微软 SAMI 格式 (.smi)</option>
                </select>
                <div class="divider" style="margin: 16px 0;"></div>
                <div class="input-field"><label>输出目录 (留空为原视频同级)</label>
                  <input v-model="store.subSettings.outputDir" placeholder="如: C:\Subtitles" style="font-size: 12px; color: #64748b;" />
                </div>
              </div>
            </div>
          </template>

          <template v-if="store.currentSubMode === 'Convert'">
            <h2 class="section-title">转换设定</h2>
            <div class="input-stack">
              <div class="input-field"><label>全局目标格式</label>
                <select v-model="store.subConvertSettings.globalTargetFormat" class="native-select">
                  <option value="srt">通用 SRT 格式 (.srt)</option>
                  <option value="ass">特效 ASS 格式 (.ass)</option>
                  <option value="vtt">网页 VTT 格式 (.vtt)</option>
                  <option value="ssa">标准特效字幕 (.ssa)</option>
                  <option value="ttml">广电/流媒体标准 (.ttml)</option>
                  <option value="smi">微软 SAMI 格式 (.smi)</option>
                </select>
              </div>
              <div class="divider" style="margin: 16px 0;"></div>
              <div class="input-field"><label>输出目录</label>
                <input v-model="store.subConvertSettings.outputDir" placeholder="(必填) 如不填请单独指定文件名" style="font-size: 12px; color: #64748b;" />
                <p class="tiny-hint" style="margin-top:4px; color:#ef4444;">必须配置此项或指定新文件名。如果批量转换请指定输出目录。</p>
              </div>
            </div>
          </template>
        </div>
        <div class="global-actions">
          <template v-if="store.currentSubMode === 'Extract'">
            <button class="btn btn-primary w-full shadow-sm" :disabled="store.subTasks.length === 0" @click="store.extractAllTasks">一键提取所有勾选</button>
            <div class="action-row mt-2"><button class="btn btn-secondary w-full" @click="$emit('select-files')">选择视频</button></div>
          </template>
          <template v-if="store.currentSubMode === 'Convert'">
            <button class="btn btn-primary w-full shadow-sm" :disabled="store.subConvertTasks.length === 0" @click="store.runAllSubConvert">一键全部转换</button>
            <div class="action-row mt-2"><button class="btn btn-secondary w-full" @click="$emit('select-files')">选择字幕文件</button></div>
          </template>
        </div>
      </aside>
      <main class="data-view" style="--wails-drop-target: drop">
        <template v-if="store.currentSubMode === 'Extract'">
          <div v-if="store.subTasks.length === 0" class="empty-state">
            <div class="empty-icon">📝</div><h3>等待导入视频</h3><p>选择带字幕流的 MKV/MP4等视频文件至此区域扫描</p>
          </div>
          <div v-else class="table-container">
            <div class="table-toolbar">
              <button class="btn btn-warning-soft" @click="store.clearErrorSubTasks" v-if="store.hasErrorSubTasks">清除无效任务</button>
              <button class="btn btn-danger-soft" @click="store.clearSubTasks">清空列表</button>
            </div>
            <table class="native-table">
              <thead><tr><th style="width: 30%;">视频文件</th><th style="width: 40%;">发现的字幕流 (勾选提取)</th><th style="width: 15%;">状态</th><th style="width: 15%; text-align: right; padding-right: 20px;">操作</th></tr></thead>
              <tbody>
                <tr v-for="(task, index) in store.subTasks" :key="task.id" :class="{'row-error': task.status === 'error' || task.status === 'no_sub'}">
                  <td class="col-new-name"><span class="text-truncate font-medium" :title="task.name">{{ task.name }}</span></td>
                  <td>
                    <div v-if="task.status === 'scanning'" class="progress-display processing"><span class="pulsing-dot"></span> 深度扫描中...</div>
                    <div v-else-if="task.status === 'no_sub'" style="color: #94a3b8; font-size: 12px;">未探测到内封字幕</div>
                    <div v-else-if="task.status === 'error'" style="color: #ef4444; font-size: 12px;">探测失败</div>
                    <div v-else class="chips-group">
                      <label v-for="stream in task.streams" :key="stream.Index"
                             class="chip"
                             :class="{ 'active': task.selectedStreams.includes(stream.Index), 'chip-warn': stream.Codec.toLowerCase() === 'microdvd' || stream.Codec.toLowerCase() === 'dvd_subtitle' }"
                             :title="(stream.Codec.toLowerCase() === 'microdvd' || stream.Codec.toLowerCase() === 'dvd_subtitle') ? '该类型需使用更专业工具来进行转换' : ''">
                        <input type="checkbox" :value="stream.Index" v-model="task.selectedStreams" style="display: none;" />
                        <span style="margin-right: 4px; font-weight: bold; opacity: 0.5;">#{{ stream.Index }}</span>
                        {{ stream.Language.toUpperCase() }} ({{ stream.Codec }})
                        <span v-if="stream.Codec.toLowerCase() === 'microdvd' || stream.Codec.toLowerCase() === 'dvd_subtitle'"> ⚠️</span>
                      </label>
                    </div>
                  </td>
                  <td>
                    <div class="progress-display" :class="task.status">
                      <span class="progress-text" :title="task.progressText">{{ task.progressText }}</span>
                    </div>
                  </td>
                  <td class="col-actions">
                    <button v-if="task.status === 'ready'" class="btn-action primary" @click="store.extractSubtitles(task)" :disabled="task.selectedStreams.length === 0">提取</button>
                    <button class="btn-action danger" @click="store.removeSubTask(index)">移除</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>

        <template v-if="store.currentSubMode === 'Convert'">
          <div v-if="store.subConvertTasks.length === 0" class="empty-state">
            <div class="empty-icon">🔄</div><h3>等待导入字幕</h3><p>选择 .srt, .ass, .vtt 等文件至此区域</p>
          </div>
          <div v-else class="table-container">
            <div class="table-toolbar">
              <button class="btn btn-danger-soft" @click="store.clearSubConvertTasks">清空列表</button>
            </div>
            <table class="native-table">
              <thead><tr><th style="width: 25%;">原字幕文件</th><th style="width: 20%;">新文件名(选填)</th><th style="width: 15%;">目标格式</th><th style="width: 25%;">状态</th><th style="width: 15%; text-align: right; padding-right: 20px;">操作</th></tr></thead>
              <tbody>
                <tr v-for="(task, index) in store.subConvertTasks" :key="task.id" :class="{'row-error': task.status === 'error'}">
                  <td class="col-old-name"><span class="text-truncate font-medium" :title="task.name">{{ task.name }}</span></td>
                  <td><input v-model="task.outputName" class="native-select mini" style="width:100%; box-sizing:border-box; text-align:left; font-weight:normal; color:#1e293b; background:#f8fafc;" placeholder="不填则保持原名" :disabled="task.status === 'processing'" /></td>
                  <td>
                    <select v-model="task.targetFormat" class="native-select mini" :disabled="task.status === 'processing'">
                      <option v-for="fmt in store.supportedSubFormats" :key="fmt" :value="fmt">{{ fmt.toUpperCase() }}</option>
                    </select>
                  </td>
                  <td>
                    <div class="progress-display" :class="task.status">
                      <span v-if="task.status === 'processing'" class="pulsing-dot"></span>
                      <span class="progress-text" :title="task.progressText">{{ task.progressText }}</span>
                    </div>
                  </td>
                  <td class="col-actions">
                    <button v-if="task.status !== 'success' && task.status !== 'processing'" class="btn-action primary" @click="store.runSubConvert(task)">转换</button>
                    <button class="btn-action danger" @click="store.removeSubConvertTask(index)">移除</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </main>
    </div>
  </main>
</template>
