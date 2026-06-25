import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  { path: '/login', name: 'Login', component: () => import('../views/LoginView.vue'), meta: { public: true } },
  { path: '/',          name: 'Home',      component: () => import('../views/HomeView.vue') },
  { path: '/exam/:id',  name: 'Exam',      component: () => import('../views/ExamView.vue') },
  { path: '/quiz/:moduleId', name: 'Quiz', component: () => import('../views/QuizView.vue') },
  { path: '/history',   name: 'History',   component: () => import('../views/HistoryView.vue') },
  { path: '/stats',     name: 'Stats',     component: () => import('../views/StatsView.vue') },
  { path: '/settings',  name: 'Settings',  component: () => import('../views/SettingsView.vue') },
  { path: '/admin/dashboard', name: 'AdminDashboard', component: () => import('../views/AdminDashboardView.vue'), meta: { admin: true } },
  { path: '/admin/exams', name: 'ExamManage', component: () => import('../views/ExamManageView.vue'), meta: { admin: true } },
  { path: '/admin/questions', name: 'QuestionManage', component: () => import('../views/QuestionManageView.vue'), meta: { admin: true } },
  { path: '/admin/users', name: 'UserManage', component: () => import('../views/UserManageView.vue'), meta: { admin: true } },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({ history: createWebHistory(), routes })

// 路由守卫 — 使用 authStore 统一状态源，避免重复解析 localStorage
router.beforeEach((to, from, next) => {
  const auth = useAuthStore()

  if (to.meta.public) {
    // 已登录用户访问登录页 → 重定向首页
    next(auth.isLoggedIn ? '/' : undefined)
  } else if (!auth.isLoggedIn) {
    next('/login')
  } else if (to.meta.admin && !auth.isAdmin) {
    next('/')
  } else if (to.path === '/' && auth.isAdmin) {
    next('/admin/dashboard')
  } else {
    next()
  }
})

export default router
