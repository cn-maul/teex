/**
 * 图表共享配置和颜色常量
 */

// 题型正确率柱状图颜色
export const TYPE_COLORS = ['#6366f1', '#8b5cf6', '#10b981', '#f59e0b']

// 难度正确率柱状图颜色
export const DIFFICULTY_COLORS = ['#10b981', '#34d399', '#f59e0b', '#f97316', '#ef4444']

// 模块答题量颜色
export const MODULE_COLORS = ['#6366f1', '#8b5cf6', '#ec4899', '#f59e0b', '#10b981', '#3b82f6', '#ef4444', '#14b8a6']

// 通用折线图配置
export const baseLineOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    y: {
      beginAtZero: true,
      max: 100,
      ticks: { callback: v => v + '%', font: { size: 10 } },
      grid: { color: 'rgba(0,0,0,0.05)' }
    },
    x: {
      ticks: { font: { size: 10 }, maxRotation: 45 },
      grid: { display: false }
    }
  }
}

// 通用柱状图配置（垂直）
export const baseBarOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: { legend: { display: false } },
  scales: {
    y: {
      beginAtZero: true,
      ticks: { stepSize: 1, font: { size: 10 } },
      grid: { color: 'rgba(0,0,0,0.05)' }
    },
    x: {
      ticks: { font: { size: 10 }, maxRotation: 45 },
      grid: { display: false }
    }
  }
}
