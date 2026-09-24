import { defineStore } from 'pinia'
import { login as loginApi } from '../api'

const TOKEN_KEY = 'token'
const USER_KEY = 'user'

// 从 localStorage 恢复用户信息（刷新页面后保持登录态）
const readUser = () => {
  try {
    return JSON.parse(localStorage.getItem(USER_KEY) || 'null')
  } catch {
    return null
  }
}

// 登录态集中管理：token 与用户信息由 store 唯一持有
export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    user: readUser()
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    isAdmin: (state) => state.user?.role === 'admin',
    username: (state) => state.user?.username || '',
    avatarText: (state) => (state.user?.username || '?').charAt(0).toUpperCase()
  },

  actions: {
    // 写入登录态并同步到 localStorage
    setAuth({ token, user }) {
      this.token = token || ''
      this.user = user || null
      if (this.token) {
        localStorage.setItem(TOKEN_KEY, this.token)
      } else {
        localStorage.removeItem(TOKEN_KEY)
      }
      if (this.user) {
        localStorage.setItem(USER_KEY, JSON.stringify(this.user))
      } else {
        localStorage.removeItem(USER_KEY)
      }
    },

    async login(payload) {
      const data = await loginApi(payload)
      this.setAuth(data)
      return data
    },

    // 清除登录态（退出登录 / token 失效）
    clearAuth() {
      this.setAuth({ token: '', user: null })
    },

    logout() {
      this.clearAuth()
    }
  }
})
