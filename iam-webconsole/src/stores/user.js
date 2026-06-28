import { defineStore } from 'pinia'
import { getMe } from '@/api/auth'

export const useUserStore = defineStore('user', {
  state: () => ({
    info: null,
    loaded: false,
  }),

  getters: {
    isAdmin: (state) => state.info?.is_admin === true,
    username: (state) => state.info?.username || '',
  },

  actions: {
    async fetchUserInfo(force = false) {
      if (this.loaded && !force) return this.info
      try {
        const res = await getMe()
        this.info = res
        this.loaded = true
        return res
      } catch {
        this.info = null
        this.loaded = false
        return null
      }
    },

    clear() {
      this.info = null
      this.loaded = false
    },
  },
})
