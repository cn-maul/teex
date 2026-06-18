<template>
  <div class="quiz-view">
    <!-- 进度条 -->
    <div class="progress-bar" v-if="questions.length > 0">
      <div class="progress-info">
        <span class="progress-text">{{ currentIndex + 1 }} / {{ questions.length }}</span>
        <div class="progress-center"></div>
        <span class="progress-score">
          <span class="score-correct">✓ {{ correctCount }}</span>
          <span class="score-wrong">✗ {{ wrongCount }}</span>
        </span>
      </div>
      <div class="progress-track">
        <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
      </div>
    </div>

    <!-- 加载状态 -->
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
    </div>

    <!-- 无题目 -->
    <div v-else-if="questions.length === 0" class="empty">
      <div class="empty-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" width="48" height="48" color="var(--text-muted)">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
          <polyline points="14 2 14 8 20 8"></polyline>
        </svg>
      </div>
      <p>暂无题目</p>
      <button class="btn btn-primary" @click="$router.back()">返回</button>
    </div>

    <!-- 刷题完成 -->
    <div v-else-if="finished" class="finished">
      <div class="finished-card">
        <div class="finished-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="48" height="48" color="var(--success)">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
            <polyline points="22 4 12 14.01 9 11.01"></polyline>
          </svg>
        </div>
        <h2>刷题完成</h2>
        <div class="result-ring">
          <svg viewBox="0 0 120 120">
            <circle class="ring-bg" cx="60" cy="60" r="50"></circle>
            <circle 
              class="ring-fill" 
              cx="60" cy="60" r="50"
              :style="{ strokeDasharray: 314, strokeDashoffset: 314 - (314 * accuracy / 100) }"
            ></circle>
          </svg>
          <div class="ring-value">
            <span class="ring-number">{{ accuracy }}</span>
            <span class="ring-unit">%</span>
          </div>
        </div>
        <div class="result-stats">
          <div class="result-item">
            <span class="result-value total">{{ questions.length }}</span>
            <span class="result-label">总题数</span>
          </div>
          <div class="result-divider"></div>
          <div class="result-item">
            <span class="result-value success">{{ correctCount }}</span>
            <span class="result-label">正确</span>
          </div>
          <div class="result-divider"></div>
          <div class="result-item">
            <span class="result-value error">{{ wrongCount }}</span>
            <span class="result-label">错误</span>
          </div>
        </div>
      </div>

      <div class="finished-actions">
        <button class="btn btn-primary" @click="restartQuiz">再来一轮</button>
        <button class="btn btn-ghost" @click="$router.back()">返回</button>
      </div>
    </div>

    <!-- 题目卡片 -->
    <div v-else class="question-card">
      <div class="question-header">
        <span class="question-type-badge">{{ getTypeLabel(currentQuestion.type) }}</span>
        <div class="difficulty">
          <span v-for="i in 5" :key="i" class="star" :class="{ filled: i <= currentQuestion.difficulty }">★</span>
        </div>
      </div>

      <div class="question-content">
        <p>{{ currentQuestion.content }}</p>
      </div>

      <div class="options">
        <div 
          v-for="(option, index) in parsedOptions" 
          :key="index"
          class="option"
          :class="getOptionClass(option)"
          @click="selectOption(option)"
        >
          <span class="option-letter">{{ getOptionLetter(option) }}</span>
          <span class="option-text">{{ getOptionText(option) }}</span>
          <svg v-if="isSelected(option) && !finished" class="option-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
        </div>
      </div>

      <!-- 即时反馈 -->
      <Transition name="slide">
        <div v-if="showFeedback" class="feedback-card" :class="isCorrect ? 'feedback-success' : 'feedback-error'">
          <div class="feedback-icon">{{ isCorrect ? '✓' : '✗' }}</div>
          <div class="feedback-body">
            <div class="feedback-title">{{ isCorrect ? '回答正确' : '回答错误' }}</div>
            <div class="feedback-answer">正确答案：{{ currentQuestion.answer }}</div>
            <div class="feedback-analysis" v-if="currentQuestion.analysis">{{ currentQuestion.analysis }}</div>
          </div>
        </div>
      </Transition>

      <!-- 底部操作 -->
      <div class="question-actions">
        <button 
          v-if="!showFeedback" 
          class="btn btn-primary btn-lg" 
          :disabled="!selectedOption"
          @click="submitSingleAnswer"
        >
          提交答案
        </button>
        <button v-else class="btn btn-primary btn-lg" @click="nextQuestion">
          {{ currentIndex < questions.length - 1 ? '下一题' : '查看结果' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { startQuiz, submitAnswer as apiSubmitAnswer } from '../api'
import { useExamStore } from '../stores/exam'

const route = useRoute()
const router = useRouter()
const examStore = useExamStore()

const questions = ref([])
const currentIndex = ref(0)
const selectedOption = ref('')
const showFeedback = ref(false)
const isCorrect = ref(false)
const loading = ref(true)
const finished = ref(false)
const correctCount = ref(0)
const wrongCount = ref(0)
const sessionId = ref(null)

const currentQuestion = computed(() => questions.value[currentIndex.value] || {})
const progressPercent = computed(() => {
  if (questions.value.length === 0) return 0
  return ((currentIndex.value + 1) / questions.value.length) * 100
})
const accuracy = computed(() => {
  const total = correctCount.value + wrongCount.value
  if (total === 0) return 0
  return Math.round((correctCount.value / total) * 100)
})

const parsedOptions = computed(() => {
  try {
    return JSON.parse(currentQuestion.value.options || '[]')
  } catch {
    return []
  }
})

onMounted(async () => {
  document.addEventListener('keydown', handleKeydown)
  await loadQuestions()
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

function handleKeydown(e) {
  // Don't fire shortcuts when typing in input fields
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') return
  // Don't fire when quiz hasn't loaded or is finished
  if (loading.value || questions.value.length === 0) return

  const key = e.key.toLowerCase()

  // A/B/C/D - Select the corresponding option
  if (['a', 'b', 'c', 'd'].includes(key) && !finished.value) {
    const letterMap = { a: 'A', b: 'B', c: 'C', d: 'D' }
    const letter = letterMap[key]
    const option = parsedOptions.value.find(opt => getOptionLetter(opt) === letter)
    if (option) {
      selectOption(option)
      if (!showFeedback.value && selectedOption.value) {
        setTimeout(() => submitSingleAnswer(), 50)
      }
    }
    return
  }

  // Enter - submit or next
  if (e.key === 'Enter') {
    e.preventDefault()
    if (showFeedback.value) {
      nextQuestion()
    } else if (selectedOption.value) {
      submitSingleAnswer()
    }
    return
  }
}

async function loadQuestions() {
  loading.value = true
  try {
    const mode = route.query.mode || 'default'
    const count = examStore.settings.quizCount || 10
    const difficulty = parseInt(route.query.difficulty) || 0
    const res = await startQuiz({
      module_id: parseInt(route.params.moduleId),
      count,
      mode,
      difficulty
    })
    questions.value = res.data.data
    sessionId.value = res.data.session_id || null
    currentIndex.value = 0
    correctCount.value = 0
    wrongCount.value = 0
    finished.value = false
    showFeedback.value = false
    selectedOption.value = ''
  } catch (err) {
    console.error('Failed to start quiz:', err)
  } finally {
    loading.value = false
  }
}

function getTypeLabel(type) {
  const labels = { single: '单选题', multi: '多选题', judge: '判断题', fill: '填空题' }
  return labels[type] || '单选题'
}

function getOptionLetter(option) {
  if (!option) return ''
  return option.charAt(0)
}

function getOptionText(option) {
  if (!option || option.length < 2) return option || ''
  return option.substring(1).replace(/^[\s.、\s]+/, '').trim()
}

function isSelected(option) {
  const letter = getOptionLetter(option)
  return selectedOption.value === letter
}

function selectOption(option) {
  if (finished.value || showFeedback.value) return
  const letter = getOptionLetter(option)
  selectedOption.value = letter
}

function getOptionClass(option) {
  const letter = getOptionLetter(option)
  const classes = []

  if (selectedOption.value === letter) classes.push('selected')
  if (showFeedback.value) {
    if (letter === currentQuestion.value.answer) classes.push('correct')
    else if (selectedOption.value === letter) classes.push('wrong')
  }

  return classes.join(' ')
}

async function submitSingleAnswer() {
  if (!selectedOption.value) return
  try {
    const res = await apiSubmitAnswer({
      question_id: currentQuestion.value.id,
      user_input: selectedOption.value,
      duration: 0
    })
    isCorrect.value = res.data.data.is_correct
    showFeedback.value = true
    if (isCorrect.value) correctCount.value++
    else wrongCount.value++
  } catch (err) {
    console.error('Failed to submit answer:', err)
  }
}

function nextQuestion() {
  if (currentIndex.value < questions.value.length - 1) {
    currentIndex.value++
    selectedOption.value = ''
    showFeedback.value = false
    isCorrect.value = false
  } else {
    finished.value = true
  }
}


async function restartQuiz() {
  await loadQuestions()
}
</script>

<style scoped>
.quiz-view {
  max-width: 720px;
  margin: 0 auto;
}

/* Progress bar */
.progress-bar {
  margin-bottom: 1.25rem;
}

.progress-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.progress-text {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text);
}

.progress-score {
  display: flex;
  gap: 0.75rem;
  font-size: 0.85rem;
  font-weight: 500;
}

.score-correct { color: var(--success); }
.score-wrong { color: var(--error); }

.progress-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.progress-track {
  height: 6px;
  background: var(--border);
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 3px;
  transition: width 0.3s ease;
}

/* Loading */
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

/* Empty */
.empty, .finished {
  text-align: center;
  padding: 3rem;
}

.empty-icon {
  margin-bottom: 1rem;
}

/* Finished */
.finished-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  padding: 2.5rem 2rem;
  margin-bottom: 2rem;
}

.finished-icon {
  margin-bottom: 0.75rem;
}

.finished-card h2 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 1.5rem;
}

.result-ring {
  position: relative;
  width: 140px;
  height: 140px;
  margin: 0 auto 1.5rem;
}

.result-ring svg {
  transform: rotate(-90deg);
  width: 100%;
  height: 100%;
}

.ring-bg {
  fill: none;
  stroke: var(--border);
  stroke-width: 8;
}

.ring-fill {
  fill: none;
  stroke: var(--primary);
  stroke-width: 8;
  stroke-linecap: round;
  transition: stroke-dashoffset 1s ease;
}

.ring-value {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.ring-number {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text);
}

.ring-unit {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-muted);
}

.result-stats {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 2rem;
}

.result-item {
  text-align: center;
}

.result-value {
  display: block;
  font-size: 1.75rem;
  font-weight: 700;
}

.result-value.total { color: var(--text); }
.result-value.success { color: var(--success); }
.result-value.error { color: var(--error); }

.result-label {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-top: 0.1rem;
}

.result-divider {
  width: 1px;
  height: 2.5rem;
  background: var(--border);
}

.finished-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
}

/* Question card */
.question-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  padding: 2rem;
}

.question-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
}

.question-type-badge {
  background: var(--primary-bg);
  color: var(--primary);
  padding: 0.2rem 0.75rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 600;
}

.difficulty {
  display: flex;
  gap: 1px;
}

.star {
  color: var(--border);
  font-size: 0.85rem;
}

.star.filled {
  color: var(--warning);
}

.question-content {
  font-size: 1.1rem;
  line-height: 1.8;
  color: var(--text);
  margin-bottom: 1.5rem;
  white-space: pre-wrap;
}

/* Options */
.options {
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  margin-bottom: 1.5rem;
}

.option {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  padding: 0.85rem 1.1rem;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-lg);
  cursor: pointer;
  transition: var(--transition);
}

.option:hover:not(.correct):not(.wrong) {
  border-color: var(--primary-light);
  background: var(--primary-bg);
}

.option.selected {
  border-color: var(--primary);
  background: var(--primary-bg);
}

.option.correct {
  border-color: var(--success);
  background: var(--success-bg);
}

.option.wrong {
  border-color: var(--error);
  background: var(--error-bg);
}

.option-letter {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-hover);
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.85rem;
  color: var(--text-secondary);
  flex-shrink: 0;
  transition: var(--transition);
}

.option.selected .option-letter {
  background: var(--primary);
  color: white;
}

.option.correct .option-letter {
  background: var(--success);
  color: white;
}

.option.wrong .option-letter {
  background: var(--error);
  color: white;
}

.option-text {
  flex: 1;
  font-size: 0.95rem;
  color: var(--text);
}

.option-check {
  width: 18px;
  height: 18px;
  color: var(--primary);
  flex-shrink: 0;
}

/* Feedback */
.feedback-card {
  display: flex;
  gap: 1rem;
  padding: 1.25rem;
  border-radius: var(--radius-lg);
  margin-bottom: 1.25rem;
}

.feedback-success {
  background: var(--success-bg);
  border: 1px solid #bbf7d0;
}

.feedback-error {
  background: var(--error-bg);
  border: 1px solid #fecaca;
}

.feedback-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.9rem;
  flex-shrink: 0;
}

.feedback-success .feedback-icon {
  background: var(--success);
  color: white;
}

.feedback-error .feedback-icon {
  background: var(--error);
  color: white;
}

.feedback-body {
  flex: 1;
}

.feedback-title {
  font-weight: 600;
  font-size: 0.95rem;
  margin-bottom: 0.25rem;
}

.feedback-success .feedback-title { color: #065f46; }
.feedback-error .feedback-title { color: #991b1b; }

.feedback-answer {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 0.25rem;
}

.feedback-analysis {
  font-size: 0.85rem;
  color: var(--text-muted);
  line-height: 1.5;
}

.slide-enter-active,
.slide-leave-active {
  transition: all 0.25s ease;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* Actions */
.question-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
  margin-top: 0.5rem;
}

.btn {
  padding: 0.6rem 1.5rem;
  border: none;
  border-radius: var(--radius);
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-lg {
  padding: 0.7rem 1.75rem;
  font-size: 0.95rem;
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

/* Mobile responsive */
@media (max-width: 768px) {
  .quiz-view {
    max-width: 100%;
  }

  .question-card {
    padding: 1.25rem;
  }

  .question-content {
    font-size: 1rem;
    line-height: 1.7;
  }

  .option {
    padding: 0.75rem 0.85rem;
  }

  .option-text {
    font-size: 0.9rem;
  }

  .question-actions {
    flex-direction: column;
    gap: 0.5rem;
  }

  .question-actions .btn {
    width: 100%;
    justify-content: center;
  }

  .progress-info {
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .finished-card {
    padding: 1.5rem 1rem;
  }

  .result-ring {
    width: 120px;
    height: 120px;
  }

  .ring-number {
    font-size: 1.75rem;
  }

  .result-stats {
    gap: 1.25rem;
  }

  .finished-actions {
    flex-direction: column;
  }

  .finished-actions .btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
