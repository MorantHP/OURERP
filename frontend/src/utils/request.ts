// src/utils/request.ts
import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

// 从环境变量获取API地址，默认为开发环境地址
const baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

const request = axios.create({
  baseURL,
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }

    // 从 localStorage 获取租户ID（避免循环依赖）
    const tenantId = localStorage.getItem('tenant_id')
    if (tenantId) {
      config.headers['X-Tenant-ID'] = tenantId
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error) => {
    const message = error.response?.data?.error || '请求失败'

    // 只在真正的认证失败时才登出（token无效或过期）
    // 缺少租户等错误不应该导致登出
    if (error.response?.status === 401) {
      const errorMsg = error.response?.data?.error || ''
      // 只有token相关的401才登出
      if (errorMsg.includes('token') || errorMsg.includes('Token') ||
          errorMsg.includes('认证') || errorMsg.includes('登录') ||
          errorMsg.includes('无效') || errorMsg.includes('过期')) {
        const userStore = useUserStore()
        userStore.logout()
        window.location.href = '/login'
      } else {
        // 其他401错误只显示消息，不登出
        ElMessage.error(message)
      }
    } else {
      ElMessage.error(message)
    }

    return Promise.reject(error)
  }
)

export default request
