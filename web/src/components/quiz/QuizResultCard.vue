<template>
  <div class="finished-card">
    <div class="finished-icon">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="48" height="48" color="var(--success)">
        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
        <polyline points="22 4 12 14.01 9 11.01"></polyline>
      </svg>
    </div>
    <h2>{{ title }}</h2>
    <ScoreRing :percentage="accuracy" />
    <div class="result-stats">
      <div class="result-item">
        <span class="result-value total">{{ total }}</span>
        <span class="result-label">总题数</span>
      </div>
      <div class="result-divider"></div>
      <div class="result-item">
        <span class="result-value success">{{ correct }}</span>
        <span class="result-label">正确</span>
      </div>
      <div class="result-divider"></div>
      <div class="result-item">
        <span class="result-value error">{{ wrong }}</span>
        <span class="result-label">错误</span>
      </div>
      <template v-if="showUnanswered">
        <div class="result-divider"></div>
        <div class="result-item">
          <span class="result-value" style="color: var(--text-muted);">{{ unanswered }}</span>
          <span class="result-label">未答</span>
        </div>
      </template>
    </div>
    <div class="result-chart-section">
      <h3 class="chart-title">题型分布</h3>
      <div class="chart-container-sm">
        <Doughnut :data="typeChartData" :options="doughnutOptions" />
      </div>
    </div>
    <div class="result-chart-section">
      <h3 class="chart-title">难度正确率</h3>
      <div class="chart-container-sm">
        <Bar :data="difficultyChartData" :options="barOptions" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { Doughnut, Bar } from 'vue-chartjs'
import ScoreRing from './ScoreRing.vue'

defineProps({
  title: { type: String, required: true },
  total: { type: Number, required: true },
  correct: { type: Number, required: true },
  wrong: { type: Number, required: true },
  unanswered: { type: Number, default: 0 },
  accuracy: { type: Number, required: true },
  typeChartData: { type: Object, required: true },
  difficultyChartData: { type: Object, required: true },
  showUnanswered: { type: Boolean, default: false },
})

const doughnutOptions = {
  responsive: true,
  maintainAspectRatio: true,
  plugins: {
    legend: {
      position: 'bottom',
      labels: { padding: 12, usePointStyle: true, pointStyleWidth: 8, font: { size: 12 } }
    }
  },
  cutout: '60%'
}

const barOptions = {
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
    y: {
      beginAtZero: true,
      max: 100,
      ticks: { callback: (v) => v + '%', font: { size: 11 } },
      grid: { color: 'rgba(0,0,0,0.05)' }
    },
    x: {
      ticks: { font: { size: 11 } },
      grid: { display: false }
    }
  }
}
</script>

<style scoped>
.finished-card {
  background: var(--card-bg);
  border-radius: 16px;
  padding: 2rem;
  text-align: center;
  margin-bottom: 1.5rem;
}

.finished-icon {
  margin-bottom: 0.5rem;
}

.finished-card h2 {
  font-size: 1.3rem;
  margin: 0 0 1rem;
  color: var(--text);
}

.result-stats {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin: 1.5rem 0;
}

.result-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 60px;
}

.result-value {
  font-size: 1.4rem;
  font-weight: 700;
}

.result-value.total { color: var(--text); }
.result-value.success { color: var(--success); }
.result-value.error { color: var(--error); }

.result-label {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: 0.15rem;
}

.result-divider {
  width: 1px;
  height: 30px;
  background: var(--border);
}

.result-chart-section {
  margin-top: 1.2rem;
}

.chart-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 0.5rem;
}

.chart-container-sm {
  max-width: 280px;
  margin: 0 auto;
}
</style>
