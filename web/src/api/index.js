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

// 请求拦截器：自动注入 token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 正在跳转登录页标记，防止 401 循环
let isRedirectingToLogin = false

// 响应拦截器：统一错误处理
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // 清除认证状态
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      // 防止循环重定向：只跳转一次
      if (!isRedirectingToLogin) {
        isRedirectingToLogin = true
        // 用 location.replace 避免在 history 中留下可回退的记录
        window.location.replace('/login')
        // 重置标记，防止导航被取消后标记永久残留
        setTimeout(() => { isRedirectingToLogin = false }, 2000)
      }
      return Promise.reject(error)
    }
    if (error.response?.status === 429) {
      showToast('请求过于频繁，请稍后再试')
      return Promise.reject(error)
    }
    if (error.response?.status === 500) {
      showToast('服务器内部错误，请稍后重试')
      return Promise.reject(error)
    }
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
export const batchDeleteQuestions = (ids) => api.delete('/questions/batch', { data: { ids } })

// 刷题
export const startQuiz = (data) => api.post('/quiz/start', data)
export const submitAnswer = (data) => api.post('/quiz/answer', data)
export const submitBatchAnswers = (data) => api.post('/quiz/submit-batch', data)

// 统计
export const getStats = () => api.get('/stats')
export const getModuleStats = (id) => api.get(`/stats/module/${id}`)
export const getExamStats = (examId) => api.get(`/exams/${examId}/stats`)
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

// 认证
export const login = (data) => api.post('/auth/login', data)
export const register = (data) => api.post('/auth/register', data)
export const getProfile = () => api.get('/profile')
export const updateProfile = (data) => api.put('/profile', data)
export const changePassword = (data) => api.put('/profile/password', data)

// 管理员
export const getUsers = () => api.get('/admin/users')
export const adminCreateUser = (data) => api.post('/admin/users', data)
export const adminUpdateUser = (id, data) => api.put(`/admin/users/${id}`, data)
export const adminDeleteUser = (id) => api.delete(`/admin/users/${id}`)

// 系统设置
export const getRegistrationStatus = () => api.get('/settings/registration')
export const setRegistrationStatus = (data) => api.put('/admin/settings/registration', data)

export default api
export { showToast }
