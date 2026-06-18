import { reactive, computed } from 'vue'
import { getExamTypes } from '../api'

const STORAGE_KEY = 'exam-quiz-current-exam'
const SETTINGS_KEY = 'exam-quiz-settings'

// 全局考试状态
const state = reactive({
  currentExamId: null,
  currentExamName: '',
  examList: [],
  loading: false,
})

// 设置状态
const defaultSettings = {
  quizCount: 10,
}

const settingsState = reactive({
  ...defaultSettings,
  ...loadSettings(),
})

function loadSettings() {
  try {
    const raw = localStorage.getItem(SETTINGS_KEY)
    return raw ? JSON.parse(raw) : {}
  } catch {
    return {}
  }
}

function saveSettings() {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify({
    quizCount: settingsState.quizCount,
  }))
}

// 初始化：从 localStorage 恢复
function initFromStorage() {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved) {
    try {
      const parsed = JSON.parse(saved)
      state.currentExamId = parsed.id
      state.currentExamName = parsed.name
    } catch {
      // ignore
    }
  }
}

initFromStorage()

export function useExamStore() {
  // 加载考试列表
  async function loadExams() {
    if (state.examList.length > 0) return
    state.loading = true
    try {
      const res = await getExamTypes()
      state.examList = res.data.data || []
      // 如果没有选中考试，自动选第一个
      if (!state.currentExamId && state.examList.length > 0) {
        setExam(state.examList[0])
      }
    } catch (err) {
      console.error('Failed to load exams:', err)
    } finally {
      state.loading = false
    }
  }

  // 切换考试
  function setExam(exam) {
    state.currentExamId = exam.id
    state.currentExamName = exam.name
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ id: exam.id, name: exam.name }))
  }

  // 当前考试对象
  const currentExam = computed(() => {
    return state.examList.find(e => e.id === state.currentExamId) || null
  })

  // 强制刷新考试列表（用于增删后同步下拉菜单）
  async function refreshExams() {
    state.loading = true
    try {
      const res = await getExamTypes()
      state.examList = res.data.data || []
      const stillExists = state.examList.some(e => e.id === state.currentExamId)
      if (!stillExists) {
        if (state.examList.length > 0) {
          setExam(state.examList[0])
        } else {
          state.currentExamId = null
          state.currentExamName = ''
          localStorage.removeItem(STORAGE_KEY)
        }
      }
    } catch (err) {
      console.error('Failed to refresh exams:', err)
    } finally {
      state.loading = false
    }
  }

  // 设置相关
  function updateQuizCount(count) {
    settingsState.quizCount = count
    saveSettings()
  }

  return {
    state,
    settings: settingsState,
    loadExams,
    refreshExams,
    setExam,
    currentExam,
    updateQuizCount,
  }
}
