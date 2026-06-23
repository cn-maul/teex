<template>
  <div class="settings-view">
    <h1>设置</h1>

    <!-- 账户信息 -->
    <div class="settings-section">
      <div class="section-header">
        <h2>账户信息</h2>
        <p class="section-desc">管理你的个人资料</p>
      </div>

      <div class="profile-card">
        <div class="profile-avatar">👤</div>
        <div class="profile-info">
          <div class="profile-name">
            {{ authStore.user?.nickname || authStore.user?.username }}
            <span class="role-badge" :class="authStore.user?.role === 'admin' ? 'role-admin' : 'role-user'">
              {{ authStore.user?.role === 'admin' ? '管理员' : '普通用户' }}
            </span>
          </div>
          <div class="profile-meta">
            <span>用户名：{{ authStore.user?.username }}</span>
            <span v-if="authStore.user?.created_at">注册时间：{{ formatDate(authStore.user.created_at) }}</span>
          </div>
        </div>
      </div>

      <!-- 修改昵称 -->
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">昵称</label>
          <p class="setting-desc">当前昵称：{{ authStore.user?.nickname || authStore.user?.username }}</p>
        </div>
        <div class="setting-control">
          <div v-if="!editingNickname" class="inline-edit">
            <button class="btn btn-ghost" @click="startEditNickname">修改</button>
          </div>
          <div v-else class="inline-edit-form">
            <input v-model="newNickname" type="text" placeholder="输入新昵称" maxlength="50" class="edit-input" />
            <button class="btn btn-primary btn-sm" @click="saveNickname" :disabled="!newNickname.trim()">保存</button>
            <button class="btn btn-ghost btn-sm" @click="editingNickname = false">取消</button>
          </div>
        </div>
      </div>

      <!-- 修改密码 -->
      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">修改密码</label>
          <p class="setting-desc">修改登录密码</p>
        </div>
        <div class="setting-control">
          <button v-if="!showPasswordForm" class="btn btn-ghost" @click="showPasswordForm = true">修改密码</button>
        </div>
      </div>

      <div v-if="showPasswordForm" class="password-form">
        <div class="form-row">
          <label>原密码</label>
          <input v-model="passwordForm.oldPassword" type="password" placeholder="输入原密码" />
        </div>
        <div class="form-row">
          <label>新密码</label>
          <input v-model="passwordForm.newPassword" type="password" placeholder="至少 6 位" />
        </div>
        <div class="form-row">
          <label>确认密码</label>
          <input v-model="passwordForm.confirmPassword" type="password" placeholder="再次输入新密码" />
        </div>
        <div class="form-actions">
          <button class="btn btn-primary btn-sm" @click="savePassword" :disabled="!canSubmitPassword">确认修改</button>
          <button class="btn btn-ghost btn-sm" @click="cancelPasswordChange">取消</button>
        </div>
      </div>
    </div>

    <!-- 个人概览 -->
    <div class="settings-section">
      <div class="section-header">
        <h2>个人概览</h2>
        <p class="section-desc">你的刷题数据</p>
      </div>

      <div v-if="statsLoading" class="stats-loading">加载中...</div>
      <div v-else class="stats-grid">
        <div class="stat-card">
          <span class="stat-value">{{ stats.total_questions }}</span>
          <span class="stat-label">总题数</span>
        </div>
        <div class="stat-card">
          <span class="stat-value">{{ stats.total_answered }}</span>
          <span class="stat-label">已做</span>
        </div>
        <div class="stat-card">
          <span class="stat-value accent">{{ stats.accuracy }}%</span>
          <span class="stat-label">正确率</span>
        </div>
        <div class="stat-card">
          <span class="stat-value">{{ stats.unanswered }}</span>
          <span class="stat-label">未做</span>
        </div>
      </div>
    </div>

    <!-- 刷题偏好（保留原有内容） -->
    <div class="settings-section">
      <div class="section-header">
        <h2>刷题偏好</h2>
        <p class="section-desc">调整刷题数量和模式</p>
      </div>

      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">每次刷题数量</label>
          <p class="setting-desc">每次开始刷题时加载的题目数</p>
        </div>
        <div class="setting-control range-control">
          <input
            type="range"
            min="5"
            max="50"
            step="5"
            :value="examStore.settings.quizCount"
            @input="handleQuizCountChange"
            class="range-input"
          />
          <span class="range-badge">{{ examStore.settings.quizCount }} 题</span>
        </div>
      </div>

      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">刷题模式</label>
          <p class="setting-desc">解析模式：逐题作答，即时反馈；考试模式：全部答完再统一评分</p>
        </div>
        <div class="setting-control mode-control">
          <button
            class="mode-btn"
            :class="{ active: examStore.settings.quizMode === 'analysis' }"
            @click="examStore.updateQuizMode('analysis')"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
            解析模式
          </button>
          <button
            class="mode-btn"
            :class="{ active: examStore.settings.quizMode === 'exam' }"
            @click="examStore.updateQuizMode('exam')"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>
            考试模式
          </button>
        </div>
      </div>
    </div>

    <!-- 数据管理 -->
    <div class="settings-section">
      <div class="section-header">
        <h2>数据管理</h2>
        <p class="section-desc">管理你的刷题数据</p>
      </div>

      <div class="setting-item">
        <div class="setting-info">
          <label class="setting-label">清空答题记录</label>
          <p class="setting-desc">清除所有答题记录，此操作不可恢复</p>
        </div>
        <div class="setting-control">
          <button class="btn btn-danger" @click="confirmClearData">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
            清空记录
          </button>
        </div>
      </div>
    </div>

    <!-- 操作提示 -->
    <Transition name="fade">
      <div v-if="toast" class="toast" :class="toast.type === 'success' ? 'toast-success' : 'toast-error'">
        {{ toast.message }}
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useExamStore } from '../stores/exam'
import { useAuthStore } from '../stores/auth.js'
import { deleteRecords, getStats, updateProfile, changePassword } from '../api'

const examStore = useExamStore()
const authStore = useAuthStore()

// 统计数据
const stats = ref({ total_questions: 0, total_answered: 0, accuracy: 0, unanswered: 0 })
const statsLoading = ref(true)

// 昵称编辑
const editingNickname = ref(false)
const newNickname = ref('')

// 密码修改
const showPasswordForm = ref(false)
const passwordForm = reactive({ oldPassword: '', newPassword: '', confirmPassword: '' })

// Toast
const toast = ref(null)

onMounted(async () => {
  try {
    const res = await getStats()
    stats.value = res.data.data || stats.value
  } catch {
    // ignore
  } finally {
    statsLoading.value = false
  }
})

function showToast(message, type = 'success') {
  toast.value = { message, type }
  setTimeout(() => { toast.value = null }, 3000)
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

// 刷题偏好
function handleQuizCountChange(event) {
  examStore.updateQuizCount(parseInt(event.target.value))
}

// 昵称
function startEditNickname() {
  newNickname.value = authStore.user?.nickname || ''
  editingNickname.value = true
}

async function saveNickname() {
  const nick = newNickname.value.trim()
  if (!nick) return
  try {
    await updateProfile({ nickname: nick })
    // 更新本地状态
    authStore.user.nickname = nick
    localStorage.setItem('user', JSON.stringify(authStore.user))
    editingNickname.value = false
    showToast('昵称修改成功')
  } catch (e) {
    showToast(e.response?.data?.error || '修改失败', 'error')
  }
}

// 密码
const canSubmitPassword = computed(() => {
  return passwordForm.oldPassword && passwordForm.newPassword.length >= 6 && passwordForm.newPassword === passwordForm.confirmPassword
})

async function savePassword() {
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    showToast('两次输入的密码不一致', 'error')
    return
  }
  try {
    await changePassword({ old_password: passwordForm.oldPassword, new_password: passwordForm.newPassword })
    showToast('密码修改成功')
    cancelPasswordChange()
  } catch (e) {
    showToast(e.response?.data?.error || '修改失败', 'error')
  }
}

function cancelPasswordChange() {
  showPasswordForm.value = false
  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
}

// 清空数据
async function confirmClearData() {
  if (confirm('确定要清空所有答题记录吗？此操作不可恢复！')) {
    try {
      await deleteRecords()
      stats.value = { total_questions: stats.value.total_questions, total_answered: 0, accuracy: 0, unanswered: stats.value.total_questions }
      showToast('答题记录已清空')
    } catch (e) {
      showToast(e.response?.data?.error || '清空失败', 'error')
    }
  }
}
</script>

<style scoped>
.settings-view {
  max-width: 640px;
  margin: 0 auto;
}

h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 1.75rem;
}

.settings-section {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  padding: 1.5rem;
  margin-bottom: 1rem;
}

.section-header {
  margin-bottom: 1.25rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border-light);
}

.section-header h2 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 0.15rem;
}

.section-desc {
  font-size: 0.85rem;
  color: var(--text-muted);
}

/* Profile card */
.profile-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding-bottom: 1rem;
  margin-bottom: 0.5rem;
  border-bottom: 1px solid var(--border-light);
}

.profile-avatar {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  background: var(--bg-hover);
  border-radius: 50%;
  flex-shrink: 0;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

.profile-name {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.role-badge {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 0.1rem 0.5rem;
  border-radius: 10px;
}

.role-admin {
  background: #fef3c7;
  color: #92400e;
}

.role-user {
  background: var(--primary-bg);
  color: var(--primary);
}

.profile-meta {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-top: 0.2rem;
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

/* Setting items */
.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
  padding: 0.85rem 0;
}

.setting-item + .setting-item {
  border-top: 1px solid var(--border-light);
}

.setting-info {
  flex: 1;
  min-width: 0;
}

.setting-label {
  display: block;
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--text);
  margin-bottom: 0.1rem;
}

.setting-desc {
  font-size: 0.8rem;
  color: var(--text-muted);
}

.setting-control {
  flex-shrink: 0;
}

/* Inline edit */
.inline-edit-form {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.edit-input {
  padding: 0.4rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 0.85rem;
  width: 180px;
  outline: none;
  transition: border-color 0.2s;
}

.edit-input:focus {
  border-color: var(--primary);
}

/* Password form */
.password-form {
  padding: 1rem 0;
  border-top: 1px solid var(--border-light);
}

.form-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.form-row label {
  width: 70px;
  font-size: 0.85rem;
  color: var(--text-secondary);
  text-align: right;
  flex-shrink: 0;
}

.form-row input {
  flex: 1;
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 0.85rem;
  outline: none;
  transition: border-color 0.2s;
  max-width: 280px;
}

.form-row input:focus {
  border-color: var(--primary);
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  margin-left: 78px;
}

/* Stats grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.75rem;
}

.stat-card {
  background: var(--bg);
  border-radius: var(--radius-lg);
  padding: 1rem 0.75rem;
  text-align: center;
}

.stat-value {
  display: block;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
  line-height: 1.2;
}

.stat-value.accent { color: var(--primary); }

.stat-label {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: 0.25rem;
}

.stats-loading {
  text-align: center;
  color: var(--text-muted);
  font-size: 0.85rem;
  padding: 1.5rem;
}

/* Range input */
.range-control {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 200px;
}

.range-input {
  flex: 1;
  height: 6px;
  -webkit-appearance: none;
  appearance: none;
  background: var(--border);
  border-radius: 3px;
  outline: none;
}

.range-input::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 18px;
  height: 18px;
  background: var(--primary);
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 1px 3px rgba(99, 102, 241, 0.3);
  transition: transform 0.15s ease;
}

.range-input::-webkit-slider-thumb:hover {
  transform: scale(1.15);
}

.range-badge {
  background: var(--primary-bg);
  color: var(--primary);
  padding: 0.2rem 0.5rem;
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  font-weight: 600;
  white-space: nowrap;
  min-width: 42px;
  text-align: center;
}

/* Buttons */
.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: var(--radius);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn svg {
  width: 15px;
  height: 15px;
}

.btn-sm {
  padding: 0.35rem 0.75rem;
  font-size: 0.8rem;
}

.btn-primary {
  background: var(--primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--primary-dark);
}

.btn-ghost {
  background: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border);
}

.btn-ghost:hover:not(:disabled) {
  background: var(--bg-hover);
  border-color: var(--text-muted);
}

.btn-danger {
  background: var(--error-bg);
  color: var(--error);
  border: 1px solid #fecaca;
}

.btn-danger:hover {
  background: #fecaca;
}

/* Mode toggle */
.mode-control {
  display: flex;
  gap: 0.5rem;
}

.mode-btn {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 1rem;
  border: 1.5px solid var(--border);
  border-radius: var(--radius-lg);
  background: transparent;
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
}

.mode-btn svg {
  width: 16px;
  height: 16px;
}

.mode-btn:hover {
  border-color: var(--primary-light);
  background: var(--primary-bg);
}

.mode-btn.active {
  border-color: var(--primary);
  background: var(--primary-bg);
  color: var(--primary);
}

/* Toast */
.toast {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  padding: 0.65rem 1.25rem;
  border-radius: var(--radius-lg);
  font-size: 0.85rem;
  font-weight: 500;
  z-index: 10000;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.toast-success {
  background: var(--success);
  color: white;
}

.toast-error {
  background: var(--error);
  color: white;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .profile-meta {
    flex-direction: column;
    gap: 0.2rem;
  }

  .form-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.3rem;
  }

  .form-row label {
    width: auto;
    text-align: left;
  }

  .form-row input {
    max-width: 100%;
    width: 100%;
  }

  .form-actions {
    margin-left: 0;
  }
}
</style>
