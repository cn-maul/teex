<template>
  <div class="settings-section">
    <div class="section-header">
      <h2>刷题偏好</h2>
      <p class="section-desc">调整刷题数量和模式</p>
    </div>

    <div class="setting-item">
      <div class="setting-info">
        <label class="setting-label">每次刷题数量</label>
        <p class="setting-desc">每次开始刷题时加载的题目数</p>
      </div>
      <div class="setting-control range-control">
        <input
          type="range"
          min="5"
          max="50"
          step="5"
          :value="quizCount"
          @input="$emit('update:quizCount', parseInt($event.target.value))"
          class="range-input"
        />
        <span class="range-badge">{{ quizCount }} 题</span>
      </div>
    </div>

    <div class="setting-item">
      <div class="setting-info">
        <label class="setting-label">刷题模式</label>
        <p class="setting-desc">解析模式：逐题作答，即时反馈；考试模式：全部答完再统一评分</p>
      </div>
      <div class="setting-control mode-control">
        <button
          class="mode-btn"
          :class="{ active: quizMode === 'analysis' }"
          @click="$emit('update:quizMode', 'analysis')"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
          解析模式
        </button>
        <button
          class="mode-btn"
          :class="{ active: quizMode === 'exam' }"
          @click="$emit('update:quizMode', 'exam')"
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
          考试模式
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  quizCount: { type: Number, default: 10 },
  quizMode: { type: String, default: 'analysis' },
})

defineEmits(['update:quizCount', 'update:quizMode'])
</script>

<style scoped>
.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  padding: 0.85rem 0;
}

.setting-item + .setting-item {
  border-top: 1px solid var(--border-light);
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

.range-control {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 200px;
}

.range-input {
  flex: 1;
  height: 6px;
  -webkit-appearance: none;
  appearance: none;
  background: var(--border);
  border-radius: 3px;
  outline: none;
}

.range-input::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 18px;
  height: 18px;
  background: var(--primary);
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(99, 102, 241, 0.3);
  transition: transform 0.15s ease;
}

.range-input::-webkit-slider-thumb:hover {
  transform: scale(1.15);
}

.range-badge {
  background: var(--primary-bg);
  color: var(--primary);
  padding: 0.2rem 0.5rem;
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  font-weight: 600;
  white-space: nowrap;
  min-width: 42px;
  text-align: center;
}

.mode-control {
  display: flex;
  gap: 0.5rem;
}

.mode-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1rem;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-lg);
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
}

.mode-btn svg {
  width: 16px;
  height: 16px;
}

.mode-btn:hover {
  border-color: var(--primary-light);
  background: var(--primary-bg);
}

.mode-btn.active {
  border-color: var(--primary);
  background: var(--primary-bg);
  color: var(--primary);
}

@media (max-width: 768px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }
}
</style>
