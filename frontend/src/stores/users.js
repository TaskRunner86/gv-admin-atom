import { defineStore } from 'pinia'
import { getUsers, createUser, updateUser, deleteUser } from '../api'

// 用户管理列表与分页状态集中管理
export const useUsersStore = defineStore('users', {
  state: () => ({
    list: [],
    total: 0,
    page: 1,
    pageSize: 10,
    keyword: '',
    loading: false,
    saving: false
  }),

  actions: {
    async load() {
      this.loading = true
      try {
        const data = await getUsers({
          page: this.page,
          pageSize: this.pageSize,
          keyword: this.keyword
        })
        this.list = data.list
        this.total = data.total
      } finally {
        this.loading = false
      }
    },

    // 条件变化后从第一页重新查询
    search() {
      this.page = 1
      return this.load()
    },

    async create(payload) {
      this.saving = true
      try {
        await createUser(payload)
      } finally {
        this.saving = false
      }
      await this.load()
    },

    async update(id, payload) {
      this.saving = true
      try {
        await updateUser(id, payload)
      } finally {
        this.saving = false
      }
      await this.load()
    },

    async remove(id) {
      await deleteUser(id)
      // 删除当前页最后一条时回退一页，避免停留在空页
      if (this.list.length === 1 && this.page > 1) {
        this.page -= 1
      }
      await this.load()
    },

    reset() {
      this.$reset()
    }
  }
})
