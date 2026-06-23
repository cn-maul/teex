import { reactive, ref, computed } from 'vue'

const token = ref(localStorage.getItem('token') || '')
function parseUser() {
  try {
    const raw = localStorage.getItem('user')
    return raw ? JSON.parse(raw) : null
  } catch {
    localStorage.removeItem('user')
    return null
  }
}
const user = ref(parseUser())

export function useAuthStore() {
  const isLoggedIn = computed(() => !!token.value)

  function setAuth(tokenVal, userVal) {
    token.value = tokenVal
    user.value = userVal
    localStorage.setItem('token', tokenVal)
    localStorage.setItem('user', JSON.stringify(userVal))
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return reactive({ isLoggedIn, user, token, setAuth, logout })
}
