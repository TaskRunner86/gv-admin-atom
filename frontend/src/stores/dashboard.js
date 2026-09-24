import { defineStore } from 'pinia'
import { getOverview } from '../api'

const emptyOverview = () => ({
  stats: { totalDevices: 0, onlineDevices: 0, todayMessages: 0, todayAlerts: 0 },
  trend: [],
  monthTrend: [],
  typeDist: [],
  statusDist: [],
  recentAlerts: []
})

// 设备看板数据集中管理
export const useDashboardStore = defineStore('dashboard', {
  state: () => ({
    overview: emptyOverview(),
    loading: false
  }),

  getters: {
    stats: (state) => state.overview.stats,
    trend: (state) => state.overview.trend,
    monthTrend: (state) => state.overview.monthTrend,
    typeDist: (state) => state.overview.typeDist,
    statusDist: (state) => state.overview.statusDist,
    recentAlerts: (state) => state.overview.recentAlerts
  },

  actions: {
    async fetchOverview() {
      this.loading = true
      try {
        this.overview = await getOverview()
      } finally {
        this.loading = false
      }
    },

    reset() {
      this.overview = emptyOverview()
    }
  }
})
