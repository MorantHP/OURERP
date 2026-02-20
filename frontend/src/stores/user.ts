import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, type UserInfo } from '@/api/auth'

// 重新导出 UserInfo 类型以保持向后兼容
export type { UserInfo } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<UserInfo | null>(null)
  const isInitialized = ref(false)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value?.is_root === true)
  const isApproved = computed(() => userInfo.value?.is_approved === true)

  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  const clearToken = () => {
    token.value = ''
    userInfo.value = null
    isInitialized.value = false
    localStorage.removeItem('token')
  }

  const fetchUserInfo = async (): Promise<UserInfo | null> => {
    if (!token.value) {
      return null
    }
    try {
      const res = await authApi.getMe()
      userInfo.value = res.user
      isInitialized.value = true
      return userInfo.value
    } catch (error) {
      clearToken()
      throw error
    }
  }

  // 初始化用户状态 - 验证token并获取用户信息
  const init = async (): Promise<boolean> => {
    if (!token.value) {
      isInitialized.value = true
      return false
    }

    try {
      await fetchUserInfo()
      return true
    } catch {
      return false
    }
  }

  const logout = () => {
    clearToken()
  }

  return {
    token,
    userInfo,
    isInitialized,
    isLoggedIn,
    isAdmin,
    isApproved,
    setToken,
    clearToken,
    fetchUserInfo,
    init,
    logout
  }
})
