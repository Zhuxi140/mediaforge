<script setup lang="ts">
import { useRenamerStore } from '../stores/renamer'
import { ApplyRename, QuickRename } from '../../wailsjs/go/controller/RenamerApp'
import { useModal } from '../composables/useModal'

const store = useRenamerStore()
const { showModal } = useModal()

const cleanCharOptions = [
  { label: '空格', value: ' ' },
  { label: '减号 (-)', value: '-' },
  { label: '下划线 (_)', value: '_' },
  { label: '点号 (.)', value: '.' },
  { label: '逗号 (,)', value: ',' }
]

const handleExecute = async () => {
  if (!store.canExecute) return
  const confirmed = await showModal(
    '确认执行',
    `即将重命名 ${store.validPreviews.length} 个文件。\n此操作不可撤销，是否继续？`,
    'warning', true
  )
  if (!confirmed) return
  try {
    const msg = await ApplyRename(store.validPreviews as any)
    if (msg === 'success') {
      showModal('执行成功', `成功重命名 ${store.validPreviews.length} 个文件！`, 'success')
      const validPaths = store.validPreviews.map(p => p.originalPath).filter((p): p is string => !!p)
      store.filePaths = store.filePaths.filter(p => !validPaths.includes(p))
    } else {
      showModal('执行失败', msg, 'error')
    }
  } catch (err: any) {
    showModal('系统异常', err, 'error')
  }
}

const startEdit = (index: number, name: string) => { store.editingIndex = index; store.editTempName = name }
const cancelEdit = () => { store.editingIndex = -1; store.editTempName = '' }
const saveEdit = async (index: number, oldPath: string) => {
  if (!store.editTempName.trim() || store.editTempName === store.previews[index].originalName) { cancelEdit(); return }
  try {
    store.filePaths[index] = await QuickRename(oldPath, store.editTempName)
    cancelEdit()
  } catch (err: any) {
    showModal('修改失败', err, 'error')
  }
}
</script>

<template>
  <main class="app-layout">
    <div class="workspace">
      <aside class="control-sidebar">
        <div class="panel-content">
          <div class="input-field" style="margin-bottom: 20px;">
            <label>重命名模式</label>
            <div class="segmented-control mini-segmented">
              <button :class="{ active: store.currentMode === 'BasicMode' }" @click="store.currentMode = 'BasicMode'" style="flex: 1;">基础替换</button>
              <button :class="{ active: store.currentMode === 'SmartMode' }" @click="store.currentMode = 'SmartMode'" style="flex: 1;">智能解析</button>
            </div>
          </div>

          <template v-if="store.currentMode === 'BasicMode'">
            <h2 class="section-title">基础规则设定</h2>
            <div class="input-stack">
              <div class="input-field"><label>文件前缀</label><input v-model="store.rule.Prefix" placeholder="如: 2026_" /></div>
              <div class="input-field"><label>文件后缀</label><input v-model="store.rule.Suffix" placeholder="如: _v1" /></div>
            </div>
            <div class="divider"></div>
            <h2 class="section-title">字符替换</h2>
            <div class="input-stack">
              <input v-model="store.rule.ReplaceOld" placeholder="查找特定字符..." />
              <div class="icon-arrow">↓</div>
              <input v-model="store.rule.ReplaceNew" placeholder="替换为新字符..." />
            </div>
          </template>

          <template v-if="store.currentMode === 'SmartMode'">
            <h2 class="section-title">智能提取变量</h2>
            <p class="helper-text">利用正则无视文件名的混乱顺序提取关键信息。</p>
            <div class="rules-list">
              <div v-for="(r, index) in store.rule.SmartRules" :key="index" class="rule-card">
                <div class="rule-header">
                  <span class="rule-num">{{ index + 1 }}</span>
                  <button class="btn-icon text-danger" @click="store.removeSmartRule(index)">✕</button>
                </div>
                <div class="rule-inputs">
                  <input v-model="r.Name" placeholder="变量名" class="code-input var-name" />
                  <input v-model="r.Pattern" placeholder="正则表达式" class="code-input pattern" />
                </div>
              </div>
            </div>
            <button class="btn-dashed w-full" @click="store.addSmartRule">＋ 新增提取变量</button>
            <div class="divider"></div>
            <h2 class="section-title">重组装设定</h2>
            <div class="input-field"><label>目标格式模板</label><input v-model="store.rule.SmartTemplate" class="code-input highlight" placeholder="{class}-{name}_{id}" /></div>
            <div class="input-field mt-3"><label>预处理: 过滤干扰符</label>
              <div class="chips-group">
                <label v-for="opt in cleanCharOptions" :key="opt.value" class="chip" :class="{ 'active': store.selectedChars.includes(opt.value) }">
                  <input type="checkbox" :value="opt.value" v-model="store.selectedChars" style="display: none;" />{{ opt.label }}
                </label>
              </div>
            </div>
          </template>
        </div>
        <div class="global-actions">
          <button class="btn btn-primary w-full shadow-sm" :disabled="!store.canExecute" @click="handleExecute">
            执行重命名 <span v-if="store.validPreviews.length" class="badge">{{ store.validPreviews.length }}</span>
          </button>
          <div class="action-row mt-2">
            <button class="btn btn-secondary w-full" @click="$emit('select-files')">选择文件</button>
          </div>
        </div>
      </aside>
      <main class="data-view" style="--wails-drop-target: drop">
        <div v-if="store.previews.length === 0" class="empty-state">
          <div class="empty-icon">📂</div><h3>没有选择文件</h3><p>请点击左侧选择文件</p>
        </div>
        <div v-else class="table-container">
          <div class="table-toolbar">
            <button class="btn btn-secondary" @click="store.refreshList">重新扫描</button>
            <button class="btn btn-danger-soft" @click="store.clearList">清空列表</button>
          </div>
          <table class="native-table">
            <thead><tr><th style="width: 35%;">原文件名</th><th style="width: 30%;">预览新文件名</th><th style="width: 20%;">状态</th><th style="width: 15%; text-align: right; padding-right: 20px;">操作</th></tr></thead>
            <tbody>
              <tr v-for="(item, index) in store.previews" :key="index" :class="{'row-error': item.formatError}">
                <td class="col-old-name">
                  <div v-if="store.editingIndex === index" class="inline-edit-wrapper">
                    <input v-model="store.editTempName" @keyup.enter="item.originalPath && saveEdit(index, item.originalPath)" @keyup.esc="cancelEdit" class="inline-input" autofocus />
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
                  <button v-if="store.editingIndex !== index && (item.formatError || item.hasConflict)" class="btn-action primary" @click="startEdit(index, item.originalName)">修正</button>
                  <button v-if="store.editingIndex === index" class="btn-action success" @click="item.originalPath && saveEdit(index, item.originalPath)">保存</button>
                  <button v-if="store.editingIndex === index" class="btn-action danger" @click="cancelEdit">取消</button>
                  <button v-if="store.editingIndex !== index" class="btn-action danger" @click="store.removeFile(index)">移除</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </main>
    </div>
  </main>
</template>
