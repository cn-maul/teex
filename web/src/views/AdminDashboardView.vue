<template>
  <div class="admin-dashboard">
    <div class="page-header">
      <h1 class="page-title">数据看板</h1>
      <p class="page-desc">全局数据概览与分析</p>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="empty">
      <p style="color: var(--error)">{{ error }}</p>
      <button class="btn btn-primary" @click="loadDashboard">重试</button>
    </div>

    <template v-else>
      <!-- 指标卡片 -->
      <div class="metrics-row">
        <div class="metric-card">
          <div class="metric-label">总用户数</div>
          <div class="metric-value">{{ dashboard.total_users }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">7 天活跃用户</div>
          <div class="metric-value accent">{{ dashboard.active_users_7d }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">总答题次数</div>
          <div class="metric-value">{{ dashboard.total_answers?.toLocaleString() }}</div>
        </div>
        <div class="metric-card">
          <div class="metric-label">全局正确率</div>
          <div class="metric-value" :style="{ color: dashboard.accuracy >= 60 ? 'var(--success)' : 'var(--error)' }">
            {{ dashboard.accuracy }}%
          </div>
        </div>
      </div>

      <!-- 每日趋势 -->
      <div class="charts-grid">
        <div class="chart-wrapper chart-wide">
          <div class="chart-header">
            <h3 class="chart-title">每日趋势（近 30 天）</h3>
          </div>
          <div class="chart-body-lg">
            <Line v-if="dailyTrendData" :data="dailyTrendData" :options="dailyTrendOptions" />
          </div>
        </div>
      </div>

      <!-- 题型 + 难度统计 -->
      <div class="charts-grid">
        <div class="chart-wrapper">
          <div class="chart-header">
            <h3 class="chart-title">按题型统计</h3>
          </div>
          <div class="chart-body-md">
            <Bar v-if="typeData" :data="typeData" :options="typeOptions" />
          </div>
        </div>
        <div class="chart-wrapper">
          <div class="chart-header">
            <h3 class="chart-title">按难度统计</h3>
          </div>
          <div class="chart-body-md">
            <Bar v-if="diffData" :data="diffData" :options="diffOptions" />
          </div>
        </div>
      </div>

      <!-- 模块统计 + 用户排行 -->
      <div class="charts-grid">
        <div class="chart-wrapper">
          <div class="chart-header">
            <h3 class="chart-title">模块答题量排行</h3>
          </div>
          <div class="chart-body-md">
            <Bar v-if="moduleData" :data="moduleData" :options="moduleOptions" />
          </div>
        </div>
        <div class="chart-wrapper">
          <div class="chart-header">
            <h3 class="chart-title">Top 10 活跃用户</h3>
          </div>
          <div class="table-wrapper">
            <table class="data-table" v-if="dashboard.top_users?.length">
              <thead>
                <tr>
                  <th>#</th>
                  <th>用户</th>
                  <th>答题数</th>
                  <th>正确率</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(u, i) in dashboard.top_users" :key="u.user_id">
                  <td class="rank">{{ i + 1 }}</td>
                  <td>{{ u.nickname || u.username }}</td>
                  <td>{{ u.total_answered }}</td>
                  <td>
                    <span :style="{ color: u.accuracy >= 60 ? 'var(--success)' : 'var(--error)' }">
                      {{ u.accuracy }}%
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
            <div v-else class="empty-small">暂无数据</div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { getAdminDashboardStats } from '../api'
import { TYPE_LABELS } from '../utils/quiz'
import { Line, Bar } from 'vue-chartjs'
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, BarElement, Tooltip, Legend } from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, BarElement, Tooltip, Legend)

const loading = ref(true)
const error = ref('')
const dashboard = ref({
  total_users: 0, active_users_7d: 0, active_users_30d: 0,
  total_questions: 0, total_answers: 0, total_correct: 0, accuracy: 0,
  daily_stats: [], accuracy_by_type: [], accuracy_by_difficulty: [],
  module_stats: [], top_users: []
})

onMounted(() => loadDashboard())

async function loadDashboard() {
  loading.value = true
  error.value = ''
  try {
    const res = await getAdminDashboardStats()
    dashboard.value = res.data.data || dashboard.value
  } catch (err) {
    error.value = '加载失败，请稍后重试'
    console.error('Failed to load admin dashboard:', err)
  } finally {
    loading.value = false
  }
}

// ===== 每日趋势（双线：答题量 + 活跃用户）=====
const dailyTrendData = computed(() => {
  const stats = dashboard.value.daily_stats || []
  if (!stats.length) return null
  return {
    labels: stats.map(d => {
      const parts = d.date.split('-')
      return `${parseInt(parts[1])}/${parseInt(parts[2])}`
    }),
    datasets: [
      {
        label: '答题量',
        data: stats.map(d => d.count),
        borderColor: '#6366f1',
        backgroundColor: 'rgba(99, 102, 241, 0.1)',
        borderWidth: 2, fill: true, tension: 0.35,
        pointRadius: 2, pointBackgroundColor: '#6366f1',
        yAxisID: 'y',
      },
      {
        label: '活跃用户',
        data: stats.map(d => d.active_users),
        borderColor: '#10b981',
        backgroundColor: 'rgba(16, 185, 129, 0.1)',
        borderWidth: 2, fill: true, tension: 0.35,
        pointRadius: 2, pointBackgroundColor: '#10b981',
        yAxisID: 'y1',
      }
    ]
  }
})

const dailyTrendOptions = {
  responsive: true, maintainAspectRatio: false,
  interaction: { mode: 'index', intersect: false },
  plugins: { legend: { position: 'top', labels: { usePointStyle: true, padding: 16 } } },
  scales: {
    y: { position: 'left', beginAtZero: true, grid: { color: 'var(--border)' }, ticks: { color: 'var(--text-muted)' } },
    y1: { position: 'right', beginAtZero: true, grid: { drawOnChartArea: false }, ticks: { color: 'var(--text-muted)' } },
    x: { grid: { display: false }, ticks: { color: 'var(--text-muted)', maxRotation: 45, maxTicksLimit: 15 } }
  }
}

// ===== 题型统计 =====
const typeData = computed(() => {
  const items = dashboard.value.accuracy_by_type || []
  if (!items.length) return null
  return {
    labels: items.map(t => TYPE_LABELS[t.type] || t.type),
    datasets: [{
      label: '正确率 %',
      data: items.map(t => t.accuracy),
      backgroundColor: ['#6366f1', '#8b5cf6', '#10b981', '#f59e0b'],
      borderRadius: 6, barThickness: 32,
    }]
  }
})

const typeOptions = {
  indexAxis: 'y', responsive: true, maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    x: { max: 100, grid: { color: 'var(--border)' }, ticks: { color: 'var(--text-muted)', callback: v => v + '%' } },
    y: { grid: { display: false }, ticks: { color: 'var(--text-muted)' } }
  }
}

// ===== 难度统计 =====
const DIFF_COLORS = ['#10b981', '#34d399', '#f59e0b', '#f97316', '#ef4444']

const diffData = computed(() => {
  const items = dashboard.value.accuracy_by_difficulty || []
  if (!items.length) return null
  const sorted = [...items].sort((a, b) => a.difficulty - b.difficulty)
  return {
    labels: sorted.map(d => `${d.difficulty} 星`),
    datasets: [{
      label: '正确率 %',
      data: sorted.map(d => d.accuracy),
      backgroundColor: sorted.map((_, i) => DIFF_COLORS[i] || '#6366f1'),
      borderRadius: 6, barThickness: 32,
    }]
  }
})

const diffOptions = {
  responsive: true, maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    y: { max: 100, grid: { color: 'var(--border)' }, ticks: { color: 'var(--text-muted)', callback: v => v + '%' } },
    x: { grid: { display: false }, ticks: { color: 'var(--text-muted)' } }
  }
}

// ===== 模块统计 =====
const MODULE_COLORS = ['#6366f1', '#8b5cf6', '#ec4899', '#f59e0b', '#10b981', '#3b82f6', '#ef4444', '#14b8a6']

const moduleData = computed(() => {
  const items = dashboard.value.module_stats || []
  if (!items.length) return null
  const sorted = [...items].sort((a, b) => b.total - a.total).slice(0, 8)
  return {
    labels: sorted.map(m => m.name),
    datasets: [{
      label: '答题量',
      data: sorted.map(m => m.total),
      backgroundColor: sorted.map((_, i) => MODULE_COLORS[i % MODULE_COLORS.length]),
      borderRadius: 6, barThickness: 24,
    }]
  }
})

const moduleOptions = {
  indexAxis: 'y', responsive: true, maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    x: { beginAtZero: true, grid: { color: 'var(--border)' }, ticks: { color: 'var(--text-muted)' } },
    y: { grid: { display: false }, ticks: { color: 'var(--text-muted)' } }
  }
}
</script>

<style scoped>
.admin-dashboard {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1.5rem;
}

.page-header { margin-bottom: 1.5rem; }
.page-title { font-size: 1.5rem; font-weight: 700; color: var(--text); margin: 0 0 0.25rem; }
.page-desc { color: var(--text-muted); font-size: 0.9rem; margin: 0; }

/* 指标卡片 */
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
  text-align: center;
}
.metric-label {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 0.5rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.metric-value {
  font-size: 1.8rem;
  font-weight: 700;
  color: var(--text);
}
.metric-value.accent { color: #6366f1; }

/* 图表网格 */
.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}
.chart-wrapper {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}
.chart-wide { grid-column: 1 / -1; }
.chart-header {
  padding: 1rem 1.25rem 0;
}
.chart-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}
.chart-body-lg { height: 280px; padding: 1rem 1.25rem; }
.chart-body-md { height: 240px; padding: 1rem 1.25rem; }

/* 表格 */
.table-wrapper {
  padding: 0.75rem 1.25rem 1.25rem;
  max-height: 300px;
  overflow-y: auto;
}
.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}
.data-table th {
  text-align: left;
  padding: 0.5rem 0.75rem;
  font-weight: 600;
  color: var(--text-muted);
  border-bottom: 1px solid var(--border);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.data-table td {
  padding: 0.6rem 0.75rem;
  color: var(--text);
  border-bottom: 1px solid var(--border);
}
.data-table tr:last-child td { border-bottom: none; }
.data-table .rank {
  font-weight: 700;
  color: var(--text-muted);
  width: 2rem;
}

.empty-small {
  padding: 2rem;
  text-align: center;
  color: var(--text-muted);
  font-size: 0.9rem;
}

/* 加载/错误 */
.loading { display: flex; justify-content: center; padding: 4rem; }
.empty { text-align: center; padding: 4rem; }

@media (max-width: 768px) {
  .metrics-row { grid-template-columns: repeat(2, 1fr); }
  .charts-grid { grid-template-columns: 1fr; }
  .metric-value { font-size: 1.4rem; }
}
</style>
