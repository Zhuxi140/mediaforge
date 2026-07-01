import { ref } from 'vue'
import type { ModalState, ModalType } from '../types'

let resolveModal: ((value: boolean) => void) | null = null

const modalState = ref<ModalState>({
  visible: false,
  title: '',
  message: '',
  type: 'info',
  isConfirm: false
})

export function useModal() {
  const showModal = (
    title: string,
    message: string,
    type: ModalType = 'info',
    isConfirm = false
  ): Promise<boolean> => {
    modalState.value = { visible: true, title, message, type, isConfirm }
    return new Promise((resolve) => { resolveModal = resolve })
  }

  const confirm = () => {
    modalState.value.visible = false
    if (resolveModal) resolveModal(true)
  }

  const cancel = () => {
    modalState.value.visible = false
    if (resolveModal) resolveModal(false)
  }

  return { modalState, showModal, confirm, cancel }
}
