<template>
  <div class="exam-view">
    <div class="header">
      <button class="back-btn" @click="$router.push('/')">← 返回</button>
      <h1>{{ examName }}</h1>
    </div>
    
    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="loadError" class="error-state">
      <p>{{ loadError }}</p>
      <button class="btn btn-primary" @click="loadError = ''; loadModules()">重试</button>
    </div>

    <div v-else class="module-grid">
      <div 
        v-for="mod in modules" 
        :key="mod.id"
        class="module-card"
      >
        <div class="module-header">
          <h3>{{ mod.name }}</h3>
          <span class="badge" :class="getBadgeClass(mod)">
            {{ getBadgeText(mod) }}
          </span>
        </div>
        
        <div class="module-stats">
          <div class="stat">
            <span class="stat-value">{{ mod.question_count }}</span>
            <span class="stat-label">总题数</span>
          </div>
          <div class="stat">
            <span class="stat-value">{{ mod.unanswered }}</span>
            <span class="stat-label">未做</span>
          </div>
        </div>

        <!-- 难度筛选 -->
        <div class="difficulty-filter">
          <label class="filter-label">难度：</label>
          <div class="filter-options">
            <button 
              v-for="d in difficultyOptions" 
              :key="d.value"
              class="filter-btn"
              :class="{ active: moduleDifficulty[mod.id] === d.value }"
              @click="setModuleDifficulty(mod.id, d.value)"
            >
              {{ d.label }}
            </button>
          </div>
        </div>

        <!-- 模式选择 + 时间设置 -->
        <div class="mode-settings">
          <div class="mode-row">
            <button 
              class="mode-btn" 
              :class="{ active: moduleMode[mod.id] === 'default' }"
              @click="setModuleMode(mod.id, 'default')"
            >
              刷题
            </button>
            <button 
              class="mode-btn" 
              :class="{ active: moduleMode[mod.id] === 'wrong' }"
              @click="setModuleMode(mod.id, 'wrong')"
            >
              错题
            </button>
          </div>
        </div>
        
        <div class="module-actions">
          <button class="btn btn-primary" @click="startQuiz(mod.id)">
            开始刷题
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getExamModules } from '../api'
import { useExamStore } from '../stores/exam'

const route = useRoute()
const router = useRouter()
const examStore = useExamStore()

const examName = ref('')
const modules = ref([])
const loading = ref(true)
const loadError = ref('')

const moduleDifficulty = reactive({})
const moduleMode = reactive({})

const difficultyOptions = [
  { value: 0, label: '全部' },
  { value: 1, label: '★' },
  { value: 2, label: '★★' },
  { value: 3, label: '★★★' },
  { value: 4, label: '★★★★' },
  { value: 5, label: '★★★★★' },
]

onMounted(async () => {
  await loadModules()
})

async function loadModules() {
  loading.value = true
  loadError.value = ''
  try {
    if (examStore.state.examList.length === 0) {
      await examStore.loadExams()
    }
    const exam = examStore.state.examList.find(e => e.id === parseInt(route.params.id))
    examName.value = exam ? exam.name : '考试'

    const res = await getExamModules(route.params.id)
    modules.value = res.data.data

    // 初始化每个模块的筛选状态
    modules.value.forEach(mod => {
      moduleDifficulty[mod.id] = 0
      moduleMode[mod.id] = 'default'
    })
  } catch (err) {
    console.error('Failed to load modules:', err)
    loadError.value = '加载失败，请检查网络后重试'
  } finally {
    loading.value = false
  }
}

function setModuleDifficulty(moduleId, difficulty) {
  moduleDifficulty[moduleId] = difficulty
}

function setModuleMode(moduleId, mode) {
  moduleMode[moduleId] = mode
}

function getBadgeClass(mod) {
  const done = mod.question_count - mod.unanswered
  if (done === 0) return 'badge-empty'
  if (done >= mod.question_count * 0.8) return 'badge-complete'
  return 'badge-progress'
}

function getBadgeText(mod) {
  const done = mod.question_count - mod.unanswered
  if (done === 0) return '未开始'
  if (done >= mod.question_count * 0.8) return '即将完成'
  return '进行中'
}

function startQuiz(moduleId) {
  const mode = moduleMode[moduleId] || 'default'
  const difficulty = moduleDifficulty[moduleId] || 0
  const query = {}

  if (mode === 'wrong') {
    query.mode = 'wrong'
  }

  if (difficulty > 0) {
    query.difficulty = difficulty
  }

  router.push({ path: `/quiz/${moduleId}`, query })
}
</script>

<style scoped>
.header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
}

.back-btn {
  background: none;
  border: none;
  font-size: 1rem;
  color: var(--primary);
  cursor: pointer;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  transition: background 0.2s;
}

.back-btn:hover {
  background: var(--bg-hover);
}

h1 {
  font-size: 1.8rem;
  color: var(--text);
}

.loading {
  text-align: center;
  padding: 2rem;
  color: var(--text-muted);
}

.error-state {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--text-muted);
}

.error-state p {
  color: var(--error);
  margin-bottom: 1rem;
  font-size: 0.95rem;
}

.module-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 1.5rem;
}

.module-card {
  background: var(--bg-card);
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: var(--shadow);
}

.module-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.module-header h3 {
  font-size: 1.2rem;
  color: var(--text);
  margin: 0;
}

.badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.badge-empty {
  background: var(--bg-hover);
  color: var(--text-muted);
}

.badge-progress {
  background: var(--warning-bg);
  color: var(--warning);
}

.badge-complete {
  background: var(--success-bg);
  color: var(--success);
}

.module-stats {
  display: flex;
  gap: 2rem;
  margin-bottom: 1.25rem;
  padding: 1rem;
  background: var(--bg);
  border-radius: 8px;
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: bold;
  color: var(--primary);
}

.stat-label {
  font-size: 0.85rem;
  color: var(--text-muted);
  margin-top: 0.25rem;
}

/* Difficulty filter */
.difficulty-filter {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.filter-label {
  font-size: 0.85rem;
  color: var(--text-secondary);
  white-space: nowrap;
}

.filter-options {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.filter-btn {
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg-card);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.15s;
  color: var(--text-secondary);
}

.filter-btn:hover {
  border-color: var(--primary);
  color: var(--primary);
}

.filter-btn.active {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}

/* Mode settings */
.mode-settings {
  margin-bottom: 1.25rem;
}

.mode-row {
  display: flex;
  gap: 6px;
  margin-bottom: 0.75rem;
}

.mode-btn {
  flex: 1;
  padding: 0.5rem;
  border: 1.5px solid var(--border);
  border-radius: 8px;
  background: var(--bg-card);
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.15s;
  color: var(--text-secondary);
}

.mode-btn:hover {
  border-color: var(--primary);
}

.mode-btn.active {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}


.module-actions {
  display: flex;
  gap: 0.75rem;
}

.btn {
  flex: 1;
  padding: 0.75rem 1rem;
  border: none;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--primary);
  color: white;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.4);
}

.btn-secondary {
  background: var(--bg-hover);
  color: var(--text-secondary);
}

.btn-secondary:hover {
  background: var(--bg-hover);
}

@media (max-width: 768px) {
  .module-grid {
    grid-template-columns: 1fr;
  }

  h1 {
    font-size: 1.3rem;
  }

  .module-stats {
    gap: 1rem;
  }
}
</style>
