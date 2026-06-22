import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 简易 Toast 通知
function showToast(message, type = 'error') {
  const existing = document.querySelector('.toast-notification')
  if (existing) existing.remove()

  const toast = document.createElement('div')
  toast.className = 'toast-notification'
  toast.style.cssText = `
    position: fixed; top: 20px; right: 20px; z-index: 10000;
    padding: 12px 20px; border-radius: 8px; font-size: 14px; font-weight: 500;
    color: white; max-width: 400px; word-break: break-word;
    box-shadow: 0 4px 12px rgba(0,0,0,0.15); animation: toast-in 0.3s ease;
    background: ${type === 'error' ? '#ef4444' : type === 'success' ? '#22c55e' : '#3b82f6'};
  `
  toast.textContent = message
  document.body.appendChild(toast)

  // 添加动画样式
  if (!document.querySelector('#toast-style')) {
    const style = document.createElement('style')
    style.id = 'toast-style'
    style.textContent = `
      @keyframes toast-in { from { opacity: 0; transform: translateY(-10px); } to { opacity: 1; transform: translateY(0); } }
      @keyframes toast-out { from { opacity: 1; } to { opacity: 0; transform: translateY(-10px); } }
    `
    document.head.appendChild(style)
  }

  setTimeout(() => {
    toast.style.animation = 'toast-out 0.3s ease forwards'
    setTimeout(() => toast.remove(), 300)
  }, 3000)
}

// 响应拦截器：统一错误处理
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data?.error
      || (error.code === 'ECONNABORTED' ? '请求超时，请稍后重试' : '网络错误，请检查连接')
    showToast(message)
    return Promise.reject(error)
  }
)

// 考试类型
export const getExamTypes = () => api.get('/exams')
export const getExamModules = (examId) => api.get(`/exams/${examId}/modules`)

// 题目
export const getQuestions = (params) => api.get('/questions', { params })
export const getQuestion = (id) => api.get(`/questions/${id}`)
export const createQuestion = (data) => api.post('/questions', data)
export const updateQuestion = (id, data) => api.put(`/questions/${id}`, data)
export const deleteQuestion = (id) => api.delete(`/questions/${id}`)

// 刷题
export const startQuiz = (data) => api.post('/quiz/start', data)
export const submitAnswer = (data) => api.post('/quiz/answer', data)
export const submitBatchAnswers = (data) => api.post('/quiz/submit-batch', data)

// 统计
export const getStats = () => api.get('/stats')
export const getModuleStats = (id) => api.get(`/stats/module/${id}`)
// 数据管理
export const deleteRecords = () => api.delete('/records')

// 批量导入
export const importQuestions = (data) => api.post('/questions/import', data)

// 考试场次
export const getSessions = (params) => api.get('/sessions', { params })
export const getSession = (id) => api.get(`/sessions/${id}`)
export const getSessionAnswers = (id) => api.get(`/sessions/${id}/answers`)

// 考试/模块管理
export const createExamType = (data) => api.post('/exams', data)
export const updateExamType = (id, data) => api.put(`/exams/${id}`, data)
export const deleteExamType = (id) => api.delete(`/exams/${id}`)
export const createModule = (data) => api.post('/modules', data)
export const updateModule = (id, data) => api.put(`/modules/${id}`, data)
export const deleteModule = (id) => api.delete(`/modules/${id}`)

export default api
export { showToast }
