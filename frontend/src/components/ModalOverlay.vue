<script setup lang="ts">
import { useModal } from '../composables/useModal'

const { modalState, confirm, cancel } = useModal()
</script>

<template>
  <Transition name="modal-fade">
    <div v-if="modalState.visible" class="custom-modal-overlay" @mousedown.self="cancel">
      <div class="custom-modal-box">
        <div class="modal-header">
          <div class="modal-icon" :class="modalState.type">
            <svg v-if="modalState.type === 'success'" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
            <svg v-else-if="modalState.type === 'warning'" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
            <svg v-else-if="modalState.type === 'error'" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
          </div>
          <h3 class="modal-title">{{ modalState.title }}</h3>
        </div>
        <div class="modal-body">{{ modalState.message }}</div>
        <div class="modal-footer">
          <button v-if="modalState.isConfirm" class="btn btn-secondary" @click="cancel">取消</button>
          <button class="btn btn-primary" :class="{'danger-btn': modalState.type === 'error' || modalState.type === 'warning'}" @click="confirm">
            {{ modalState.isConfirm ? (modalState.type === 'warning' ? '强行丢弃' : '确定') : '我知道了' }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>
