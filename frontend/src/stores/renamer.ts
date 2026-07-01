import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'
import type { RenameRule, RenamePreview } from '../types'
import { PreviewRename } from '../../wailsjs/go/controller/RenamerApp'

export const useRenamerStore = defineStore('renamer', () => {
  const filePaths = ref<string[]>([])
  const previews = ref<RenamePreview[]>([])
  const currentMode = ref<'BasicMode' | 'SmartMode'>('BasicMode')
  const rule = ref<RenameRule>({
    Mode: 'BasicMode', Prefix: '', Suffix: '', ReplaceOld: '', ReplaceNew: '',
    SmartRules: [
      { Name: 'id', Pattern: '202\\d{8}' },
      { Name: 'class', Pattern: '大数据.*?[0-9]*班' },
      { Name: 'name', Pattern: '\\p{Han}{2,4}' }
    ],
    SmartTemplate: '{class}-{name}_{id}',
    CleanChars: '-_ '
  })
  const selectedChars = ref<string[]>([' ', '-', '_'])
  const editingIndex = ref(-1)
  const editTempName = ref('')

  const validPreviews = computed(() =>
    previews.value.filter(p => !p.hasConflict && !p.formatError && p.originalName !== p.newName)
  )
  const canExecute = computed(() => validPreviews.value.length > 0)

  watch(selectedChars, (v) => { rule.value.CleanChars = v.join('') }, { immediate: true })
  watch(currentMode, (m) => { rule.value.Mode = m })

  watch([filePaths, rule], async () => {
    if (filePaths.value.length === 0) { previews.value = []; return }
    try {
      previews.value = await PreviewRename(filePaths.value, rule.value as any) || []
    } catch { previews.value = [] }
  }, { deep: true })

  const addSmartRule = () => rule.value.SmartRules.push({ Name: '', Pattern: '' })
  const removeSmartRule = (i: number) => rule.value.SmartRules.splice(i, 1)
  const clearList = () => { filePaths.value = []; previews.value = [] }
  const removeFile = (i: number) => filePaths.value.splice(i, 1)
  const refreshList = () => { if (filePaths.value.length) filePaths.value = [...filePaths.value] }

  return {
    filePaths, previews, currentMode, rule, selectedChars,
    editingIndex, editTempName,
    validPreviews, canExecute,
    addSmartRule, removeSmartRule, clearList, removeFile, refreshList
  }
})
