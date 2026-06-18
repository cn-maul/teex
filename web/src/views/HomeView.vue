<template>
  <div class="home">
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <div v-else-if="modules.length === 0" class="empty">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="48" height="48" color="var(--text-muted)">
          <rect x="3" y="3" width="7" height="7"></rect>
          <rect x="14" y="3" width="7" height="7"></rect>
          <rect x="14" y="14" width="7" height="7"></rect>
          <rect x="3" y="14" width="7" height="7"></rect>
        </svg>
      </div>
      <p>暂无模块数据</p>
    </div>

    <div v-else class="module-grid">
      <div 
        v-for="(mod, index) in modules" 
        :key="mod.id"
        class="module-card"
        :style="{ '--delay': index * 0.05 + 's' }"
      >
        <div class="module-card-header">
          <div class="module-icon" :class="'icon-' + (index % 4)">
            {{ getModuleIcon(mod.name) }}
          </div>
          <div class="module-meta">
            <h3>{{ mod.name }}</h3>
            <div class="module-progress-info">
              <span>{{ mod.question_count - mod.unanswered }} / {{ mod.question_count }} 已完成</span>
            </div>
          </div>
        </div>

        <div class="module-progress">
          <div class="progress-track">
            <div 
              class="progress-fill" 
              :style="{ width: mod.question_count > 0 ? ((mod.question_count - mod.unanswered) / mod.question_count * 100) + '%' : '0%' }"
            ></div>
          </div>
        </div>

        <div class="module-actions">
          <button class="btn btn-primary" @click="startQuiz(mod.id, 'default')">
            开始刷题
          </button>
          <button 
            class="btn btn-ghost" 
            @click="startQuiz(mod.id, 'wrong')" 
            :disabled="mod.question_count === mod.unanswered"
          >
            错题重做
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getExamModules } from '../api'
import { useExamStore } from '../stores/exam'

const router = useRouter()
const examStore = useExamStore()

const modules = ref([])
const loading = ref(false)

function getModuleIcon(name) {
  if (!name) return '—'
  return name.charAt(0)
}

async function loadModules() {
  if (!examStore.state.currentExamId) return
  loading.value = true
  try {
    const res = await getExamModules(examStore.state.currentExamId)
    modules.value = res.data.data || []
  } catch (err) {
    console.error('Failed to load modules:', err)
  } finally {
    loading.value = false
  }
}

watch(() => examStore.state.currentExamId, () => {
  loadModules()
}, { immediate: true })



function startQuiz(moduleId, mode) {
  router.push({
    path: `/quiz/${moduleId}`,
    query: mode !== 'default' ? { mode } : {}
  })
}
</script>

<style scoped>
.home {}

.loading {
  text-align: center;
  padding: 4rem;
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty {
  text-align: center;
  padding: 4rem 2rem;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.module-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.module-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 1.25rem;
  transition: var(--transition);
  animation: fadeIn 0.3s ease backwards;
  animation-delay: var(--delay);
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
}

.module-card:hover {
  border-color: var(--primary-light);
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.module-card-header {
  display: flex;
  gap: 0.85rem;
  margin-bottom: 1rem;
}

.module-icon {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text-secondary);
  border-radius: var(--radius-lg);
  flex-shrink: 0;
}

.icon-0 { background: #eef2ff; }
.icon-1 { background: #fef3c7; }
.icon-2 { background: #d1fae5; }
.icon-3 { background: #ede9fe; }

.module-meta {
  flex: 1;
  min-width: 0;
}

.module-meta h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 0.15rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.module-progress-info {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.module-progress {
  margin-bottom: 1rem;
}

.progress-track {
  height: 5px;
  background: var(--border);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 3px;
  transition: width 0.5s ease;
}

.module-actions {
  display: flex;
  gap: 0.5rem;
}

.btn {
  flex: 1;
  padding: 0.55rem 0.875rem;
  border: none;
  border-radius: var(--radius);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-primary {
  background: var(--primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--primary-dark);
}

.btn-ghost {
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border);
}

.btn-ghost:hover:not(:disabled) {
  background: var(--bg-hover);
  border-color: var(--text-muted);
}
</style>
