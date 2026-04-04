<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed } from 'vue';
import {
  SelectFiles, PreviewRename, ApplyRename, QuickRename,
  SubmitMediaTask, CancelMediaTask, ScanSubtitles, ExtractSubtitle,
  CheckInputFile // ✨ 引入后端的格式雷达
} from '../wailsjs/go/main/App';
import { OnFileDrop, OnFileDropOff, EventsOn, EventsOff } from '../wailsjs/runtime/runtime';

// ==========================================
// 💎 现代化自定义弹窗系统 (Custom Modal API)
// ==========================================
let resolveModal: ((value: boolean) => void) | null = null;
const modalState = ref({
  visible: false,
  title: '',
  message: '',
  type: 'info' as 'success' | 'warning' | 'error' | 'info',
  isConfirm: false
});

const showModal = (title: string, message: string, type: 'success' | 'warning' | 'error' | 'info' = 'info', isConfirm = false): Promise<boolean> => {
  modalState.value = { visible: true, title, message, type, isConfirm };
  return new Promise((resolve) => { resolveModal = resolve; });
};

const handleModalConfirm = () => { modalState.value.visible = false; if (resolveModal) resolveModal(true); };
const handleModalCancel = () => { modalState.value.visible = false; if (resolveModal) resolveModal(false); };


// ==========================================
// 🌌 App Shell 全局状态
// ==========================================
const activeApp = ref<'renamer' | 'ffmpeg' | 'subtitle'>('subtitle'); // 默认打开字幕便于测试

// ==========================================
// 🗂️ 模块一：批量重命名 (Renamer)
// ==========================================
const filePaths = ref<string[]>([]);
const previews = ref<any[]>([]);
const currentMode = ref<'BasicMode' | 'SmartMode'>('BasicMode');
const rule = ref({ Mode: 'BasicMode', Prefix: '', Suffix: '', ReplaceOld: '', ReplaceNew: '', SmartRules: [{ Name: 'id', Pattern: '202\\d{8}' }, { Name: 'class', Pattern: '大数据.*?[0-9]*班' }, { Name: 'name', Pattern: '\\p{Han}{2,4}' }], SmartTemplate: '{class}-{name}_{id}', CleanChars: '-_ ' });
const cleanCharOptions = [ { label: '空格', value: ' ' }, { label: '减号 (-)', value: '-' }, { label: '下划线 (_)', value: '_' }, { label: '点号 (.)', value: '.' }, { label: '逗号 (,)', value: ',' } ];
const selectedChars = ref([' ', '-', '_']);
watch(selectedChars, (newVal) => { rule.value.CleanChars = newVal.join(''); }, { immediate: true, deep: true });
watch(currentMode, (newMode) => { rule.value.Mode = newMode; });
const addSmartRule = () => { rule.value.SmartRules.push({ Name: '', Pattern: '' }); };
const removeSmartRule = (index: number) => { rule.value.SmartRules.splice(index, 1); };
const handleSelectFiles = async () => { const files = await SelectFiles(); if (files && files.length > 0) { if (activeApp.value === 'renamer') { const newFiles = files.filter(f => !filePaths.value.includes(f)); filePaths.value = [...filePaths.value, ...newFiles]; } else if (activeApp.value === 'ffmpeg') { addMediaFiles(files); } else if (activeApp.value === 'subtitle') { addSubtitleFiles(files); } } };
const clearList = () => { filePaths.value = []; previews.value = []; };
const removeFile = (index: number) => { filePaths.value.splice(index, 1); };
const refreshList = async () => { if (filePaths.value.length > 0) filePaths.value = [...filePaths.value]; };
watch([filePaths, rule], async () => { if (filePaths.value.length === 0) { previews.value = []; return; } try { previews.value = await PreviewRename(filePaths.value, rule.value as any) || []; } catch (err) { console.error("生成预览失败:", err); } }, { deep: true });
const validPreviews = computed(() => previews.value.filter(p => !p.hasConflict && !p.formatError && p.originalName !== p.newName));
const canExecute = computed(() => validPreviews.value.length > 0);

const handleExecute = async () => {
  if (!canExecute.value) return;
  try {
    const msg = await ApplyRename(validPreviews.value as any);
    if (msg === "success") {
      showModal('执行成功', `🎉 成功重命名 ${validPreviews.value.length} 个文件！`, 'success');
      const validPaths = validPreviews.value.map(p => p.originalPath);
      filePaths.value = filePaths.value.filter(p => !validPaths.includes(p));
    } else {
      showModal('执行失败', msg, 'error');
    }
  } catch (error: any) {
    showModal('系统异常', error, 'error');
  }
};

const editingIndex = ref<number>(-1);
const editTempName = ref<string>('');
const startEdit = (index: number, currentName: string) => { editingIndex.value = index; editTempName.value = currentName; };
const cancelEdit = () => { editingIndex.value = -1; editTempName.value = ''; };
const saveEdit = async (index: number, oldPath: string) => {
  if (!editTempName.value.trim() || editTempName.value === previews.value[index].originalName) { cancelEdit(); return; }
  try {
    filePaths.value[index] = await QuickRename(oldPath, editTempName.value);
    cancelEdit();
  } catch (err: any) {
    showModal('修改失败', err, 'error');
  }
};


// ==========================================
// 🎬 模块二：影音工厂 (FFmpeg)
// ==========================================
interface MediaTask { id: string; path: string; name: string; targetFormat: string; status: 'pending' | 'processing' | 'success' | 'error'; progressTime: string; errorMessage?: string; }
const mediaTasks = ref<MediaTask[]>([]);
const videoFormats = ['mp4', 'mkv', 'avi', 'gif'];
const audioFormats = ['mp3', 'wav', 'aac', 'flac'];
const qualityOptions = [ { label: '极速/普清 (Low)', value: 'Low' }, { label: '平衡 (Medium)', value: 'Medium' }, { label: '高质量 (High)', value: 'High' }, { label: '⚙️ 自定义 (Custom)', value: 'Custom' } ];
const hwAccelOptions = [ { label: 'CPU 软解 (兼容性最强)', value: 'cpu' }, { label: 'NVIDIA 显卡加速 (NVENC)', value: 'nvidia' }, { label: 'Intel 核显加速 (QSV)', value: 'intel' }, { label: 'AMD 显卡加速 (AMF)', value: 'amd' }, { label: 'Apple Mac 加速', value: 'mac' } ];
const mediaSettings = ref({ mediaType: 'video', globalTargetFormat: 'mp4', globalQuality: 'Medium', videoCRF: '23', audioVBR: '2', outputDir: '', hwAccel: 'cpu' });
const currentFormatOptions = computed(() => { return mediaSettings.value.mediaType === 'video' ? videoFormats : audioFormats; });

watch(() => mediaSettings.value.mediaType, (newType) => {
  mediaSettings.value.globalTargetFormat = newType === 'video' ? 'mp4' : 'mp3';
  mediaTasks.value.forEach(t => { if (t.status === 'pending') t.targetFormat = mediaSettings.value.globalTargetFormat; });
});

const generateId = () => Math.random().toString(36).substring(2, 9) + Date.now().toString(36);

// ✨ 完美安检版：影音工厂导入文件
const addMediaFiles = async (paths: string[]) => {
  for (const p of paths) {
    if (mediaTasks.value.some(task => task.path === p)) continue;

    const name = p.split('\\').pop() || p.split('/').pop() || p;

    mediaTasks.value.push({
      id: generateId(),
      path: p,
      name: name,
      targetFormat: mediaSettings.value.globalTargetFormat,
      status: 'pending',
      progressTime: '等待中...'
    });

    const proxyTask = mediaTasks.value[mediaTasks.value.length - 1];

    try {
      // 🚀 呼叫后端雷达瞬间鉴定格式
      const checkResult = await CheckInputFile(p);

      if (checkResult !== "success") {
        proxyTask.status = 'error';
        proxyTask.errorMessage = checkResult;
        proxyTask.progressTime = `❌ ${checkResult}`; // 精准回显被拦截的格式
      }
    } catch (err) {
      console.error("鉴定格式通信失败:", err);
    }
  }
};

const startMediaTask = async (task: MediaTask, forceDropSubtitle = false) => {
  if (task.status === 'processing') return;
  task.status = 'processing'; task.progressTime = '启动引擎...'; task.errorMessage = '';

  EventsOn('ffmpeg-progress-' + task.id, (timeStr: string) => { task.progressTime = `处理中: ${timeStr}`; });
  EventsOn('ffmpeg-done-' + task.id, () => { task.status = 'success'; task.progressTime = '✅ 完成'; clearMediaListeners(task.id); });
  EventsOn('ffmpeg-error-' + task.id, (err: string) => {
    task.status = 'error'; task.errorMessage = err;
    if (err === '任务已取消') {
      task.progressTime = '⏹️ 已取消';
    } else {
      task.progressTime = '❌ 引擎崩溃 (查看详情)';
      showModal('转码失败', `文件: ${task.name}\n\n原因: ${err}`, 'error');
    }
    clearMediaListeners(task.id);
  });

  const passCRF = mediaSettings.value.mediaType === 'video' ? String(mediaSettings.value.videoCRF) : "";
  const passVBR = mediaSettings.value.mediaType === 'audio' ? String(mediaSettings.value.audioVBR) : "";

  try {
    await SubmitMediaTask({
      ID: task.id,
      InputPath: task.path,
      OutputPath: mediaSettings.value.outputDir,
      Format: task.targetFormat,
      Quality: mediaSettings.value.globalQuality,
      VideoCRF: passCRF,
      AudioVBR: passVBR,
      HwAccel: mediaSettings.value.hwAccel,
      ForceDropSubtitle: forceDropSubtitle
    });
  } catch (err: any) {
    if (err === "WARN_SUBTITLE") {
      clearMediaListeners(task.id);
      const userAgree = await showModal(
          '⚠️ 预检拦截',
          `文件 [${task.name}]\n包含 MP4 不兼容的内封字幕流！\n强行转换可能导致程序崩溃或丢弃字幕。\n\n点击【强行丢弃】同意剥离字幕继续转换，或点击【取消】稍后单独提取。`,
          'warning',
          true
      );
      if (userAgree) { task.status = 'pending'; startMediaTask(task, true); } else { task.status = 'pending'; task.progressTime = '⏸️ 等待处理字幕'; }
    } else {
      task.status = 'error'; task.progressTime = `❌ ${err}`;
      showModal('系统异常', err, 'error');
      clearMediaListeners(task.id);
    }
  }
};

const stopMediaTask = (task: MediaTask) => { if (task.status === 'processing') { task.progressTime = '正在停止...'; CancelMediaTask(task.id); } };
const startAllMediaTasks = () => { mediaTasks.value.forEach(task => { if (task.status === 'pending' || task.status === 'error' || task.progressTime === '⏸️ 等待处理字幕') { startMediaTask(task); } }); };
const removeMediaTask = (index: number) => { const task = mediaTasks.value[index]; clearMediaListeners(task.id); mediaTasks.value.splice(index, 1); };
const clearMediaTasks = () => { mediaTasks.value.forEach(t => clearMediaListeners(t.id)); mediaTasks.value = []; };
const clearMediaListeners = (id: string) => { EventsOff('ffmpeg-progress-' + id); EventsOff('ffmpeg-done-' + id); EventsOff('ffmpeg-error-' + id); };
watch(() => mediaSettings.value.globalTargetFormat, (newFormat) => { mediaTasks.value.forEach(t => { if (t.status === 'pending') t.targetFormat = newFormat; }); });

// ✨ 影音工厂：一键清除错误任务
const hasErrorMediaTasks = computed(() => mediaTasks.value.some(t => t.status === 'error'));
const clearErrorMediaTasks = () => {
  mediaTasks.value.filter(t => t.status === 'error').forEach(t => clearMediaListeners(t.id));
  mediaTasks.value = mediaTasks.value.filter(t => t.status !== 'error');
};


// ==========================================
// 📝 模块三：字幕剥离工厂
// ==========================================
interface SubtitleStreamInfo { Index: string; Language: string; Codec: string; }
interface SubtitleTask { id: string; path: string; name: string; status: 'scanning' | 'ready' | 'processing' | 'success' | 'error' | 'no_sub'; streams: SubtitleStreamInfo[]; selectedStreams: string[]; progressText: string; }
const subTasks = ref<SubtitleTask[]>([]);
const subSettings = ref({ outputDir: '' });

// ✨ 显示真实拦截原因的扫描函数
const addSubtitleFiles = async (paths: string[]) => {
  for (const p of paths) {
    if (subTasks.value.some(t => t.path === p)) continue;

    subTasks.value.push({
      id: generateId(), path: p, name: p.split('\\').pop() || p.split('/').pop() || p,
      status: 'scanning', streams: [], selectedStreams: [], progressText: '🔍 扫描中...'
    });

    const proxyTask = subTasks.value[subTasks.value.length - 1];

    try {
      const streams = await ScanSubtitles(p);
      if (streams && streams.length > 0) {
        proxyTask.streams = streams;
        proxyTask.status = 'ready';
        proxyTask.progressText = `发现 ${streams.length} 条字幕`;
        proxyTask.selectedStreams = streams.map((s: any) => s.Index);
      } else {
        proxyTask.status = 'no_sub';
        proxyTask.progressText = '未发现字幕流';
      }
    } catch (err) {
      proxyTask.status = 'error';
      // 直接显示 Go 抛过来的错误文本！
      proxyTask.progressText = `❌ ${err}`;
    }
  }
};

// ✨ 绝杀死锁版：并发提取 + 状态回传机制
const extractSubtitles = (task: SubtitleTask) => {
  if (task.selectedStreams.length === 0) {
    showModal('提示', "请至少选择一条字幕流进行提取！", 'warning');
    return;
  }

  task.status = 'processing';
  task.progressText = '提取中...';

  const expectedCount = task.selectedStreams.length;
  let finishedCount = 0;
  let successCount = 0;

  // 定义统一结账函数
  const checkDone = () => {
    if (finishedCount === expectedCount) {
      task.status = 'success';
      task.progressText = `✅ 已提取 ${successCount} 条`;
    }
  };

  // 抛弃 await 阻塞，全量派发任务让后端并行处理
  for (const streamIdx of task.selectedStreams) {
    const subId = task.id + "_" + streamIdx;

    EventsOff('ffmpeg-done-' + subId);
    EventsOff('ffmpeg-error-' + subId);

    EventsOn('ffmpeg-done-' + subId, () => {
      successCount++;
      finishedCount++;
      checkDone();
    });

    EventsOn('ffmpeg-error-' + subId, (err) => {
      showModal('提取出错', String(err), 'error');
      finishedCount++;
      checkDone();
    });

    ExtractSubtitle(subId, task.path, streamIdx, subSettings.value.outputDir).catch(err => {
      console.error("IPC 提取调用失败:", err);
      finishedCount++;
      checkDone();
    });
  }
};

const extractAllTasks = () => { subTasks.value.forEach(task => { if (task.status === 'ready') extractSubtitles(task); }); };
const removeSubTask = (index: number) => { subTasks.value.splice(index, 1); };
const clearSubTasks = () => { subTasks.value = []; };

// ✨ 字幕工厂：一键清除错误或无字幕任务
const hasErrorSubTasks = computed(() => subTasks.value.some(t => t.status === 'error' || t.status === 'no_sub'));
const clearErrorSubTasks = () => {
  subTasks.value = subTasks.value.filter(t => t.status !== 'error' && t.status !== 'no_sub');
};

// ==========================================
// 🌌 全局拖拽接管
// ==========================================
onMounted(() => {
  OnFileDrop((x: number, y: number, paths: string[]) => {
    if (activeApp.value === 'renamer') { const newFiles = paths.filter(p => !filePaths.value.includes(p)); filePaths.value = [...filePaths.value, ...newFiles]; }
    else if (activeApp.value === 'ffmpeg') { addMediaFiles(paths); }
    else if (activeApp.value === 'subtitle') { addSubtitleFiles(paths); }
  }, true);
});
onUnmounted(() => { OnFileDropOff(); });
</script>

<template>
  <div class="app-root">

    <nav class="global-sidebar">
      <div class="brand"><div class="logo"><svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="url(#logo-grad)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><defs><linearGradient id="logo-grad" x1="0%" y1="0%" x2="100%" y2="100%"><stop offset="0%" stop-color="#3b82f6" /><stop offset="100%" stop-color="#8b5cf6" /></linearGradient></defs><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg></div></div>
      <div class="nav-menu">
        <button class="nav-item" :class="{ active: activeApp === 'renamer' }" @click="activeApp = 'renamer'"><span class="nav-icon"><svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 13.5V4a2 2 0 0 1 2-2h8.5L20 7.5V20a2 2 0 0 1-2 2h-5.5"/><polyline points="14 2 14 8 20 8"/><path d="M10.42 12.61a2.1 2.1 0 1 1 2.97 2.97L7.95 21 4 22l.99-3.95 5.43-5.44Z"/></svg></span><span class="nav-text">重命名</span></button>
        <button class="nav-item" :class="{ active: activeApp === 'ffmpeg' }" @click="activeApp = 'ffmpeg'"><span class="nav-icon"><svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.2 6 3 11l-.9-2.4c-.3-1.1.4-2.2 1.5-2.5l13.5-4c1.1-.3 2.2.4 2.5 1.5z"/><path d="m6.2 5.3 3.1 3.9"/><path d="m12.4 3.4 3.1 4"/><path d="M3 11h18v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z"/></svg></span><span class="nav-text">影音工厂</span></button>
        <button class="nav-item" :class="{ active: activeApp === 'subtitle' }" @click="activeApp = 'subtitle'"><span class="nav-icon"><svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/><line x1="9" y1="9" x2="15" y2="9"/><line x1="9" y1="13" x2="15" y2="13"/></svg></span><span class="nav-text">字幕剥离</span></button>
      </div>
    </nav>

    <main v-show="activeApp === 'renamer'" class="app-layout">
      <header class="app-header">
        <div class="header-title"><h1>批量重命名引擎</h1></div>
        <div class="segmented-control">
          <button :class="{ active: currentMode === 'BasicMode' }" @click="currentMode = 'BasicMode'">基础替换</button>
          <button :class="{ active: currentMode === 'SmartMode' }" @click="currentMode = 'SmartMode'">智能解析</button>
        </div>
      </header>
      <div class="workspace">
        <aside class="control-sidebar">
          <div v-if="currentMode === 'BasicMode'" class="panel-content">
            <h2 class="section-title">基础规则设定</h2>
            <div class="input-stack"><div class="input-field"><label>文件前缀</label><input v-model="rule.Prefix" placeholder="如: 2026_" /></div><div class="input-field"><label>文件后缀</label><input v-model="rule.Suffix" placeholder="如: _v1" /></div></div>
            <div class="divider"></div><h2 class="section-title">字符替换</h2>
            <div class="input-stack"><input v-model="rule.ReplaceOld" placeholder="查找特定字符..." /><div class="icon-arrow">↓</div><input v-model="rule.ReplaceNew" placeholder="替换为新字符..." /></div>
          </div>
          <div v-if="currentMode === 'SmartMode'" class="panel-content">
            <h2 class="section-title">智能提取变量</h2><p class="helper-text">利用正则无视文件名的混乱顺序提取关键信息。</p>
            <div class="rules-list"><div v-for="(r, index) in rule.SmartRules" :key="index" class="rule-card"><div class="rule-header"><span class="rule-num">{{ index + 1 }}</span><button class="btn-icon text-danger" @click="removeSmartRule(index)">✕</button></div><div class="rule-inputs"><input v-model="r.Name" placeholder="变量名" class="code-input var-name" /><input v-model="r.Pattern" placeholder="正则表达式" class="code-input pattern" /></div></div></div>
            <button class="btn-dashed w-full" @click="addSmartRule">＋ 新增提取变量</button>
            <div class="divider"></div><h2 class="section-title">重组装设定</h2>
            <div class="input-field"><label>目标格式模板</label><input v-model="rule.SmartTemplate" class="code-input highlight" placeholder="{class}-{name}_{id}" /></div>
            <div class="input-field mt-3"><label>预处理: 过滤干扰符</label><div class="chips-group"><label v-for="opt in cleanCharOptions" :key="opt.value" class="chip" :class="{ 'active': selectedChars.includes(opt.value) }"><input type="checkbox" :value="opt.value" v-model="selectedChars" style="display: none;" />{{ opt.label }}</label></div></div>
          </div>
          <div class="global-actions">
            <button class="btn btn-primary w-full shadow-sm" :disabled="!canExecute" @click="handleExecute">🚀 执行重命名 <span v-if="validPreviews.length" class="badge">{{ validPreviews.length }}</span></button>
            <div class="action-row mt-2">
              <button class="btn btn-secondary w-full" @click="handleSelectFiles">选择文件</button>
            </div>
          </div>
        </aside>
        <main class="data-view" style="--wails-drop-target: drop">
          <div v-if="previews.length === 0" class="empty-state"><div class="empty-icon">📂</div><h3>没有选择文件</h3><p>请点击左侧选择文件，或将文件拖拽至此区域</p></div>
          <div v-else class="table-container">
            <div class="table-toolbar">
              <button class="btn btn-secondary" @click="refreshList">重新扫描</button>
              <button class="btn btn-danger-soft" @click="clearList">清空列表</button>
            </div>
            <table class="native-table">
              <thead><tr><th style="width: 35%;">原文件名</th><th style="width: 30%;">预览新文件名</th><th style="width: 20%;">状态</th><th style="width: 15%; text-align: right; padding-right: 20px;">操作</th></tr></thead>
              <tbody>
              <tr v-for="(item, index) in previews" :key="index" :class="{'row-error': item.formatError}">
                <td class="col-old-name"><div v-if="editingIndex === index" class="inline-edit-wrapper"><input v-model="editTempName" @keyup.enter="saveEdit(index, item.originalPath)" @keyup.esc="cancelEdit" class="inline-input" autofocus /></div><span v-else class="text-truncate" :title="item.originalName">{{ item.originalName }}</span></td>
                <td class="col-new-name"><span class="text-truncate font-medium" :title="item.newName">{{ item.newName }}</span></td>
                <td><span v-if="item.formatError" class="status-badge error" :title="item.formatError">{{ item.formatError }}</span><span v-else-if="item.hasConflict" class="status-badge error">同名冲突</span><span v-else-if="item.originalName === item.newName" class="status-badge neutral">跳过</span><span v-else class="status-badge success">待修改</span></td>
                <td class="col-actions"><button v-if="editingIndex !== index && (item.formatError || item.hasConflict)" class="btn-action primary" @click="startEdit(index, item.originalName)">修正</button><button v-if="editingIndex === index" class="btn-action success" @click="saveEdit(index, item.originalPath)">保存</button><button v-if="editingIndex === index" class="btn-action danger" @click="cancelEdit">取消</button><button v-if="editingIndex !== index" class="btn-action danger" @click="removeFile(index)">移除</button></td>
              </tr>
              </tbody>
            </table>
          </div>
        </main>
      </div>
    </main>

    <main v-show="activeApp === 'ffmpeg'" class="app-layout">
      <header class="app-header"><div class="header-title"><h1>FFmpeg 影音铸造厂</h1></div></header>
      <div class="workspace">
        <aside class="control-sidebar">
          <div class="panel-content">
            <h2 class="section-title">全局转换设定</h2>
            <div class="input-stack">
              <div class="input-field"><label>媒体类型</label><div class="segmented-control mini-segmented"><button :class="{ active: mediaSettings.mediaType === 'video' }" @click="mediaSettings.mediaType = 'video'" style="flex: 1;">🎬 视频处理</button><button :class="{ active: mediaSettings.mediaType === 'audio' }" @click="mediaSettings.mediaType = 'audio'" style="flex: 1;">🎵 音频提取</button></div></div>
              <div class="input-field"><label>目标格式</label><select v-model="mediaSettings.globalTargetFormat" class="native-select"><option v-for="fmt in currentFormatOptions" :key="fmt" :value="fmt">{{ fmt.toUpperCase() }}</option></select></div>
              <div v-if="mediaSettings.mediaType === 'video'" class="input-field mt-1"><label>编码加速引擎</label><select v-model="mediaSettings.hwAccel" class="native-select" style="font-weight: 500; color: #2563eb;"><option v-for="opt in hwAccelOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option></select></div>
              <div class="input-field mt-1"><label>输出质量参数</label><select v-model="mediaSettings.globalQuality" class="native-select"><option v-for="q in qualityOptions" :key="q.value" :value="q.value">{{ q.label }}</option></select></div>
              <div v-if="mediaSettings.globalQuality === 'Custom'" class="geek-panel">
                <div v-if="mediaSettings.mediaType === 'video'" class="input-field"><label>视频 CRF</label><p class="tiny-hint">画质参数, 越小越清，损失越少，推荐(18-28)</p><input v-model="mediaSettings.videoCRF" type="number" min="0" max="51" class="code-input" /></div>
                <div v-if="mediaSettings.mediaType === 'audio'" class="input-field mt-2"><label>音频 VBR</label><p class="tiny-hint">音质参数, 越小越好，损失越少，推荐(0-5)</p><input v-model="mediaSettings.audioVBR" type="number" min="0" max="9" class="code-input" /></div>
              </div>
              <div class="divider" style="margin: 16px 0;"></div>
              <div class="input-field"><label>输出目录 (留空为原文件同级)</label><div style="display: flex; gap: 8px;"><input v-model="mediaSettings.outputDir" placeholder="如: C:\Videos\Output" style="font-size: 12px; color: #64748b;" /></div></div>
            </div>
          </div>
          <div class="global-actions">
            <button class="btn btn-primary w-full shadow-sm" :disabled="mediaTasks.length === 0" @click="startAllMediaTasks">▶️ 一键全部转换</button>
            <div class="action-row mt-2">
              <button class="btn btn-secondary w-full" @click="handleSelectFiles">选择媒体文件</button>
            </div>
          </div>
        </aside>
        <main class="data-view" style="--wails-drop-target: drop">
          <div v-if="mediaTasks.length === 0" class="empty-state"><div class="empty-icon">🎞️</div><h3>等待导入媒体</h3><p>拖拽视频或音频文件至此区域开始转换</p></div>
          <div v-else class="table-container">
            <div class="table-toolbar">
              <button class="btn btn-warning-soft" @click="clearErrorMediaTasks" v-if="hasErrorMediaTasks">清除失败任务</button>
              <button class="btn btn-danger-soft" @click="clearMediaTasks">清空列表</button>
            </div>
            <table class="native-table">
              <thead><tr><th style="width: 35%;">文件名称</th><th style="width: 15%;">目标格式</th><th style="width: 35%;">转换进度</th><th style="width: 15%; text-align: right; padding-right: 20px;">操作</th></tr></thead>
              <tbody>
              <tr v-for="(task, index) in mediaTasks" :key="task.id" :class="{'row-error': task.status === 'error'}">
                <td class="col-new-name"><span class="text-truncate font-medium" :title="task.name">{{ task.name }}</span></td>
                <td><select v-model="task.targetFormat" class="native-select mini" :disabled="task.status === 'processing'"><option v-for="fmt in currentFormatOptions" :key="fmt" :value="fmt">{{ fmt.toUpperCase() }}</option></select></td>
                <td>
                  <div class="progress-display" :class="task.status">
                    <span v-if="task.status === 'processing'" class="pulsing-dot"></span>
                    <span class="progress-text" :title="task.errorMessage || task.progressTime">{{ task.progressTime }}</span>
                  </div>
                </td>
                <td class="col-actions"><button v-if="task.status !== 'processing' && task.status !== 'success'" class="btn-action primary" @click="startMediaTask(task)">▶️ 转换</button><button v-if="task.status === 'processing'" class="btn-action danger" @click="stopMediaTask(task)">⏹️ 停止</button><button v-if="task.status !== 'processing'" class="btn-action danger" @click="removeMediaTask(index)">移除</button></td>
              </tr>
              </tbody>
            </table>
          </div>
        </main>
      </div>
    </main>

    <main v-show="activeApp === 'subtitle'" class="app-layout">
      <header class="app-header"><div class="header-title"><h1>智能字幕剥离工厂</h1></div></header>
      <div class="workspace">
        <aside class="control-sidebar">
          <div class="panel-content">
            <h2 class="section-title">剥离设定</h2>
            <div class="input-stack">
              <div class="input-field"><label>输出格式</label><select disabled class="native-select" style="background: #e2e8f0; color: #64748b; font-weight: bold;"><option>通用 SRT 格式 (.srt)</option></select><p class="tiny-hint" style="margin-top: 4px;">将各种内置字幕统一剥离为纯文本 SRT，方便修改或压制。</p></div>
              <div class="divider" style="margin: 16px 0;"></div>
              <div class="input-field"><label>输出目录 (留空为原视频同级)</label><div style="display: flex; gap: 8px;"><input v-model="subSettings.outputDir" placeholder="如: C:\Subtitles" style="font-size: 12px; color: #64748b;" /></div></div>
            </div>
          </div>
          <div class="global-actions">
            <button class="btn btn-primary w-full shadow-sm" :disabled="subTasks.length === 0" @click="extractAllTasks">📥 一键提取所有勾选</button>
            <div class="action-row mt-2">
              <button class="btn btn-secondary w-full" @click="handleSelectFiles">选择视频</button>
            </div>
          </div>
        </aside>
        <main class="data-view" style="--wails-drop-target: drop">
          <div v-if="subTasks.length === 0" class="empty-state"><div class="empty-icon">📝</div><h3>等待导入视频</h3><p>拖拽带字幕的 MKV/MP4 文件至此区域扫描</p></div>
          <div v-else class="table-container">
            <div class="table-toolbar">
              <button class="btn btn-warning-soft" @click="clearErrorSubTasks" v-if="hasErrorSubTasks">清除无效任务</button>
              <button class="btn btn-danger-soft" @click="clearSubTasks">清空列表</button>
            </div>
            <table class="native-table">
              <thead><tr><th style="width: 30%;">视频文件</th><th style="width: 40%;">发现的字幕流 (勾选提取)</th><th style="width: 15%;">状态</th><th style="width: 15%; text-align: right; padding-right: 20px;">操作</th></tr></thead>
              <tbody>
              <tr v-for="(task, index) in subTasks" :key="task.id" :class="{'row-error': task.status === 'error' || task.status === 'no_sub'}">
                <td class="col-new-name"><span class="text-truncate font-medium" :title="task.name">{{ task.name }}</span></td>
                <td>
                  <div v-if="task.status === 'scanning'" class="progress-display processing"><span class="pulsing-dot"></span> 深度扫描中...</div>
                  <div v-else-if="task.status === 'no_sub'" style="color: #94a3b8; font-size: 12px;">未探测到内封字幕</div>
                  <div v-else-if="task.status === 'error'" style="color: #ef4444; font-size: 12px;">探测被拦截</div>
                  <div v-else class="chips-group"><label v-for="stream in task.streams" :key="stream.Index" class="chip" :class="{ 'active': task.selectedStreams.includes(stream.Index) }"><input type="checkbox" :value="stream.Index" v-model="task.selectedStreams" style="display: none;" /><span style="margin-right: 4px; font-weight: bold; opacity: 0.5;">#{{ stream.Index }}</span> {{ stream.Language.toUpperCase() }} ({{ stream.Codec }})</label></div>
                </td>
                <td>
                  <div class="progress-display" :class="task.status">
                    <span class="progress-text" :title="task.progressText">{{ task.progressText }}</span>
                  </div>
                </td>
                <td class="col-actions"><button v-if="task.status === 'ready'" class="btn-action primary" @click="extractSubtitles(task)" :disabled="task.selectedStreams.length === 0">提取</button><button class="btn-action danger" @click="removeSubTask(index)">移除</button></td>
              </tr>
              </tbody>
            </table>
          </div>
        </main>
      </div>
    </main>

    <Transition name="modal-fade">
      <div v-if="modalState.visible" class="custom-modal-overlay" @mousedown.self="handleModalCancel">
        <div class="custom-modal-box">
          <div class="modal-header">
            <div class="modal-icon" :class="modalState.type">
              <svg v-if="modalState.type === 'success'" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
              <svg v-else-if="modalState.type === 'warning'" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>
              <svg v-else-if="modalState.type === 'error'" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="15" y1="9" x2="9" y2="15"></line><line x1="9" y1="9" x2="15" y2="15"></line></svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="16" x2="12" y2="12"></line><line x1="12" y1="8" x2="12.01" y2="8"></line></svg>
            </div>
            <h3 class="modal-title">{{ modalState.title }}</h3>
          </div>
          <div class="modal-body">{{ modalState.message }}</div>
          <div class="modal-footer">
            <button v-if="modalState.isConfirm" class="btn btn-secondary" @click="handleModalCancel">取消</button>
            <button class="btn btn-primary" :class="{'danger-btn': modalState.type === 'error' || modalState.type === 'warning'}" @click="handleModalConfirm">
              {{ modalState.isConfirm ? (modalState.type === 'warning' ? '强行丢弃' : '确定') : '我知道了' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

  </div>
</template>

<style>
/* ================= 全局样式与组件 ================= */
:global(body), :global(html) { margin: 0; padding: 0; height: 100vh; overflow: hidden; background-color: #f8fafc; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; color: #1e293b; -webkit-font-smoothing: antialiased; }
.app-root { display: flex; height: 100vh; width: 100vw; overflow: hidden; }

/* 左侧深色极简导航栏 */
.global-sidebar { width: 76px; min-width: 76px; background: #ffffff; display: flex; flex-direction: column; align-items: center; padding: 20px 0; z-index: 50; border-right: 1px solid #e2e8f0; -webkit-app-region: drag; flex-shrink: 0; }
.brand { margin-bottom: 32px; display: flex; justify-content: center; width: 100%;}
.logo { display: flex; align-items: center; justify-content: center; }
.nav-menu { display: flex; flex-direction: column; gap: 8px; flex: 1; -webkit-app-region: no-drag; width: 100%; padding: 0 10px; box-sizing: border-box;}
.nav-item { background: transparent; border: none; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px; color: #64748b; cursor: pointer; transition: all 0.2s; padding: 10px 4px; border-radius: 12px; width: 100%; box-sizing: border-box; }
.nav-item:hover { color: #0f172a; background: #f1f5f9; }
.nav-item.active { color: #2563eb; background: #eff6ff; }
.nav-item.active .nav-text { font-weight: 600; }
.nav-icon { display: flex; align-items: center; justify-content: center; transition: transform 0.2s;}
.nav-item:active .nav-icon { transform: scale(0.9); }
.nav-text { font-size: 11px; text-align: center;}

/* 主视图区域 */
.app-layout { display: flex; flex-direction: column; flex: 1; height: 100vh; background: #f8fafc; min-width: 0; }
.app-header { height: 60px; min-height: 60px; background: #ffffff; display: flex; align-items: center; justify-content: space-between; padding: 0 24px; border-bottom: 1px solid #e2e8f0; -webkit-app-region: drag; flex-shrink: 0; }
.header-title h1 { margin: 0; font-size: 16px; font-weight: 600; color: #0f172a; }

.segmented-control { display: flex; background: #f1f5f9; padding: 3px; border-radius: 8px; -webkit-app-region: no-drag; }
.segmented-control button { border: none; background: transparent; padding: 6px 16px; font-size: 13px; font-weight: 500; color: #64748b; border-radius: 6px; cursor: pointer; transition: all 0.2s ease; }
.segmented-control button.active { background: #ffffff; color: #0f172a; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.mini-segmented { width: 100%; box-sizing: border-box; }

.workspace { display: flex; flex: 1; overflow: hidden; height: calc(100vh - 60px); }

/* 侧边控制面板：绝对禁止挤压 */
.control-sidebar { width: 320px; min-width: 320px; background: #ffffff; border-right: 1px solid #e2e8f0; display: flex; flex-direction: column; z-index: 10; flex-shrink: 0; }
.panel-content { flex: 1; overflow-y: auto; padding: 24px 20px; }
.global-actions { padding: 16px 20px; background: #ffffff; border-top: 1px solid #e2e8f0; box-shadow: 0 -4px 10px rgba(0,0,0,0.02); flex-shrink: 0; }

/* 右侧数据面板：修复 Flex 撑爆屏幕的绝症 */
.data-view { flex: 1; display: flex; flex-direction: column; position: relative; background: #f8fafc; min-width: 0; }
.table-container { flex: 1; overflow: auto; padding: 16px; }

/* 表格核心布局：开启 fixed 布局防止重叠 */
.native-table { width: 100%; min-width: 800px; table-layout: fixed; border-collapse: separate; border-spacing: 0; background: #ffffff; border: 1px solid #e2e8f0; border-radius: 8px; box-shadow: 0 1px 2px rgba(0,0,0,0.02); }
.native-table th { background: #f8fafc; color: #475569; font-size: 12px; font-weight: 600; text-align: left; padding: 10px 16px; position: sticky; top: 0; z-index: 10; border-bottom: 1px solid #e2e8f0; }
.native-table td { padding: 10px 16px; font-size: 13px; border-bottom: 1px solid #f1f5f9; vertical-align: middle; word-wrap: break-word; }
.native-table tr:last-child td { border-bottom: none; }
.native-table tr:hover td { background: #f8fafc; }

/* 文字溢出截断：配合 :title 实现 Hover 悬浮显示完整内容 */
.text-truncate { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; width: 100%; }

/* 小组件与表单样式 */
.section-title { font-size: 13px; font-weight: 600; color: #0f172a; text-transform: uppercase; letter-spacing: 0.5px; margin: 0 0 16px 0; }
.helper-text { font-size: 12px; color: #64748b; margin-top: -10px; margin-bottom: 16px; line-height: 1.5; }
.tiny-hint { font-size: 11px; color: #64748b; margin: 0 0 4px 0; }
.divider { height: 1px; background: #e2e8f0; margin: 24px 0; flex-shrink: 0; }
.input-stack { display: flex; flex-direction: column; gap: 12px; }
.input-field { display: flex; flex-direction: column; gap: 6px; }
.input-field label { font-size: 12px; font-weight: 500; color: #475569; }
input { width: 100%; box-sizing: border-box; padding: 8px 12px; font-size: 13px; color: #1e293b; background: #f8fafc; border: 1px solid #cbd5e1; border-radius: 6px; outline: none; transition: all 0.2s; }
input:focus { background: #ffffff; border-color: #3b82f6; box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1); }

/* 按钮组与排版：安全折行，拒绝溢出 */
.action-row { display: flex; gap: 8px; flex-wrap: wrap; align-items: center;}
.table-toolbar { display: flex; gap: 10px; margin-bottom: 12px; align-items: center; flex-wrap: wrap; }

.btn { display: inline-flex; align-items: center; justify-content: center; gap: 6px; box-sizing: border-box; padding: 8px 16px; font-size: 13px; font-weight: 500; border-radius: 6px; cursor: pointer; border: 1px solid transparent; transition: all 0.15s; white-space: nowrap; }
.w-full { width: 100%; }
.flex-1 { flex: 1; }
.mt-1 { margin-top: 4px; }
.mt-2 { margin-top: 8px; }
.mt-3 { margin-top: 16px; }
.btn-primary { background: #2563eb; color: #fff; }
.btn-primary:hover:not(:disabled) { background: #1d4ed8; }
.btn-primary:disabled { background: #cbd5e1; color: #f8fafc; cursor: not-allowed; }
.btn-secondary { background: #f1f5f9; color: #334155; border-color: #e2e8f0; }
.btn-secondary:hover { background: #e2e8f0; }
.btn-danger-soft { background: #fef2f2; color: #ef4444; border-color: #fee2e2; }
.btn-danger-soft:hover { background: #fee2e2; border-color: #fca5a5; }
.btn-warning-soft { background: #fef3c7; color: #d97706; border-color: #fde68a; }
.btn-warning-soft:hover { background: #fde68a; border-color: #fcd34d; }
.btn-action { background: transparent; border: none; font-size: 12px; font-weight: 500; padding: 4px 8px; border-radius: 4px; cursor: pointer; transition: all 0.2s; }
.btn-action.primary { color: #2563eb; } .btn-action.primary:hover { background: #eff6ff; }
.btn-action.success { color: #059669; } .btn-action.success:hover { background: #ecfdf5; }
.btn-action.danger { color: #ef4444; } .btn-action.danger:hover { background: #fef2f2; }

/* 滚动条美化 */
::-webkit-scrollbar { width: 8px; height: 8px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 4px; border: 2px solid transparent; background-clip: padding-box; }
::-webkit-scrollbar-thumb:hover { background-color: #94a3b8; }

/* 各种杂项样式 */
.row-error td { background-color: #fef2f2 !important; }
.font-medium { font-weight: 500; }
.col-old-name { color: #64748b; text-decoration: line-through; }
.col-new-name { color: #0f172a; }
.col-actions { display: flex; gap: 4px; justify-content: flex-end; padding-right: 20px !important; flex-wrap: wrap; }
.chips-group { display: flex; flex-wrap: wrap; gap: 6px; }
.chip { font-size: 11px; padding: 4px 10px; border-radius: 12px; background: #f1f5f9; color: #475569; border: 1px solid transparent; cursor: pointer; transition: all 0.15s; user-select: none; }
.chip:hover { background: #e2e8f0; }
.chip.active { background: #eff6ff; color: #2563eb; border-color: #bfdbfe; font-weight: 600; }
.status-badge { display: inline-flex; padding: 4px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; white-space: nowrap; }
.status-badge.neutral { background: #f1f5f9; color: #64748b; }
.status-badge.success { background: #dcfce7; color: #15803d; }
.status-badge.error { background: #fee2e2; color: #b91c1c; }
.native-select { width: 100%; box-sizing: border-box; padding: 8px 12px; font-size: 13px; color: #1e293b; background: #f8fafc; border: 1px solid #cbd5e1; border-radius: 6px; outline: none; cursor: pointer; appearance: auto; }
.native-select:focus { border-color: #3b82f6; box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1); }
.native-select.mini { padding: 4px 8px; width: 80px; font-weight: 600; color: #2563eb; background: #eff6ff; border-color: #bfdbfe; }
.geek-panel { background: #f1f5f9; padding: 12px; border-radius: 6px; margin-top: 8px; border: 1px dashed #cbd5e1; }

/* 进度文字防堆叠截断 */
.progress-display { display: flex; align-items: center; gap: 8px; font-family: "JetBrains Mono", Consolas, monospace; font-size: 12px; font-weight: 500; width: 100%; overflow: hidden;}
.progress-text { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; width: 100%; }

.progress-display.pending { color: #94a3b8; }
.progress-display.success { color: #10b981; }
.progress-display.error { color: #ef4444; }
.progress-display.processing { color: #2563eb; }
.pulsing-dot { width: 8px; height: 8px; border-radius: 50%; background-color: #3b82f6; animation: pulse 1.5s infinite ease-in-out; flex-shrink: 0; }
@keyframes pulse {
  0% { transform: scale(0.8); box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 6px rgba(59, 130, 246, 0); }
  100% { transform: scale(0.8); box-shadow: 0 0 0 0 rgba(59, 130, 246, 0); }
}
.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; color: #94a3b8; }
.empty-icon { font-size: 48px; margin-bottom: 16px; filter: grayscale(1) opacity(0.5); }
.empty-state h3 { font-size: 16px; font-weight: 600; color: #475569; margin: 0 0 8px 0; }
.empty-state p { font-size: 13px; margin: 0; }
.wails-drop-target-active::after { content: "释放文件以添加"; position: absolute; inset: 12px; background: rgba(239, 246, 255, 0.9); border: 2px dashed #3b82f6; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 18px; font-weight: 600; color: #2563eb; z-index: 100; pointer-events: none; }

/* 弹窗系统 */
.custom-modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(15, 23, 42, 0.4); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 9999; }
.custom-modal-box { background: #ffffff; width: 400px; max-width: 90vw; border-radius: 12px; box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1), 0 10px 10px -5px rgba(0,0,0,0.04); overflow: hidden; display: flex; flex-direction: column; }
.modal-header { display: flex; align-items: center; gap: 12px; padding: 20px 24px 16px; }
.modal-icon { display: flex; align-items: center; justify-content: center; width: 40px; height: 40px; border-radius: 50%; flex-shrink: 0; }
.modal-icon.success { background: #dcfce7; color: #16a34a; }
.modal-icon.warning { background: #fef3c7; color: #d97706; }
.modal-icon.error { background: #fee2e2; color: #dc2626; }
.modal-icon.info { background: #e0e7ff; color: #2563eb; }
.modal-title { margin: 0; font-size: 18px; font-weight: 600; color: #0f172a; }
.modal-body { padding: 0 24px 24px; font-size: 14px; color: #475569; line-height: 1.6; white-space: pre-wrap; max-height: 50vh; overflow-y: auto; }
.modal-footer { background: #f8fafc; padding: 16px 24px; display: flex; justify-content: flex-end; gap: 12px; border-top: 1px solid #e2e8f0; }
.btn.danger-btn { background: #ef4444; color: white; border-color: #ef4444; }
.btn.danger-btn:hover { background: #dc2626; }
.modal-fade-enter-active, .modal-fade-leave-active { transition: all 0.2s ease; }
.modal-fade-enter-from, .modal-fade-leave-to { opacity: 0; transform: scale(0.95); }
</style>