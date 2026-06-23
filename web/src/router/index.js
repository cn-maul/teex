import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/login', name: 'Login', component: () => import('../views/LoginView.vue'), meta: { public: true } },
  { path: '/',          name: 'Home',      component: () => import('../views/HomeView.vue') },
  { path: '/exam/:id',  name: 'Exam',      component: () => import('../views/ExamView.vue') },
  { path: '/quiz/:moduleId', name: 'Quiz', component: () => import('../views/QuizView.vue') },
  { path: '/history',   name: 'History',   component: () => import('../views/HistoryView.vue') },
  { path: '/settings',  name: 'Settings',  component: () => import('../views/SettingsView.vue') },
  { path: '/admin/exams', name: 'ExamManage', component: () => import('../views/ExamManageView.vue'), meta: { admin: true } },
  { path: '/admin/questions', name: 'QuestionManage', component: () => import('../views/QuestionManageView.vue'), meta: { admin: true } },
  { path: '/admin/users', name: 'UserManage', component: () => import('../views/UserManageView.vue'), meta: { admin: true } },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({ history: createWebHistory(), routes })

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (!to.meta.public && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/')
  } else if (to.meta.admin) {
    // 管理页面需要管理员权限
    try {
      const user = JSON.parse(localStorage.getItem('user') || '{}')
      if (user.role !== 'admin') {
        next('/')
        return
      }
    } catch {
      next('/')
      return
    }
    next()
  } else {
    next()
  }
})

export default router
