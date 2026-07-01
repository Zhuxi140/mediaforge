# ADR-005: Promise 弹窗系统

**状态**: ✅ 已采纳  
**日期**: 2026-06-29  

## 背景

应用需要在多种场景下弹出模态对话框（操作确认、错误提示、字幕兼容性警告等），需要一个灵活且调用简洁的弹窗系统。

## 可选方案

1. **浏览器原生 confirm/alert** — 简单但无法自定义样式
2. **UI 库（Element Plus/Naive UI）** — 功能全但增加依赖体积
3. **自定义弹窗 + Promise 链** — 自定义样式 + async/await 调用
4. **自定义弹窗 + 回调** — 自定义样式 + 回调函数

## 决策

选择方案 **3（自定义弹窗 + Promise 链）**。

## 接口设计

```ts
// 调用方式
const confirmed = await showModal('标题', '消息', 'warning', true)
if (confirmed) { /* 用户点了确定 */ }

// 内部实现
let resolveModal: ((value: boolean) => void) | null = null

function showModal(title, message, type, isConfirm): Promise<boolean> {
  modalState.value = { visible: true, title, message, type, isConfirm }
  return new Promise((resolve) => { resolveModal = resolve })
}
```

## 理由

- `async/await` 调用方式代码线性、易读，避免回调地狱
- 自定义样式与现有 UI 风格一致
- 确认模式支持"强行丢弃/取消"双按钮，定制灵活
- 作为 composable 导出，任何组件都可直接使用，无需 props 传递
- 不需要引入额外 UI 库

## 后果

- `resolveModal` 为模块级变量，不支持同时弹出多个弹窗（当前无此需求）
- 弹窗样式需要手动维护
- 弹窗内容当前为纯文本，如需富文本需扩展 `ModalState`
