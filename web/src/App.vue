<template>
  <div id="app">
    <!-- 未登录：只显示 router-view（登录页自己管理布局） -->
    <template v-if="!authStore.isLoggedIn">
      <router-view />
    </template>

    <!-- 已登录：完整布局 -->
    <template v-else>
      <!-- 顶部导航栏 -->
      <nav class="navbar">
        <div class="nav-left">
          <div class="nav-brand" @click="$router.push('/')">
            <span class="brand-icon"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20"><path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z"></path><path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z"></path></svg></span>
            <span class="brand-text">公考刷题</span>
          </div>

          <!-- 考试类型自定义下拉 -->
          <div class="exam-selector" v-if="examStore.state.examList.length > 0" ref="dropdownRef">
            <button class="exam-trigger" @click="dropdownOpen = !dropdownOpen">
              <span class="exam-trigger-text">{{ examStore.state.currentExamName || '选择考试' }}</span>
              <span class="exam-trigger-arrow" :class="{ open: dropdownOpen }">▾</span>
            </button>
            <Transition name="dropdown">
              <div class="exam-dropdown" v-if="dropdownOpen">
                <button
                  v-for="exam in examStore.state.examList"
                  :key="exam.id"
                  class="exam-option"
                  :class="{ active: exam.id === examStore.state.currentExamId }"
                  @click="selectExam(exam)"
                >
                  <span class="exam-option-dot" v-if="exam.id === examStore.state.currentExamId"></span>
                  {{ exam.name }}
                </button>
              </div>
            </Transition>
          </div>
        </div>

        <div class="nav-right">
          <span class="exam-badge" v-if="examStore.state.currentExamName">
            {{ examStore.state.currentExamName }}
          </span>
        </div>
      </nav>

      <!-- 主体布局 -->
      <div class="layout">
        <Sidebar />
        <main class="main-content">
          <router-view />
        </main>
        <StatsPanel />
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useExamStore } from './stores/exam'
import { useAuthStore } from './stores/auth.js'
import Sidebar from './components/Sidebar.vue'
import StatsPanel from './components/StatsPanel.vue'

const router = useRouter()
const examStore = useExamStore()
const authStore = useAuthStore()

const dropdownOpen = ref(false)
const dropdownRef = ref(null)

// 监听登录状态变化：登录后加载考试列表，加载完成再允许路由跳转
watch(() => authStore.isLoggedIn, async (loggedIn) => {
  if (loggedIn) {
    await examStore.loadExams()
  }
}, { immediate: true })

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

function handleClickOutside(e) {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target)) {
    dropdownOpen.value = false
  }
}

function selectExam(exam) {
  examStore.setExam(exam)
  dropdownOpen.value = false
  const currentRoute = router.currentRoute.value
  if (currentRoute.name === 'Exam' || currentRoute.name === 'Quiz') {
    router.push('/')
  }
}
</script>

<style scoped>
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 1.25rem;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
  z-index: 200;
}

.nav-left {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  user-select: none;
}

.brand-icon {
  font-size: 1.4rem;
}

.brand-text {
  font-size: 1.15rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.02em;
}

.nav-right {
  display: flex;
  align-items: center;
}

/* Custom dropdown */
.exam-selector {
  position: relative;
}

.exam-trigger {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.75rem;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  color: var(--text);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
}

.exam-trigger:hover {
  border-color: var(--primary-light);
}

.exam-trigger-arrow {
  font-size: 0.7rem;
  transition: transform 0.2s ease;
  color: var(--text-muted);
}

.exam-trigger-arrow.open {
  transform: rotate(180deg);
}

.exam-dropdown {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  min-width: 200px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: 0.35rem;
  z-index: 300;
}

.exam-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.55rem 0.75rem;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.875rem;
  border-radius: var(--radius-sm);
  cursor: pointer;
  text-align: left;
  transition: var(--transition);
}

.exam-option:hover {
  background: var(--bg-hover);
  color: var(--text);
}

.exam-option.active {
  background: var(--primary-bg);
  color: var(--primary);
  font-weight: 500;
}

.exam-option-dot {
  width: 6px;
  height: 6px;
  background: var(--primary);
  border-radius: 50%;
  flex-shrink: 0;
}

/* Dropdown transition */
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

.exam-badge {
  background: var(--primary-bg);
  color: var(--primary);
  padding: 0.25rem 0.65rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
  white-space: nowrap;
}

.layout {
  display: flex;
  min-height: 100vh;
  padding-top: 56px;
}

.main-content {
  flex: 1;
  margin-left: 220px;
  margin-right: 260px;
  padding: 1.75rem 2rem;
}
</style>
