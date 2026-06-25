<template>
  <div class="settings-section">
    <div class="section-header">
      <h2>账户信息</h2>
      <p class="section-desc">管理你的登录密码</p>
    </div>

    <div class="setting-item">
      <div class="setting-info">
        <label class="setting-label">修改密码</label>
        <p class="setting-desc">修改登录密码</p>
      </div>
      <div class="setting-control">
        <button v-if="!showForm" class="btn btn-ghost" @click="showForm = true">修改密码</button>
      </div>
    </div>

    <div v-if="showForm" class="password-form">
      <div class="form-row">
        <label>原密码</label>
        <input v-model="form.oldPassword" type="password" placeholder="输入原密码" />
      </div>
      <div class="form-row">
        <label>新密码</label>
        <input v-model="form.newPassword" type="password" placeholder="至少 6 位" />
      </div>
      <div class="form-row">
        <label>确认密码</label>
        <input v-model="form.confirmPassword" type="password" placeholder="再次输入新密码" />
      </div>
      <div class="form-actions">
        <button class="btn btn-primary btn-sm" @click="handleSubmit" :disabled="!canSubmit">确认修改</button>
        <button class="btn btn-ghost btn-sm" @click="cancel">取消</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'

const props = defineProps({
  saved: { type: Boolean, default: false }
})

const emit = defineEmits(['save-password'])

const showForm = ref(false)
const form = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

watch(() => props.saved, (val) => {
  if (val) cancel()
})

const canSubmit = computed(() => {
  return form.oldPassword && form.newPassword.length >= 6 && form.newPassword === form.confirmPassword
})

function handleSubmit() {
  if (form.newPassword !== form.confirmPassword) return
  emit('save-password', {
    old_password: form.oldPassword,
    new_password: form.newPassword,
  })
}

function cancel() {
  showForm.value = false
  form.oldPassword = ''
  form.newPassword = ''
  form.confirmPassword = ''
}
</script>

<style scoped>
.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  padding: 0.85rem 0;
}

.setting-info {
  flex: 1;
  min-width: 0;
}

.setting-label {
  display: block;
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text);
  margin-bottom: 0.1rem;
}

.setting-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.password-form {
  padding: 1rem 0;
  border-top: 1px solid var(--border-light);
}

.form-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.form-row label {
  width: 70px;
  font-size: 0.85rem;
  color: var(--text-secondary);
  text-align: right;
  flex-shrink: 0;
}

.form-row input {
  flex: 1;
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 0.85rem;
  outline: none;
  transition: border-color 0.2s;
  max-width: 280px;
}

.form-row input:focus {
  border-color: var(--primary);
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  margin-left: 78px;
}

@media (max-width: 768px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .form-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.3rem;
  }

  .form-row label {
    width: auto;
    text-align: left;
  }

  .form-row input {
    max-width: 100%;
    width: 100%;
  }

  .form-actions {
    margin-left: 0;
  }
}
</style>
