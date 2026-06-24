<template>
  <AdminDashboardView v-if="authStore.isAdmin" />

  <div v-else class="home">
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

    <div v-else class="page-header">
      <div>
        <h1>刷题练习</h1>
        <p class="page-desc">{{ examStore.state.currentExamName || '请选择考试类型' }}</p>
      </div>
      <div class="quick-stats" v-if="modules.length > 0">
        <div class="quick-stat">
          <span class="quick-stat-value">{{ totalQuestions }}</span>
          <span class="quick-stat-label">总题数</span>
        </div>
        <div class="quick-stat">
          <span class="quick-stat-value">{{ totalAnswered }}</span>
          <span class="quick-stat-label">已完成</span>
        </div>
        <div class="quick-stat">
          <span class="quick-stat-value accent">{{ overallAccuracy }}%</span>
          <span class="quick-stat-label">正确率</span>
        </div>
        <div class="quick-stat">
          <span class="quick-stat-value">{{ completionRate }}%</span>
          <span class="quick-stat-label">完成率</span>
        </div>
      </div>
    </div>

    <div v-if="modules.length > 0" class="module-grid">
      <div 
        v-for="(mod, index) in modules" 
        :key="mod.id"
        class="module-card"
        :style="{ '--delay': index * 0.05 + 's' }"
      >
        <div class="module-card-header">
          <div class="module-icon" :class="'icon-' + (index % 6)">
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

        <div class="module-actions" v-if="!authStore.isAdmin">
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
import { ref, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { getExamModules, getExamStats } from '../api'
import { useExamStore } from '../stores/exam'
import { useAuthStore } from '../stores/auth'
import AdminDashboardView from './AdminDashboardView.vue'

const router = useRouter()
const examStore = useExamStore()
const authStore = useAuthStore()

const modules = ref([])
const loading = ref(false)
const correctCount = ref(0)
const examAnswered = ref(0)

const totalQuestions = computed(() => modules.value.reduce((s, m) => s + (m.question_count || 0), 0))
const totalAnswered = computed(() => examAnswered.value)
const overallAccuracy = computed(() => {
  const total = totalAnswered.value
  if (total === 0) return 0
  return Math.round(correctCount.value / total * 100)
})
const completionRate = computed(() => {
  const total = totalQuestions.value
  if (total === 0) return 0
  return Math.round(totalAnswered.value / total * 100)
})

function getModuleIcon(name) {
  if (!name) return '📚'
  if (name.includes('言语')) return '📝'
  if (name.includes('数量') || name.includes('数学')) return '🔢'
  if (name.includes('判断') || name.includes('逻辑')) return '🧩'
  if (name.includes('资料')) return '📊'
  if (name.includes('常识')) return '💡'
  if (name.includes('申论')) return '✍️'
  if (name.includes('政治')) return '⚖️'
  if (name.includes('法律')) return '📜'
  if (name.includes('经济')) return '💰'
  if (name.includes('科技') || name.includes('计算机')) return '💻'
  if (name.includes('历史')) return '🏛️'
  if (name.includes('地理')) return '🌍'
  if (name.includes('农业') || name.includes('农村')) return '🌾'
  if (name.includes('公文')) return '📋'
  return '📚'
}

async function loadModules() {
  if (!examStore.state.currentExamId) return
  if (!localStorage.getItem('token')) return
  loading.value = true
  try {
    const [modRes, statsRes] = await Promise.all([
      getExamModules(examStore.state.currentExamId),
      getExamStats(examStore.state.currentExamId)
    ])
    modules.value = modRes.data.data || []
    // Both correctCount and totalAnswered come from the same stats API for consistent scope
    const examStats = statsRes.data.data || []
    correctCount.value = examStats.reduce((sum, m) => sum + (m.correct_count || 0), 0)
    examAnswered.value = examStats.reduce((sum, m) => sum + (m.total_answered || 0), 0)
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

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.75rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.page-header h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.15rem;
}

.page-desc {
  font-size: 0.9rem;
  color: var(--text-muted);
}

.quick-stats {
  display: flex;
  gap: 1.5rem;
}

.quick-stat {
  text-align: center;
}

.quick-stat-value {
  display: block;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text);
}

.quick-stat-value.accent { color: var(--primary); }

.quick-stat-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.module-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.module-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-top: 3px solid var(--primary);
  border-radius: var(--radius-lg);
  padding: 1.25rem;
  transition: var(--transition);
  animation: fadeIn 0.3s ease backwards;
  animation-delay: var(--delay);
}

.module-card:nth-child(6n+1) { border-top-color: #6366f1; }
.module-card:nth-child(6n+2) { border-top-color: #f59e0b; }
.module-card:nth-child(6n+3) { border-top-color: #10b981; }
.module-card:nth-child(6n+4) { border-top-color: #8b5cf6; }
.module-card:nth-child(6n+5) { border-top-color: #ec4899; }
.module-card:nth-child(6n+6) { border-top-color: #3b82f6; }

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
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--text-secondary);
  border-radius: var(--radius-lg);
  flex-shrink: 0;
}

.icon-0 { background: #eef2ff; }  /* indigo */
.icon-1 { background: #fef3c7; }  /* amber */
.icon-2 { background: #d1fae5; }  /* emerald */
.icon-3 { background: #ede9fe; }  /* violet */
.icon-4 { background: #fce7f3; }  /* pink */
.icon-5 { background: #dbeafe; }  /* blue */

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

.module-actions {
  display: flex;
  gap: 0.5rem;
}

.btn {
  flex: 1;
  padding: 0.55rem 0.875rem;
  border-radius: var(--radius);
  font-size: 0.85rem;
  font-weight: 500;
}
</style>
