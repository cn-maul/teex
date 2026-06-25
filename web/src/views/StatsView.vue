<template>
  <div class="stats-view">
    <div class="page-header">
      <div>
        <h1>学习统计</h1>
        <p class="page-desc">全面了解你的学习状况</p>
      </div>
    </div>

    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>加载统计数据...</p>
    </div>

    <div v-else-if="error" class="empty">
      <p>{{ error }}</p>
      <button class="btn btn-primary" @click="loadDashboard">重试</button>
    </div>

    <template v-else>
      <!-- 顶部指标卡片 -->
      <div class="metrics-row">
        <div class="metric-card">
          <div class="metric-icon">🔥</div>
          <div class="metric-body">
            <span class="metric-value">{{ dashboard.streak_days }}</span>
            <span class="metric-label">连续学习天数</span>
          </div>
        </div>
        <div class="metric-card">
          <div class="metric-icon">📊</div>
          <div class="metric-body">
            <span class="metric-value">{{ dashboard.total_answered }}</span>
            <span class="metric-label">累计答题</span>
          </div>
        </div>
        <div class="metric-card">
          <div class="metric-icon">🎯</div>
          <div class="metric-body">
            <span class="metric-value accent">{{ dashboard.accuracy }}%</span>
            <span class="metric-label">总体正确率</span>
          </div>
        </div>
        <div class="metric-card">
          <div class="metric-icon">📝</div>
          <div class="metric-body">
            <span class="metric-value">{{ dashboard.total_questions }}</span>
            <span class="metric-label">总题数</span>
          </div>
        </div>
      </div>

      <!-- 图表区域 -->
      <div class="charts-grid">
        <!-- 正确率趋势 -->
        <div class="chart-card chart-wide">
          <h3 class="chart-card-title">📈 正确率趋势（近30天）</h3>
          <div class="chart-wrapper-lg">
            <Line v-if="dashboard.daily_stats?.length > 0" :data="trendData" :options="trendOptions" />
            <div v-else class="chart-empty">暂无数据</div>
          </div>
        </div>

        <!-- 雷达图：模块能力 -->
        <div class="chart-card">
          <h3 class="chart-card-title">🎯 模块能力雷达</h3>
          <div class="chart-wrapper-md">
            <Radar v-if="dashboard.module_stats?.length >= 3" :data="moduleRadarData" :options="radarOptions" />
            <div v-else class="chart-empty">需要至少 3 个科目</div>
          </div>
        </div>

        <!-- 题型正确率 -->
        <div class="chart-card">
          <h3 class="chart-card-title">📋 题型正确率</h3>
          <div class="chart-wrapper-md">
            <Bar v-if="dashboard.accuracy_by_type?.length > 0" :data="typeBarData" :options="horizontalBarOptions" />
            <div v-else class="chart-empty">暂无数据</div>
          </div>
        </div>

        <!-- 难度分布 -->
        <div class="chart-card">
          <h3 class="chart-card-title">⚡ 难度正确率</h3>
          <div class="chart-wrapper-md">
            <Bar v-if="dashboard.accuracy_by_difficulty?.length > 0" :data="difficultyBarData" :options="verticalBarOptions" />
            <div v-else class="chart-empty">暂无数据</div>
          </div>
        </div>

        <!-- 每日练习量 -->
        <div class="chart-card">
          <h3 class="chart-card-title">📊 每日练习量</h3>
          <div class="chart-wrapper-md">
            <Bar v-if="dashboard.daily_stats?.length > 0" :data="dailyBarData" :options="dailyBarOptions" />
            <div v-else class="chart-empty">暂无数据</div>
          </div>
        </div>
      </div>

      <!-- 最近练习记录 -->
      <div class="recent-section" v-if="dashboard.recent_sessions?.length > 0">
        <h3 class="section-title">最近练习</h3>
        <div class="recent-list">
          <div v-for="s in dashboard.recent_sessions" :key="s.id" class="recent-item">
            <div class="recent-info">
              <span class="recent-module">{{ s.module_name }}</span>
              <span class="recent-mode" :class="'mode-' + s.mode">
                {{ s.mode === 'exam' ? '考试' : s.mode === 'wrong' ? '错题' : '刷题' }}
              </span>
            </div>
            <div class="recent-stats">
              <span class="recent-accuracy" :style="{ color: getAccuracyColor(s.accuracy) }">{{ s.accuracy }}%</span>
              <span class="recent-detail">{{ s.correct_count }}/{{ s.total_count }}</span>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getDashboardStats } from '../api'
import { TYPE_LABELS } from '../utils/quiz'
import { getAccuracyColor } from '../utils/format'
import { Line, Bar, Radar } from 'vue-chartjs'

const loading = ref(true)
const error = ref('')
const dashboard = ref({
  total_answered: 0, total_correct: 0, accuracy: 0, total_questions: 0,
  streak_days: 0, daily_stats: [], accuracy_by_type: [],
  accuracy_by_difficulty: [], recent_sessions: [], module_stats: []
})

onMounted(() => loadDashboard())

async function loadDashboard() {
  loading.value = true
  error.value = ''
  try {
    const res = await getDashboardStats()
    dashboard.value = res.data.data || dashboard.value
  } catch (err) {
    error.value = '加载失败，请稍后重试'
    console.error('Failed to load dashboard:', err)
  } finally {
    loading.value = false
  }
}

// ===== Chart Data =====

const trendData = computed(() => {
  const stats = dashboard.value.daily_stats || []
  return {
    labels: stats.map(d => {
      const parts = d.date.split('-')
      return `${parseInt(parts[1])}/${parseInt(parts[2])}`
    }),
    datasets: [{
      label: '正确率',
      data: stats.map(d => d.count > 0 ? Math.round(d.correct / d.count * 100) : 0),
      borderColor: '#6366f1',
      backgroundColor: 'rgba(99, 102, 241, 0.1)',
      borderWidth: 2,
      fill: true,
      tension: 0.35,
      pointRadius: 3,
      pointBackgroundColor: '#6366f1',
      pointBorderColor: '#fff',
      pointBorderWidth: 2,
    }]
  }
})

const trendOptions = {
  responsive: true, maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    y: { beginAtZero: true, max: 100, ticks: { callback: v => v + '%' }, grid: { color: 'rgba(0,0,0,0.05)' } },
    x: { ticks: { maxRotation: 0, autoSkip: true, maxTicksLimit: 10 }, grid: { display: false } }
  }
}

const moduleRadarData = computed(() => {
  const ms = dashboard.value.module_stats || []
  return {
    labels: ms.map(m => m.name.length > 5 ? m.name.slice(0, 5) + '…' : m.name),
    datasets: [{
      label: '正确率',
      data: ms.map(m => m.accuracy || 0),
      backgroundColor: 'rgba(99, 102, 241, 0.15)',
      borderColor: '#6366f1',
      borderWidth: 2,
      pointBackgroundColor: '#6366f1',
      pointRadius: 3,
    }]
  }
})

const radarOptions = {
  responsive: true, maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    r: { beginAtZero: true, max: 100, ticks: { display: false }, grid: { color: 'rgba(0,0,0,0.06)' }, pointLabels: { font: { size: 10 } } }
  }
}

const typeBarData = computed(() => {
  const types = dashboard.value.accuracy_by_type || []
  return {
    labels: types.map(t => TYPE_LABELS[t.type] || t.type),
    datasets: [{
      label: '正确率',
      data: types.map(t => t.accuracy),
      backgroundColor: ['#6366f1', '#8b5cf6', '#10b981', '#f59e0b'],
      borderRadius: 6, borderSkipped: false,
    }]
  }
})

const horizontalBarOptions = {
  responsive: true, maintainAspectRatio: false, indexAxis: 'y',
  plugins: { legend: { display: false } },
  scales: {
    x: { beginAtZero: true, max: 100, ticks: { callback: v => v + '%' }, grid: { color: 'rgba(0,0,0,0.05)' } },
    y: { grid: { display: false } }
  }
}

const difficultyBarData = computed(() => {
  const diffs = dashboard.value.accuracy_by_difficulty || []
  return {
    labels: diffs.map(d => `难度 ${d.difficulty}`),
    datasets: [{
      label: '正确率',
      data: diffs.map(d => d.accuracy),
      backgroundColor: ['#10b981', '#34d399', '#f59e0b', '#f97316', '#ef4444'],
      borderRadius: 6, borderSkipped: false,
    }]
  }
})

const verticalBarOptions = {
  responsive: true, maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    y: { beginAtZero: true, max: 100, ticks: { callback: v => v + '%' }, grid: { color: 'rgba(0,0,0,0.05)' } },
    x: { grid: { display: false } }
  }
}

const dailyBarData = computed(() => {
  const stats = dashboard.value.daily_stats || []
  return {
    labels: stats.map(d => {
      const parts = d.date.split('-')
      return `${parseInt(parts[1])}/${parseInt(parts[2])}`
    }),
    datasets: [{
      label: '答题数',
      data: stats.map(d => d.count),
      backgroundColor: '#818cf8',
      borderRadius: 4, borderSkipped: false,
    }]
  }
})

const dailyBarOptions = {
  responsive: true, maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    y: { beginAtZero: true, ticks: { stepSize: 1 }, grid: { color: 'rgba(0,0,0,0.05)' } },
    x: { ticks: { maxRotation: 0, autoSkip: true, maxTicksLimit: 10 }, grid: { display: false } }
  }
}
</script>

<style scoped>
.stats-view {
  max-width: 960px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 1.75rem;
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

/* Metrics */
.metrics-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.metric-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 0.85rem;
}

.metric-icon {
  font-size: 1.75rem;
  flex-shrink: 0;
}

.metric-value {
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  line-height: 1.2;
}

.metric-value.accent { color: var(--primary); }

.metric-label {
  font-size: 0.8rem;
  color: var(--text-muted);
}

/* Charts grid */
.charts-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.chart-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 1.25rem;
}

.chart-wide {
  grid-column: 1 / -1;
}

.chart-card-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 0.75rem;
}

.chart-wrapper-lg {
  height: 260px;
  position: relative;
}

.chart-wrapper-md {
  height: 220px;
  position: relative;
}

.chart-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-muted);
  font-size: 0.85rem;
}

/* Recent sessions */
.recent-section {
  margin-bottom: 2rem;
}

.section-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 0.75rem;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.recent-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: 0.75rem 1rem;
}

.recent-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.recent-module {
  font-weight: 500;
  color: var(--text);
  font-size: 0.9rem;
}

.recent-mode {
  font-size: 0.7rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  font-weight: 500;
}

.mode-default { background: var(--primary-bg); color: var(--primary); }
.mode-wrong { background: var(--warning-bg); color: #d97706; }
.mode-exam { background: var(--error-bg); color: #dc2626; }

.recent-stats {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.recent-accuracy {
  font-weight: 700;
  font-size: 0.95rem;
}

.recent-detail {
  font-size: 0.8rem;
  color: var(--text-muted);
}

@media (max-width: 768px) {
  .metrics-row {
    grid-template-columns: repeat(2, 1fr);
  }

  .charts-grid {
    grid-template-columns: 1fr;
  }

  .metric-card {
    padding: 1rem;
  }

  .metric-icon {
    font-size: 1.25rem;
  }

  .metric-value {
    font-size: 1.2rem;
  }
}
</style>
