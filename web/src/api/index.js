import axios from 'axios'
import { showToast } from '../utils/toast.js'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

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
      showToast('请求过于频繁，请稍后再试', 'error')
      return Promise.reject(error)
    }
    if (error.response?.status === 403) {
      showToast(error.response?.data?.error || '没有权限执行此操作', 'error')
      return Promise.reject(error)
    }
    if (error.response?.status === 500) {
      showToast('服务器内部错误，请稍后重试', 'error')
      return Promise.reject(error)
    }
    const message = error.response?.data?.error
      || (error.code === 'ECONNABORTED' ? '请求超时，请稍后重试' : '网络错误，请检查连接')
    showToast(message, 'error')
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
export const getDashboardStats = () => api.get('/stats/dashboard')
export const getAdminDashboardStats = () => api.get('/admin/dashboard')

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
export const getUsers = () => api.get('/users')
export const adminCreateUser = (data) => api.post('/users', data)
export const adminUpdateUser = (id, data) => api.put(`/users/${id}`, data)
export const adminDeleteUser = (id) => api.delete(`/users/${id}`)

// 系统设置
export const getRegistrationStatus = () => api.get('/settings/registration')
export const setRegistrationStatus = (data) => api.put('/settings/registration', data)

export default api
