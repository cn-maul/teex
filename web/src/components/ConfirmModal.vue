<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="confirm-overlay" @click.self="cancel">
        <div class="confirm-dialog">
          <div class="confirm-header">
            <h3>{{ title }}</h3>
          </div>
          <div class="confirm-body">
            <p>{{ message }}</p>
          </div>
          <div class="confirm-footer">
            <button class="btn btn-ghost" @click="cancel">取消</button>
            <button class="btn" :class="dangerMode ? 'btn-danger' : 'btn-primary'" @click="confirm">
              {{ confirmText }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
defineProps({
  visible: { type: Boolean, default: false },
  title: { type: String, default: '确认操作' },
  message: { type: String, default: '' },
  confirmText: { type: String, default: '确定' },
  dangerMode: { type: Boolean, default: false }
})

const emit = defineEmits(['confirm', 'cancel'])

const confirm = () => emit('confirm')
const cancel = () => emit('cancel')
</script>

<style scoped>
.confirm-overlay {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.5); display: flex;
  align-items: center; justify-content: center;
  z-index: 2000;
}
.confirm-dialog {
  background: var(--bg-card, #fff); border-radius: 12px;
  padding: 24px; max-width: 420px; width: 90%;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);
}
.confirm-header h3 { margin: 0 0 12px; font-size: 1.1rem; color: var(--text, #1a1a2e); }
.confirm-body p { margin: 0; color: var(--text-secondary, #666); line-height: 1.6; }
.confirm-footer { display: flex; gap: 12px; justify-content: flex-end; margin-top: 24px; }
.btn { padding: 8px 20px; border-radius: 8px; border: none; cursor: pointer; font-size: 0.9rem; }
.btn-primary { background: var(--primary, #667eea); color: #fff; }
.btn-danger { background: var(--error, #ef4444); color: #fff; }
.btn-ghost { background: transparent; border: 1px solid var(--border, #e2e8f0); color: var(--text, #333); }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from .confirm-dialog, .modal-leave-to .confirm-dialog { transform: scale(0.95); }
</style>
