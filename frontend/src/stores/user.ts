import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  // State
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<any>(null)
  
  // Getters
  const isLoggedIn = computed(() => !!token.value)
  
  // Actions
  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }
  
  const clearToken = () => {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
  }
  
  const login = async (email: string, password: string) => {
    // TODO: 调用登录API
    const mockToken = 'mock-token-' + Date.now()
    setToken(mockToken)
    userInfo.value = { email, name: 'Test User' }
    return true
  }
  
  const logout = () => {
    clearToken()
  }
  
  return {
    token,
    userInfo,
    isLoggedIn,
    setToken,
    clearToken,
    login,
    logout
  }
})
