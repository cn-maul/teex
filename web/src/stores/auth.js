import { reactive, computed } from 'vue'

const state = reactive({
  token: localStorage.getItem('token') || '',
  user: JSON.parse(localStorage.getItem('user') || 'null')
})

export function useAuthStore() {
  const isLoggedIn = computed(() => !!state.token)
  const user = computed(() => state.user)
  const token = computed(() => state.token)

  function setAuth(tokenVal, userVal) {
    state.token = tokenVal
    state.user = userVal
    localStorage.setItem('token', tokenVal)
    localStorage.setItem('user', JSON.stringify(userVal))
  }

  function logout() {
    state.token = ''
    state.user = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return { isLoggedIn, user, token, setAuth, logout }
}
