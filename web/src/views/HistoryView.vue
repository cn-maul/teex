<template>
  <div class="history-view">
    <h1>历史记录</h1>
    
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <p>加载中...</p>
    </div>
    
    <div v-else-if="sessions.length === 0" class="empty">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="48" height="48" color="var(--text-muted)">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
          <line x1="3" y1="9" x2="21" y2="9"></line>
          <line x1="9" y1="21" x2="9" y2="9"></line>
        </svg>
      </div>
      <p class="empty-text">暂无历史记录</p>
      <p class="empty-desc">开始刷题后，这里会显示你的学习记录</p>
      <button class="btn btn-primary" @click="$router.push('/')">去刷题</button>
    </div>
    
    <div v-else>
      <div class="timeline">
        <div 
          v-for="session in sessions" 
          :key="session.id"
          class="timeline-item"
          @click="viewDetail(session)"
        >
          <div class="timeline-dot" :class="getDotClass(session)"></div>
          <div class="timeline-content">
            <div class="session-header">
              <div class="session-info">
                <span class="session-module">{{ session.module?.name || '未知模块' }}</span>
                <span class="session-mode" :class="'mode-' + session.mode">
                  {{ session.mode === 'exam' ? '考试' : session.mode === 'wrong' ? '错题' : '刷题' }}
                </span>
              </div>
              <span class="session-time">{{ formatTime(session.started_at) }}</span>
            </div>
            
            <div class="session-stats">
              <div class="stat-ring-small" v-if="session.total_count > 0">
                <svg viewBox="0 0 40 40">
                  <circle class="ring-bg-sm" cx="20" cy="20" r="16"></circle>
                  <circle 
                    class="ring-fill-sm" 
                    cx="20" cy="20" r="16"
                    :style="{ 
                      strokeDasharray: 100, 
                      strokeDashoffset: 100 - (100 * (session.total_count > 0 ? session.correct_count / session.total_count : 0)),
                      stroke: getAccuracyColor(session)
                    }"
                  ></circle>
                </svg>
                <span class="ring-text-sm">{{ session.total_count > 0 ? Math.round(session.correct_count / session.total_count * 100) : 0 }}%</span>
              </div>
              
              <div class="session-detail">
                <span class="detail-item">
                  <strong>{{ session.correct_count }}</strong>/{{ session.total_count }} 正确
                </span>
                <span class="detail-item" v-if="session.duration > 0">
                  {{ formatDuration(session.duration) }}
                </span>
                <span class="detail-item" v-if="session.finished_at">
                  已完成
                </span>
                <span class="detail-item unfinished" v-else>
                  进行中
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="pagination" v-if="total > pageSize">
        <button class="btn btn-ghost" :disabled="page <= 1" @click="loadPage(page - 1)">← 上一页</button>
        <input
          type="number"
          :value="page"
          @keyup.enter="goToPage($event.target.value)"
          min="1"
          :max="Math.ceil(total / pageSize)"
          class="page-input"
          placeholder="页码"
        />
        <span class="page-info">第 {{ page }} / {{ Math.ceil(total / pageSize) }} 页</span>
        <button class="btn btn-ghost" :disabled="page >= Math.ceil(total / pageSize)" @click="loadPage(page + 1)">下一页 →</button>
      </div>
    </div>

    <!-- 详情弹窗 -->
    <Transition name="modal">
      <div v-if="detailSession" class="modal-overlay" @click.self="detailSession = null">
        <div class="modal-content">
          <div class="modal-header">
            <h2>{{ detailSession.module?.name || '刷题记录' }}</h2>
            <button class="modal-close" @click="detailSession = null">✕</button>
          </div>
          
          <div class="modal-stats">
            <div class="modal-stat">
              <span class="modal-stat-value">{{ detailSession.correct_count }}/{{ detailSession.total_count }}</span>
              <span class="modal-stat-label">正确/总题</span>
            </div>
            <div class="modal-stat">
              <span class="modal-stat-value" :style="{ color: getAccuracyColor(detailSession) }">
                {{ detailSession.total_count > 0 ? Math.round(detailSession.correct_count / detailSession.total_count * 100) : 0 }}%
              </span>
              <span class="modal-stat-label">正确率</span>
            </div>
            <div class="modal-stat" v-if="detailSession.duration > 0">
              <span class="modal-stat-value">{{ formatDuration(detailSession.duration) }}</span>
              <span class="modal-stat-label">用时</span>
            </div>
          </div>
          
          <div v-if="detailLoading" class="detail-loading">加载中...</div>
          
          <div v-else-if="detailAnswers.length === 0" class="detail-empty">暂无答题记录</div>
          
          <div v-else class="detail-answers">
            <div 
              v-for="(answer, idx) in detailAnswers" 
              :key="answer.id"
              class="detail-answer"
              :class="answer.is_correct ? 'answer-correct' : 'answer-wrong'"
            >
              <div class="answer-header">
                <span class="answer-index">第 {{ idx + 1 }} 题</span>
                <span class="answer-badge" :class="answer.is_correct ? 'badge-success' : 'badge-error'">
                  {{ answer.is_correct ? '正确' : '错误' }}
                </span>
              </div>
              <div class="answer-content" v-if="answer.question">{{ answer.question.content }}</div>
              <div class="answer-detail">
                <span>你的答案：<strong :class="answer.is_correct ? 'text-success' : 'text-error'">{{ answer.user_input || '未作答' }}</strong></span>
                <span v-if="!answer.is_correct && answer.question">正确答案：<strong class="text-success">{{ answer.question.answer }}</strong></span>
              </div>
              <div class="answer-analysis" v-if="answer.question?.analysis">
                {{ answer.question.analysis }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getSessions, getSessionAnswers } from '../api'
import { formatDuration } from '../utils/format'

const sessions = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(true)

const detailSession = ref(null)
const detailAnswers = ref([])
const detailLoading = ref(false)

onMounted(async () => {
  await loadSessions()
})

async function loadSessions() {
  loading.value = true
  try {
    const res = await getSessions({ page: page.value, size: pageSize })
    sessions.value = res.data.data || []
    total.value = res.data.total || 0
  } catch (err) {
    console.error('Failed to load sessions:', err)
  } finally {
    loading.value = false
  }
}

function loadPage(p) {
  page.value = p
  loadSessions()
}

function goToPage(val) {
  const p = parseInt(val, 10)
  const max = Math.ceil(total.value / pageSize)
  if (isNaN(p) || p < 1 || p > max) return
  loadPage(p)
}

async function viewDetail(session) {
  detailSession.value = session
  detailAnswers.value = []
  detailLoading.value = true
  try {
    const res = await getSessionAnswers(session.id)
    detailAnswers.value = res.data.data || []
  } catch (err) {
    console.error('Failed to load session answers:', err)
  } finally {
    detailLoading.value = false
  }
}

function getDotClass(session) {
  if (!session.finished_at) return 'dot-unfinished'
  const accuracy = session.total_count > 0 ? session.correct_count / session.total_count : 0
  if (accuracy >= 0.8) return 'dot-good'
  if (accuracy >= 0.6) return 'dot-ok'
  return 'dot-bad'
}

function getAccuracyColor(session) {
  const accuracy = session.total_count > 0 ? session.correct_count / session.total_count : 0
  if (accuracy >= 0.8) return '#10b981'
  if (accuracy >= 0.6) return '#f59e0b'
  return '#ef4444'
}

function formatTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now - d
  
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)} 天前`
  
  return `${d.getMonth() + 1}月${d.getDate()}日 ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

</script>

<style scoped>
.history-view {
  max-width: 800px;
  margin: 0 auto;
}

h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 1.75rem;
}

.loading {
  text-align: center;
  padding: 3rem;
  color: var(--text-muted);
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid var(--border);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty {
  text-align: center;
  padding: 4rem 2rem;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
}

.empty-text {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 0.35rem;
}

.empty-desc {
  font-size: 0.9rem;
  color: var(--text-muted);
  margin-bottom: 1.5rem;
}

/* Timeline */
.timeline {
  position: relative;
}

.timeline-item {
  display: flex;
  gap: 1rem;
  padding: 1.25rem 0;
  border-bottom: 1px solid var(--border-light);
  cursor: pointer;
  transition: background 0.15s;
}

.timeline-item:hover {
  background: var(--bg-hover);
  margin: 0 -1rem;
  padding-left: 1rem;
  padding-right: 1rem;
  border-radius: 8px;
}

.timeline-item:last-child {
  border-bottom: none;
}

.timeline-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  margin-top: 0.35rem;
  flex-shrink: 0;
}

.dot-good { background: var(--success); }
.dot-ok { background: var(--warning); }
.dot-bad { background: var(--error); }
.dot-unfinished { background: var(--text-muted); }

.timeline-content {
  flex: 1;
  min-width: 0;
}

.session-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.session-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.session-module {
  font-weight: 600;
  color: var(--text);
  font-size: 0.95rem;
}

.session-mode {
  font-size: 0.75rem;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  font-weight: 500;
}

.mode-default { background: var(--primary-bg); color: var(--primary); }
.mode-wrong { background: #fef3c7; color: #d97706; }
.mode-exam { background: #fee2e2; color: #dc2626; }

.session-time {
  font-size: 0.8rem;
  color: var(--text-muted);
  white-space: nowrap;
}

.session-stats {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-ring-small {
  position: relative;
  width: 40px;
  height: 40px;
  flex-shrink: 0;
}

.stat-ring-small svg {
  transform: rotate(-90deg);
  width: 100%;
  height: 100%;
}

.ring-bg-sm {
  fill: none;
  stroke: var(--border);
  stroke-width: 4;
}

.ring-fill-sm {
  fill: none;
  stroke-width: 4;
  stroke-linecap: round;
  transition: stroke-dashoffset 0.5s ease;
}

.ring-text-sm {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 0.65rem;
  font-weight: 700;
  color: var(--text);
}

.session-detail {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.detail-item {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.detail-item strong {
  color: var(--text);
}

.detail-item.unfinished {
  color: var(--text-muted);
  font-style: italic;
}

/* Pagination */
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin-top: 1.5rem;
}

.page-info {
  font-size: 0.85rem;
  color: var(--text-muted);
}

.page-input {
  width: 60px;
  padding: 0.4rem 0.5rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  background: var(--bg-card);
  color: var(--text);
  font-size: 0.85rem;
  text-align: center;
  outline: none;
  transition: border-color 0.2s;
}

.page-input:focus {
  border-color: var(--primary);
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: var(--bg-card);
  border-radius: 16px;
  width: 100%;
  max-width: 700px;
  max-height: 85vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border);
}

.modal-header h2 {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--text);
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.25rem;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 4px;
}

.modal-close:hover {
  background: var(--bg-hover);
  color: var(--text);
}

.modal-stats {
  display: flex;
  justify-content: center;
  gap: 2.5rem;
  padding: 1.5rem;
  border-bottom: 1px solid var(--border-light);
}

.modal-stat {
  text-align: center;
}

.modal-stat-value {
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
}

.modal-stat-label {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-top: 0.1rem;
}

.detail-loading, .detail-empty {
  text-align: center;
  padding: 2rem;
  color: var(--text-muted);
  font-size: 0.9rem;
}

.detail-answers {
  padding: 1rem 1.5rem;
}

.detail-answer {
  padding: 1rem;
  border-radius: 8px;
  margin-bottom: 0.75rem;
  border-left: 3px solid;
}

.answer-correct {
  background: #f0fdf4;
  border-color: var(--success);
}

.answer-wrong {
  background: #fef2f2;
  border-color: var(--error);
}

.answer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.answer-index {
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text);
}

.answer-badge {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
}

.badge-success { background: var(--success-bg); color: var(--success); }
.badge-error { background: var(--error-bg); color: var(--error); }

.answer-content {
  font-size: 0.9rem;
  color: var(--text-secondary);
  line-height: 1.5;
  margin-bottom: 0.5rem;
}

.answer-detail {
  display: flex;
  gap: 1rem;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.text-success { color: var(--success); }
.text-error { color: var(--error); }

.answer-analysis {
  font-size: 0.85rem;
  color: var(--text-muted);
  line-height: 1.5;
  margin-top: 0.5rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--border-light);
}

/* Buttons */
.btn {
  padding: 0.6rem 1.25rem;
  border: none;
  border-radius: var(--radius);
  font-size: 0.9rem;
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

.btn-primary:hover {
  background: var(--primary-dark);
}

.btn-ghost {
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border);
}

.btn-ghost:hover:not(:disabled) {
  background: var(--bg-hover);
}

/* Modal transition */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .modal-content,
.modal-leave-active .modal-content {
  transition: transform 0.2s ease;
}

.modal-enter-from .modal-content {
  transform: scale(0.95);
}

.modal-leave-to .modal-content {
  transform: scale(0.95);
}

/* Mobile responsive */
@media (max-width: 768px) {
  .history-view {
    padding: 0;
  }

  h1 {
    font-size: 1.25rem;
  }

  .timeline-item {
    padding: 1rem 0;
  }

  .session-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.35rem;
  }

  .session-info {
    flex-wrap: wrap;
  }

  .session-module {
    font-size: 0.9rem;
  }

  .session-time {
    font-size: 0.75rem;
  }

  .session-stats {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .session-detail {
    gap: 0.5rem;
  }

  .modal-content {
    max-width: 100%;
    margin: 0.5rem;
    max-height: 90vh;
  }

  .modal-stats {
    gap: 1.5rem;
    padding: 1rem;
  }

  .modal-stat-value {
    font-size: 1.25rem;
  }

  .detail-answers {
    padding: 0.75rem 1rem;
  }

  .answer-detail {
    flex-direction: column;
    gap: 0.25rem;
  }

  .pagination {
    gap: 0.5rem;
  }

  .pagination .btn {
    padding: 0.5rem 0.85rem;
    font-size: 0.8rem;
  }
}
</style>
