<template>
  <div class="settings-view">
    <h1>设置</h1>

    <!-- 账户信息 -->
    <ProfileSection
      :user="authStore.user"
      :editing-nickname="editingNickname"
      :new-nickname="newNickname"
      @start-edit-nickname="startEditNickname"
      @update:new-nickname="newNickname = $event"
      @save-nickname="saveNickname"
      @cancel-edit-nickname="editingNickname = false"
    />

    <!-- 修改密码 -->
    <PasswordSection :saved="passwordSaved" @save-password="savePassword" />

    <!-- 系统设置（仅管理员） -->
    <AdminSection
      v-if="authStore.isAdmin"
      :registration-enabled="registrationEnabled"
      :registration-loading="registrationLoading"
      :batch-limit="batchLimit"
      :batch-limit-saving="batchLimitSaving"
      :rate-limit="rateLimit"
      :rate-limit-saving="rateLimitSaving"
      @toggle-registration="toggleRegistration"
      @save-batch-limit="saveBatchLimit"
      @save-rate-limit="saveRateLimit"
    />

    <!-- 个人概览（仅普通用户） -->
    <div v-if="!authStore.isAdmin" class="settings-section">
      <div class="section-header">
        <h2>个人概览</h2>
        <p class="section-desc">你的刷题数据</p>
      </div>

      <div v-if="statsLoading" class="stats-loading">加载中...</div>
      <div v-else class="stats-grid">
        <div class="stat-card">
          <div class="stat-ring-mini">
            <Doughnut :data="completionData" :options="miniDoughnutOptions" />
          </div>
          <span class="stat-value">{{ stats.total_answered }}</span>
          <span class="stat-label">已做 / {{ stats.total_questions }} 题</span>
        </div>
        <div class="stat-card accent-card">
          <div class="stat-ring-mini">
            <Doughnut :data="accuracyData" :options="miniDoughnutOptions" />
          </div>
          <span class="stat-value accent">{{ stats.accuracy }}%</span>
          <span class="stat-label">正确率</span>
        </div>
        <div class="stat-card">
          <div class="stat-icon-lg">📚</div>
          <span class="stat-value">{{ stats.total_questions }}</span>
          <span class="stat-label">总题数</span>
        </div>
        <div class="stat-card">
          <div class="stat-icon-lg">⏳</div>
          <span class="stat-value">{{ stats.unanswered }}</span>
          <span class="stat-label">待完成</span>
        </div>
      </div>
    </div>

    <!-- 刷题偏好 -->
    <QuizPreferenceSection
      :quiz-count="examStore.settings.quizCount"
      :quiz-mode="examStore.settings.quizMode"
      @update:quiz-count="examStore.updateQuizCount($event)"
      @update:quiz-mode="examStore.updateQuizMode($event)"
    />

    <!-- 数据管理 -->
    <DataSection v-if="!authStore.isAdmin" @clear-data="confirmClearData" />
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useExamStore } from '../stores/exam'
import { useAuthStore } from '../stores/auth.js'
import { deleteRecords, getStats, updateProfile, changePassword, getRegistrationStatus, setRegistrationStatus, getBatchLimit, setBatchLimit, getRateLimit, setRateLimit } from '../api'
import { showToast } from '../utils/toast'
import { useConfirm } from '../utils/confirm'
import ProfileSection from '../components/settings/ProfileSection.vue'
import PasswordSection from '../components/settings/PasswordSection.vue'
import AdminSection from '../components/settings/AdminSection.vue'
import QuizPreferenceSection from '../components/settings/QuizPreferenceSection.vue'
import DataSection from '../components/settings/DataSection.vue'
import { Doughnut } from 'vue-chartjs'

const { showConfirm } = useConfirm()

const examStore = useExamStore()
const authStore = useAuthStore()

// 统计数据
const stats = ref({ total_questions: 0, total_answered: 0, accuracy: 0, unanswered: 0 })
const statsLoading = ref(true)

// 昵称编辑
const editingNickname = ref(false)
const newNickname = ref('')

// 注册开关
const registrationEnabled = ref(false)
const registrationLoading = ref(true)

// 批量操作上限
const batchLimit = ref(500)
const batchLimitSaving = ref(false)

// 请求频率限制
const rateLimit = ref(120)
const rateLimitSaving = ref(false)

// 密码修改状态
const passwordSaved = ref(false)

const completionData = computed(() => ({
  labels: ['已完成', '未完成'],
  datasets: [{
    data: [stats.value.total_answered, stats.value.unanswered],
    backgroundColor: ['#6366f1', '#e2e8f0'],
    borderWidth: 0,
    hoverOffset: 2
  }]
}))

const accuracyData = computed(() => ({
  labels: ['正确', '错误'],
  datasets: [{
    data: [stats.value.accuracy, 100 - stats.value.accuracy],
    backgroundColor: ['#10b981', '#e2e8f0'],
    borderWidth: 0,
    hoverOffset: 2
  }]
}))

const miniDoughnutOptions = {
  responsive: true,
  maintainAspectRatio: true,
  cutout: '72%',
  plugins: {
    legend: { display: false },
    tooltip: { enabled: false }
  }
}

onMounted(async () => {
  try {
    const res = await getStats()
    stats.value = res.data.data || stats.value
  } catch {
    // ignore
  } finally {
    statsLoading.value = false
  }

  if (authStore.isAdmin) {
    try {
      const regRes = await getRegistrationStatus()
      registrationEnabled.value = regRes.data.data?.enabled ?? false
    } catch { /* ignore */ } finally {
      registrationLoading.value = false
    }

    try {
      const limitRes = await getBatchLimit()
      batchLimit.value = limitRes.data.data?.limit ?? 500
    } catch { /* ignore */ }

    try {
      const rlRes = await getRateLimit()
      rateLimit.value = rlRes.data.data?.limit ?? 120
    } catch { /* ignore */ }
  }
})

async function toggleRegistration() {
  try {
    const newVal = !registrationEnabled.value
    await setRegistrationStatus({ enabled: newVal })
    registrationEnabled.value = newVal
    showToast(newVal ? '注册已开放' : '注册已关闭')
  } catch (e) {
    showToast(e.response?.data?.error || '操作失败', 'error')
  }
}

async function saveBatchLimit(val) {
  if (val < 1 || val > 10000) {
    showToast('批量操作上限必须在 1 ~ 10000 之间', 'error')
    return
  }
  batchLimitSaving.value = true
  try {
    await setBatchLimit({ limit: val })
    batchLimit.value = val
    showToast('批量操作上限已更新')
  } catch (e) {
    showToast(e.response?.data?.error || '操作失败', 'error')
  } finally {
    batchLimitSaving.value = false
  }
}

async function saveRateLimit(val) {
  if (val < 10 || val > 10000) {
    showToast('请求频率限制必须在 10 ~ 10000 之间', 'error')
    return
  }
  rateLimitSaving.value = true
  try {
    await setRateLimit({ limit: val })
    rateLimit.value = val
    showToast('请求频率限制已更新')
  } catch (e) {
    showToast(e.response?.data?.error || '操作失败', 'error')
  } finally {
    rateLimitSaving.value = false
  }
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
    authStore.user.nickname = nick
    localStorage.setItem('user', JSON.stringify(authStore.user))
    editingNickname.value = false
    showToast('昵称修改成功')
  } catch (e) {
    showToast(e.response?.data?.error || '修改失败', 'error')
  }
}

// 密码
async function savePassword(data) {
  try {
    await changePassword(data)
    showToast('密码修改成功')
    passwordSaved.value = true
    // Reset after a tick so the watcher in PasswordSection can react
    setTimeout(() => { passwordSaved.value = false }, 0)
  } catch (e) {
    showToast(e.response?.data?.error || '修改失败', 'error')
  }
}

// 清空数据
async function confirmClearData() {
  if (!await showConfirm({ message: '确定要清空所有答题记录吗？此操作不可恢复！', dangerMode: true })) return
  try {
    await deleteRecords()
    stats.value = { total_questions: stats.value.total_questions, total_answered: 0, accuracy: 0, unanswered: stats.value.total_questions }
    showToast('答题记录已清空')
  } catch (e) {
    showToast(e.response?.data?.error || '清空失败', 'error')
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

.stat-ring-mini {
  width: 64px;
  height: 64px;
  margin: 0 auto 0.5rem;
}

.stat-icon-lg {
  font-size: 1.5rem;
  margin-bottom: 0.35rem;
}

.accent-card {
  border: 1px solid var(--primary-light);
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

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
