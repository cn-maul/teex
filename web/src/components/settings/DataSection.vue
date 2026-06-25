<template>
  <div class="settings-section">
    <div class="section-header">
      <h2>数据管理</h2>
      <p class="section-desc">管理你的刷题数据</p>
    </div>

    <div class="setting-item">
      <div class="setting-info">
        <label class="setting-label">清空答题记录</label>
        <p class="setting-desc">清除所有答题记录，此操作不可恢复</p>
      </div>
      <div class="setting-control">
        <button class="btn btn-danger" @click="$emit('clear-data')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
          清空记录
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineEmits(['clear-data'])
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

.btn svg {
  width: 15px;
  height: 15px;
}

.btn-danger {
  background: var(--error-bg);
  color: var(--error);
  border: 1px solid #fecaca;
}

.btn-danger:hover {
  background: #fecaca;
}

@media (max-width: 768px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }
}
</style>
