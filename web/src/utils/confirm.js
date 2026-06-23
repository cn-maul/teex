import { ref } from 'vue'

const confirmState = ref({
  visible: false,
  title: '确认操作',
  message: '',
  confirmText: '确定',
  dangerMode: false,
  _resolve: null
})

export function useConfirm() {
  function showConfirm({ title = '确认操作', message = '', confirmText = '确定', dangerMode = false } = {}) {
    // 拒绝上一个未完成的 confirm
    if (confirmState.value._resolve) {
      confirmState.value._resolve(false)
    }
    return new Promise((resolve) => {
      confirmState.value = {
        visible: true, title, message, confirmText, dangerMode, _resolve: resolve
      }
    })
  }

  function handleConfirm() {
    const resolve = confirmState.value._resolve
    confirmState.value.visible = false
    if (resolve) resolve(true)
  }

  function handleCancel() {
    const resolve = confirmState.value._resolve
    confirmState.value.visible = false
    if (resolve) resolve(false)
  }

  return { confirmState, showConfirm, handleConfirm, handleCancel }
}
