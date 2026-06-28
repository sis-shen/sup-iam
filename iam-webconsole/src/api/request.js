import axios from 'axios'
import { ElMessage } from 'element-plus'
import { getAccessToken, getRefreshToken, setTokens, clearTokens } from '@/utils/auth'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor: attach JWT token
request.interceptors.request.use(
  (config) => {
    const token = getAccessToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// Response interceptor: handle 401 and token refresh
let isRefreshing = false
let refreshSubscribers = []

function onRefreshed(token) {
  refreshSubscribers.forEach((cb) => cb(token))
  refreshSubscribers = []
}

function addRefreshSubscriber(cb) {
  refreshSubscribers.push(cb)
}

request.interceptors.response.use(
  (response) => response.data,
  async (error) => {
    const { response, config } = error
    if (!response) {
      ElMessage.error('网络连接失败，请检查网络')
      return Promise.reject(error)
    }

    const { status, data } = response

    // Try token refresh on 401
    if (status === 401 && !config._retry) {
      const refreshToken = getRefreshToken()
      if (refreshToken) {
        if (!isRefreshing) {
          isRefreshing = true
          config._retry = true
          try {
            const res = await axios.post('/api/v1/auth/refresh', {
              refresh_token: refreshToken,
            })
            const newToken = res.data.access_token
            setTokens(newToken, null, res.data.expires_in)
            isRefreshing = false
            onRefreshed(newToken)
            config.headers.Authorization = `Bearer ${newToken}`
            return request(config)
          } catch {
            isRefreshing = false
            refreshSubscribers = []
            clearTokens()
            window.location.href = '/login'
            return Promise.reject(error)
          }
        } else {
          return new Promise((resolve) => {
            addRefreshSubscriber((token) => {
              config.headers.Authorization = `Bearer ${token}`
              resolve(request(config))
            })
          })
        }
      } else {
        clearTokens()
        window.location.href = '/login'
      }
    }

    // Display error message from API (skip silent "not found" errors)
    if (status !== 401 || !config._retry) {
      const msg = data?.error_description || data?.error || '请求失败'
      if (msg !== 'repository: not found') {
        ElMessage.error(msg)
      }
    }
    return Promise.reject(error)
  }
)

export default request
