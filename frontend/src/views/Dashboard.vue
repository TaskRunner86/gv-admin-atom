<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="16" class="stat-row">
      <el-col v-for="card in statCards" :key="card.label" :xs="12" :sm="12" :md="8">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-body">
            <div class="stat-icon" :style="{ backgroundColor: card.color + '1a', color: card.color }">
              <el-icon :size="26"><component :is="card.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ card.value }}</div>
              <div class="stat-label">{{ card.label }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 上报趋势折线 + 设备类型饼图 -->
    <el-row :gutter="16" class="chart-row">
      <el-col :xs="24" :md="14">
        <el-card shadow="hover">
          <template #header>近 7 天数据上报趋势</template>
          <EChart :option="trendOption" height="320px" />
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card shadow="hover">
          <template #header>设备类型占比</template>
          <EChart :option="typeOption" height="320px" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 30 天上报柱状 + 设备状态 -->
    <el-row :gutter="16" class="chart-row">
      <el-col :xs="24" :md="14">
        <el-card shadow="hover">
          <template #header>近 30 天数据上报量</template>
          <EChart :option="monthOption" height="300px" />
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card shadow="hover">
          <template #header>设备状态分布</template>
          <EChart :option="statusOption" height="300px" />
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useDashboardStore } from '../stores'
import EChart from '../components/EChart.vue'

const dashboardStore = useDashboardStore()
const { overview } = storeToRefs(dashboardStore)

const formatNumber = (n) => Number(n || 0).toLocaleString('zh-CN')
const formatCompact = (n) => {
  const v = Number(n || 0)
  return v >= 10000 ? (v / 10000).toFixed(1) + ' 万' : formatNumber(v)
}

const statCards = computed(() => [
  { label: '接入设备总数', value: formatNumber(overview.value.stats.totalDevices), icon: 'Monitor', color: '#409eff' },
  { label: '在线设备数', value: formatNumber(overview.value.stats.onlineDevices), icon: 'Connection', color: '#67c23a' },
  { label: '今日上报消息', value: formatCompact(overview.value.stats.todayMessages), icon: 'DataLine', color: '#e6a23c' }
])

const trendOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  legend: { data: ['消息上报量', '在线设备数'] },
  grid: { left: 60, right: 55, top: 40, bottom: 30 },
  xAxis: { type: 'category', boundaryGap: false, data: overview.value.trend.map((p) => p.day.slice(5)) },
  yAxis: [
    { type: 'value', name: '条' },
    { type: 'value', name: '台', min: 0, splitLine: { show: false } }
  ],
  series: [
    {
      name: '消息上报量',
      type: 'line',
      smooth: true,
      data: overview.value.trend.map((p) => p.messages),
      areaStyle: { opacity: 0.15 },
      itemStyle: { color: '#409eff' }
    },
    {
      name: '在线设备数',
      type: 'line',
      smooth: true,
      yAxisIndex: 1,
      data: overview.value.trend.map((p) => p.devices),
      itemStyle: { color: '#67c23a' }
    }
  ]
}))

const monthOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 60, right: 20, top: 30, bottom: 30 },
  xAxis: { type: 'category', data: overview.value.monthTrend.map((p) => p.day.slice(5)) },
  yAxis: {
    type: 'value',
    axisLabel: { formatter: (v) => (v >= 10000 ? (v / 10000).toFixed(1) + '万' : v) }
  },
  series: [
    {
      name: '上报消息量',
      type: 'bar',
      barMaxWidth: 22,
      data: overview.value.monthTrend.map((p) => p.messages),
      itemStyle: { color: '#409eff', borderRadius: [3, 3, 0, 0] }
    }
  ]
}))

const typeOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} 台 ({d}%)' },
  legend: { bottom: 0, type: 'scroll' },
  series: [
    {
      name: '设备类型',
      type: 'pie',
      radius: ['40%', '68%'],
      center: ['50%', '44%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: false },
      emphasis: { label: { show: true, fontSize: 14, fontWeight: 'bold' } },
      data: overview.value.typeDist.map((t) => ({ name: t.name, value: t.value }))
    }
  ]
}))

const statusOption = computed(() => ({
  tooltip: { trigger: 'item', formatter: '{b}: {c} 台 ({d}%)' },
  legend: { bottom: 0 },
  series: [
    {
      name: '设备状态',
      type: 'pie',
      radius: '68%',
      center: ['50%', '44%'],
      itemStyle: { borderRadius: 6, borderColor: '#fff', borderWidth: 2 },
      label: { show: true, formatter: '{b}: {c}' },
      data: overview.value.statusDist.map((s) => ({ name: s.name, value: s.value }))
    }
  ]
}))

onMounted(async () => {
  try {
    await dashboardStore.fetchOverview()
  } catch {
    /* 拦截器已提示 */
  }
})
</script>

<style scoped>
.stat-row {
  margin-bottom: 16px;
}

.stat-card :deep(.el-card__body) {
  padding: 18px;
}

.stat-body {
  display: flex;
  align-items: center;
  gap: 14px;
}

.stat-icon {
  width: 54px;
  height: 54px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  margin-top: 4px;
  font-size: 13px;
  color: #909399;
}

.chart-row {
  margin-bottom: 16px;
}
</style>
