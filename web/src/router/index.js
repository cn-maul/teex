import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/',          name: 'Home',      component: () => import('../views/HomeView.vue') },
  { path: '/exam/:id',  name: 'Exam',      component: () => import('../views/ExamView.vue') },
  { path: '/quiz/:moduleId', name: 'Quiz', component: () => import('../views/QuizView.vue') },
  { path: '/history',   name: 'History',   component: () => import('../views/HistoryView.vue') },
  { path: '/settings',  name: 'Settings',  component: () => import('../views/SettingsView.vue') },
  { path: '/admin/exams', name: 'ExamManage', component: () => import('../views/ExamManageView.vue') },
  { path: '/admin/questions', name: 'QuestionManage', component: () => import('../views/QuestionManageView.vue') },
]

const router = createRouter({ history: createWebHistory(), routes })
export default router
