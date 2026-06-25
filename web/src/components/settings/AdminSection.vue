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

    <div class="setting-item">
      <div class="setting-info">
        <label class="setting-label">批量操作上限</label>
        <p class="setting-desc">单次导入、删除、提交答案的最大数量（1 ~ 10000）</p>
      </div>
      <div class="setting-control number-control">
        <input
          v-model.number="localBatchLimit"
          type="number"
          :min="1"
          :max="10000"
          class="number-input"
          :disabled="batchLimitSaving"
        />
        <button
          class="save-btn"
          :disabled="batchLimitSaving || !isBatchLimitChanged"
          @click="$emit('save-batch-limit', localBatchLimit)"
        >
          {{ batchLimitSaving ? '保存中...' : '保存' }}
        </button>
      </div>
    </div>

    <div class="setting-item">
      <div class="setting-info">
        <label class="setting-label">请求频率限制</label>
        <p class="setting-desc">每分钟每个 IP 最大请求数（10 ~ 10000）。过低可能影响正常使用</p>
      </div>
      <div class="setting-control number-control">
        <input
          v-model.number="localRateLimit"
          type="number"
          :min="10"
          :max="10000"
          class="number-input"
          :disabled="rateLimitSaving"
        />
        <button
          class="save-btn"
          :disabled="rateLimitSaving || !isRateLimitChanged"
          @click="$emit('save-rate-limit', localRateLimit)"
        >
          {{ rateLimitSaving ? '保存中...' : '保存' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  registrationEnabled: { type: Boolean, default: false },
  registrationLoading: { type: Boolean, default: true },
  batchLimit: { type: Number, default: 500 },
  batchLimitSaving: { type: Boolean, default: false },
  rateLimit: { type: Number, default: 120 },
  rateLimitSaving: { type: Boolean, default: false },
})

defineEmits(['toggle-registration', 'save-batch-limit', 'save-rate-limit'])

const localBatchLimit = ref(props.batchLimit)
const isBatchLimitChanged = ref(false)

watch(() => props.batchLimit, (val) => {
  localBatchLimit.value = val
  isBatchLimitChanged.value = false
})

watch(localBatchLimit, (val) => {
  isBatchLimitChanged.value = val !== props.batchLimit && val >= 1 && val <= 10000
})

const localRateLimit = ref(props.rateLimit)
const isRateLimitChanged = ref(false)

watch(() => props.rateLimit, (val) => {
  localRateLimit.value = val
  isRateLimitChanged.value = false
})

watch(localRateLimit, (val) => {
  isRateLimitChanged.value = val !== props.rateLimit && val >= 10 && val <= 10000
})
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

.number-control {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.number-input {
  width: 100px;
  padding: 0.4rem 0.6rem;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  color: var(--text);
  font-size: 0.85rem;
  text-align: center;
  transition: var(--transition);
}

.number-input:focus {
  outline: none;
  border-color: var(--primary);
}

.number-input:disabled {
  opacity: 0.5;
}

/* 隐藏 number input 的上下箭头 */
.number-input::-webkit-inner-spin-button,
.number-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
.number-input[type=number] {
  -moz-appearance: textfield;
}

.save-btn {
  padding: 0.4rem 1rem;
  border: 1.5px solid var(--primary);
  border-radius: var(--radius-lg);
  background: var(--primary);
  color: #fff;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
  min-width: 60px;
}

.save-btn:hover:not(:disabled) {
  opacity: 0.85;
}

.save-btn:disabled {
  opacity: 0.4;
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
