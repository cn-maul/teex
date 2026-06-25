import { defineStore } from 'pinia'
import { getExamTypes } from '../api'

const STORAGE_KEY = 'exam-quiz-current-exam'
const SETTINGS_KEY = 'exam-quiz-settings'
const CACHE_TTL = 5 * 60 * 1000 // 5 分钟

function loadSettings() {
  try {
    const raw = localStorage.getItem(SETTINGS_KEY)
    if (!raw) return {}
    const parsed = JSON.parse(raw)
    if (parsed.quizMode && parsed.quizMode !== 'analysis' && parsed.quizMode !== 'exam') {
      delete parsed.quizMode
    }
    return parsed
  } catch {
    return {}
  }
}

function loadExamFromStorage() {
  if (!localStorage.getItem('token')) return {}
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return {}
    const parsed = JSON.parse(raw)
    return { currentExamId: parsed.id, currentExamName: parsed.name }
  } catch {
    return {}
  }
}

function saveSettings(quizCount, quizMode) {
  localStorage.setItem(SETTINGS_KEY, JSON.stringify({ quizCount, quizMode }))
}

export const useExamStore = defineStore('exam', {
  state: () => {
    const settings = loadSettings()
    const saved = loadExamFromStorage()
    return {
      // 考试列表
      currentExamId: saved.currentExamId || null,
      currentExamName: saved.currentExamName || '',
      examList: [],
      loading: false,
      lastFetchedAt: 0,
      // 刷题设置
      quizCount: settings.quizCount ?? 10,
      quizMode: settings.quizMode ?? 'analysis',
    }
  },

  getters: {
    currentExam: (state) => {
      return state.examList.find(e => e.id === state.currentExamId) || null
    },
    // 兼容旧代码中的 store.settings 访问模式
    settings: (state) => ({
      quizCount: state.quizCount,
      quizMode: state.quizMode,
    }),
  },

  actions: {
    async loadExams() {
      if (this.examList.length > 0 && Date.now() - this.lastFetchedAt < CACHE_TTL) return
      if (!localStorage.getItem('token')) return
      await this._fetchExams()
      if (!this.currentExamId && this.examList.length > 0) {
        this.setExam(this.examList[0])
      }
    },

    async _fetchExams() {
      this.loading = true
      try {
        const res = await getExamTypes()
        this.examList = res.data.data || []
        this.lastFetchedAt = Date.now()
      } catch (err) {
        console.error('Failed to load exams:', err)
      } finally {
        this.loading = false
      }
    },

    setExam(exam) {
      this.currentExamId = exam.id
      this.currentExamName = exam.name
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ id: exam.id, name: exam.name }))
    },

    async refreshExams() {
      await this._fetchExams()
      const stillExists = this.examList.some(e => e.id === this.currentExamId)
      if (!stillExists) {
        if (this.examList.length > 0) {
          this.setExam(this.examList[0])
        } else {
          this.currentExamId = null
          this.currentExamName = ''
          localStorage.removeItem(STORAGE_KEY)
        }
      }
    },

    updateQuizCount(count) {
      this.quizCount = count
      saveSettings(this.quizCount, this.quizMode)
    },

    updateQuizMode(mode) {
      if (mode !== 'analysis' && mode !== 'exam') return
      this.quizMode = mode
      saveSettings(this.quizCount, this.quizMode)
    },
  },
})
