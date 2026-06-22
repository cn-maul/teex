<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>📝 Teex</h1>
        <p>公考刷题工具</p>
      </div>

      <div class="tab-switch">
        <button :class="{ active: mode === 'login' }" @click="mode = 'login'">登录</button>
        <button :class="{ active: mode === 'register' }" @click="mode = 'register'">注册</button>
      </div>

      <form @submit.prevent="handleSubmit" class="login-form">
        <div class="form-group">
          <input v-model="form.username" type="text" placeholder="用户名" required autocomplete="username" />
        </div>
        <div class="form-group">
          <input v-model="form.password" type="password" placeholder="密码" required autocomplete="current-password" />
        </div>
        <div class="form-group" v-if="mode === 'register'">
          <input v-model="form.nickname" type="text" placeholder="昵称（可选）" />
        </div>

        <button type="submit" class="btn-submit" :disabled="loading">
          {{ loading ? '处理中...' : (mode === 'login' ? '登录' : '注册') }}
        </button>

        <p class="error-msg" v-if="error">{{ error }}</p>
      </form>

      <p class="hint" v-if="mode === 'login'">
        默认管理员：admin / admin123
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { login, register } from '../api/index.js'
import { useAuthStore } from '../stores/auth.js'
import { useExamStore } from '../stores/exam.js'

const router = useRouter()
const authStore = useAuthStore()
const examStore = useExamStore()
const mode = ref('login')
const loading = ref(false)
const error = ref('')

const form = reactive({
  username: '',
  password: '',
  nickname: ''
})

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    const fn = mode.value === 'login' ? login : register
    const res = await fn({
      username: form.username,
      password: form.password,
      ...(mode.value === 'register' ? { nickname: form.nickname } : {})
    })
    const { token, user } = res.data.data
    authStore.setAuth(token, user)
    // 先加载考试数据，再跳转首页，避免数据竞态
    await examStore.loadExams()
    router.push('/')
  } catch (e) {
    error.value = e.response?.data?.error || '操作失败'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}
.login-card {
  background: white;
  border-radius: 16px;
  padding: 40px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.2);
}
.login-header {
  text-align: center;
  margin-bottom: 30px;
}
.login-header h1 {
  font-size: 28px;
  margin: 0 0 8px 0;
  color: #333;
}
.login-header p {
  color: #888;
  margin: 0;
  font-size: 14px;
}
.tab-switch {
  display: flex;
  margin-bottom: 24px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
}
.tab-switch button {
  flex: 1;
  padding: 10px;
  border: none;
  background: #f9fafb;
  cursor: pointer;
  font-size: 14px;
  color: #666;
  transition: all 0.2s;
}
.tab-switch button.active {
  background: #667eea;
  color: white;
}
.form-group {
  margin-bottom: 16px;
}
.form-group input {
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
}
.form-group input:focus {
  border-color: #667eea;
}
.btn-submit {
  width: 100%;
  padding: 12px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 15px;
  cursor: pointer;
  transition: background 0.2s;
}
.btn-submit:hover { background: #5a6fd6; }
.btn-submit:disabled { opacity: 0.6; cursor: not-allowed; }
.error-msg {
  color: #ef4444;
  font-size: 13px;
  text-align: center;
  margin-top: 12px;
}
.hint {
  text-align: center;
  color: #aaa;
  font-size: 12px;
  margin-top: 16px;
}
</style>
