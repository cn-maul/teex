<template>
  <div class="settings-section">
    <div class="section-header">
      <h2>系统设置</h2>
      <p class="section-desc">管理系统全局配置</p>
    </div>

    <div class="setting-item">
      <div class="setting-info">
        <label class="setting-label">开放注册</label>
        <p class="setting-desc">开启后，用户可以自行注册账号（注册的用户为普通角色）</p>
      </div>
      <div class="setting-control">
        <button
          class="toggle-btn"
          :class="{ active: registrationEnabled }"
          :disabled="registrationLoading"
          @click="$emit('toggle-registration')"
        >
          {{ registrationLoading ? '加载中...' : (registrationEnabled ? '已开放' : '已关闭') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  registrationEnabled: { type: Boolean, default: false },
  registrationLoading: { type: Boolean, default: true },
})

defineEmits(['toggle-registration'])
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

.toggle-btn {
  padding: 0.4rem 1rem;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  color: var(--text-muted);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
  min-width: 80px;
}

.toggle-btn:hover:not(:disabled) {
  border-color: var(--primary);
}

.toggle-btn.active {
  background: var(--success-bg);
  border-color: var(--success);
  color: var(--success);
}

.toggle-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }
}
</style>
