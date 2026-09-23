import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

const http = axios.create({
  baseURL: '/api',
  timeout: 15000
})

// 请求拦截：附加 token
http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截：统一处理业务码与 401
http.interceptors.response.use(
  (res) => {
    const body = res.data
    if (body && body.code !== 0) {
      ElMessage.error(body.message || '请求失败')
      return Promise.reject(new Error(body.message))
    }
    return body.data
  },
  (err) => {
    const status = err.response?.status
    const message = err.response?.data?.message
    if (status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      router.push('/login')
      ElMessage.error(message || '登录已过期，请重新登录')
    } else {
      ElMessage.error(message || '网络异常，请稍后重试')
    }
    return Promise.reject(err)
  }
)

// ===== 认证 =====
export const login = (data) => http.post('/login', data)
export const getProfile = () => http.get('/user/profile')

// ===== 看板 =====
export const getOverview = () => http.get('/dashboard/overview')

// ===== 用户管理 =====
export const getUsers = (params) => http.get('/users', { params })
export const createUser = (data) => http.post('/users', data)
export const updateUser = (id, data) => http.put(`/users/${id}`, data)
export const deleteUser = (id) => http.delete(`/users/${id}`)

export default http
