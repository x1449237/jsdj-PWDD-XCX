import { defineStore } from 'pinia'
import { getToken, setToken, removeToken, getNeedInit, setNeedInit } from '@/utils/auth'
import {
  login as loginApi,
  getAdminInfo,
  getInitStatus,
  logout as logoutApi
} from '@/api/auth'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: getToken() || '',
    adminInfo: null,
    isInitialized: false,
    // 当前账号是否需要完成首次登录初始化
    needInit: getNeedInit(),
    // 系统是否已完成初始化（首次登录修改密码/绑定邮箱）
    systemInitialized: null
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    username: (state) => state.adminInfo?.username || '',
    avatar: (state) => state.adminInfo?.avatar || '',
    roles: (state) => state.adminInfo?.roles || [],
    permissions: (state) => state.adminInfo?.permissions || []
  },

  actions: {
    // 登录：存储 token，并根据返回判断是否需要初始化
    async login(loginForm) {
      const res = await loginApi(loginForm)
      const data = res.data || {}
      const token = data.token
      if (token) {
        setToken(token)
        this.token = token
      }
      // needInit 表示当前账号需要完成首次登录初始化
      this.needInit = !!data.needInit
      setNeedInit(this.needInit)
      if (this.needInit) {
        return { needInit: true }
      }
      await this.getInfo()
      return { needInit: false }
    },

    // 获取当前管理员信息
    async getInfo() {
      const res = await getAdminInfo()
      this.adminInfo = res.data
      this.isInitialized = true
    },

    // 完成首次登录初始化：清除标记并加载用户信息
    async finishInit() {
      this.needInit = false
      setNeedInit(false)
      try {
        await this.getInfo()
      } catch (err) {
        console.error('加载用户信息失败:', err)
      }
    },

    // 检查系统初始化状态
    async checkInitStatus() {
      try {
        const res = await getInitStatus()
        this.systemInitialized = !!res.data?.initialized
        return this.systemInitialized
      } catch (err) {
        this.systemInitialized = false
        return false
      }
    },

    // 登出：清理本地状态
    async logout() {
      try {
        if (this.token) {
          await logoutApi()
        }
      } catch (err) {
        // 即使接口失败也清理本地状态
        console.error('登出失败:', err)
      } finally {
        this.resetState()
        removeToken()
        setNeedInit(false)
      }
    },

    hasPermission(perm) {
      if (!this.permissions || this.permissions.length === 0) return false
      return this.permissions.includes(perm)
    },

    resetState() {
      this.token = ''
      this.adminInfo = null
      this.isInitialized = false
      this.needInit = false
    }
  }
})
