<template>
  <aside class="sidebar">
    <nav class="sidebar-nav">
      <router-link to="/" class="sidebar-item">
        <span class="sidebar-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg>
        </span>
        <span class="sidebar-label">首页</span>
      </router-link>

      <router-link to="/history" class="sidebar-item">
        <span class="sidebar-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
        </span>
        <span class="sidebar-label">历史记录</span>
      </router-link>

      <div class="sidebar-divider"></div>

      <router-link to="/settings" class="sidebar-item">
        <span class="sidebar-icon">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
        </span>
        <span class="sidebar-label">设置</span>
      </router-link>

      <template v-if="authStore.user?.role === 'admin'">
        <div class="sidebar-divider"></div>

        <router-link to="/admin/exams" class="sidebar-item">
          <span class="sidebar-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>
          </span>
          <span class="sidebar-label">考试管理</span>
        </router-link>

        <router-link to="/admin/questions" class="sidebar-item">
          <span class="sidebar-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
          </span>
          <span class="sidebar-label">题目管理</span>
        </router-link>

        <router-link to="/admin/users" class="sidebar-item">
          <span class="sidebar-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path></svg>
          </span>
          <span class="sidebar-label">人员管理</span>
        </router-link>
      </template>
    </nav>

    <div class="sidebar-user" v-if="authStore.isLoggedIn">
      <div class="user-info">
        <span class="user-avatar">👤</span>
        <span class="user-name">{{ authStore.user?.nickname || authStore.user?.username }}</span>
      </div>
      <button class="btn-logout" @click="handleLogout" title="退出登录">⏻</button>
    </div>
  </aside>
</template>

<script setup>
import { useAuthStore } from '../stores/auth.js'
import { useRouter } from 'vue-router'
import { useConfirm } from '../utils/confirm'

const { showConfirm } = useConfirm()
const authStore = useAuthStore()
const router = useRouter()

async function handleLogout() {
  if (!await showConfirm({ message: '确定要退出登录吗？' })) return
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.sidebar {
  position: fixed;
  left: 0;
  top: 56px;
  bottom: 0;
  width: 220px;
  background: var(--bg-card);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  z-index: 100;
  overflow: hidden;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  padding: 0.25rem 0.5rem;
  flex: 1;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.75rem;
  color: var(--text-secondary);
  text-decoration: none;
  border-radius: var(--radius);
  transition: var(--transition);
  white-space: nowrap;
  position: relative;
  margin-bottom: 2px;
}

.sidebar-item:hover {
  background: var(--bg-hover);
  color: var(--text);
}

.sidebar-item.router-link-exact-active {
  background: var(--primary-bg);
  color: var(--primary);
  font-weight: 500;
}

.sidebar-item.router-link-exact-active::before {
  content: '';
  position: absolute;
  left: -0.5rem;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  background: var(--primary);
  border-radius: 0 3px 3px 0;
}

.sidebar-icon {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
}

.sidebar-icon svg {
  width: 100%;
  height: 100%;
}

.sidebar-label {
  font-size: 0.875rem;
}

.sidebar-divider {
  height: 1px;
  background: var(--border);
  margin: 0.5rem 0.75rem;
}

.sidebar-user {
  margin-top: auto;
  padding: 12px;
  border-top: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-secondary);
  font-size: 13px;
  overflow: hidden;
}
.user-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.btn-logout {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 16px;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
}
.btn-logout:hover {
  color: #ef4444;
  background: rgba(239,68,68,0.1);
}
</style>
