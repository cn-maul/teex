<template>
  <aside class="stats-panel">
    <div v-if="!examStore.state.currentExamId" class="stats-empty">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="32" height="32" color="var(--text-muted)">
        <path d="M18 20V10"></path><path d="M12 20V4"></path><path d="M6 20v-6"></path>
      </svg>
      <p>请选择考试</p>
    </div>

    <div v-else-if="loading" class="stats-loading">
      <div class="spinner"></div>
    </div>

    <template v-else>
      <div class="stats-header">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16" color="var(--primary)">
          <path d="M18 20V10"></path><path d="M12 20V4"></path><path d="M6 20v-6"></path>
        </svg>
        <span>{{ examStore.state.currentExamName }}</span>
      </div>

      <!-- 总览卡片 -->
      <div class="stats-overview">
        <div class="stat-box">
          <div class="stat-icon">📝</div>
          <span class="stat-value">{{ summary.total }}</span>
          <span class="stat-label">总题数</span>
        </div>
        <div class="stat-box">
          <div class="stat-icon">✅</div>
          <span class="stat-value">{{ summary.answered }}</span>
          <span class="stat-label">已做</span>
        </div>
        <div class="stat-box">
          <div class="stat-icon">🎯</div>
          <span class="stat-value accent">{{ summary.accuracy }}%</span>
          <span class="stat-label">正确率</span>
        </div>
        <div class="stat-box">
          <div class="stat-icon">📋</div>
          <span class="stat-value">{{ summary.unanswered }}</span>
          <span class="stat-label">未做</span>
        </div>
      </div>

      <!-- 雷达图：各科目正确率 -->
      <div class="stats-section-title">能力雷达</div>
      <div class="radar-chart-wrapper" v-if="moduleStats.length >= 3">
        <Radar :data="radarData" :options="radarOptions" />
      </div>
      <div class="stats-empty-hint" v-else>需要至少 3 个科目才有雷达图</div>

      <!-- 各科目进度 -->
      <div class="stats-section-title">各科目进度</div>
      <div class="module-list">
        <div v-for="mod in moduleStats" :key="mod.id" class="module-progress-item">
          <div class="module-progress-header">
            <span class="module-progress-name">{{ mod.name }}</span>
            <span class="module-progress-count">{{ mod.answered }}/{{ mod.total }}</span>
          </div>
          <div class="progress-track">
            <div class="progress-fill" :style="{ width: mod.total > 0 ? (mod.answered / mod.total * 100) + '%' : '0%' }"></div>
          </div>
        </div>
        <div v-if="moduleStats.length === 0" class="stats-empty-hint">暂无科目数据</div>
      </div>
    </template>
  </aside>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import { getExamStats } from '../api'
import { useExamStore } from '../stores/exam'
import { Radar } from 'vue-chartjs'
import { Chart as ChartJS, RadialLinearScale, PointElement, LineElement, Filler, Tooltip, Legend } from 'chart.js'

ChartJS.register(RadialLinearScale, PointElement, LineElement, Filler, Tooltip, Legend)

const examStore = useExamStore()

const loading = ref(false)
const moduleStats = ref([])
const summary = ref({ total: 0, answered: 0, accuracy: 0, unanswered: 0 })

const radarData = computed(() => ({
  labels: moduleStats.value.map(m => m.name.length > 4 ? m.name.slice(0, 4) + '…' : m.name),
  datasets: [{
    label: '正确率',
    data: moduleStats.value.map(m => m.accuracy || 0),
    backgroundColor: 'rgba(99, 102, 241, 0.15)',
    borderColor: '#6366f1',
    borderWidth: 2,
    pointBackgroundColor: '#6366f1',
    pointBorderColor: '#fff',
    pointBorderWidth: 1,
    pointRadius: 3,
  }]
}))

const radarOptions = {
  responsive: true,
  maintainAspectRatio: true,
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx) => `正确率: ${ctx.raw}%`
      }
    }
  },
  scales: {
    r: {
      beginAtZero: true,
      max: 100,
      ticks: { stepSize: 20, display: false },
      grid: { color: 'rgba(0,0,0,0.06)' },
      angleLines: { color: 'rgba(0,0,0,0.06)' },
      pointLabels: { font: { size: 10 }, color: '#64748b' }
    }
  }
}

watch(() => examStore.state.currentExamId, async (id) => {
  if (!id) {
    moduleStats.value = []
    summary.value = { total: 0, answered: 0, accuracy: 0, unanswered: 0 }
    return
  }
  await loadStats(id)
}, { immediate: true })

async function loadStats(examId) {
  loading.value = true
  try {
    const res = await getExamStats(examId)
    const modules = res.data.data || []

    if (modules.length === 0) {
      moduleStats.value = []
      summary.value = { total: 0, answered: 0, accuracy: 0, unanswered: 0 }
      return
    }

    const results = modules.map(mod => ({
      id: mod.id,
      name: mod.name,
      total: mod.total_questions || 0,
      answered: mod.total_answered || 0,
      correct: mod.correct_count || 0,
      accuracy: mod.accuracy || 0,
      unanswered: mod.unanswered ?? ((mod.total_questions || 0) - (mod.total_answered || 0)),
    }))

    moduleStats.value = results

    const totalQ = results.reduce((s, m) => s + m.total, 0)
    const totalA = results.reduce((s, m) => s + m.answered, 0)
    const totalC = results.reduce((s, m) => s + m.correct, 0)
    summary.value = {
      total: totalQ,
      answered: totalA,
      accuracy: totalA > 0 ? Math.round((totalC / totalA) * 100) : 0,
      unanswered: totalQ - totalA,
    }
  } catch (err) {
    console.error('Failed to load stats:', err)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.stats-panel {
  position: fixed;
  right: 0;
  top: 56px;
  bottom: 0;
  width: 260px;
  background: var(--bg-card);
  border-left: 1px solid var(--border);
  z-index: 100;
  overflow-y: auto;
  padding: 1rem;
}

.stats-empty, .stats-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 0.75rem;
  color: var(--text-muted);
  font-size: 0.85rem;
}

.spinner {
  width: 28px;
  height: 28px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }

.stats-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 1rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border);
}

.stats-overview {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
  margin-bottom: 1.25rem;
}

.stat-box {
  background: var(--bg);
  border-radius: var(--radius);
  padding: 0.65rem 0.5rem;
  text-align: center;
}

.stat-value {
  display: block;
  font-size: 1.2rem;
  font-weight: 700;
  color: var(--text);
  line-height: 1.2;
}

.stat-value.accent { color: var(--primary); }

.stat-icon {
  font-size: 1.1rem;
  margin-bottom: 0.15rem;
}

.stat-label {
  font-size: 0.7rem;
  color: var(--text-muted);
  margin-top: 0.1rem;
}

.stats-section-title {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin-bottom: 0.75rem;
}

.module-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.module-progress-item {}

.module-progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.3rem;
}

.module-progress-name {
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 160px;
}

.module-progress-count {
  font-size: 0.7rem;
  color: var(--text-muted);
  flex-shrink: 0;
}

.progress-track {
  height: 4px;
  background: var(--border);
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 2px;
  transition: width 0.5s ease;
}

.radar-chart-wrapper {
  margin-bottom: 1.25rem;
  padding: 0.25rem;
}

.stats-empty-hint {
  text-align: center;
  color: var(--text-muted);
  font-size: 0.8rem;
  padding: 1rem 0;
}
</style>
