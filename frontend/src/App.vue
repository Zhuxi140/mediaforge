<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed } from 'vue';
import { SelectFiles, PreviewRename, ApplyRename, QuickRename } from '../wailsjs/go/main/App';
import { OnFileDrop, OnFileDropOff } from '../wailsjs/runtime/runtime';

// ==========================================
// 🌌 App Shell 全局状态
// ==========================================
const activeApp = ref<'renamer' | 'ffmpeg' | 'settings'>('renamer');

// ==========================================
// 🗂️ 模块一：批量重命名 (Renamer) 状态
// ==========================================
const filePaths = ref<string[]>([]);
const previews = ref<any[]>([]);
const currentMode = ref<'BasicMode' | 'SmartMode'>('BasicMode');

const rule = ref({
  Mode: 'BasicMode',
  Prefix: '', Suffix: '', ReplaceOld: '', ReplaceNew: '',
  SmartRules: [
    { Name: 'id', Pattern: '202\\d{8}' },
    { Name: 'class', Pattern: '大数据.*?[0-9]*班' },
    { Name: 'name', Pattern: '\\p{Han}{2,4}' }
  ],
  SmartTemplate: '{class}-{name}_{id}',
  CleanChars: '-_ '
});

const cleanCharOptions = [
  { label: '空格', value: ' ' }, { label: '减号 (-)', value: '-' },
  { label: '下划线 (_)', value: '_' }, { label: '点号 (.)', value: '.' }, { label: '逗号 (,)', value: ',' }
];
const selectedChars = ref([' ', '-', '_']);

watch(selectedChars, (newVal) => { rule.value.CleanChars = newVal.join(''); }, { immediate: true, deep: true });
watch(currentMode, (newMode) => { rule.value.Mode = newMode; });

const addSmartRule = () => { rule.value.SmartRules.push({ Name: '', Pattern: '' }); };
const removeSmartRule = (index: number) => { rule.value.SmartRules.splice(index, 1); };

onMounted(() => {
  OnFileDrop((x: number, y: number, paths: string[]) => {
    if (activeApp.value === 'renamer') {
      const newFiles = paths.filter(p => !filePaths.value.includes(p));
      filePaths.value = [...filePaths.value, ...newFiles];
    }
  }, true);
});

onUnmounted(() => { OnFileDropOff(); });

const handleSelectFiles = async () => {
  const files = await SelectFiles();
  if (files && files.length > 0) {
    const newFiles = files.filter(f => !filePaths.value.includes(f));
    filePaths.value = [...filePaths.value, ...newFiles];
  }
};

const clearList = () => { filePaths.value = []; previews.value = []; };
const removeFile = (index: number) => { filePaths.value.splice(index, 1); };
const refreshList = async () => { if (filePaths.value.length > 0) filePaths.value = [...filePaths.value]; };

watch([filePaths, rule], async () => {
  if (filePaths.value.length === 0) { previews.value = []; return; }
  try {
    const result = await PreviewRename(filePaths.value, rule.value as any);
    previews.value = result || [];
  } catch (err) { console.error("生成预览失败:", err); }
}, { deep: true });

const validPreviews = computed(() => previews.value.filter(p => !p.hasConflict && !p.formatError && p.originalName !== p.newName));
const canExecute = computed(() => validPreviews.value.length > 0);

const handleExecute = async () => {
  if (!canExecute.value) return;
  try {
    const msg = await ApplyRename(validPreviews.value as any);
    if (msg === "success") {
      alert(`🎉 成功重命名 ${validPreviews.value.length} 个文件！`);
      const validPaths = validPreviews.value.map(p => p.originalPath);
      filePaths.value = filePaths.value.filter(p => !validPaths.includes(p));
    } else { alert("执行失败: " + msg); }
  } catch (error) { alert("系统异常: " + error); }
};

const editingIndex = ref<number>(-1);
const editTempName = ref<string>('');

const startEdit = (index: number, currentName: string) => { editingIndex.value = index; editTempName.value = currentName; };
const cancelEdit = () => { editingIndex.value = -1; editTempName.value = ''; };
const saveEdit = async (index: number, oldPath: string) => {
  if (!editTempName.value.trim() || editTempName.value === previews.value[index].originalName) { cancelEdit(); return; }
  try {
    const newPath = await QuickRename(oldPath, editTempName.value);
    filePaths.value[index] = newPath;
    cancelEdit();
  } catch (err) { alert("修改失败: " + err); }
};
</script>

<template>
  <div class="app-root">

    <nav class="global-sidebar">
      <div class="brand">
        <div class="logo">
          <svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="url(#logo-grad)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <defs>
              <linearGradient id="logo-grad" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stop-color="#3b82f6" />
                <stop offset="100%" stop-color="#8b5cf6" />
              </linearGradient>
            </defs>
            <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon>
          </svg>
        </div>
      </div>

      <div class="nav-menu">
        <button class="nav-item" :class="{ active: activeApp === 'renamer' }" @click="activeApp = 'renamer'">
          <span class="nav-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 13.5V4a2 2 0 0 1 2-2h8.5L20 7.5V20a2 2 0 0 1-2 2h-5.5"/><polyline points="14 2 14 8 20 8"/><path d="M10.42 12.61a2.1 2.1 0 1 1 2.97 2.97L7.95 21 4 22l.99-3.95 5.43-5.44Z"/></svg>
          </span>
          <span class="nav-text">重命名</span>
        </button>

        <button class="nav-item" :class="{ active: activeApp === 'ffmpeg' }" @click="activeApp = 'ffmpeg'">
          <span class="nav-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20.2 6 3 11l-.9-2.4c-.3-1.1.4-2.2 1.5-2.5l13.5-4c1.1-.3 2.2.4 2.5 1.5z"/><path d="m6.2 5.3 3.1 3.9"/><path d="m12.4 3.4 3.1 4"/><path d="M3 11h18v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z"/></svg>
          </span>
          <span class="nav-text">影音工厂</span>
        </button>
      </div>

      <div class="nav-bottom">
        <button class="nav-item" :class="{ active: activeApp === 'settings' }" @click="activeApp = 'settings'">
          <span class="nav-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>
          </span>
          <span class="nav-text">设置</span>
        </button>
      </div>
    </nav>

    <main v-if="activeApp === 'renamer'" class="app-layout">
      <header class="app-header">
        <div class="header-title">
          <h1>批量重命名引擎</h1>
        </div>
        <div class="segmented-control">
          <button :class="{ active: currentMode === 'BasicMode' }" @click="currentMode = 'BasicMode'">基础替换</button>
          <button :class="{ active: currentMode === 'SmartMode' }" @click="currentMode = 'SmartMode'">智能解析</button>
        </div>
      </header>

      <div class="workspace">
        <aside class="control-sidebar">
          <div v-if="currentMode === 'BasicMode'" class="panel-content">
            <h2 class="section-title">基础规则设定</h2>
            <div class="input-stack">
              <div class="input-field"><label>文件前缀</label><input v-model="rule.Prefix" placeholder="如: 2026_" /></div>
              <div class="input-field"><label>文件后缀</label><input v-model="rule.Suffix" placeholder="如: _v1" /></div>
            </div>
            <div class="divider"></div>
            <h2 class="section-title">字符替换</h2>
            <div class="input-stack">
              <input v-model="rule.ReplaceOld" placeholder="查找特定字符..." />
              <div class="icon-arrow">↓</div>
              <input v-model="rule.ReplaceNew" placeholder="替换为新字符..." />
            </div>
          </div>

          <div v-if="currentMode === 'SmartMode'" class="panel-content">
            <h2 class="section-title">智能提取变量</h2>
            <p class="helper-text">利用正则无视文件名的混乱顺序提取关键信息。</p>
            <div class="rules-list">
              <div v-for="(r, index) in rule.SmartRules" :key="index" class="rule-card">
                <div class="rule-header">
                  <span class="rule-num">{{ index + 1 }}</span>
                  <button class="btn-icon text-danger" @click="removeSmartRule(index)">✕</button>
                </div>
                <div class="rule-inputs">
                  <input v-model="r.Name" placeholder="变量名" class="code-input var-name" />
                  <input v-model="r.Pattern" placeholder="正则表达式" class="code-input pattern" />
                </div>
              </div>
            </div>
            <button class="btn-dashed w-full" @click="addSmartRule">＋ 新增提取变量</button>
            <div class="divider"></div>
            <h2 class="section-title">重组装设定</h2>
            <div class="input-field">
              <label>目标格式模板</label>
              <input v-model="rule.SmartTemplate" class="code-input highlight" placeholder="{class}-{name}_{id}" />
            </div>
            <div class="input-field mt-3">
              <label>预处理: 过滤干扰符</label>
              <div class="chips-group">
                <label v-for="opt in cleanCharOptions" :key="opt.value" class="chip" :class="{ 'active': selectedChars.includes(opt.value) }">
                  <input type="checkbox" :value="opt.value" v-model="selectedChars" style="display: none;" />
                  {{ opt.label }}
                </label>
              </div>
            </div>
          </div>

          <div class="global-actions">
            <button class="btn btn-primary w-full shadow-sm" :disabled="!canExecute" @click="handleExecute">
              🚀 执行重命名 <span v-if="validPreviews.length" class="badge">{{ validPreviews.length }}</span>
            </button>
            <div class="action-row mt-2">
              <button class="btn btn-secondary flex-1" @click="handleSelectFiles">选择文件</button>
              <button class="btn btn-secondary" @click="refreshList" v-if="filePaths.length > 0" title="重新扫描磁盘">🔄</button>
              <button class="btn btn-danger-soft" @click="clearList" v-if="filePaths.length > 0" title="清空列表">🗑️</button>
            </div>
          </div>
        </aside>

        <main class="data-view" style="--wails-drop-target: drop">
          <div v-if="previews.length === 0" class="empty-state">
            <div class="empty-icon">📂</div>
            <h3>没有选择文件</h3>
            <p>请点击左侧选择文件，或将文件拖拽至此区域</p>
          </div>
          <div v-else class="table-container">
            <table class="native-table">
              <thead>
              <tr>
                <th style="width: 35%;">原文件名</th>
                <th style="width: 30%;">预览新文件名</th>
                <th style="width: 20%;">状态</th>
                <th style="width: 15%; text-align: right; padding-right: 20px;">操作</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="(item, index) in previews" :key="index" :class="{'row-error': item.formatError}">
                <td class="col-old-name">
                  <div v-if="editingIndex === index" class="inline-edit-wrapper">
                    <input v-model="editTempName" @keyup.enter="saveEdit(index, item.originalPath)" @keyup.esc="cancelEdit" class="inline-input" autofocus />
                  </div>
                  <span v-else class="text-truncate" :title="item.originalName">{{ item.originalName }}</span>
                </td>
                <td class="col-new-name"><span class="text-truncate font-medium" :title="item.newName">{{ item.newName }}</span></td>
                <td>
                  <span v-if="item.formatError" class="status-badge error" :title="item.formatError">{{ item.formatError }}</span>
                  <span v-else-if="item.hasConflict" class="status-badge error">同名冲突</span>
                  <span v-else-if="item.originalName === item.newName" class="status-badge neutral">跳过</span>
                  <span v-else class="status-badge success">待修改</span>
                </td>
                <td class="col-actions">
                  <button v-if="editingIndex !== index && (item.formatError || item.hasConflict)" class="btn-action primary" @click="startEdit(index, item.originalName)">修正</button>
                  <button v-if="editingIndex === index" class="btn-action success" @click="saveEdit(index, item.originalPath)">保存</button>
                  <button v-if="editingIndex === index" class="btn-action danger" @click="cancelEdit">取消</button>
                  <button v-if="editingIndex !== index" class="btn-action danger" @click="removeFile(index)">移除</button>
                </td>
              </tr>
              </tbody>
            </table>
          </div>
        </main>
      </div>
    </main>

    <main v-if="activeApp === 'ffmpeg'" class="app-layout">
      <header class="app-header">
        <div class="header-title"><h1>FFmpeg 影音铸造厂</h1></div>
      </header>
      <div class="workspace" style="align-items: center; justify-content: center; background: #f8fafc;">
        <div style="text-align: center; color: #64748b; display: flex; flex-direction: column; align-items: center;">
          <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" style="color: #cbd5e1; margin-bottom: 16px;"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>
          <h2 style="color: #475569; margin: 0 0 8px 0;">子进程引擎未连接</h2>
          <p style="margin: 0;">请完成 Go 后端的 FFmpeg 调用模块后开启此功能</p>
        </div>
      </div>
    </main>

    <main v-if="activeApp === 'settings'" class="app-layout">
      <header class="app-header">
        <div class="header-title"><h1>全局设置</h1></div>
      </header>
      <div class="workspace" style="padding: 40px; background: #f8fafc;">
        <h3 style="color: #475569;">开发中...</h3>
      </div>
    </main>

  </div>
</template>

<style>
/* ================= 🌍 顶级容器与 App Shell ================= */
:global(body), :global(html) {
  margin: 0; padding: 0; height: 100vh; overflow: hidden;
  background-color: #f8fafc; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  color: #1e293b; -webkit-font-smoothing: antialiased;
}

.app-root { display: flex; height: 100vh; width: 100vw; overflow: hidden; }

/* ✨ 极其清爽的亮色左侧边栏 */
.global-sidebar {
  width: 76px; min-width: 76px; background: #ffffff; display: flex; flex-direction: column;
  align-items: center; padding: 20px 0; z-index: 50; border-right: 1px solid #e2e8f0;
  -webkit-app-region: drag; /* 允许拖拽 */
}
.brand { margin-bottom: 32px; display: flex; justify-content: center; width: 100%;}
.logo { display: flex; align-items: center; justify-content: center; }

.nav-menu { display: flex; flex-direction: column; gap: 8px; flex: 1; -webkit-app-region: no-drag; width: 100%; padding: 0 10px; box-sizing: border-box;}
.nav-bottom { margin-top: auto; -webkit-app-region: no-drag; width: 100%; padding: 0 10px; box-sizing: border-box;}

/* 极简风导航按钮 */
.nav-item {
  background: transparent; border: none; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px;
  color: #64748b; cursor: pointer; transition: all 0.2s; padding: 10px 4px; border-radius: 12px; width: 100%; box-sizing: border-box;
}
.nav-item:hover { color: #0f172a; background: #f1f5f9; }
.nav-item.active { color: #2563eb; background: #eff6ff; }
.nav-item.active .nav-text { font-weight: 600; }
.nav-icon { display: flex; align-items: center; justify-content: center; transition: transform 0.2s;}
.nav-item:active .nav-icon { transform: scale(0.9); }
.nav-text { font-size: 11px; }

/* ================= 🗂️ 右侧内容区通用样式 ================= */
.app-layout { display: flex; flex-direction: column; flex: 1; height: 100vh; background: #f8fafc; }

.app-header {
  height: 60px; min-height: 60px; background: #ffffff; display: flex; align-items: center; justify-content: space-between;
  padding: 0 24px; border-bottom: 1px solid #e2e8f0; -webkit-app-region: drag;
}
.header-title h1 { margin: 0; font-size: 16px; font-weight: 600; color: #0f172a; }

.segmented-control { display: flex; background: #f1f5f9; padding: 3px; border-radius: 8px; -webkit-app-region: no-drag; }
.segmented-control button {
  border: none; background: transparent; padding: 6px 16px; font-size: 13px; font-weight: 500; color: #64748b;
  border-radius: 6px; cursor: pointer; transition: all 0.2s ease;
}
.segmented-control button.active { background: #ffffff; color: #0f172a; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }

.workspace { display: flex; flex: 1; overflow: hidden; height: calc(100vh - 60px); }

/* ================= 原有 Renamer 专属样式 ================= */
::-webkit-scrollbar { width: 8px; height: 8px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 4px; border: 2px solid transparent; background-clip: padding-box; }
::-webkit-scrollbar-thumb:hover { background-color: #94a3b8; }

.control-sidebar { width: 320px; min-width: 320px; background: #ffffff; border-right: 1px solid #e2e8f0; display: flex; flex-direction: column; z-index: 10; }
.panel-content { flex: 1; overflow-y: auto; padding: 24px 20px; }
.section-title { font-size: 13px; font-weight: 600; color: #0f172a; text-transform: uppercase; letter-spacing: 0.5px; margin: 0 0 16px 0; }
.helper-text { font-size: 12px; color: #64748b; margin-top: -10px; margin-bottom: 16px; line-height: 1.5; }
.divider { height: 1px; background: #e2e8f0; margin: 24px 0; }

.input-stack { display: flex; flex-direction: column; gap: 12px; }
.input-field { display: flex; flex-direction: column; gap: 6px; }
.input-field label { font-size: 12px; font-weight: 500; color: #475569; }
input { width: 100%; box-sizing: border-box; padding: 8px 12px; font-size: 13px; color: #1e293b; background: #f8fafc; border: 1px solid #cbd5e1; border-radius: 6px; outline: none; transition: all 0.2s; }
input:focus { background: #ffffff; border-color: #3b82f6; box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1); }
.icon-arrow { text-align: center; color: #94a3b8; font-size: 16px; }

.rules-list { display: flex; flex-direction: column; gap: 10px; margin-bottom: 12px; }
.rule-card { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 10px; position: relative; }
.rule-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.rule-num { font-size: 11px; font-weight: 600; color: #64748b; background: #e2e8f0; padding: 2px 6px; border-radius: 10px; }
.rule-inputs { display: flex; gap: 8px; }
.code-input { font-family: "JetBrains Mono", Consolas, monospace; font-size: 12px; }
.var-name { width: 35%; color: #059669; }
.pattern { width: 65%; color: #2563eb; }
.highlight { border-color: #93c5fd; background: #eff6ff; font-weight: 600; color: #1e3a8a; }

.chips-group { display: flex; flex-wrap: wrap; gap: 6px; }
.chip { font-size: 11px; padding: 4px 10px; border-radius: 12px; background: #f1f5f9; color: #475569; border: 1px solid transparent; cursor: pointer; transition: all 0.15s; user-select: none; }
.chip:hover { background: #e2e8f0; }
.chip.active { background: #eff6ff; color: #2563eb; border-color: #bfdbfe; font-weight: 600; }

.global-actions { padding: 16px 20px; background: #ffffff; border-top: 1px solid #e2e8f0; box-shadow: 0 -4px 10px rgba(0,0,0,0.02); }
.action-row { display: flex; gap: 8px; }
.w-full { width: 100%; }
.flex-1 { flex: 1; }
.mt-2 { margin-top: 8px; }
.mt-3 { margin-top: 16px; }

.btn { display: inline-flex; align-items: center; justify-content: center; gap: 6px; box-sizing: border-box; padding: 8px 16px; font-size: 13px; font-weight: 500; border-radius: 6px; cursor: pointer; border: 1px solid transparent; transition: all 0.15s; }
.btn-primary { background: #2563eb; color: #fff; }
.btn-primary:hover:not(:disabled) { background: #1d4ed8; }
.btn-primary:disabled { background: #cbd5e1; color: #f8fafc; cursor: not-allowed; }
.btn-secondary { background: #f1f5f9; color: #334155; border-color: #e2e8f0; }
.btn-secondary:hover { background: #e2e8f0; }
.btn-danger-soft { background: #fef2f2; color: #ef4444; border-color: #fee2e2; }
.btn-danger-soft:hover { background: #fee2e2; border-color: #fca5a5; }
.btn-dashed { background: transparent; border: 1px dashed #cbd5e1; color: #64748b; font-size: 12px; padding: 8px; border-radius: 6px; cursor: pointer; transition: all 0.2s; }
.btn-dashed:hover { border-color: #94a3b8; color: #475569; background: #f8fafc; }
.badge { background: #ffffff; color: #2563eb; font-size: 11px; font-weight: 700; padding: 2px 6px; border-radius: 10px; margin-left: 4px; }
.btn-icon { background: transparent; border: none; cursor: pointer; padding: 4px; border-radius: 4px; display: flex; align-items: center; justify-content: center;}
.btn-icon:hover { background: #f1f5f9; }
.text-danger { color: #ef4444; }

.data-view { flex: 1; display: flex; flex-direction: column; position: relative; background: #f8fafc; }
.wails-drop-target-active::after { content: "释放文件以添加"; position: absolute; inset: 12px; background: rgba(239, 246, 255, 0.9); border: 2px dashed #3b82f6; border-radius: 12px; display: flex; align-items: center; justify-content: center; font-size: 18px; font-weight: 600; color: #2563eb; z-index: 100; pointer-events: none; }
.empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; color: #94a3b8; }
.empty-icon { font-size: 48px; margin-bottom: 16px; filter: grayscale(1) opacity(0.5); }
.empty-state h3 { font-size: 16px; font-weight: 600; color: #475569; margin: 0 0 8px 0; }
.empty-state p { font-size: 13px; margin: 0; }

.table-container { flex: 1; overflow: auto; padding: 16px; }
.native-table { width: 100%; border-collapse: separate; border-spacing: 0; background: #ffffff; border: 1px solid #e2e8f0; border-radius: 8px; box-shadow: 0 1px 2px rgba(0,0,0,0.02); }
.native-table th { background: #f8fafc; color: #475569; font-size: 12px; font-weight: 600; text-align: left; padding: 10px 16px; position: sticky; top: 0; z-index: 10; border-bottom: 1px solid #e2e8f0; }
.native-table th:first-child { border-top-left-radius: 8px; }
.native-table th:last-child { border-top-right-radius: 8px; }
.native-table td { padding: 10px 16px; font-size: 13px; border-bottom: 1px solid #f1f5f9; vertical-align: middle; }
.native-table tr:last-child td { border-bottom: none; }
.native-table tr:hover td { background: #f8fafc; }

.row-error td { background-color: #fef2f2 !important; }
.text-truncate { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 250px; }
.col-old-name { color: #64748b; text-decoration: line-through; }
.col-new-name { color: #0f172a; }
.font-medium { font-weight: 500; }

.status-badge { display: inline-flex; padding: 4px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; white-space: nowrap; }
.status-badge.neutral { background: #f1f5f9; color: #64748b; }
.status-badge.success { background: #dcfce7; color: #15803d; }
.status-badge.error { background: #fee2e2; color: #b91c1c; }

.col-actions { display: flex; gap: 4px; justify-content: flex-end; padding-right: 20px !important; }
.btn-action { background: transparent; border: none; font-size: 12px; font-weight: 500; padding: 4px 8px; border-radius: 4px; cursor: pointer; transition: all 0.2s; }
.btn-action.primary { color: #2563eb; } .btn-action.primary:hover { background: #eff6ff; }
.btn-action.success { color: #059669; } .btn-action.success:hover { background: #ecfdf5; }
.btn-action.danger { color: #ef4444; } .btn-action.danger:hover { background: #fef2f2; }

.inline-edit-wrapper { margin: -6px -8px; }
.inline-input { width: 100%; border: 2px solid #3b82f6; border-radius: 4px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); padding: 4px 8px;}
</style>