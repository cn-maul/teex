<template>
  <div class="quiz-view">
    <!-- 顶部条：模式 + 进度 -->
    <div class="top-bar" v-if="questions.length > 0 && !finished">
      <div class="mode-badge" :class="quizMode === 'exam' ? 'mode-exam' : 'mode-analysis'">
        {{ quizMode === 'exam' ? '考试模式' : '解析模式' }}
      </div>
      <div class="top-progress">
        <span class="progress-text">{{ quizMode === 'exam' ? answeredCount + ' / ' + questions.length : currentIndex + 1 + ' / ' + questions.length }}</span>
      </div>
      <div class="top-stats" v-if="quizMode === 'analysis'">
        <span class="score-correct">✓ {{ correctCount }}</span>
        <span class="score-wrong">✗ {{ wrongCount }}</span>
      </div>
    </div>
    <div class="progress-track" v-if="questions.length > 0 && !finished">
      <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
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

    <!-- ========== 考试模式：答题页面 ========== -->
    <div v-else-if="quizMode === 'exam' && !finished && !showExamResults" class="exam-container">
      <div class="exam-progress-hint">
        <span>已答 {{ answeredCount }} 题，共 {{ questions.length }} 题</span>
      </div>

      <div class="exam-questions">
        <div
          v-for="(q, idx) in questions"
          :key="q.id"
          class="exam-question-card"
          :class="{ 'exam-answered': examSelectedAnswers[idx] }"
        >
          <div class="eq-header">
            <span class="eq-number">{{ idx + 1 }}</span>
            <span class="question-type-badge">{{ getTypeLabel(q.type) }}</span>
            <div class="difficulty">
              <span v-for="i in 5" :key="i" class="star" :class="{ filled: i <= q.difficulty }">★</span>
            </div>
          </div>

          <div class="eq-content">{{ q.content }}</div>

          <div class="options">
            <div
              v-for="(option, oi) in parseOptions(q.options)"
              :key="oi"
              class="option"
              :class="{ selected: examSelectedAnswers[idx] === getOptionLetter(option) }"
              @click="selectExamOption(idx, option)"
            >
              <span class="option-letter">{{ getOptionLetter(option) }}</span>
              <span class="option-text">{{ getOptionText(option) }}</span>
              <svg v-if="examSelectedAnswers[idx] === getOptionLetter(option)" class="option-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
            </div>
          </div>
        </div>
      </div>

      <div class="exam-actions">
        <button class="btn btn-primary btn-lg" :disabled="answeredCount === 0" @click="submitExam">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
          交卷
        </button>
        <span class="exam-actions-hint" v-if="answeredCount < questions.length">还有 {{ questions.length - answeredCount }} 题未作答</span>
      </div>
    </div>

    <!-- ========== 考试模式：交卷后的结果 ========== -->
    <div v-else-if="quizMode === 'exam' && showExamResults" class="finished">
      <div class="finished-card">
        <div class="finished-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="48" height="48" color="var(--success)">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
            <polyline points="22 4 12 14.01 9 11.01"></polyline>
          </svg>
        </div>
        <h2>考试完成</h2>
        <div class="result-ring">
          <svg viewBox="0 0 120 120">
            <circle class="ring-bg" cx="60" cy="60" r="50"></circle>
            <circle
              class="ring-fill"
              cx="60" cy="60" r="50"
              :style="{ strokeDasharray: 314, strokeDashoffset: 314 - (314 * examAccuracy / 100) }"
            ></circle>
          </svg>
          <div class="ring-value">
            <span class="ring-number">{{ examAccuracy }}</span>
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
            <span class="result-value success">{{ examCorrect }}</span>
            <span class="result-label">正确</span>
          </div>
          <div class="result-divider"></div>
          <div class="result-item">
            <span class="result-value error">{{ examWrong }}</span>
            <span class="result-label">错误</span>
          </div>
          <div class="result-divider"></div>
          <div class="result-item">
            <span class="result-value" style="color: var(--text-muted);">{{ examUnanswered }}</span>
            <span class="result-label">未答</span>
          </div>
        </div>
      </div>

      <!-- 全部题目的解析 -->
      <div class="exam-analysis-list">
        <h3 class="analysis-title">题目解析</h3>
        <div
          v-for="(q, idx) in questions"
          :key="q.id"
          class="analysis-item"
          :class="{ 'analysis-correct': examResults[idx]?.is_correct, 'analysis-wrong': examResults[idx] && !examResults[idx].is_correct, 'analysis-unanswered': !examResults[idx] }"
        >
          <div class="analysis-header">
            <span class="analysis-number">{{ idx + 1 }}</span>
            <span class="analysis-type">{{ getTypeLabel(q.type) }}</span>
            <span class="analysis-status" v-if="!examResults[idx]">未作答</span>
            <span class="analysis-status status-correct" v-else-if="examResults[idx].is_correct">✓ 正确</span>
            <span class="analysis-status status-wrong" v-else>✗ 错误</span>
          </div>
          <div class="analysis-content">{{ q.content }}</div>
          <div class="analysis-detail">
            <span class="analysis-user-answer" v-if="examResults[idx]">你的答案：<strong>{{ examResults[idx].user_input || '（未作答）' }}</strong></span>
            <span class="analysis-correct-answer">正确答案：<strong>{{ q.answer }}</strong></span>
          </div>
          <div class="analysis-explanation" v-if="q.analysis">{{ q.analysis }}</div>
          <div class="analysis-options">
            <div
              v-for="(option, oi) in parseOptions(q.options)"
              :key="oi"
              class="analysis-option"
              :class="{
                'aopt-selected': examSelectedAnswers[idx] === getOptionLetter(option),
                'aopt-correct': getOptionLetter(option) === q.answer,
                'aopt-wrong': examSelectedAnswers[idx] === getOptionLetter(option) && getOptionLetter(option) !== q.answer
              }"
            >
              <span class="aopt-letter">{{ getOptionLetter(option) }}</span>
              <span class="aopt-text">{{ getOptionText(option) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="finished-actions">
        <button class="btn btn-primary" @click="restartQuiz">再来一轮</button>
        <button class="btn btn-ghost" @click="$router.back()">返回</button>
      </div>
    </div>

    <!-- ========== 解析模式：逐题作答 ========== -->
    <template v-else-if="quizMode === 'analysis'">
      <!-- 解析模式完成页 -->
      <div v-if="finished" class="finished">
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

        <!-- 全部题目的解析 -->
        <div class="exam-analysis-list">
          <h3 class="analysis-title">题目解析</h3>
          <div
            v-for="(q, idx) in questions"
            :key="q.id"
            class="analysis-item"
            :class="{ 'analysis-correct': analysisResults[idx]?.is_correct, 'analysis-wrong': analysisResults[idx] && !analysisResults[idx].is_correct }"
          >
            <div class="analysis-header">
              <span class="analysis-number">{{ idx + 1 }}</span>
              <span class="analysis-type">{{ getTypeLabel(q.type) }}</span>
              <span class="analysis-status status-correct" v-if="analysisResults[idx]?.is_correct">✓ 正确</span>
              <span class="analysis-status status-wrong" v-else-if="analysisResults[idx]">✗ 错误</span>
            </div>
            <div class="analysis-content">{{ q.content }}</div>
            <div class="analysis-detail">
              <span class="analysis-user-answer" v-if="analysisResults[idx]">你的答案：<strong>{{ analysisResults[idx].user_input }}</strong></span>
              <span class="analysis-correct-answer">正确答案：<strong>{{ q.answer }}</strong></span>
            </div>
            <div class="analysis-explanation" v-if="q.analysis">{{ q.analysis }}</div>
            <div class="analysis-options">
              <div
                v-for="(option, oi) in parseOptions(q.options)"
                :key="oi"
                class="analysis-option"
                :class="{
                  'aopt-selected': getAnalysisSelected(idx) === getOptionLetter(option),
                  'aopt-correct': getOptionLetter(option) === q.answer,
                  'aopt-wrong': getAnalysisSelected(idx) === getOptionLetter(option) && getOptionLetter(option) !== q.answer
                }"
              >
                <span class="aopt-letter">{{ getOptionLetter(option) }}</span>
                <span class="aopt-text">{{ getOptionText(option) }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="finished-actions">
          <button class="btn btn-primary" @click="restartQuiz">再来一轮</button>
          <button class="btn btn-ghost" @click="$router.back()">返回</button>
        </div>
      </div>

      <!-- 解析模式：题目卡片 -->
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
          <button
            v-if="!showFeedback && currentIndex > 0"
            class="btn btn-ghost btn-lg"
            @click="submitEarly"
          >
            交卷
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { startQuiz, submitAnswer as apiSubmitAnswer, submitBatchAnswers } from '../api'
import { useExamStore } from '../stores/exam'

const route = useRoute()
const router = useRouter()
const examStore = useExamStore()

// ====== 通用状态 ======
const questions = ref([])
const loading = ref(true)
const finished = ref(false)
const quizMode = ref(examStore.settings.quizMode || 'analysis')

// ====== 解析模式状态 ======
const currentIndex = ref(0)
const selectedOption = ref('')
const showFeedback = ref(false)
const isCorrect = ref(false)
const correctCount = ref(0)
const wrongCount = ref(0)
const sessionId = ref(null)
const questionStartTime = ref(0)
// 解析模式累积的答题结果（用于终页解析）
const analysisResults = ref([])

// ====== 考试模式状态 ======
const examSelectedAnswers = ref({}) // { [questionIndex]: 'A' | 'B' | ... }
const examResults = ref([])          // AnswerResult[]
const showExamResults = ref(false)
const examSessionId = ref(null)
const examStartTime = ref(0)

// ====== 计算属性 ======

// 解析模式
const currentQuestion = computed(() => questions.value[currentIndex.value] || {})
const progressPercent = computed(() => {
  if (questions.value.length === 0) return 0
  if (quizMode.value === 'exam') {
    return (answeredCount.value / questions.value.length) * 100
  }
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

// 考试模式
const answeredCount = computed(() => {
  return Object.keys(examSelectedAnswers.value).length
})

const examAccuracy = computed(() => {
  const answered = examResults.value.filter(r => r !== undefined)
  if (answered.length === 0) return 0
  const correct = answered.filter(r => r.is_correct).length
  return Math.round((correct / answered.length) * 100)
})

const examCorrect = computed(() => {
  return examResults.value.filter(r => r && r.is_correct).length
})

const examWrong = computed(() => {
  return examResults.value.filter(r => r && !r.is_correct).length
})

const examUnanswered = computed(() => {
  return questions.value.length - examResults.value.length
})

// ====== 生命周期 ======

onMounted(async () => {
  document.addEventListener('keydown', handleKeydown)
  await loadQuestions()
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})

// ====== 键盘快捷键 ======

function handleKeydown(e) {
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') return
  if (loading.value || questions.value.length === 0) return

  const key = e.key.toLowerCase()

  // 考试模式下，快捷键映射到当前可见题目的选项
  if (quizMode.value === 'exam' && !showExamResults.value) {
    if (['a', 'b', 'c', 'd', 'e'].includes(key) && !finished.value) {
      const letterMap = { a: 'A', b: 'B', c: 'C', d: 'D', e: 'E' }
      const letter = letterMap[key]
      // 找到第一个未选中的题目，选择对应选项
      for (let i = 0; i < questions.value.length; i++) {
        if (!examSelectedAnswers.value[i]) {
          const q = questions.value[i]
          const options = parseOptions(q.options)
          const option = options.find(opt => getOptionLetter(opt) === letter)
          if (option) {
            selectExamOption(i, option)
          }
          break
        }
      }
      return
    }
    return
  }

  // 解析模式快捷键
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

// ====== 工具函数 ======

function parseOptions(optionsStr) {
  try {
    return JSON.parse(optionsStr || '[]')
  } catch {
    return []
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

function startTimer() {
  questionStartTime.value = Date.now()
}

function getDuration() {
  if (!questionStartTime.value) return 0
  return Math.floor((Date.now() - questionStartTime.value) / 1000)
}

// ====== 题目加载 ======

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
    examSessionId.value = res.data.session_id || null
    currentIndex.value = 0
    correctCount.value = 0
    wrongCount.value = 0
    finished.value = false
    showFeedback.value = false
    selectedOption.value = ''
    examSelectedAnswers.value = {}
    examResults.value = []
    showExamResults.value = false
    analysisResults.value = []
    startTimer()
    examStartTime.value = Date.now()
  } catch (err) {
    console.error('Failed to start quiz:', err)
  } finally {
    loading.value = false
  }
}

// ====== 解析模式逻辑 ======

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
      duration: getDuration()
    })
    isCorrect.value = res.data.data.is_correct
    showFeedback.value = true
    if (isCorrect.value) correctCount.value++
    else wrongCount.value++
    // 记录结果
    analysisResults.value[currentIndex.value] = res.data.data
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
    startTimer()
  } else {
    finished.value = true
  }
}

// 解析模式：提前交卷
async function submitEarly() {
  finished.value = true
}

// 解析模式结果页获取某题的选项
function getAnalysisSelected(idx) {
  const result = analysisResults.value[idx]
  if (!result) return ''
  return result.user_input || ''
}

// ====== 考试模式逻辑 ======

function selectExamOption(idx, option) {
  if (showExamResults.value) return
  const letter = getOptionLetter(option)
  if (examSelectedAnswers.value[idx] === letter) {
    // 点击已选的则取消
    const newAnswers = { ...examSelectedAnswers.value }
    delete newAnswers[idx]
    examSelectedAnswers.value = newAnswers
  } else {
    examSelectedAnswers.value = { ...examSelectedAnswers.value, [idx]: letter }
  }
}

async function submitExam() {
  if (answeredCount.value === 0) return

  // 构建批量提交数据
  const answers = []
  for (const [idx, userInput] of Object.entries(examSelectedAnswers.value)) {
    const q = questions.value[parseInt(idx)]
    if (!q) continue
    answers.push({
      question_id: q.id,
      user_input: userInput,
      duration: Math.floor((Date.now() - examStartTime.value) / 1000)
    })
  }

  try {
    const res = await submitBatchAnswers({
      answers,
      session_id: examSessionId.value
    })
    // 按 question_id 映射结果
    const resultMap = {}
    res.data.data.forEach((r, i) => {
      resultMap[answers[i].question_id] = r
    })
    const orderedResults = []
    for (let i = 0; i < questions.value.length; i++) {
      const q = questions.value[i]
      const result = resultMap[q.id] || null
      orderedResults[i] = result
    }
    examResults.value = orderedResults
    showExamResults.value = true
  } catch (err) {
    console.error('Failed to submit batch answers:', err)
  }
}

// ====== 公共 ======

async function restartQuiz() {
  await loadQuestions()
}
</script>

<style scoped>
.quiz-view {
  max-width: 720px;
  margin: 0 auto;
}

/* ====== Top bar ====== */
.top-bar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.mode-badge {
  font-size: 0.78rem;
  font-weight: 600;
  padding: 0.15rem 0.6rem;
  border-radius: 20px;
  flex-shrink: 0;
}

.mode-analysis {
  background: #dbeafe;
  color: #2563eb;
}

.mode-exam {
  background: #fce7f3;
  color: #db2777;
}

.top-progress {
  flex: 1;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text);
}

.top-stats {
  display: flex;
  gap: 0.75rem;
  font-size: 0.85rem;
  font-weight: 500;
}

.score-correct { color: var(--success); }
.score-wrong { color: var(--error); }

.progress-track {
  height: 6px;
  background: var(--border);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 1.25rem;
}

.progress-fill {
  height: 100%;
  background: var(--primary);
  border-radius: 3px;
  transition: width 0.3s ease;
}

/* ====== Loading ====== */
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

/* ====== Empty ====== */
.empty, .finished {
  text-align: center;
  padding: 1rem 0;
}

.empty-icon {
  margin-bottom: 1rem;
}

/* ====== Finished card ====== */
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
  margin-top: 2rem;
}

/* ====== Analysis list (shared between modes) ====== */
.exam-analysis-list {
  text-align: left;
  margin-top: 1rem;
}

.analysis-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 1rem;
  text-align: left;
  padding-left: 0.25rem;
}

.analysis-item {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 1.25rem;
  margin-bottom: 1rem;
}

.analysis-correct {
  border-left: 4px solid var(--success);
}

.analysis-wrong {
  border-left: 4px solid var(--error);
}

.analysis-unanswered {
  border-left: 4px solid var(--text-muted);
  opacity: 0.8;
}

.analysis-header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  margin-bottom: 0.6rem;
}

.analysis-number {
  width: 26px;
  height: 26px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-hover);
  border-radius: 50%;
  font-weight: 700;
  font-size: 0.8rem;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.analysis-type {
  font-size: 0.75rem;
  color: var(--text-muted);
  background: var(--bg-hover);
  padding: 0.1rem 0.45rem;
  border-radius: 10px;
}

.analysis-status {
  font-size: 0.8rem;
  font-weight: 600;
  margin-left: auto;
}

.status-correct { color: var(--success); }
.status-wrong { color: var(--error); }

.analysis-content {
  font-size: 0.92rem;
  line-height: 1.6;
  color: var(--text);
  margin-bottom: 0.5rem;
  white-space: pre-wrap;
}

.analysis-detail {
  display: flex;
  gap: 1.25rem;
  font-size: 0.82rem;
  color: var(--text-secondary);
  margin-bottom: 0.4rem;
}

.analysis-user-answer {
  color: var(--text-secondary);
}

.analysis-correct-answer {
  color: var(--success);
}

.analysis-explanation {
  font-size: 0.85rem;
  color: var(--text-muted);
  line-height: 1.5;
  background: var(--bg-hover);
  padding: 0.6rem 0.8rem;
  border-radius: var(--radius-sm);
  margin-bottom: 0.6rem;
}

/* Analysis options */
.analysis-options {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.analysis-option {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 0.85rem;
}

.aopt-letter {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-hover);
  border-radius: 6px;
  font-weight: 600;
  font-size: 0.78rem;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.aopt-selected.aopt-correct {
  border-color: var(--success);
  background: var(--success-bg);
}
.aopt-selected.aopt-correct .aopt-letter {
  background: var(--success);
  color: white;
}

.aopt-selected.aopt-wrong {
  border-color: var(--error);
  background: var(--error-bg);
}
.aopt-selected.aopt-wrong .aopt-letter {
  background: var(--error);
  color: white;
}

.aopt-correct:not(.aopt-selected) {
  border-color: var(--success);
  border-style: dashed;
}
.aopt-correct:not(.aopt-selected) .aopt-letter {
  background: var(--success);
  color: white;
}

.aopt-text {
  flex: 1;
  color: var(--text);
}

/* ====== Exam mode ====== */
.exam-container {
  margin-top: 0.25rem;
}

.exam-progress-hint {
  font-size: 0.82rem;
  color: var(--text-muted);
  margin-bottom: 1rem;
  text-align: center;
}

.exam-questions {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.exam-question-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  padding: 1.5rem;
  transition: border-color 0.2s ease;
}

.exam-answered {
  border-color: var(--primary-light);
}

.eq-header {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  margin-bottom: 0.75rem;
}

.eq-number {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--primary-bg);
  color: var(--primary);
  border-radius: 50%;
  font-weight: 700;
  font-size: 0.82rem;
  flex-shrink: 0;
}

.eq-content {
  font-size: 1rem;
  line-height: 1.7;
  color: var(--text);
  margin-bottom: 1rem;
  white-space: pre-wrap;
}

.exam-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  padding: 1.5rem 0;
  border-top: 1px solid var(--border-light);
}

.exam-actions-hint {
  font-size: 0.8rem;
  color: var(--text-muted);
}

/* ====== Question card (analysis mode) ====== */
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

/* Options (shared) */
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

/* Mobile */
@media (max-width: 768px) {
  .quiz-view {
    max-width: 100%;
  }

  .question-card,
  .exam-question-card {
    padding: 1.25rem;
  }

  .question-content,
  .eq-content {
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

  .exam-questions {
    gap: 0.75rem;
  }

  .analysis-detail {
    flex-direction: column;
    gap: 0.25rem;
  }
}
</style>
