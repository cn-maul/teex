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
      <p v-if="loadError" style="color: var(--error); margin-bottom: 0.5rem;">{{ loadError }}</p>
      <p v-else>暂无题目</p>
      <div class="empty-actions">
        <button class="btn btn-primary" @click="loadError ? loadQuestions() : $router.back()">{{ loadError ? '重试' : '返回' }}</button>
        <button v-if="loadError" class="btn btn-ghost" @click="$router.back()">返回</button>
      </div>
    </div>

    <!-- ========== 考试模式：答题页面 ========== -->
    <div v-else-if="quizMode === 'exam' && !finished && !showExamResults" class="exam-container">
      <div class="exam-progress-hint">
        <span>已答 {{ answeredCount }} 题，共 {{ questions.length }} 题</span>
      </div>

      <div class="exam-questions">
        <template v-for="(q, idx) in questions" :key="q.id">
          <div
            v-show="Math.abs(idx - currentExamIndex) <= 10 || examSelectedAnswers[idx]"
            class="exam-question-card"
            :class="{ 'exam-answered': examSelectedAnswers[idx] }"
            :data-exam-idx="idx"
            :ref="el => { if (el) examQuestionRefs[idx] = el }"
          >
            <div class="eq-header">
              <span class="eq-number">{{ idx + 1 }}</span>
              <span class="question-type-badge">{{ getTypeLabel(q.type) }}</span>
              <div class="difficulty">
                <span v-for="i in 5" :key="i" class="star" :class="{ filled: i <= q.difficulty }">★</span>
              </div>
            </div>

            <div class="eq-content">{{ q.content || '（题目内容加载失败）' }}</div>

            <div v-if="q.type === 'fill'" class="fill-input-wrapper">
              <input
                :value="examSelectedAnswers[idx] || ''"
                @input="examSelectedAnswers = { ...examSelectedAnswers, [idx]: $event.target.value }"
                type="text"
                class="fill-input"
                placeholder="请输入答案"
                :disabled="showExamResults"
              />
            </div>
            <div v-else class="options">
              <div
                v-for="(option, oi) in parseOptions(q.options)"
                :key="oi"
                class="option"
                :class="{ selected: isExamOptionSelected(idx, option) }"
                @click="selectExamOption(idx, option)"
              >
                <span class="option-letter">{{ getOptionLetter(option) }}</span>
                <span class="option-text">{{ getOptionText(option) }}</span>
                <svg v-if="isExamOptionSelected(idx, option)" class="option-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
            </div>
          </div>
        </template>
      </div>

      <div class="exam-actions">
        <button class="btn btn-primary btn-lg" :disabled="answeredCount === 0 || submitting" @click="submitExam">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" width="16" height="16"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
          交卷
        </button>
        <span class="exam-actions-hint" v-if="answeredCount < questions.length">还有 {{ questions.length - answeredCount }} 题未作答</span>
      </div>
    </div>

    <!-- ========== 考试模式：交卷后的结果 ========== -->
    <div v-else-if="quizMode === 'exam' && showExamResults" class="finished">
      <QuizResultCard
        title="考试完成"
        :total="questions.length"
        :correct="examCorrect"
        :wrong="examWrong"
        :unanswered="examUnanswered"
        :accuracy="examAccuracy"
        :type-chart-data="typeChartData"
        :difficulty-chart-data="examDifficultyChartData"
        :show-unanswered="true"
      />

      <QuizAnalysisList
        :questions="questions"
        :results="examResults"
        :is-exam-mode="true"
      />

      <div class="finished-actions">
        <button class="btn btn-primary" @click="restartQuiz">再来一轮</button>
        <button class="btn btn-ghost" @click="$router.back()">返回</button>
      </div>
    </div>

    <!-- ========== 解析模式：逐题作答 ========== -->
    <template v-else-if="quizMode === 'analysis'">
      <!-- 解析模式完成页 -->
      <div v-if="finished" class="finished">
        <QuizResultCard
          title="刷题完成"
          :total="questions.length"
          :correct="correctCount"
          :wrong="wrongCount"
          :accuracy="accuracy"
          :type-chart-data="typeChartData"
          :difficulty-chart-data="analysisDifficultyChartData"
        />

        <QuizAnalysisList
          :questions="questions"
          :results="analysisResults"
        />

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
          <p>{{ currentQuestion.content || '（题目内容加载失败）' }}</p>
        </div>

        <!-- 填空题输入框 -->
        <div v-if="currentQuestion.type === 'fill'" class="fill-input-wrapper">
          <input
            v-model="fillAnswer"
            type="text"
            class="fill-input"
            placeholder="请输入答案"
            @keyup.enter="showFeedback ? nextQuestion() : submitSingleAnswer()"
            :disabled="showFeedback"
          />
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
            :disabled="currentQuestion.type === 'fill' ? !fillAnswer.trim() : selectedOptions.size === 0"
            @click="submitSingleAnswer"
          >
            提交答案
          </button>
          <button v-else class="btn btn-primary btn-lg" @click="nextQuestion">
            {{ currentIndex < questions.length - 1 ? '下一题' : '查看结果' }}
          </button>
        </div>
      </div>
    </template>

    <!-- 兜底：防止未知模式白屏 -->
    <div v-else class="empty">
      <p>加载异常，请刷新页面重试</p>
      <button class="btn btn-primary" @click="$router.back()">返回</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { startQuiz, submitAnswer as apiSubmitAnswer, submitBatchAnswers } from '../api'
import { useExamStore } from '../stores/exam'
import { useAuthStore } from '../stores/auth'
import { getTypeLabel, parseOptions, getOptionLetter, getOptionText } from '../utils/quiz'
import { showToast } from '../utils/toast'
import { calcAccuracy } from '../utils/format'
import { useConfirm } from '../utils/confirm'
import QuizResultCard from '../components/quiz/QuizResultCard.vue'
import QuizAnalysisList from '../components/quiz/QuizAnalysisList.vue'

const route = useRoute()
const router = useRouter()
const examStore = useExamStore()
const authStore = useAuthStore()
const { showConfirm } = useConfirm()

// ====== 通用状态 ======
const questions = ref([])
const loading = ref(true)
const finished = ref(false)
const loadError = ref('')

// 管理员不能答题，直接拦截（必须在 ref 声明之后）
if (authStore.isAdmin) {
  loading.value = false
  loadError.value = '管理员不能参与答题，请使用管理员面板查看数据'
}
// 只允许 'analysis' 或 'exam'，其他值一律回退到 'analysis'
const quizMode = computed(() => {
  const mode = examStore.quizMode
  return mode === 'exam' || mode === 'analysis' ? mode : 'analysis'
})

// ====== 解析模式状态 ======
const currentIndex = ref(0)
const selectedOptions = ref(new Set()) // 多选题支持多选
const showFeedback = ref(false)
const isCorrect = ref(false)
const correctCount = ref(0)
const wrongCount = ref(0)
const sessionId = ref(null)
const questionStartTime = ref(0)
const fillAnswer = ref('')
// 解析模式累积的答题结果（用于终页解析）
const analysisResults = ref([])

// ====== 考试模式状态 ======
const examSelectedAnswers = ref({}) // { [questionIndex]: 'A' | 'B' | ... }
const examResults = ref([])          // AnswerResult[]
const showExamResults = ref(false)
const examSessionId = ref(null)
const examStartTime = ref(0)
const currentExamIndex = ref(0)      // 当前可视题目索引（用于虚拟渲染优化）
const submitting = ref(false)         // 防止交卷重复点击
const examQuestionRefs = {}          // 题目 DOM 引用
let examObserver = null              // IntersectionObserver 实例

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
const accuracy = computed(() => calcAccuracy(correctCount.value, correctCount.value + wrongCount.value))

const parsedOptions = computed(() => {
  return parseOptions(currentQuestion.value.options)
})

// 考试模式
const answeredCount = computed(() => {
  return Object.keys(examSelectedAnswers.value).length
})

const examAccuracy = computed(() => {
  if (questions.value.length === 0) return 0
  const correct = examResults.value.filter(r => r && r.is_correct).length
  return Math.round((correct / questions.value.length) * 100)
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

// ====== Chart 数据 ======

const typeChartData = computed(() => {
  const counts = { single: 0, multi: 0, judge: 0, fill: 0 }
  questions.value.forEach(q => { if (counts[q.type] !== undefined) counts[q.type]++ })
  return {
    labels: ['单选题', '多选题', '判断题', '填空题'],
    datasets: [{
      data: [counts.single, counts.multi, counts.judge, counts.fill],
      backgroundColor: ['#6366f1', '#8b5cf6', '#10b981', '#f59e0b'],
      borderWidth: 0,
      hoverOffset: 4
    }]
  }
})

const examDifficultyChartData = computed(() => {
  const diffMap = {}
  questions.value.forEach((q, idx) => {
    const d = q.difficulty || 1
    if (!diffMap[d]) diffMap[d] = { correct: 0, total: 0 }
    if (examResults.value[idx]) {
      diffMap[d].total++
      if (examResults.value[idx].is_correct) diffMap[d].correct++
    }
  })
  const levels = [1, 2, 3, 4, 5]
  return {
    labels: levels.map(l => `难度 ${l}`),
    datasets: [{
      label: '正确率',
      data: levels.map(l => {
        const d = diffMap[l]
        return d && d.total > 0 ? Math.round(d.correct / d.total * 100) : 0
      }),
      backgroundColor: ['#10b981', '#34d399', '#f59e0b', '#f97316', '#ef4444'],
      borderRadius: 6,
      borderSkipped: false,
    }]
  }
})

const analysisDifficultyChartData = computed(() => {
  const diffMap = {}
  questions.value.forEach((q, idx) => {
    const d = q.difficulty || 1
    if (!diffMap[d]) diffMap[d] = { correct: 0, total: 0 }
    if (analysisResults.value[idx]) {
      diffMap[d].total++
      if (analysisResults.value[idx].is_correct) diffMap[d].correct++
    }
  })
  const levels = [1, 2, 3, 4, 5]
  return {
    labels: levels.map(l => `难度 ${l}`),
    datasets: [{
      label: '正确率',
      data: levels.map(l => {
        const d = diffMap[l]
        return d && d.total > 0 ? Math.round(d.correct / d.total * 100) : 0
      }),
      backgroundColor: ['#10b981', '#34d399', '#f59e0b', '#f97316', '#ef4444'],
      borderRadius: 6,
      borderSkipped: false,
    }]
  }
})

// ====== 生命周期 ======

onMounted(async () => {
  document.addEventListener('keydown', handleKeydown)
  await loadQuestions()
})

watch(() => route.params.moduleId, (newId, oldId) => {
  if (newId && newId !== oldId) {
    loadQuestions()
  }
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  disconnectExamObserver()
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
      // 使用当前可视题目（由 IntersectionObserver 追踪的 currentExamIndex）
      const i = currentExamIndex.value
      if (i >= 0 && i < questions.value.length && !examSelectedAnswers.value[i]) {
        const q = questions.value[i]
        const options = parseOptions(q.options)
        const option = options.find(opt => getOptionLetter(opt) === letter)
        if (option) {
          selectExamOption(i, option)
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
    }
    return
  }

  // Enter - submit or next
  if (e.key === 'Enter') {
    e.preventDefault()
    if (showFeedback.value) {
      nextQuestion()
    } else if (selectedOptions.value.size > 0) {
      submitSingleAnswer()
    }
    return
  }
}

// ====== 工具函数 ======

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
  loadError.value = ''
  for (const key in examQuestionRefs) {
    delete examQuestionRefs[key]
  }
  try {
    const mode = route.query.mode || 'default'
    const count = examStore.quizCount || 10
    const difficulty = parseInt(route.query.difficulty) || 0
    const moduleId = parseInt(route.params.moduleId)
    if (!moduleId || isNaN(moduleId)) {
      loadError.value = '无效的模块ID'
      loading.value = false
      return
    }
    const res = await startQuiz({
      module_id: moduleId,
      count,
      mode,
      difficulty
    })
    questions.value = res.data.data || []
    sessionId.value = res.data.session_id || null
    examSessionId.value = res.data.session_id || null
    currentIndex.value = 0
    correctCount.value = 0
    wrongCount.value = 0
    finished.value = false
    showFeedback.value = false
    selectedOptions.value = new Set()
    fillAnswer.value = ''
    examSelectedAnswers.value = {}
    examResults.value = []
    showExamResults.value = false
    analysisResults.value = []
    currentExamIndex.value = 0
    startTimer()
    examStartTime.value = Date.now()
    // 考试模式下设置 IntersectionObserver 追踪可视题目
    if (quizMode.value === 'exam') {
      nextTick(() => setupExamObserver())
    }
  } catch (err) {
    console.error('[QuizView] Failed to start quiz:', err)
    console.error('[QuizView] Error response:', err.response?.data)
    console.error('[QuizView] Route params:', route.params, 'Query:', route.query)
    const serverMsg = err.response?.data?.error
    if (err.response?.status === 401) {
      loadError.value = '登录已过期，请重新登录'
    } else if (serverMsg) {
      loadError.value = serverMsg
    } else {
      loadError.value = '加载题目失败，请检查网络后重试'
    }
  } finally {
    loading.value = false
  }
}

// ====== 解析模式逻辑 ======

function isSelected(option) {
  const letter = getOptionLetter(option)
  return selectedOptions.value.has(letter)
}

function selectOption(option) {
  if (finished.value || showFeedback.value) return
  if (currentQuestion.value.type === 'fill') return
  const letter = getOptionLetter(option)
  if (currentQuestion.value.type === 'multi') {
    // 多选题：切换选中状态
    const newSet = new Set(selectedOptions.value)
    if (newSet.has(letter)) {
      newSet.delete(letter)
    } else {
      newSet.add(letter)
    }
    selectedOptions.value = newSet
  } else {
    // 单选/判断：直接替换
    selectedOptions.value = new Set([letter])
  }
}

function getOptionClass(option) {
  const letter = getOptionLetter(option)
  const classes = []

  if (selectedOptions.value.has(letter)) classes.push('selected')
  if (showFeedback.value) {
    // 多选题的正确答案可能包含多个字母
    const correctLetters = new Set((currentQuestion.value.answer || '').split(',').map(s => s.trim()))
    if (correctLetters.has(letter)) classes.push('correct')
    else if (selectedOptions.value.has(letter)) classes.push('wrong')
  }

  return classes.join(' ')
}

async function submitSingleAnswer() {
  let userInput
  if (currentQuestion.value.type === 'fill') {
    userInput = fillAnswer.value.trim()
  } else if (currentQuestion.value.type === 'multi') {
    // 多选题：排序后用逗号拼接
    userInput = [...selectedOptions.value].sort().join(',')
  } else {
    userInput = [...selectedOptions.value][0] || ''
  }
  if (!userInput) return
  try {
    const res = await apiSubmitAnswer({
      question_id: currentQuestion.value.id,
      user_input: userInput,
      duration: getDuration(),
      session_id: sessionId.value || 0
    })
    isCorrect.value = res.data.data.is_correct
    showFeedback.value = true
    if (isCorrect.value) correctCount.value++
    else wrongCount.value++
    // 记录结果
    analysisResults.value[currentIndex.value] = res.data.data
  } catch (err) {
    console.error('Failed to submit answer:', err)
    showToast('提交失败，请检查网络后重试', 'error')
  }
}

function nextQuestion() {
  if (currentIndex.value < questions.value.length - 1) {
    currentIndex.value++
    selectedOptions.value = new Set()
    fillAnswer.value = ''
    showFeedback.value = false
    isCorrect.value = false
    startTimer()
  } else {
    finished.value = true
  }
}

// ====== 考试模式可视范围优化 ======

function setupExamObserver() {
  disconnectExamObserver()
  if (!window.IntersectionObserver) return
  examObserver = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          const idx = parseInt(entry.target.dataset.examIdx, 10)
          if (!isNaN(idx)) {
            currentExamIndex.value = idx
          }
        }
      }
    },
    { rootMargin: '-10% 0px -60% 0px' }
  )
  nextTick(() => {
    for (const idx in examQuestionRefs) {
      const el = examQuestionRefs[idx]
      if (el) examObserver.observe(el)
    }
  })
}

function disconnectExamObserver() {
  if (examObserver) {
    examObserver.disconnect()
    examObserver = null
  }
}

watch(currentExamIndex, () => {
  nextTick(() => {
    if (!examObserver) return
    for (const idx in examQuestionRefs) {
      const el = examQuestionRefs[idx]
      if (el) examObserver.observe(el)
    }
  })
})

// ====== 考试模式逻辑 ======

function isExamOptionSelected(idx, option) {
  const letter = getOptionLetter(option)
  const answer = examSelectedAnswers.value[idx] || ''
  return answer.split(',').includes(letter)
}

function isExamOptionCorrect(idx, option) {
  const letter = getOptionLetter(option)
  const q = questions.value[idx]
  if (!q) return false
  return (q.answer || '').split(',').map(s => s.trim()).includes(letter)
}

function selectExamOption(idx, option) {
  if (showExamResults.value) return
  const letter = getOptionLetter(option)
  const q = questions.value[idx]
  if (q && q.type === 'multi') {
    // 多选题：切换选中状态，逗号分隔存储
    const current = examSelectedAnswers.value[idx] || ''
    const selected = current ? current.split(',') : []
    const newSet = new Set(selected)
    if (newSet.has(letter)) {
      newSet.delete(letter)
    } else {
      newSet.add(letter)
    }
    const newAnswers = { ...examSelectedAnswers.value }
    const sorted = [...newSet].sort()
    if (sorted.length === 0) {
      delete newAnswers[idx]
    } else {
      newAnswers[idx] = sorted.join(',')
    }
    examSelectedAnswers.value = newAnswers
  } else {
    // 单选/判断：直接替换
    if (examSelectedAnswers.value[idx] === letter) {
      const newAnswers = { ...examSelectedAnswers.value }
      delete newAnswers[idx]
      examSelectedAnswers.value = newAnswers
    } else {
      examSelectedAnswers.value = { ...examSelectedAnswers.value, [idx]: letter }
    }
  }
}

async function submitExam() {
  if (answeredCount.value === 0 || submitting.value) return
  const unanswered = questions.value.length - answeredCount.value
  const msg = unanswered > 0
    ? `还有 ${unanswered} 题未作答，确定要交卷吗？`
    : '确定要交卷吗？'
  if (!await showConfirm({ message: msg })) return
  submitting.value = true

  // 构建批量提交数据（将总时间均分给每道已答题）
  const answers = []
  const totalTime = Math.floor((Date.now() - examStartTime.value) / 1000)
  const perQuestionTime = Math.max(1, Math.floor(totalTime / answeredCount.value))

  for (const [idx, userInput] of Object.entries(examSelectedAnswers.value)) {
    const q = questions.value[parseInt(idx)]
    if (!q) continue
    answers.push({
      question_id: q.id,
      user_input: userInput,
      duration: perQuestionTime
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
    disconnectExamObserver()
    showExamResults.value = true
  } catch (err) {
    console.error('Failed to submit batch answers:', err)
    const status = err.response?.status
    const serverMsg = err.response?.data?.error
    if (status === 409) {
      showToast(serverMsg || '该场次已结束，无法再次交卷', 'error')
    } else if (status === 404) {
      showToast(serverMsg || '考试场次不存在', 'error')
    } else {
      showToast('交卷失败，请检查网络后重试', 'error')
    }
  } finally {
    submitting.value = false
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

/* ====== Empty ====== */
.empty, .finished {
  text-align: center;
  padding: 1rem 0;
}

.empty-icon {
  margin-bottom: 1rem;
}

.empty-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
  margin-top: 1rem;
}

/* ====== Finished card ====== */
.finished-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
  margin-top: 2rem;
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
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.feedback-error {
  background: var(--error-bg);
  border: 1px solid rgba(239, 68, 68, 0.3);
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

/* Fill type input */
.fill-input-wrapper {
  margin-bottom: 1.5rem;
}

.fill-input {
  width: 100%;
  padding: 0.85rem 1.1rem;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-lg);
  font-size: 0.95rem;
  color: var(--text);
  background: var(--bg-card);
  outline: none;
  transition: var(--transition);
  box-sizing: border-box;
}

.fill-input:focus {
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
}

.fill-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.fill-input::placeholder {
  color: var(--text-muted);
}
</style>
